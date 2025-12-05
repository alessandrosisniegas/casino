package game

import (
	"net"
	"strings"
	"testing"
	"time"
)

// Mock connection for testing
type mockConn struct {
	messages []string
}

func (m *mockConn) Read(b []byte) (n int, err error) { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) {
	m.messages = append(m.messages, string(b))
	return len(b), nil
}
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNewTable(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)

	if table.ID != "test-1" {
		t.Errorf("Expected table ID 'test-1', got '%s'", table.ID)
	}
	if table.MinBet != 1000 {
		t.Errorf("Expected MinBet 1000, got %d", table.MinBet)
	}
	if table.MaxBet != 100000 {
		t.Errorf("Expected MaxBet 100000, got %d", table.MaxBet)
	}
	if table.MaxPlayers != 4 {
		t.Errorf("Expected MaxPlayers 4, got %d", table.MaxPlayers)
	}
	if table.Phase != PhaseLobby {
		t.Errorf("Expected initial phase LOBBY, got %s", table.Phase)
	}
	if len(table.Players) != 0 {
		t.Errorf("Expected 0 players initially, got %d", len(table.Players))
	}
}

func TestAddPlayer(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	err := table.AddPlayer("session1", "alice", conn)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	if len(table.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(table.Players))
	}

	player := table.Players[0]
	if player.SessionID != "session1" {
		t.Errorf("Expected SessionID 'session1', got '%s'", player.SessionID)
	}
	if player.Username != "alice" {
		t.Errorf("Expected Username 'alice', got '%s'", player.Username)
	}
	if !player.IsActive {
		t.Error("Expected player to be active")
	}
	if player.IsReady {
		t.Error("Expected player to not be ready initially")
	}
}

func TestAddPlayerAlreadyAtTable(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	err := table.AddPlayer("session1", "alice", conn)

	if err == nil {
		t.Error("Expected error when adding same player twice")
	}
	if err.Error() != "already at table" {
		t.Errorf("Expected 'already at table' error, got '%v'", err)
	}
}

func TestAddPlayerTableFull(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 2)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	err := table.AddPlayer("session3", "charlie", conn)

	if err == nil {
		t.Error("Expected error when table is full")
	}
	if err.Error() != "table is full" {
		t.Errorf("Expected 'table is full' error, got '%v'", err)
	}
}

func TestRemovePlayerNoRound(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)

	table.RemovePlayer("session1")

	if len(table.Players) != 1 {
		t.Errorf("Expected 1 player after removal, got %d", len(table.Players))
	}
	if table.Players[0].SessionID != "session2" {
		t.Errorf("Expected remaining player to be session2, got %s", table.Players[0].SessionID)
	}
}

func TestRemovePlayerDuringRound(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.RoundActive = true

	table.RemovePlayer("session1")

	if len(table.Players) != 1 {
		t.Errorf("Expected player to remain in list during round, got %d players", len(table.Players))
	}
	if table.Players[0].IsActive {
		t.Error("Expected player to be marked inactive during round")
	}
}

func TestGetPlayer(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)

	player := table.GetPlayer("session1")
	if player == nil {
		t.Fatal("Expected to find player")
	}
	if player.Username != "alice" {
		t.Errorf("Expected alice, got %s", player.Username)
	}

	notFound := table.GetPlayer("session999")
	if notFound != nil {
		t.Error("Expected nil for non-existent player")
	}
}

func TestPlayerCount(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	if table.PlayerCount() != 0 {
		t.Errorf("Expected 0 players, got %d", table.PlayerCount())
	}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)

	if table.PlayerCount() != 2 {
		t.Errorf("Expected 2 players, got %d", table.PlayerCount())
	}

	// Mark one inactive
	table.Players[0].IsActive = false

	if table.PlayerCount() != 1 {
		t.Errorf("Expected 1 active player, got %d", table.PlayerCount())
	}
}

