package game

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// TablePhase represents the current phase of the multiplayer table
type TablePhase string

const (
	PhaseLobby   TablePhase = "LOBBY"   // Waiting for players
	PhaseBetting TablePhase = "BETTING" // Players placing bets
	PhaseDealing TablePhase = "DEALING" // Initial cards dealt
	PhasePayouts TablePhase = "PAYOUTS" // Calculating results
)

// Reuse GamePhase constants for player and dealer turns
const (
	TablePhasePlayerTurn TablePhase = "PLAYER_TURN"
	TablePhaseDealerTurn TablePhase = "DEALER_TURN"
)

// PlayerSeat represents a player's seat at the multiplayer table
type PlayerSeat struct {
	SessionID  string
	Username   string
	Connection net.Conn
	Hand       *Hand
	Bet        int64
	IsReady    bool
	IsActive   bool // false if disconnected
	HasActed   bool // true after player completes their turn
	Result     GameResult
	Payout     int64
}

// Table represents a multiplayer Blackjack table
type Table struct {
	ID          string
	Dealer      *Hand
	Deck        *Deck
	Players     []*PlayerSeat
	Phase       TablePhase
	CurrentTurn int // Index of player whose turn it is
	MinBet      int64
	MaxBet      int64
	MaxPlayers  int
	RoundActive bool
	mu          sync.RWMutex
}

// NewTable creates a new multiplayer table
func NewTable(id string, minBet, maxBet int64, maxPlayers int) *Table {
	return &Table{
		ID:         id,
		Dealer:     NewHand(),
		Deck:       NewDeck(),
		Players:    make([]*PlayerSeat, 0, maxPlayers),
		Phase:      PhaseLobby,
		MinBet:     minBet,
		MaxBet:     maxBet,
		MaxPlayers: maxPlayers,
	}
}

// AddPlayer adds a player to the table
func (t *Table) AddPlayer(sessionID, username string, conn net.Conn) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if already at table
	for _, p := range t.Players {
		if p.SessionID == sessionID {
			return fmt.Errorf("already at table")
		}
	}

	// Check if table is full
	if len(t.Players) >= t.MaxPlayers {
		return fmt.Errorf("table is full")
	}

	seat := &PlayerSeat{
		SessionID:  sessionID,
		Username:   username,
		Connection: conn,
		Hand:       NewHand(),
		IsActive:   true,
		IsReady:    false,
	}

	t.Players = append(t.Players, seat)
	return nil
}

// RemovePlayer removes a player from the table
func (t *Table) RemovePlayer(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, p := range t.Players {
		if p.SessionID == sessionID {
			// Mark as inactive if round is active
			if t.RoundActive {
				p.IsActive = false
			} else {
				// Remove from table if no round active
				t.Players = append(t.Players[:i], t.Players[i+1:]...)
			}
			break
		}
	}
}

// GetPlayer returns the player seat for a given session ID
func (t *Table) GetPlayer(sessionID string) *PlayerSeat {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Players {
		if p.SessionID == sessionID {
			return p
		}
	}
	return nil
}

// PlayerCount returns the number of active players
func (t *Table) PlayerCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, p := range t.Players {
		if p.IsActive {
			count++
		}
	}
	return count
}

// SetReady marks a player as ready
func (t *Table) SetReady(sessionID string, ready bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if t.Phase != PhaseLobby {
		return fmt.Errorf("round already in progress")
	}

	player.IsReady = ready
	return nil
}

// AllPlayersReady checks if all active players are ready
func (t *Table) AllPlayersReady() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	activeCount := 0
	readyCount := 0

	for _, p := range t.Players {
		if p.IsActive {
			activeCount++
			if p.IsReady {
				readyCount++
			}
		}
	}

	// Need at least 2 players and all must be ready
	return activeCount >= 2 && activeCount == readyCount
}

// PlaceBet places a bet for a player
func (t *Table) PlaceBet(sessionID string, amount int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Phase != PhaseBetting {
		return fmt.Errorf("not in betting phase")
	}

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if amount < t.MinBet || amount > t.MaxBet {
		return fmt.Errorf("bet must be between $%.2f and $%.2f",
			float64(t.MinBet)/100, float64(t.MaxBet)/100)
	}

	player.Bet = amount
	return nil
}

// HasPlayerBet checks if a player has placed a bet
func (t *Table) HasPlayerBet(sessionID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	player := t.getPlayerUnsafe(sessionID)
	return player != nil && player.Bet > 0
}

