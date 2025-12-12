package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alessandrosisniegas/casino/core/game"
	"github.com/alessandrosisniegas/casino/core/security"
	"github.com/alessandrosisniegas/casino/core/vault"
)

type GameMode string

const (
	ModeLobby       GameMode = "LOBBY"
	ModeSolo        GameMode = "SOLO"
	ModeMultiplayer GameMode = "MULTIPLAYER"
)

type ClientState struct {
	conn      net.Conn
	sessionID string
	user      *vault.User
	mode      GameMode

	soloGame  *game.Game
	tableSeat *game.PlayerSeat
}

type Server struct {
	authService *security.AuthService
	db          *vault.DB
	mainTable   *game.Table
}

func main() {
	// Initialize database (use absolute path from project root)
	dbPath := filepath.Join("..", "..", "data", "casino.db")
	if err := os.MkdirAll(filepath.Join("..", "..", "data"), 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := vault.NewDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize auth service
	authService := security.NewAuthService(db)

	mainTable := game.NewTable("main", 10_00, 1000_00, 4)

	server := &Server{
		authService: authService,
		db:          db,
		mainTable:   mainTable,
	}

	// Bind address:
	// - Default is local only where we bind 127.0.0.1
	// - If user wants to run it on LAN we bind 0.0.0.0
	addr := "127.0.0.1:9090"
	if os.Getenv("LAN") == "1" {
		addr = "0.0.0.0:9090"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("Casino Server listening on", ln.Addr().String())
	fmt.Println("Database initialized at", dbPath)
	fmt.Println("Type 'help' for server commands, 'quit' to shutdown")
	fmt.Print("server> ")

	// Channel to signal server shutdown
	shutdown := make(chan bool)

	// Periodically cleanup expired sessions
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := db.CleanupExpiredSessions(); err != nil {
				log.Println("Failed to cleanup expired sessions:", err)
			}
		}
	}()

	// Handle server commands from stdin
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			command := strings.TrimSpace(strings.ToUpper(scanner.Text()))
			switch command {
			case "QUIT", "EXIT", "STOP":
				fmt.Println("Shutting down server...")
				shutdown <- true
				return
			case "HELP":
				fmt.Println("Server commands:")
				fmt.Println("  help  - Show this help")
				fmt.Println("  stats - Show server statistics")
				fmt.Println("  users - List all users")
				fmt.Println("  quit  - Shutdown server")
			case "STATS":
				server.showStats()
			case "USERS":
				server.showUsers()
			case "":
			default:
				fmt.Printf("Unknown command: %s (type 'help' for commands)\n", command)
			}
			fmt.Print("server> ")
		}
	}()

	// Accept connections until shutdown signal
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// Check if we're shutting down
				select {
				case <-shutdown:
					return
				default:
					log.Println("Accept error:", err)
					continue
				}
			}

			// Handle each client concurrently so slow clients do not block others
			go server.handleClient(conn)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	fmt.Println("Server stopped.")
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	// Set connection timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Minute))

	client := &ClientState{
		conn: conn,
		mode: ModeLobby,
	}
	scanner := bufio.NewScanner(conn)

	s.writeResponse(client, "OK Welcome to Casino! Use SIGNUP <username> <password> or LOGIN <username> <password>")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Reset read deadline on each command
		conn.SetReadDeadline(time.Now().Add(30 * time.Minute))

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		s.handleCommand(client, command, parts[1:])
	}

	if err := scanner.Err(); err != nil {
		log.Println("Client connection error:", err)
	}
}

func (s *Server) handleCommand(client *ClientState, command string, args []string) {
	switch command {
	case "SIGNUP", "REGISTER":
		s.handleSignup(client, args)
	case "LOGIN":
		s.handleLogin(client, args)
	case "LOGOUT":
		s.handleLogout(client, args)
	case "BALANCE":
		s.handleBalance(client, args)
	case "STATS":
		s.handleStats(client, args)
	case "WHOAMI":
		s.handleWhoami(client, args)
	case "SOLO":
		s.handleSolo(client, args)
	case "MULTIPLAYER", "MP":
		s.handleMultiplayer(client, args)
	case "LEAVE":
		s.handleLeave(client, args)
	case "READY":
		s.handleReady(client, args)
	case "PLAYERS":
		s.handlePlayers(client, args)
	case "BET":
		s.handleBet(client, args)
	case "HIT":
		s.handleHit(client, args)
	case "STAND":
		s.handleStand(client, args)
	case "DOUBLEDOWN", "DOUBLE":
		s.handleDoubleDown(client, args)
	case "SURRENDER":
		s.handleSurrender(client, args)
	case "PROBSTATS":
		s.handleProbStats(client, args)
	case "QUIT", "EXIT":
		s.writeResponse(client, "OK Goodbye!")
		client.conn.Close()
	case "HELP":
		s.handleHelp(client, args)
	default:
		s.writeResponse(client, "ERROR Unknown command. Type HELP for available commands.")
	}
}