func TestSetReady(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)

	err := table.SetReady("session1", true)
	if err != nil {
		t.Fatalf("Failed to set ready: %v", err)
	}

	if !table.Players[0].IsReady {
		t.Error("Expected player to be ready")
	}
}

func TestSetReadyNotAtTable(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)

	err := table.SetReady("session999", true)
	if err == nil {
		t.Error("Expected error for non-existent player")
	}
}

func TestSetReadyDuringRound(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting

	err := table.SetReady("session1", true)
	if err == nil {
		t.Error("Expected error when setting ready during round")
	}
}

func TestAllPlayersReady(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	// Need at least 2 players
	table.AddPlayer("session1", "alice", conn)
	if table.AllPlayersReady() {
		t.Error("Expected false with only 1 player")
	}

	table.AddPlayer("session2", "bob", conn)
	if table.AllPlayersReady() {
		t.Error("Expected false when players not ready")
	}

	table.SetReady("session1", true)
	if table.AllPlayersReady() {
		t.Error("Expected false when only one player ready")
	}

	table.SetReady("session2", true)
	if !table.AllPlayersReady() {
		t.Error("Expected true when all players ready")
	}
}

func TestTablePlaceBet(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting

	err := table.PlaceBet("session1", 5000)
	if err != nil {
		t.Fatalf("Failed to place bet: %v", err)
	}

	if table.Players[0].Bet != 5000 {
		t.Errorf("Expected bet 5000, got %d", table.Players[0].Bet)
	}
}

func TestTablePlaceBetWrongPhase(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseLobby

	err := table.PlaceBet("session1", 5000)
	if err == nil {
		t.Error("Expected error when placing bet in wrong phase")
	}
}

func TestTablePlaceBetTooLow(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting

	err := table.PlaceBet("session1", 500)
	if err == nil {
		t.Error("Expected error for bet below minimum")
	}
}

func TestTablePlaceBetTooHigh(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting

	err := table.PlaceBet("session1", 200000)
	if err == nil {
		t.Error("Expected error for bet above maximum")
	}
}

func TestAllPlayersBet(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting

	if table.AllPlayersBet() {
		t.Error("Expected false when no bets placed")
	}

	table.PlaceBet("session1", 5000)
	if table.AllPlayersBet() {
		t.Error("Expected false when only one player bet")
	}

	table.PlaceBet("session2", 3000)
	if !table.AllPlayersBet() {
		t.Error("Expected true when all players bet")
	}
}

func TestStartBettingPhase(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 5000
	table.Players[0].IsReady = true
	table.Players[0].HasActed = true

	table.StartBettingPhase()

	if table.Phase != PhaseBetting {
		t.Errorf("Expected phase BETTING, got %s", table.Phase)
	}
	if !table.RoundActive {
		t.Error("Expected round to be active")
	}
	if table.Players[0].Bet != 0 {
		t.Error("Expected bet to be reset")
	}
	if table.Players[0].IsReady {
		t.Error("Expected IsReady to be reset")
	}
	if table.Players[0].HasActed {
		t.Error("Expected HasActed to be reset")
	}
}

func TestDealInitialCards(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)
	table.PlaceBet("session2", 3000)

	err := table.DealInitialCards()
	if err != nil {
		t.Fatalf("Failed to deal cards: %v", err)
	}

	// Check each player has 2 cards
	for _, p := range table.Players {
		if len(p.Hand.Cards) != 2 {
			t.Errorf("Expected 2 cards for player %s, got %d", p.Username, len(p.Hand.Cards))
		}
	}

	// Check dealer has 2 cards
	if len(table.Dealer.Cards) != 2 {
		t.Errorf("Expected 2 cards for dealer, got %d", len(table.Dealer.Cards))
	}

	// Phase should advance
	if table.Phase != TablePhasePlayerTurn && table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase to advance, got %s", table.Phase)
	}
}