// AllPlayersBet checks if all active players have placed bets
func (t *Table) AllPlayersBet() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Players {
		if p.IsActive && p.Bet == 0 {
			return false
		}
	}
	return true
}

// StartBettingPhase starts the betting phase
func (t *Table) StartBettingPhase() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Phase = PhaseBetting
	t.RoundActive = true

	// Reset all player states
	for _, p := range t.Players {
		p.Hand = NewHand()
		p.Bet = 0
		p.HasActed = false
		p.Result = ""
		p.Payout = 0
		p.IsReady = false
	}
}

// DealInitialCards deals 2 cards to each player and dealer
func (t *Table) DealInitialCards() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Phase = PhaseDealing

	// Shuffle deck
	t.Deck.Shuffle()

	// Deal first card to each player
	for _, p := range t.Players {
		if p.IsActive && p.Bet > 0 {
			card, err := t.Deck.Draw()
			if err != nil {
				return fmt.Errorf("deck exhausted during deal: %w", err)
			}
			p.Hand.AddCard(card)
		}
	}

	// Deal first card to dealer
	card, err := t.Deck.Draw()
	if err != nil {
		return fmt.Errorf("deck exhausted during deal: %w", err)
	}
	t.Dealer.AddCard(card)

	// Deal second card to each player
	for _, p := range t.Players {
		if p.IsActive && p.Bet > 0 {
			card, err := t.Deck.Draw()
			if err != nil {
				return fmt.Errorf("deck exhausted during deal: %w", err)
			}
			p.Hand.AddCard(card)
		}
	}

	// Deal second card to dealer
	card, err = t.Deck.Draw()
	if err != nil {
		return fmt.Errorf("deck exhausted during deal: %w", err)
	}
	t.Dealer.AddCard(card)

	// Check for dealer blackjack
	if t.Dealer.IsBlackjack() {
		t.Phase = TablePhaseDealerTurn
		for _, p := range t.Players {
			if p.IsActive && p.Bet > 0 {
				if p.Hand.IsBlackjack() {
					p.Result = ResultPush
				} else {
					p.Result = ResultDealerWin
				}
			}
		}
		return nil
	}

	// Check for player blackjacks
	allBlackjack := true
	for _, p := range t.Players {
		if p.IsActive && p.Bet > 0 {
			if p.Hand.IsBlackjack() {
				p.Result = ResultPlayerBlackjack
				p.HasActed = true
			} else {
				allBlackjack = false
			}
		}
	}

	// If all players have blackjack, go straight to dealer turn
	if allBlackjack {
		t.Phase = TablePhaseDealerTurn
	} else {
		t.Phase = TablePhasePlayerTurn
		t.CurrentTurn = 0
	}

	return nil
}

// GetCurrentPlayer returns the player whose turn it is
func (t *Table) GetCurrentPlayer() *PlayerSeat {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Phase != TablePhasePlayerTurn {
		return nil
	}

	if t.CurrentTurn >= len(t.Players) {
		return nil
	}

	return t.Players[t.CurrentTurn]
}

// AdvanceTurn moves to the next player's turn
func (t *Table) AdvanceTurn() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find next active player who hasn't acted
	for i := t.CurrentTurn + 1; i < len(t.Players); i++ {
		if t.Players[i].IsActive && t.Players[i].Bet > 0 && !t.Players[i].HasActed {
			t.CurrentTurn = i
			return true
		}
	}

	// No more players, move to dealer turn
	t.Phase = TablePhaseDealerTurn
	return false
}

// PlayerHit handles a player hitting
func (t *Table) PlayerHit(sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Phase != TablePhasePlayerTurn {
		return fmt.Errorf("not player turn phase")
	}

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if t.Players[t.CurrentTurn].SessionID != sessionID {
		return fmt.Errorf("not your turn")
	}

	if player.HasActed {
		return fmt.Errorf("already acted this round")
	}

	card, err := t.Deck.Draw()
	if err != nil {
		return fmt.Errorf("deck exhausted: %w", err)
	}

	player.Hand.AddCard(card)

	// Check for bust
	if player.Hand.Value() > 21 {
		player.Result = ResultPlayerBust
		player.HasActed = true
	}

	return nil
}

// PlayerStand handles a player standing
func (t *Table) PlayerStand(sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Phase != TablePhasePlayerTurn {
		return fmt.Errorf("not player turn phase")
	}

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if t.Players[t.CurrentTurn].SessionID != sessionID {
		return fmt.Errorf("not your turn")
	}

	if player.HasActed {
		return fmt.Errorf("already acted this round")
	}

	player.HasActed = true
	return nil
}