func (s *Server) handleSignup(client *ClientState, args []string) {
	if len(args) != 2 {
		s.writeResponse(client, "ERROR Usage: SIGNUP <username> <password>")
		return
	}

	username, password := args[0], args[1]
	user, err := s.authService.RegisterUser(username, password)
	if err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
		return
	}

	s.writeResponse(client, fmt.Sprintf("OK Account created for %s with balance $%.2f", user.Username, float64(user.Balance)/100))
}

func (s *Server) handleLogin(client *ClientState, args []string) {
	if len(args) != 2 {
		s.writeResponse(client, "ERROR Usage: LOGIN <username> <password>")
		return
	}

	username, password := args[0], args[1]
	sessionID, user, err := s.authService.LoginUser(username, password)
	if err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
		return
	}

	client.sessionID = sessionID
	client.user = user
	client.mode = ModeLobby

	response := fmt.Sprintf("OK Welcome back, %s! Balance: $%.2f\n", user.Username, float64(user.Balance)/100)
	response += "Available modes: SOLO, MULTIPLAYER\n"
	response += "Use SOLO or MULTIPLAYER to choose a mode, or BET <amount> for quick solo play."
	s.writeResponse(client, response)
}

func (s *Server) handleLogout(client *ClientState, _ []string) {
	if client.sessionID == "" {
		s.writeResponse(client, "ERROR Not logged in")
		return
	}

	if err := s.authService.LogoutUser(client.sessionID); err != nil {
		log.Println("Failed to logout user:", err)
	}

	client.sessionID = ""
	client.user = nil
	s.writeResponse(client, "OK Logged out successfully")
}

func (s *Server) handleBalance(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	// Refresh user data from database
	user, err := s.authService.ValidateSession(client.sessionID)
	if err != nil {
		s.writeResponse(client, "ERROR Session expired, please login again")
		client.sessionID = ""
		client.user = nil
		return
	}

	client.user = user
	s.writeResponse(client, fmt.Sprintf("OK Balance: $%.2f", float64(user.Balance)/100))
}

func (s *Server) handleStats(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	stats, err := s.authService.GetUserStats(client.user.ID)
	if err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR Failed to get stats: %s", err.Error()))
		return
	}

	winRate := float64(0)
	avgBet := float64(0)
	if stats.GamesPlayed > 0 {
		winRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
		avgBet = float64(stats.TotalBet) / float64(stats.GamesPlayed) / 100
	}

	response := fmt.Sprintf("OK Stats for %s:\n", client.user.Username)
	response += fmt.Sprintf("  Games Played: %d\n", stats.GamesPlayed)
	response += fmt.Sprintf("  Games Won: %d\n", stats.GamesWon)
	response += fmt.Sprintf("  Games Lost: %d\n", stats.GamesLost)
	response += fmt.Sprintf("  Win Rate: %.1f%%\n", winRate)
	response += fmt.Sprintf("  Total Bet: $%.2f\n", float64(stats.TotalBet)/100)
	response += fmt.Sprintf("  Total Won: $%.2f\n", float64(stats.TotalWon)/100)
	response += fmt.Sprintf("  Net: $%.2f\n", float64(stats.TotalWon-stats.TotalBet)/100)
	response += fmt.Sprintf("  Avg Bet: $%.2f\n", avgBet)
	response += fmt.Sprintf("  Biggest Win: $%.2f\n", float64(stats.BiggestWin)/100)
	response += fmt.Sprintf("  Biggest Loss: $%.2f", float64(stats.BiggestLoss)/100)

	s.writeResponse(client, response)
}

func (s *Server) handleWhoami(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Not logged in")
		return
	}

	s.writeResponse(client, fmt.Sprintf("OK Logged in as: %s (ID: %d, Balance: $%.2f)",
		client.user.Username, client.user.ID, float64(client.user.Balance)/100))
}