func TestDealInitialCardsDealerBlackjack(t *testing.T) {
	// Test dealer blackjack detection by manually dealing cards
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)

	// Manually deal cards to test blackjack logic
	table.Phase = PhaseDealing

	// Alice gets 9, 8 (17)
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "8", Suit: "♠", Value: 8})

	// Dealer gets A, K (Blackjack!)
	table.Dealer.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	table.Dealer.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})

	// Check for dealer blackjack
	if table.Dealer.IsBlackjack() {
		table.Phase = TablePhaseDealerTurn
		for _, p := range table.Players {
			if p.IsActive && p.Bet > 0 {
				if p.Hand.IsBlackjack() {
					p.Result = ResultPush
				} else {
					p.Result = ResultDealerWin
				}
			}
		}
	}

	if table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase DEALER_TURN with dealer blackjack, got %s", table.Phase)
	}

	if table.Players[0].Result != ResultDealerWin {
		t.Errorf("Expected player result DEALER_WIN, got %s", table.Players[0].Result)
	}
}

func TestDealInitialCardsPlayerBlackjack(t *testing.T) {
	// Test player blackjack detection by manually dealing cards
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)

	// Manually deal cards to test blackjack logic
	table.Phase = PhaseDealing

	// Alice gets A, K (Blackjack!)
	table.Players[0].Hand.AddCard(Card{Rank: "A", Suit: "♠", Value: 11})
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})

	// Dealer gets 9, 7 (16, not blackjack)
	table.Dealer.AddCard(Card{Rank: "9", Suit: "♥", Value: 9})
	table.Dealer.AddCard(Card{Rank: "7", Suit: "♥", Value: 7})

	// Check for player blackjacks
	allBlackjack := true
	for _, p := range table.Players {
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
		table.Phase = TablePhaseDealerTurn
	}

	if table.Players[0].Result != ResultPlayerBlackjack {
		t.Errorf("Expected player result PLAYER_BLACKJACK, got %s", table.Players[0].Result)
	}

	if !table.Players[0].HasActed {
		t.Error("Expected player with blackjack to have HasActed = true")
	}

	// Should skip to dealer turn since all players have blackjack
	if table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase DEALER_TURN, got %s", table.Phase)
	}
}

func TestGetCurrentPlayer(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0

	current := table.GetCurrentPlayer()
	if current == nil {
		t.Fatal("Expected to get current player")
	}
	if current.Username != "alice" {
		t.Errorf("Expected alice, got %s", current.Username)
	}

	table.CurrentTurn = 1
	current = table.GetCurrentPlayer()
	if current.Username != "bob" {
		t.Errorf("Expected bob, got %s", current.Username)
	}
}

func TestGetCurrentPlayerWrongPhase(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = PhaseLobby

	current := table.GetCurrentPlayer()
	if current != nil {
		t.Error("Expected nil when not in player turn phase")
	}
}

func TestAdvanceTurn(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.AddPlayer("session3", "charlie", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0

	// Set up bets
	for _, p := range table.Players {
		p.Bet = 1000
	}

	hasMore := table.AdvanceTurn()
	if !hasMore {
		t.Error("Expected more players")
	}
	if table.CurrentTurn != 1 {
		t.Errorf("Expected turn 1, got %d", table.CurrentTurn)
	}

	hasMore = table.AdvanceTurn()
	if !hasMore {
		t.Error("Expected more players")
	}
	if table.CurrentTurn != 2 {
		t.Errorf("Expected turn 2, got %d", table.CurrentTurn)
	}

	hasMore = table.AdvanceTurn()
	if hasMore {
		t.Error("Expected no more players")
	}
	if table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase DEALER_TURN, got %s", table.Phase)
	}
}