// PlayerDoubleDown handles a player doubling down
func (t *Table) PlayerDoubleDown(sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Phase != TablePhasePlayerTurn {
		return fmt.Errorf("not player turn phase")
	}

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if t.Players[t.CurrentTurn].SessionID != sessionID {
		return fmt.Errorf("not your turn")
	}

	if player.HasActed {
		return fmt.Errorf("already acted this round")
	}

	if len(player.Hand.Cards) != 2 {
		return fmt.Errorf("can only double down on initial hand")
	}

	// Double the bet
	player.Bet *= 2

	// Draw one card
	card, err := t.Deck.Draw()
	if err != nil {
		return fmt.Errorf("deck exhausted: %w", err)
	}

	player.Hand.AddCard(card)

	// Check for bust
	if player.Hand.Value() > 21 {
		player.Result = ResultPlayerBust
	}

	player.HasActed = true
	return nil
}

// PlayerSurrender handles a player surrendering
func (t *Table) PlayerSurrender(sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Phase != TablePhasePlayerTurn {
		return fmt.Errorf("not player turn phase")
	}

	player := t.getPlayerUnsafe(sessionID)
	if player == nil {
		return fmt.Errorf("not at table")
	}

	if t.Players[t.CurrentTurn].SessionID != sessionID {
		return fmt.Errorf("not your turn")
	}

	if player.HasActed {
		return fmt.Errorf("already acted this round")
	}

	if len(player.Hand.Cards) != 2 {
		return fmt.Errorf("can only surrender on initial hand")
	}

	player.Result = ResultSurrender
	player.HasActed = true
	return nil
}

// PlayDealerHand plays out the dealer's hand
func (t *Table) PlayDealerHand() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Phase = TablePhaseDealerTurn

	// Check if any player is still in play (not bust, not surrendered, not blackjack)
	anyPlayerInPlay := false
	for _, p := range t.Players {
		if p.IsActive && p.Bet > 0 && p.Result == "" {
			anyPlayerInPlay = true
			break
		}
	}

	// If no players in play, dealer doesn't need to draw
	if !anyPlayerInPlay {
		return nil
	}

	// Dealer draws until 17 or higher
	for t.Dealer.Value() < 17 {
		card, err := t.Deck.Draw()
		if err != nil {
			// Deck exhausted - treat as push for remaining players
			for _, p := range t.Players {
				if p.IsActive && p.Bet > 0 && p.Result == "" {
					p.Result = ResultPush
				}
			}
			return nil
		}
		t.Dealer.AddCard(card)
	}

	return nil
}

// CalculatePayouts calculates results and payouts for all players
func (t *Table) CalculatePayouts() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Phase = PhasePayouts

	dealerValue := t.Dealer.Value()
	dealerBust := dealerValue > 21

	for _, p := range t.Players {
		if !p.IsActive || p.Bet == 0 {
			continue
		}

		// Skip if result already determined (blackjack, bust, surrender)
		if p.Result != "" {
			switch p.Result {
			case ResultPlayerBlackjack:
				p.Payout = p.Bet + (p.Bet * 3 / 2)
			case ResultSurrender:
				p.Payout = p.Bet / 2
			case ResultPush:
				p.Payout = p.Bet
			default:
				p.Payout = 0
			}
			continue
		}

		playerValue := p.Hand.Value()

		// Determine result
		if dealerBust {
			p.Result = ResultPlayerWin
			p.Payout = p.Bet * 2
		} else if playerValue > dealerValue {
			p.Result = ResultPlayerWin
			p.Payout = p.Bet * 2
		} else if playerValue == dealerValue {
			p.Result = ResultPush
			p.Payout = p.Bet
		} else {
			p.Result = ResultDealerWin
			p.Payout = 0
		}
	}
}

// EndRound ends the current round and returns to lobby
func (t *Table) EndRound() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Remove inactive players
	activePlayers := make([]*PlayerSeat, 0)
	for _, p := range t.Players {
		if p.IsActive {
			activePlayers = append(activePlayers, p)
		}
	}
	t.Players = activePlayers

	t.Phase = PhaseLobby
	t.RoundActive = false
	t.CurrentTurn = 0
	t.Dealer = NewHand()
}