func (s *Server) handleProbStats(client *ClientState, args []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	// Get user preferences (create if not exists for old accounts)
	prefs, err := s.db.GetUserPreferences(client.user.ID)
	if err != nil {
		// Preferences don't exist - create them with default value
		if err := s.db.InitUserPreferences(client.user.ID); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to initialize preferences: %s", err.Error()))
			return
		}
		// Try to get them again
		prefs, err = s.db.GetUserPreferences(client.user.ID)
		if err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to get preferences: %s", err.Error()))
			return
		}
	}

	// If no arguments, show current status
	if len(args) == 0 {
		status := "OFF"
		if prefs.ShowStats {
			status = "ON"
		}
		s.writeResponse(client, fmt.Sprintf("OK Probability stats display: %s", status))
		return
	}

	// Parse ON/OFF argument
	arg := strings.ToUpper(args[0])
	var newValue bool

	switch arg {
	case "ON":
		newValue = true
	case "OFF":
		newValue = false
	default:
		s.writeResponse(client, "ERROR Usage: PROBSTATS [ON|OFF]")
		return
	}

	// Update preferences
	if err := s.db.UpdateUserPreferences(client.user.ID, newValue); err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR Failed to update preferences: %s", err.Error()))
		return
	}

	status := "disabled"
	if newValue {
		status = "enabled"
	}
	s.writeResponse(client, fmt.Sprintf("OK Probability stats display %s", status))
}

func (s *Server) handleHelp(client *ClientState, _ []string) {
	help := "OK Available commands:\n"
	help += "\nAccount Management:\n"
	help += "  SIGNUP <username> <password> - Create a new account\n"
	help += "  LOGIN <username> <password>  - Login to your account\n"
	help += "  LOGOUT                       - Logout from your account\n"
	help += "  BALANCE                      - Check your current balance\n"
	help += "  STATS                        - View your game statistics\n"
	help += "  WHOAMI                       - Show current login status\n"

	if client.user != nil {
		help += "\nGame Modes:\n"
		if client.mode == ModeLobby {
			help += "  SOLO                         - Enter solo play mode\n"
			help += "  MULTIPLAYER                  - Join multiplayer table\n"
			help += "  BET <amount>                 - Quick solo play (auto-enter solo mode)\n"
		} else if client.mode == ModeSolo {
			help += "  Current mode: SOLO\n"
			help += "  LEAVE                        - Return to lobby\n"
		} else if client.mode == ModeMultiplayer {
			help += "  Current mode: MULTIPLAYER\n"
			help += "  LEAVE                        - Leave table and return to lobby\n"
			help += "  READY                        - Mark ready to start round\n"
			help += "  PLAYERS                      - Show players at table\n"
		}

		help += "\nBlackjack Actions:\n"
		help += "  BET <amount>                 - Place bet (in dollars)\n"
		help += "  HIT                          - Draw another card\n"
		help += "  STAND                        - End your turn\n"
		help += "  DOUBLEDOWN                   - Double bet, draw one card, end turn\n"
		help += "  SURRENDER                    - Forfeit hand, get half bet back\n"

		help += "\nProbability Analysis:\n"
		help += "  PROBSTATS [ON|OFF]           - Toggle probability stats display\n"
		help += "                                 (shows bust odds, card count, strategy)\n"
	}

	help += "\nOther:\n"
	help += "  HELP                         - Show this help message\n"
	help += "  QUIT                         - Disconnect from server\n"

	if client.user == nil {
		help += "\nUsername & Password requirements:\n"
		help += "  - 2-30 characters long\n"
		help += "  - Letters, numbers, and underscores only\n"
		help += "  - No whitespace allowed\n"
		help += "  - Password cannot be the same as username"
	}

	s.writeResponse(client, help)
}

func (s *Server) writeResponse(client *ClientState, message string) {
	client.conn.Write([]byte(message + "\n"))
}

func (s *Server) showStats() {
	fmt.Println("Server Statistics:")
	fmt.Println("  Server: Running")
	fmt.Println("  Database: Connected")
	fmt.Printf("  Address: %s\n", "127.0.0.1:9090")
}

func (s *Server) showUsers() {
	fmt.Println("Use SQLite to view users:")
	fmt.Println("  sqlite3 data/casino.db \"SELECT id, username, balance/100.0, created_at FROM users;\"")
}

// formatProbabilityStats generates the probability analysis display
func (s *Server) formatProbabilityStats(playerHand *game.Hand, dealerUpcard game.Card, deck *game.Deck, counter *game.CardCounter, canDoubleDown bool, canSurrender bool) string {
	// Calculate bust probability
	bustProb := game.CalculateBustProbability(playerHand, deck)
	bustCards := game.GetBustCards(playerHand.Value())

	// Get optimal action
	optimalAction := game.GetOptimalAction(playerHand, dealerUpcard, canDoubleDown, canSurrender)

	// Calculate card counting metrics
	decksRemaining := game.CalculateDecksRemaining(1, counter.GetCardsDealt())
	trueCount := counter.GetTrueCount(decksRemaining)
	advantage := counter.GetAdvantage(decksRemaining)

	// Format the display
	stats := "\nPROBABILITY ANALYSIS:\n"
	stats += fmt.Sprintf("  %s\n", game.FormatBustProbability(bustProb, bustCards))
	stats += fmt.Sprintf("  Running count: %+d | True count: %+.1f | ", counter.GetRunningCount(), trueCount)

	if advantage > 0 {
		stats += fmt.Sprintf("Player edge: ~%.1f%%\n", advantage)
	} else if advantage < 0 {
		stats += fmt.Sprintf("House edge: ~%.1f%%\n", -advantage)
	} else {
		stats += "Even odds\n"
	}

	stats += fmt.Sprintf("  Basic strategy: %s\n", optimalAction)

	return stats
}