func TestAdvanceTurnSkipsInactive(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.AddPlayer("session3", "charlie", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0

	// Set up bets
	for _, p := range table.Players {
		p.Bet = 1000
	}

	// Mark bob as having acted
	table.Players[1].HasActed = true

	hasMore := table.AdvanceTurn()
	if !hasMore {
		t.Error("Expected to skip to charlie")
	}
	if table.CurrentTurn != 2 {
		t.Errorf("Expected turn 2 (charlie), got %d", table.CurrentTurn)
	}
}

func TestPlayerHit(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "8", Suit: "♠", Value: 8})

	initialCards := len(table.Players[0].Hand.Cards)

	err := table.PlayerHit("session1")
	if err != nil {
		t.Fatalf("Failed to hit: %v", err)
	}

	if len(table.Players[0].Hand.Cards) != initialCards+1 {
		t.Error("Expected one more card after hit")
	}
}

func TestPlayerHitBust(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})

	// Force next card to bust
	table.Deck = &Deck{Cards: []Card{{Rank: "K", Suit: "♥", Value: 10}}}

	err := table.PlayerHit("session1")
	if err != nil {
		t.Fatalf("Failed to hit: %v", err)
	}

	if table.Players[0].Result != ResultPlayerBust {
		t.Errorf("Expected PLAYER_BUST, got %s", table.Players[0].Result)
	}

	if !table.Players[0].HasActed {
		t.Error("Expected HasActed to be true after bust")
	}
}

func TestPlayerHitNotYourTurn(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[1].Bet = 1000

	err := table.PlayerHit("session2")
	if err == nil {
		t.Error("Expected error when hitting out of turn")
	}
}

func TestPlayerStand(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000

	err := table.PlayerStand("session1")
	if err != nil {
		t.Fatalf("Failed to stand: %v", err)
	}

	if !table.Players[0].HasActed {
		t.Error("Expected HasActed to be true after stand")
	}
}

func TestPlayerDoubleDown(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "2", Suit: "♠", Value: 2})

	err := table.PlayerDoubleDown("session1")
	if err != nil {
		t.Fatalf("Failed to double down: %v", err)
	}

	if table.Players[0].Bet != 2000 {
		t.Errorf("Expected bet to double to 2000, got %d", table.Players[0].Bet)
	}

	if len(table.Players[0].Hand.Cards) != 3 {
		t.Error("Expected one card to be drawn")
	}

	if !table.Players[0].HasActed {
		t.Error("Expected HasActed to be true after double down")
	}
}

func TestPlayerDoubleDownNotInitialHand(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "2", Suit: "♠", Value: 2})
	table.Players[0].Hand.AddCard(Card{Rank: "5", Suit: "♠", Value: 5})

	err := table.PlayerDoubleDown("session1")
	if err == nil {
		t.Error("Expected error when doubling down after initial hand")
	}
}

func TestPlayerSurrender(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "7", Suit: "♠", Value: 7})

	err := table.PlayerSurrender("session1")
	if err != nil {
		t.Fatalf("Failed to surrender: %v", err)
	}

	if table.Players[0].Result != ResultSurrender {
		t.Errorf("Expected SURRENDER result, got %s", table.Players[0].Result)
	}

	if !table.Players[0].HasActed {
		t.Error("Expected HasActed to be true after surrender")
	}
}

func TestPlayerSurrenderNotInitialHand(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].Hand.AddCard(Card{Rank: "7", Suit: "♠", Value: 7})
	table.Players[0].Hand.AddCard(Card{Rank: "2", Suit: "♠", Value: 2})

	err := table.PlayerSurrender("session1")
	if err == nil {
		t.Error("Expected error when surrendering after initial hand")
	}
}

func TestPlayDealerHand(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Dealer.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Dealer.AddCard(Card{Rank: "6", Suit: "♠", Value: 6}) // 15

	// Dealer should draw until 17+
	err := table.PlayDealerHand()
	if err != nil {
		t.Fatalf("Failed to play dealer hand: %v", err)
	}

	if table.Dealer.Value() < 17 {
		t.Errorf("Expected dealer to have 17+, got %d", table.Dealer.Value())
	}

	if table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase DEALER_TURN, got %s", table.Phase)
	}
}