// GetTableState returns a formatted string of the current table state
func (t *Table) GetTableState(hideDealer bool) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state := ""

	// Show all players
	for i, p := range t.Players {
		if !p.IsActive || p.Bet == 0 {
			continue
		}

		prefix := fmt.Sprintf("Seat %d - %s:", i+1, p.Username)
		turnMarker := ""
		if t.Phase == TablePhasePlayerTurn && t.CurrentTurn == i {
			turnMarker = " [YOUR TURN]"
		}

		state += fmt.Sprintf("%s %s (Value: %d)", prefix, p.Hand.String(), p.Hand.Value())
		if p.Bet > 0 {
			state += fmt.Sprintf(" | Bet: $%.2f", float64(p.Bet)/100)
		}
		state += turnMarker + "\n"
	}

	// Show dealer
	if hideDealer && t.Phase == TablePhasePlayerTurn && len(t.Dealer.Cards) > 0 {
		state += fmt.Sprintf("Dealer: [%s%s] [Hidden]\n", t.Dealer.Cards[0].Rank, t.Dealer.Cards[0].Suit)
	} else {
		state += fmt.Sprintf("Dealer: %s (Value: %d)\n", t.Dealer.String(), t.Dealer.Value())
	}

	return state
}

// GetTableStateWithoutTurnMarker returns table state without [YOUR TURN] marker
func (t *Table) GetTableStateWithoutTurnMarker(hideDealer bool) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state := ""

	// Show all players without turn marker
	for i, p := range t.Players {
		if !p.IsActive || p.Bet == 0 {
			continue
		}

		prefix := fmt.Sprintf("Seat %d - %s:", i+1, p.Username)

		state += fmt.Sprintf("%s %s (Value: %d)", prefix, p.Hand.String(), p.Hand.Value())
		if p.Bet > 0 {
			state += fmt.Sprintf(" | Bet: $%.2f", float64(p.Bet)/100)
		}
		state += "\n"
	}

	// Show dealer
	if hideDealer && t.Phase == TablePhasePlayerTurn && len(t.Dealer.Cards) > 0 {
		state += fmt.Sprintf("Dealer: [%s%s] [Hidden]\n", t.Dealer.Cards[0].Rank, t.Dealer.Cards[0].Suit)
	} else {
		state += fmt.Sprintf("Dealer: %s (Value: %d)\n", t.Dealer.String(), t.Dealer.Value())
	}

	return state
}

// GetPlayerList returns a formatted list of players at the table
func (t *Table) GetPlayerList() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	list := fmt.Sprintf("Players at table (%d/%d):\n", len(t.Players), t.MaxPlayers)
	for i, p := range t.Players {
		status := ""
		if p.IsReady {
			status = " [READY]"
		}
		if !p.IsActive {
			status = " [DISCONNECTED]"
		}
		list += fmt.Sprintf("  Seat %d: %s%s\n", i+1, p.Username, status)
	}
	return list
}

// BroadcastToAll sends a message to all active players at the table with prompt redisplay
func (t *Table) BroadcastToAll(message string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Players {
		if p.IsActive && p.Connection != nil {
			p.Connection.Write([]byte(message + "\n"))
			p.Connection.Write([]byte("<<<PROMPT>>>\n"))
		}
	}
}

// BroadcastToAllNoPrompt sends a message to all active players without prompt redisplay
// Use this for sequential game flow messages that should appear together
func (t *Table) BroadcastToAllNoPrompt(message string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Players {
		if p.IsActive && p.Connection != nil {
			p.Connection.Write([]byte(message + "\n"))
		}
	}
}

// BroadcastToOthers sends a message to all players except the specified one
func (t *Table) BroadcastToOthers(excludeSessionID string, message string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Players {
		if p.IsActive && p.SessionID != excludeSessionID && p.Connection != nil {
			p.Connection.Write([]byte(message + "\n"))
			p.Connection.Write([]byte("<<<PROMPT>>>\n"))
		}
	}
}

// getPlayerUnsafe returns a player without locking (caller must hold lock)
func (t *Table) getPlayerUnsafe(sessionID string) *PlayerSeat {
	for _, p := range t.Players {
		if p.SessionID == sessionID {
			return p
		}
	}
	return nil
}

// WaitForBets waits for all players to place bets or times out
func (t *Table) WaitForBets(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if t.AllPlayersBet() {
				return
			}
			if time.Now().After(deadline) {
				// Timeout - remove players who didn't bet
				t.mu.Lock()
				for _, p := range t.Players {
					if p.IsActive && p.Bet == 0 {
						p.IsActive = false
					}
				}
				t.mu.Unlock()
				return
			}
		}
	}
}