func (s *Server) handleSolo(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		s.mainTable.RemovePlayer(client.sessionID)
		s.mainTable.BroadcastToAll(fmt.Sprintf("\n%s left the table", client.user.Username))
	}

	client.mode = ModeSolo
	client.soloGame = nil
	client.tableSeat = nil

	s.writeResponse(client, "OK Entered solo mode. Place a bet to start playing.")
}

func (s *Server) handleMultiplayer(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		s.writeResponse(client, "ERROR Already in multiplayer mode")
		return
	}

	if err := s.mainTable.AddPlayer(client.sessionID, client.user.Username, client.conn); err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
		return
	}

	client.mode = ModeMultiplayer
	client.soloGame = nil

	playerCount := s.mainTable.PlayerCount()
	response := fmt.Sprintf("OK Joined multiplayer table (%d/%d players)\n", playerCount, s.mainTable.MaxPlayers)
	response += s.mainTable.GetPlayerList()
	response += "Type READY when you're ready to play."
	s.writeResponse(client, response)

	s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s joined the table (%d/%d players)",
		client.user.Username, playerCount, s.mainTable.MaxPlayers))
}

func (s *Server) handleLeave(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeLobby {
		s.writeResponse(client, "ERROR Already in lobby")
		return
	}

	modeName := string(client.mode)

	if client.mode == ModeMultiplayer {
		s.mainTable.RemovePlayer(client.sessionID)
		s.mainTable.BroadcastToAll(fmt.Sprintf("\n%s left the table", client.user.Username))
	}

	client.mode = ModeLobby
	client.soloGame = nil
	client.tableSeat = nil

	response := fmt.Sprintf("OK Left %s mode. Back in lobby.\n", strings.ToLower(modeName))
	response += "Available modes: SOLO, MULTIPLAYER"
	s.writeResponse(client, response)
}

func (s *Server) handleReady(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode != ModeMultiplayer {
		s.writeResponse(client, "ERROR Only available in multiplayer mode")
		return
	}

	if err := s.mainTable.SetReady(client.sessionID, true); err != nil {
		s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
		return
	}

	s.writeResponse(client, "OK Marked ready. Waiting for other players...")
	s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s is ready!", client.user.Username))

	if s.mainTable.AllPlayersReady() {
		go s.startMultiplayerRound()
	}
}

func (s *Server) handlePlayers(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode != ModeMultiplayer {
		s.writeResponse(client, "ERROR Only available in multiplayer mode")
		return
	}

	response := "OK " + s.mainTable.GetPlayerList()
	s.writeResponse(client, response)
}

// Blackjack game handlers

