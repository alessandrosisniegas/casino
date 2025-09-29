package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alessandrosisniegas/casino/core/security"
	"github.com/alessandrosisniegas/casino/core/vault"
)

type ClientState struct {
	conn      net.Conn
	sessionID string
	user      *vault.User
}

type Server struct {
	authService *security.AuthService
	db          *vault.DB
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

	server := &Server{
		authService: authService,
		db:          db,
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

	client := &ClientState{conn: conn}
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

	s.writeResponse(client, fmt.Sprintf("OK Welcome back, %s! Balance: $%.2f", user.Username, float64(user.Balance)/100))
}

func (s *Server) handleLogout(client *ClientState, args []string) {
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

func (s *Server) handleBalance(client *ClientState, args []string) {
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

func (s *Server) handleStats(client *ClientState, args []string) {
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
	if stats.GamesPlayed > 0 {
		winRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
	}

	response := fmt.Sprintf("OK Stats for %s:\n", client.user.Username)
	response += fmt.Sprintf("  Games Played: %d\n", stats.GamesPlayed)
	response += fmt.Sprintf("  Games Won: %d\n", stats.GamesWon)
	response += fmt.Sprintf("  Games Lost: %d\n", stats.GamesLost)
	response += fmt.Sprintf("  Win Rate: %.1f%%\n", winRate)
	response += fmt.Sprintf("  Total Bet: $%.2f\n", float64(stats.TotalBet)/100)
	response += fmt.Sprintf("  Total Won: $%.2f\n", float64(stats.TotalWon)/100)
	response += fmt.Sprintf("  Net: $%.2f\n", float64(stats.TotalWon-stats.TotalBet)/100)
	response += fmt.Sprintf("  Biggest Win: $%.2f\n", float64(stats.BiggestWin)/100)
	response += fmt.Sprintf("  Biggest Loss: $%.2f", float64(stats.BiggestLoss)/100)

	s.writeResponse(client, response)
}

func (s *Server) handleWhoami(client *ClientState, args []string) {
	if client.user == nil {
		s.writeResponse(client, "ERROR Not logged in")
		return
	}

	s.writeResponse(client, fmt.Sprintf("OK Logged in as: %s (ID: %d, Balance: $%.2f)",
		client.user.Username, client.user.ID, float64(client.user.Balance)/100))
}

func (s *Server) handleHelp(client *ClientState, args []string) {
	help := "OK Available commands:\n"
	help += "  SIGNUP <username> <password> - Create a new account\n"
	help += "  LOGIN <username> <password>  - Login to your account\n"
	help += "  LOGOUT                       - Logout from your account\n"
	help += "  BALANCE                      - Check your current balance\n"
	help += "  STATS                        - View your game statistics\n"
	help += "  WHOAMI                       - Show current login status\n"
	help += "  HELP                         - Show this help message\n"
	help += "  QUIT                         - Disconnect from server\n"
	help += "\nUsername & Password requirements:\n"
	help += "  - 2-30 characters long\n"
	help += "  - Letters, numbers, and underscores only\n"
	help += "  - No whitespace allowed\n"
	help += "  - Cannot be the same as username"

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