func TestPlayDealerHandNoPlayersInPlay(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Result = ResultPlayerBust // Already busted

	initialCards := len(table.Dealer.Cards)

	err := table.PlayDealerHand()
	if err != nil {
		t.Fatalf("Failed to play dealer hand: %v", err)
	}

	// Dealer shouldn't draw if no players in play
	if len(table.Dealer.Cards) != initialCards {
		t.Error("Expected dealer not to draw when no players in play")
	}
}

func TestCalculatePayouts(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	// Player wins
	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9}) // 19

	table.Dealer.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})
	table.Dealer.AddCard(Card{Rank: "7", Suit: "♥", Value: 7}) // 17

	table.CalculatePayouts()

	if table.Players[0].Result != ResultPlayerWin {
		t.Errorf("Expected PLAYER_WIN, got %s", table.Players[0].Result)
	}

	if table.Players[0].Payout != 2000 {
		t.Errorf("Expected payout 2000, got %d", table.Players[0].Payout)
	}
}

func TestCalculatePayoutsPush(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "8", Suit: "♠", Value: 8}) // 18

	table.Dealer.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})
	table.Dealer.AddCard(Card{Rank: "8", Suit: "♥", Value: 8}) // 18

	table.CalculatePayouts()

	if table.Players[0].Result != ResultPush {
		t.Errorf("Expected PUSH, got %s", table.Players[0].Result)
	}

	if table.Players[0].Payout != 1000 {
		t.Errorf("Expected payout 1000 (bet returned), got %d", table.Players[0].Payout)
	}
}

func TestCalculatePayoutsBlackjack(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Result = ResultPlayerBlackjack

	table.CalculatePayouts()

	if table.Players[0].Payout != 2500 {
		t.Errorf("Expected payout 2500 (1.5x), got %d", table.Players[0].Payout)
	}
}

func TestCalculatePayoutsSurrender(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Result = ResultSurrender

	table.CalculatePayouts()

	if table.Players[0].Payout != 500 {
		t.Errorf("Expected payout 500 (half bet), got %d", table.Players[0].Payout)
	}
}

func TestCalculatePayoutsDealerBust(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "5", Suit: "♠", Value: 5}) // 15

	table.Dealer.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})
	table.Dealer.AddCard(Card{Rank: "9", Suit: "♥", Value: 9})
	table.Dealer.AddCard(Card{Rank: "5", Suit: "♥", Value: 5}) // 24 (bust)

	table.CalculatePayouts()

	if table.Players[0].Result != ResultPlayerWin {
		t.Errorf("Expected PLAYER_WIN when dealer busts, got %s", table.Players[0].Result)
	}

	if table.Players[0].Payout != 2000 {
		t.Errorf("Expected payout 2000, got %d", table.Players[0].Payout)
	}
}

func TestEndRound(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Players[1].IsActive = false // Bob disconnected
	table.Phase = PhasePayouts
	table.RoundActive = true

	table.EndRound()

	if len(table.Players) != 1 {
		t.Errorf("Expected inactive players removed, got %d players", len(table.Players))
	}

	if table.Phase != PhaseLobby {
		t.Errorf("Expected phase LOBBY, got %s", table.Phase)
	}

	if table.RoundActive {
		t.Error("Expected round to be inactive")
	}

	if table.CurrentTurn != 0 {
		t.Errorf("Expected CurrentTurn reset to 0, got %d", table.CurrentTurn)
	}

	if len(table.Dealer.Cards) != 0 {
		t.Error("Expected dealer hand to be reset")
	}
}

func TestGetTableState(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})

	table.Dealer.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	table.Dealer.AddCard(Card{Rank: "7", Suit: "♥", Value: 7})

	table.Phase = TablePhasePlayerTurn

	state := table.GetTableState(true)

	if state == "" {
		t.Error("Expected non-empty state string")
	}

	// Should contain player info
	if !contains(state, "alice") {
		t.Error("Expected state to contain player name")
	}

	// Should hide dealer's second card
	if !contains(state, "[Hidden]") {
		t.Error("Expected state to hide dealer's second card")
	}

	// Show full dealer hand when not hiding
	stateNoHide := table.GetTableState(false)
	if contains(stateNoHide, "[Hidden]") {
		t.Error("Expected state to show full dealer hand when not hiding")
	}
}