func (s *Server) handleBet(client *ClientState, args []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if len(args) != 1 {
		s.writeResponse(client, "ERROR Usage: BET <amount> (e.g., BET 10 for $10)")
		return
	}

	// Parse bet amount in dollars
	betDollars, err := strconv.ParseFloat(args[0], 64)
	if err != nil || betDollars <= 0 {
		s.writeResponse(client, "ERROR Invalid bet amount")
		return
	}

	betCents := int64(betDollars * 100)

	// Refresh user balance from database
	user, err := s.authService.ValidateSession(client.sessionID)
	if err != nil {
		s.writeResponse(client, "ERROR Session expired, please login again")
		client.sessionID = ""
		client.user = nil
		return
	}
	client.user = user

	if client.user.Balance < betCents {
		s.writeResponse(client, fmt.Sprintf("ERROR Insufficient balance. You have $%.2f", float64(client.user.Balance)/100))
		return
	}

	if client.mode == ModeMultiplayer {
		if err := s.mainTable.PlaceBet(client.sessionID, betCents); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		newBalance := client.user.Balance - betCents
		if err := s.authService.UpdateBalance(client.user.ID, newBalance); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to update balance: %s", err.Error()))
			return
		}
		client.user.Balance = newBalance

		s.writeResponse(client, fmt.Sprintf("OK Bet placed: $%.2f", float64(betCents)/100))
		s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s bets $%.2f", client.user.Username, float64(betCents)/100))

	} else {
		if client.mode == ModeLobby {
			client.mode = ModeSolo
		}

		client.soloGame = game.NewGame()
		if err := client.soloGame.PlaceBet(betCents); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		newBalance := client.user.Balance - betCents
		if err := s.authService.UpdateBalance(client.user.ID, newBalance); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to update balance: %s", err.Error()))
			return
		}
		client.user.Balance = newBalance

		response := fmt.Sprintf("OK Game started!\n%s", client.soloGame.GetGameState(true))

		if client.soloGame.Phase == game.PhaseGameOver {
			s.handleSoloGameOver(client)
		} else {
			// Show probability stats if enabled
			prefs, err := s.db.GetUserPreferences(client.user.ID)
			if err == nil && prefs.ShowStats && len(client.soloGame.DealerHand.Cards) > 0 {
				dealerUpcard := client.soloGame.DealerHand.Cards[0]
				canDoubleDown := len(client.soloGame.PlayerHand.Cards) == 2
				canSurrender := len(client.soloGame.PlayerHand.Cards) == 2
				response += s.formatProbabilityStats(client.soloGame.PlayerHand, dealerUpcard, client.soloGame.Deck, client.soloGame.Counter, canDoubleDown, canSurrender)
			}

			validActions := client.soloGame.GetValidActions()
			if len(validActions) > 0 {
				response += "\nActions: " + strings.Join(validActions, ", ")
			}
		}

		s.writeResponse(client, response)
	}
}

func (s *Server) handleHit(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		if err := s.mainTable.PlayerHit(client.sessionID); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		player := s.mainTable.GetPlayer(client.sessionID)
		card := player.Hand.Cards[len(player.Hand.Cards)-1]

		s.writeResponse(client, fmt.Sprintf("OK You hit and got %s\nYour hand: %s (Value: %d)",
			card.String(), player.Hand.String(), player.Hand.Value()))

		s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s hits and gets %s (now at %d)",
			client.user.Username, card.String(), player.Hand.Value()))

		if player.HasActed {
			go s.advanceMultiplayerTurn(true, true)
		} else {
			client.conn.Write([]byte("\nYOUR TURN (30 seconds)\n"))

			tableState := s.mainTable.GetTableStateWithoutTurnMarker(true)
			client.conn.Write([]byte(tableState))

			// Show probability stats if enabled (only to current player)
			prefs, err := s.db.GetUserPreferences(client.user.ID)
			if err == nil && prefs.ShowStats && len(s.mainTable.Dealer.Cards) > 0 {
				dealerUpcard := s.mainTable.Dealer.Cards[0]
				canDoubleDown := len(player.Hand.Cards) == 2
				canSurrender := len(player.Hand.Cards) == 2
				statsDisplay := s.formatProbabilityStats(player.Hand, dealerUpcard, s.mainTable.Deck, s.mainTable.Counter, canDoubleDown, canSurrender)
				client.conn.Write([]byte(statsDisplay))
			}

			actions := "Actions: HIT, STAND\n"
			if len(player.Hand.Cards) == 2 {
				actions = "Actions: HIT, STAND, DOUBLEDOWN, SURRENDER\n"
			}
			client.conn.Write([]byte(actions))
		}
	} else {
		if client.soloGame == nil {
			s.writeResponse(client, "ERROR No active game. Use BET <amount> to start a game")
			return
		}

		if err := client.soloGame.Hit(); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		response := fmt.Sprintf("OK\n%s", client.soloGame.GetGameState(true))

		if client.soloGame.Phase == game.PhaseGameOver {
			s.handleSoloGameOver(client)
		} else {
			// Show probability stats if enabled
			prefs, err := s.db.GetUserPreferences(client.user.ID)
			if err == nil && prefs.ShowStats && len(client.soloGame.DealerHand.Cards) > 0 {
				dealerUpcard := client.soloGame.DealerHand.Cards[0]
				canDoubleDown := len(client.soloGame.PlayerHand.Cards) == 2
				canSurrender := len(client.soloGame.PlayerHand.Cards) == 2
				response += s.formatProbabilityStats(client.soloGame.PlayerHand, dealerUpcard, client.soloGame.Deck, client.soloGame.Counter, canDoubleDown, canSurrender)
			}

			validActions := client.soloGame.GetValidActions()
			if len(validActions) > 0 {
				response += "\nActions: " + strings.Join(validActions, ", ")
			}
		}

		s.writeResponse(client, response)
	}
}

func (s *Server) handleStand(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		if err := s.mainTable.PlayerStand(client.sessionID); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		player := s.mainTable.GetPlayer(client.sessionID)
		s.writeResponse(client, fmt.Sprintf("OK You stand with %d", player.Hand.Value()))
		s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s stands with %d",
			client.user.Username, player.Hand.Value()))

		go s.advanceMultiplayerTurn(true, true)
	} else {
		if client.soloGame == nil {
			s.writeResponse(client, "ERROR No active game. Use BET <amount> to start a game")
			return
		}

		if err := client.soloGame.Stand(); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		response := fmt.Sprintf("OK\n%s", client.soloGame.GetGameState(false))
		s.writeResponse(client, response)

		s.handleSoloGameOver(client)
	}
}

func (s *Server) handleDoubleDown(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		player := s.mainTable.GetPlayer(client.sessionID)
		if player == nil {
			s.writeResponse(client, "ERROR Not at table")
			return
		}

		if client.user.Balance < player.Bet {
			s.writeResponse(client, fmt.Sprintf("ERROR Insufficient balance to double down. You need $%.2f more", float64(player.Bet)/100))
			return
		}

		newBalance := client.user.Balance - player.Bet
		if err := s.authService.UpdateBalance(client.user.ID, newBalance); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to update balance: %s", err.Error()))
			return
		}
		client.user.Balance = newBalance

		if err := s.mainTable.PlayerDoubleDown(client.sessionID); err != nil {
			// Refund on error
			s.authService.UpdateBalance(client.user.ID, client.user.Balance+player.Bet)
			client.user.Balance += player.Bet
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		card := player.Hand.Cards[len(player.Hand.Cards)-1]
		s.writeResponse(client, fmt.Sprintf("OK Doubled down! Got %s\nYour hand: %s (Value: %d)",
			card.String(), player.Hand.String(), player.Hand.Value()))
		s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s doubles down and gets %s (now at %d)",
			client.user.Username, card.String(), player.Hand.Value()))

		go s.advanceMultiplayerTurn(true, true) // true = add newline, true = show table state
	} else {
		// Solo mode
		if client.soloGame == nil {
			s.writeResponse(client, "ERROR No active game. Use BET <amount> to start a game")
			return
		}

		if client.user.Balance < client.soloGame.Bet {
			s.writeResponse(client, fmt.Sprintf("ERROR Insufficient balance to double down. You need $%.2f more", float64(client.soloGame.Bet)/100))
			return
		}

		newBalance := client.user.Balance - client.soloGame.Bet
		if err := s.authService.UpdateBalance(client.user.ID, newBalance); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR Failed to update balance: %s", err.Error()))
			return
		}
		client.user.Balance = newBalance

		if err := client.soloGame.DoubleDown(); err != nil {
			s.authService.UpdateBalance(client.user.ID, client.user.Balance+client.soloGame.Bet/2)
			client.user.Balance += client.soloGame.Bet / 2
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		response := fmt.Sprintf("OK Doubled down!\n%s", client.soloGame.GetGameState(false))
		s.writeResponse(client, response)

		s.handleSoloGameOver(client)
	}
}

func (s *Server) handleSurrender(client *ClientState, _ []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Please login first")
		return
	}

	if client.mode == ModeMultiplayer {
		// Multiplayer mode
		if err := s.mainTable.PlayerSurrender(client.sessionID); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		s.writeResponse(client, "OK Surrendered! You'll get half your bet back.")
		s.mainTable.BroadcastToOthers(client.sessionID, fmt.Sprintf("\n%s surrenders", client.user.Username))

		go s.advanceMultiplayerTurn(true, true) // true = add newline, true = show table state
	} else {
		// Solo mode
		if client.soloGame == nil {
			s.writeResponse(client, "ERROR No active game. Use BET <amount> to start a game")
			return
		}

		if err := client.soloGame.Surrender(); err != nil {
			s.writeResponse(client, fmt.Sprintf("ERROR %s", err.Error()))
			return
		}

		response := fmt.Sprintf("OK Surrendered!\n%s", client.soloGame.GetGameState(false))
		s.writeResponse(client, response)

		s.handleSoloGameOver(client)
	}
}