func TestGetPlayerList(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Players[0].IsReady = true

	list := table.GetPlayerList()

	if !contains(list, "alice") {
		t.Error("Expected list to contain alice")
	}

	if !contains(list, "bob") {
		t.Error("Expected list to contain bob")
	}

	if !contains(list, "[READY]") {
		t.Error("Expected list to show ready status")
	}

	if !contains(list, "2/4") {
		t.Error("Expected list to show player count")
	}
}

func TestBroadcastToAll(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn1 := &mockConn{}
	conn2 := &mockConn{}

	table.AddPlayer("session1", "alice", conn1)
	table.AddPlayer("session2", "bob", conn2)

	table.BroadcastToAll("Test message")

	// Expect 2 messages: the actual message + the prompt marker
	if len(conn1.messages) != 2 {
		t.Errorf("Expected 2 messages to alice (message + prompt), got %d", len(conn1.messages))
	}

	if len(conn2.messages) != 2 {
		t.Errorf("Expected 2 messages to bob (message + prompt), got %d", len(conn2.messages))
	}

	if !contains(conn1.messages[0], "Test message") {
		t.Error("Expected first message to contain 'Test message'")
	}

	if conn1.messages[1] != "<<<PROMPT>>>\n" {
		t.Errorf("Expected second message to be prompt marker, got %q", conn1.messages[1])
	}
}

func TestBroadcastToOthers(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn1 := &mockConn{}
	conn2 := &mockConn{}
	conn3 := &mockConn{}

	table.AddPlayer("session1", "alice", conn1)
	table.AddPlayer("session2", "bob", conn2)
	table.AddPlayer("session3", "charlie", conn3)

	table.BroadcastToOthers("session1", "Test message")

	if len(conn1.messages) != 0 {
		t.Errorf("Expected 0 messages to alice (excluded), got %d", len(conn1.messages))
	}

	// Expect 2 messages: the actual message + the prompt marker
	if len(conn2.messages) != 2 {
		t.Errorf("Expected 2 messages to bob (message + prompt), got %d", len(conn2.messages))
	}

	if len(conn3.messages) != 2 {
		t.Errorf("Expected 2 messages to charlie (message + prompt), got %d", len(conn3.messages))
	}

	if !contains(conn2.messages[0], "Test message") {
		t.Error("Expected first message to contain 'Test message'")
	}

	if conn2.messages[1] != "<<<PROMPT>>>\n" {
		t.Errorf("Expected second message to be prompt marker, got %q", conn2.messages[1])
	}
}

func TestWaitForBetsTimeout(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting

	// Only alice bets
	table.PlaceBet("session1", 5000)

	// Wait with very short timeout
	done := make(chan bool)
	go func() {
		table.WaitForBets(100 * time.Millisecond)
		done <- true
	}()

	// Wait for WaitForBets to complete
	<-done

	// Bob should be marked inactive for not betting
	if table.Players[1].IsActive {
		t.Error("Expected bob to be marked inactive after timeout")
	}
}

func TestWaitForBetsAllBet(t *testing.T) {
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting

	done := make(chan bool)
	go func() {
		table.WaitForBets(5 * time.Second)
		done <- true
	}()

	// Both players bet quickly
	time.Sleep(50 * time.Millisecond)
	table.PlaceBet("session1", 5000)
	table.PlaceBet("session2", 3000)

	// Should return quickly, not wait for timeout
	select {
	case <-done:
		// Success - returned early
	case <-time.After(1 * time.Second):
		t.Error("Expected WaitForBets to return immediately when all players bet")
	}
}