func (s *Server) handleSoloGameOver(client *ClientState) {
	payout := client.soloGame.CalculatePayout()

	newBalance := client.user.Balance + payout
	if err := s.authService.UpdateBalance(client.user.ID, newBalance); err != nil {
		log.Printf("Failed to update balance after game: %v", err)
	}
	client.user.Balance = newBalance

	stats, err := s.authService.GetUserStats(client.user.ID)
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		return
	}

	stats.GamesPlayed++
	stats.TotalBet += client.soloGame.Bet

	switch client.soloGame.Result {
	case game.ResultPlayerWin, game.ResultPlayerBlackjack:
		stats.GamesWon++
		// Add full payout to TotalWon (includes returned bet + profit)
		stats.TotalWon += payout
		// BiggestWin tracks the profit amount only
		winAmount := payout - client.soloGame.Bet
		if winAmount > stats.BiggestWin {
			stats.BiggestWin = winAmount
		}
	case game.ResultDealerWin:
		stats.GamesLost++
		lossAmount := client.soloGame.Bet
		if lossAmount > stats.BiggestLoss {
			stats.BiggestLoss = lossAmount
		}
	case game.ResultSurrender:
		stats.GamesLost++
		// Add the half-bet payout to TotalWon
		stats.TotalWon += payout
		// Loss is half the bet
		lossAmount := client.soloGame.Bet / 2
		if lossAmount > stats.BiggestLoss {
			stats.BiggestLoss = lossAmount
		}
	case game.ResultPush:
		// Push returns the bet, add to TotalWon
		stats.TotalWon += client.soloGame.Bet
	}

	if err := s.db.UpdateUserStats(stats); err != nil {
		log.Printf("Failed to update user stats: %v", err)
	}

	// Clear the game
	client.soloGame = nil
}

// Multiplayer round flow functions

func (s *Server) startMultiplayerRound() {
	time.Sleep(2 * time.Second) // Brief delay before starting

	s.mainTable.StartBettingPhase()
	s.mainTable.BroadcastToAll("\n--- NEW ROUND ---\nPlace your bets! (30 seconds)\nUse: BET <amount>")

	// Wait for bets with timeout
	s.mainTable.WaitForBets(30 * time.Second)

	// Check if enough players bet
	activePlayers := 0
	for _, p := range s.mainTable.Players {
		if p.IsActive && p.Bet > 0 {
			activePlayers++
		}
	}

	if activePlayers < 2 {
		s.mainTable.BroadcastToAll("Not enough players bet. Returning to lobby...")
		s.mainTable.EndRound()
		return
	}

	// Deal cards
	s.dealMultiplayerCards()
}

func (s *Server) dealMultiplayerCards() {
	s.mainTable.BroadcastToAllNoPrompt("\n--- DEALING ---")

	if err := s.mainTable.DealInitialCards(); err != nil {
		s.mainTable.BroadcastToAll(fmt.Sprintf("ERROR: %s", err.Error()))
		s.mainTable.EndRound()
		return
	}

	time.Sleep(1 * time.Second)

	// Show table state to everyone (without turn indicators)
	tableState := s.mainTable.GetTableStateWithoutTurnMarker(true)
	s.mainTable.BroadcastToAllNoPrompt(tableState)

	// Check if game is already over (dealer blackjack or all player blackjacks)
	if s.mainTable.Phase == game.TablePhaseDealerTurn {
		s.playMultiplayerDealerTurn()
		return
	}

	// Start player turns
	s.promptCurrentPlayer(false, false) // false = no leading newline, false = don't show table (already shown)
}

func (s *Server) promptCurrentPlayer(addLeadingNewline bool, showTableState bool) {
	currentPlayer := s.mainTable.GetCurrentPlayer()
	if currentPlayer == nil {
		// All players done, move to dealer turn
		s.playMultiplayerDealerTurn()
		return
	}

	// Skip if player already acted (blackjack or bust)
	if currentPlayer.HasActed {
		s.advanceMultiplayerTurn(addLeadingNewline, showTableState) // Pass through both flags
		return
	}

	// Announce whose turn it is to other players
	turnAnnouncement := fmt.Sprintf("%s's turn...", currentPlayer.Username)
	if addLeadingNewline {
		turnAnnouncement = "\n" + turnAnnouncement
	}
	s.mainTable.BroadcastToOthers(currentPlayer.SessionID, turnAnnouncement)

	// Show detailed turn info to the current player
	turnMsg := "YOUR TURN (30 seconds)\n"
	if addLeadingNewline {
		turnMsg = "\n" + turnMsg
	}
	currentPlayer.Connection.Write([]byte(turnMsg))

	// Show current table state so player can see all hands (only if requested)
	if showTableState {
		tableState := s.mainTable.GetTableStateWithoutTurnMarker(true) // true = hide dealer's hole card
		currentPlayer.Connection.Write([]byte(tableState))
	}

	// Show probability stats if enabled (only to current player)
	session, err := s.db.GetSession(currentPlayer.SessionID)
	if err == nil {
		prefs, err := s.db.GetUserPreferences(session.UserID)
		if err == nil && prefs.ShowStats && len(s.mainTable.Dealer.Cards) > 0 {
			dealerUpcard := s.mainTable.Dealer.Cards[0]
			canDoubleDown := len(currentPlayer.Hand.Cards) == 2
			canSurrender := len(currentPlayer.Hand.Cards) == 2
			statsDisplay := s.formatProbabilityStats(currentPlayer.Hand, dealerUpcard, s.mainTable.Deck, s.mainTable.Counter, canDoubleDown, canSurrender)
			currentPlayer.Connection.Write([]byte(statsDisplay))
		}
	}

	currentPlayer.Connection.Write([]byte("Actions: HIT, STAND, DOUBLEDOWN, SURRENDER\n"))
	currentPlayer.Connection.Write([]byte("<<<PROMPT>>>\n"))
}

func (s *Server) advanceMultiplayerTurn(addLeadingNewline bool, showTableState bool) {
	time.Sleep(500 * time.Millisecond) // Brief pause between turns

	if s.mainTable.AdvanceTurn() {
		// More players to act
		s.promptCurrentPlayer(addLeadingNewline, showTableState)
	} else {
		// All players done, dealer's turn
		s.playMultiplayerDealerTurn()
	}
}

func (s *Server) playMultiplayerDealerTurn() {
	time.Sleep(1 * time.Second)

	s.mainTable.BroadcastToAllNoPrompt("\n--- DEALER TURN ---")

	if err := s.mainTable.PlayDealerHand(); err != nil {
		log.Printf("Error playing dealer hand: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Show final dealer hand
	dealerState := fmt.Sprintf("Dealer: %s (Value: %d)", s.mainTable.Dealer.String(), s.mainTable.Dealer.Value())
	s.mainTable.BroadcastToAllNoPrompt(dealerState)

	// Calculate payouts
	s.mainTable.CalculatePayouts()

	time.Sleep(1 * time.Second)

	// Show results and update balances
	s.mainTable.BroadcastToAllNoPrompt("--- RESULTS ---")

	for _, player := range s.mainTable.Players {
		if !player.IsActive || player.Bet == 0 {
			continue
		}

		// Update player balance - get user by session
		session, err := s.db.GetSession(player.SessionID)
		if err != nil {
			log.Printf("Failed to get session %s: %v", player.SessionID, err)
			continue
		}

		user, err := s.db.GetUserByID(session.UserID)
		if err != nil {
			log.Printf("Failed to get user %d: %v", session.UserID, err)
			continue
		}

		newBalance := user.Balance + player.Payout
		if err := s.authService.UpdateBalance(user.ID, newBalance); err != nil {
			log.Printf("Failed to update balance for user %d: %v", user.ID, err)
		}

		// Update stats
		stats, err := s.authService.GetUserStats(user.ID)
		if err != nil {
			log.Printf("Failed to get stats for user %d: %v", user.ID, err)
			continue
		}

		stats.GamesPlayed++
		stats.TotalBet += player.Bet

		switch player.Result {
		case game.ResultPlayerWin, game.ResultPlayerBlackjack:
			stats.GamesWon++
			stats.TotalWon += player.Payout
			winAmount := player.Payout - player.Bet
			if winAmount > stats.BiggestWin {
				stats.BiggestWin = winAmount
			}
		case game.ResultDealerWin, game.ResultPlayerBust:
			stats.GamesLost++
			if player.Bet > stats.BiggestLoss {
				stats.BiggestLoss = player.Bet
			}
		case game.ResultSurrender:
			stats.GamesLost++
			stats.TotalWon += player.Payout
			lossAmount := player.Bet / 2
			if lossAmount > stats.BiggestLoss {
				stats.BiggestLoss = lossAmount
			}
		case game.ResultPush:
			stats.TotalWon += player.Bet
		}

		if err := s.db.UpdateUserStats(stats); err != nil {
			log.Printf("Failed to update stats for user %d: %v", user.ID, err)
		}

		// Broadcast result
		resultMsg := fmt.Sprintf("%s: %s → %s", player.Username, player.Result, formatPayout(player.Payout, player.Bet))
		s.mainTable.BroadcastToAllNoPrompt(resultMsg)
	}

	// End round and return to lobby
	time.Sleep(3 * time.Second)
	s.mainTable.EndRound()
	s.mainTable.BroadcastToAll("Round complete! Type READY to play again.")
}

func formatPayout(payout, bet int64) string {
	if payout == 0 {
		return fmt.Sprintf("LOSS (-$%.2f)", float64(bet)/100)
	} else if payout == bet {
		return "PUSH ($0.00)"
	} else {
		profit := payout - bet
		return fmt.Sprintf("WIN (+$%.2f)", float64(profit)/100)
	}
}