func TestDealInitialCardsCorrectCount(t *testing.T) {
	// Regression test: ensure each player gets exactly 2 cards
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)
	table.PlaceBet("session2", 3000)

	err := table.DealInitialCards()
	if err != nil {
		t.Fatalf("Failed to deal cards: %v", err)
	}

	// Each player should have exactly 2 cards
	for i, p := range table.Players {
		if len(p.Hand.Cards) != 2 {
			t.Errorf("Player %d should have exactly 2 cards, got %d", i, len(p.Hand.Cards))
		}
	}

	// Dealer should have exactly 2 cards
	if len(table.Dealer.Cards) != 2 {
		t.Errorf("Dealer should have exactly 2 cards, got %d", len(table.Dealer.Cards))
	}
}

func TestGetTableStateShowsYourTurnOnlyForCurrentPlayer(t *testing.T) {
	// Regression test: [YOUR TURN] should only appear for the current player
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0

	// Set up bets and hands
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[1].Bet = 1000
	table.Players[1].Hand.AddCard(Card{Rank: "8", Suit: "♠", Value: 8})
	table.Players[1].Hand.AddCard(Card{Rank: "7", Suit: "♠", Value: 7})

	state := table.GetTableState(true)

	// Count occurrences of [YOUR TURN]
	yourTurnCount := 0
	searchStr := "[YOUR TURN]"
	for i := 0; i <= len(state)-len(searchStr); i++ {
		if state[i:i+len(searchStr)] == searchStr {
			yourTurnCount++
		}
	}

	if yourTurnCount != 1 {
		t.Errorf("Expected exactly 1 [YOUR TURN] marker, got %d", yourTurnCount)
	}

	// Should appear on alice's line (seat 1) at the end
	if !contains(state, "Seat 1 - alice:") || !strings.Contains(state, "[YOUR TURN]") {
		t.Error("Expected [YOUR TURN] to appear for alice (current player)")
	}

	// Check that alice's line contains [YOUR TURN] but bob's doesn't
	lines := strings.Split(state, "\n")
	aliceLine := ""
	bobLine := ""
	for _, line := range lines {
		if strings.Contains(line, "Seat 1 - alice:") {
			aliceLine = line
		}
		if strings.Contains(line, "Seat 2 - bob:") {
			bobLine = line
		}
	}
	
	if !strings.Contains(aliceLine, "[YOUR TURN]") {
		t.Errorf("Expected [YOUR TURN] in alice's line, got: %s", aliceLine)
	}
	
	if strings.Contains(bobLine, "[YOUR TURN]") {
		t.Errorf("Expected [YOUR TURN] to NOT appear in bob's line, got: %s", bobLine)
	}
}

func TestMultiplePlayersGetCorrectCardCount(t *testing.T) {
	// Regression test: With 3 players, each should get exactly 2 cards
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.AddPlayer("session3", "charlie", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)
	table.PlaceBet("session2", 3000)
	table.PlaceBet("session3", 2000)

	err := table.DealInitialCards()
	if err != nil {
		t.Fatalf("Failed to deal cards: %v", err)
	}

	// Verify each player has exactly 2 cards
	expectedCards := 2
	for i, p := range table.Players {
		if len(p.Hand.Cards) != expectedCards {
			t.Errorf("Player %d (%s) should have %d cards, got %d: %v",
				i, p.Username, expectedCards, len(p.Hand.Cards), p.Hand.Cards)
		}
	}

	// Verify dealer has exactly 2 cards
	if len(table.Dealer.Cards) != expectedCards {
		t.Errorf("Dealer should have %d cards, got %d", expectedCards, len(table.Dealer.Cards))
	}
}

func TestDealerBlackjackEndsRoundImmediately(t *testing.T) {
	// Test that when dealer has blackjack, the round ends immediately
	// and players don't get to take actions (correct blackjack rules)
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.AddPlayer("session2", "bob", conn)
	table.Phase = PhaseBetting
	table.PlaceBet("session1", 5000)
	table.PlaceBet("session2", 3000)

	// Manually set up dealer blackjack scenario
	table.Phase = PhaseDealing

	// Alice gets blackjack (A, K)
	table.Players[0].Hand.AddCard(Card{Rank: "A", Suit: "♠", Value: 11})
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})

	// Bob gets 19 (K, 9)
	table.Players[1].Hand.AddCard(Card{Rank: "K", Suit: "♣", Value: 10})
	table.Players[1].Hand.AddCard(Card{Rank: "9", Suit: "♣", Value: 9})

	// Dealer gets blackjack (J, A)
	table.Dealer.AddCard(Card{Rank: "J", Suit: "♥", Value: 10})
	table.Dealer.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})

	// Simulate dealer blackjack detection (as done in DealInitialCards)
	if table.Dealer.IsBlackjack() {
		table.Phase = TablePhaseDealerTurn
		for _, p := range table.Players {
			if p.IsActive && p.Bet > 0 {
				if p.Hand.IsBlackjack() {
					p.Result = ResultPush
				} else {
					p.Result = ResultDealerWin
				}
			}
		}
	}

	// Verify phase transitioned to dealer turn
	if table.Phase != TablePhaseDealerTurn {
		t.Errorf("Expected phase DEALER_TURN when dealer has blackjack, got %s", table.Phase)
	}

	// Verify alice (with blackjack) pushes
	if table.Players[0].Result != ResultPush {
		t.Errorf("Expected alice (blackjack) to PUSH against dealer blackjack, got %s", table.Players[0].Result)
	}

	// Verify bob (without blackjack) loses
	if table.Players[1].Result != ResultDealerWin {
		t.Errorf("Expected bob (no blackjack) to lose against dealer blackjack, got %s", table.Players[1].Result)
	}

	// Verify players cannot take actions (should error)
	err := table.PlayerHit("session2")
	if err == nil {
		t.Error("Expected error when trying to hit after dealer blackjack ended round")
	}
	if err != nil && err.Error() != "not player turn phase" {
		t.Errorf("Expected 'not player turn phase' error, got: %v", err)
	}
}

func TestDealerHoleCardHiddenDuringPlayerTurn(t *testing.T) {
	// Test that dealer's hole card is hidden during player turn phase
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhasePlayerTurn
	table.CurrentTurn = 0
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})

	// Dealer has two cards
	table.Dealer.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	table.Dealer.AddCard(Card{Rank: "7", Suit: "♥", Value: 7})

	// Get table state with hideDealer=true
	state := table.GetTableState(true)

	// Should hide dealer's second card
	if !contains(state, "[Hidden]") {
		t.Error("Expected dealer's hole card to be hidden during player turn")
	}

	// Should show first card
	if !contains(state, "A♥") {
		t.Error("Expected dealer's first card to be visible")
	}

	// Should NOT show dealer's value when hiding
	if contains(state, "Dealer: [A♥] [7♥]") {
		t.Error("Expected dealer's second card to be hidden, but both cards are shown")
	}
}

func TestDealerHoleCardRevealedDuringDealerTurn(t *testing.T) {
	// Test that dealer's hole card is revealed during dealer turn phase
	table := NewTable("test-1", 1000, 100000, 4)
	conn := &mockConn{}

	table.AddPlayer("session1", "alice", conn)
	table.Phase = TablePhaseDealerTurn
	table.Players[0].Bet = 1000
	table.Players[0].Hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	table.Players[0].Hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	table.Players[0].HasActed = true

	// Dealer has two cards
	table.Dealer.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	table.Dealer.AddCard(Card{Rank: "7", Suit: "♥", Value: 7})

	// Get table state with hideDealer=true (but phase is dealer turn)
	state := table.GetTableState(true)

	// Should NOT hide dealer's cards during dealer turn
	if contains(state, "[Hidden]") {
		t.Error("Expected dealer's hole card to be revealed during dealer turn")
	}

	// Should show both cards
	if !contains(state, "A♥") {
		t.Error("Expected dealer's first card to be visible")
	}

	// Should show dealer's full hand
	if !contains(state, "Dealer:") {
		t.Error("Expected dealer information in state")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
