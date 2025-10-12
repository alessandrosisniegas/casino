package game

import (
	"strings"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	if len(deck.Cards) != 52 {
		t.Errorf("expected 52 cards, got %d", len(deck.Cards))
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, card := range deck.Cards {
		key := card.Rank + card.Suit
		if seen[key] {
			t.Errorf("duplicate card found: %s%s", card.Rank, card.Suit)
		}
		seen[key] = true
	}
}

func TestDeckShuffle(t *testing.T) {
	deck1 := NewDeck()
	originalOrder := make([]Card, len(deck1.Cards))
	copy(originalOrder, deck1.Cards)

	// Shuffle once and verify it's a permutation
	deck1.Shuffle()

	// Verify same number of cards
	if len(deck1.Cards) != 52 {
		t.Errorf("shuffle changed deck size to %d", len(deck1.Cards))
	}

	// Verify it's a permutation (same multiset)
	originalSet := make(map[string]int)
	shuffledSet := make(map[string]int)

	for _, card := range originalOrder {
		key := card.Rank + card.Suit
		originalSet[key]++
	}

	for _, card := range deck1.Cards {
		key := card.Rank + card.Suit
		shuffledSet[key]++
	}

	// Check same cards exist
	for key, count := range originalSet {
		if shuffledSet[key] != count {
			t.Errorf("shuffle changed card composition: %s appears %d times, expected %d", key, shuffledSet[key], count)
		}
	}

	// Note: We don't test that order changed because with a fixed seed,
	// shuffle is deterministic. In practice, the randomness is sufficient.
}

func TestDeckDraw(t *testing.T) {
	deck := NewDeck()
	initialCount := len(deck.Cards)

	card, err := deck.Draw()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if card.Rank == "" {
		t.Error("drew an empty card")
	}

	if len(deck.Cards) != initialCount-1 {
		t.Errorf("expected %d cards after draw, got %d", initialCount-1, len(deck.Cards))
	}
}

func TestDeckDrawEmpty(t *testing.T) {
	deck := &Deck{Cards: []Card{}}

	_, err := deck.Draw()
	if err == nil {
		t.Error("expected error when drawing from empty deck")
	}
}

func TestPlaceBetWithInsufficientCards(t *testing.T) {
	// Deck with only 3 cards (need 4 for initial deal)
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10},
		{Rank: "9", Suit: "♠", Value: 9},
		{Rank: "8", Suit: "♠", Value: 8},
	}

	game := NewGameWithDeck(deck)
	err := game.PlaceBetNoShuffle(1000)

	if err == nil {
		t.Error("expected error when dealing with insufficient cards")
	}

	if !strings.Contains(err.Error(), "failed to deal") {
		t.Errorf("expected 'failed to deal' error, got: %v", err)
	}
}

func TestHandValue(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		expected int
	}{
		{
			name:     "simple hand",
			cards:    []Card{{Rank: "5", Value: 5}, {Rank: "7", Value: 7}},
			expected: 12,
		},
		{
			name:     "face cards",
			cards:    []Card{{Rank: "K", Value: 10}, {Rank: "Q", Value: 10}},
			expected: 20,
		},
		{
			name:     "ace as 11",
			cards:    []Card{{Rank: "A", Value: 11}, {Rank: "9", Value: 9}},
			expected: 20,
		},
		{
			name:     "ace as 1",
			cards:    []Card{{Rank: "A", Value: 11}, {Rank: "K", Value: 10}, {Rank: "5", Value: 5}},
			expected: 16, // Ace counts as 1
		},
		{
			name:     "multiple aces",
			cards:    []Card{{Rank: "A", Value: 11}, {Rank: "A", Value: 11}, {Rank: "9", Value: 9}},
			expected: 21, // One ace as 11, one as 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := NewHand()
			for _, card := range tt.cards {
				hand.AddCard(card)
			}

			if hand.Value() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, hand.Value())
			}
		})
	}
}

func TestHandIsBusted(t *testing.T) {
	hand := NewHand()
	hand.AddCard(Card{Rank: "K", Value: 10})
	hand.AddCard(Card{Rank: "Q", Value: 10})
	hand.AddCard(Card{Rank: "5", Value: 5})

	if !hand.IsBusted() {
		t.Error("expected hand to be busted")
	}
}

func TestHandIsBlackjack(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		expected bool
	}{
		{
			name:     "blackjack",
			cards:    []Card{{Rank: "A", Value: 11}, {Rank: "K", Value: 10}},
			expected: true,
		},
		{
			name:     "21 but not blackjack",
			cards:    []Card{{Rank: "7", Value: 7}, {Rank: "7", Value: 7}, {Rank: "7", Value: 7}},
			expected: false,
		},
		{
			name:     "two cards but not 21",
			cards:    []Card{{Rank: "K", Value: 10}, {Rank: "9", Value: 9}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := NewHand()
			for _, card := range tt.cards {
				hand.AddCard(card)
			}

			if hand.IsBlackjack() != tt.expected {
				t.Errorf("expected IsBlackjack to be %v", tt.expected)
			}
		})
	}
}

func TestNewGame(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("NewGame returned nil")
	}

	if game.Phase != PhaseWaitingForBet {
		t.Errorf("expected phase %s, got %s", PhaseWaitingForBet, game.Phase)
	}

	if game.Bet != 0 {
		t.Errorf("expected bet 0, got %d", game.Bet)
	}

	if len(game.PlayerHand.Cards) != 0 {
		t.Error("expected empty player hand")
	}

	if len(game.DealerHand.Cards) != 0 {
		t.Error("expected empty dealer hand")
	}
}

func TestPlaceBet(t *testing.T) {
	// Use deterministic deck that won't produce blackjack
	deck := []Card{
		{Rank: "5", Suit: "♠", Value: 5}, // P1
		{Rank: "6", Suit: "♠", Value: 6}, // D1
		{Rank: "7", Suit: "♠", Value: 7}, // P2
		{Rank: "8", Suit: "♠", Value: 8}, // D2
	}

	game := NewGameWithDeck(deck)

	err := game.PlaceBetNoShuffle(1000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if game.Bet != 1000 {
		t.Errorf("expected bet 1000, got %d", game.Bet)
	}

	if len(game.PlayerHand.Cards) != 2 {
		t.Errorf("expected 2 cards in player hand, got %d", len(game.PlayerHand.Cards))
	}

	if len(game.DealerHand.Cards) != 2 {
		t.Errorf("expected 2 cards in dealer hand, got %d", len(game.DealerHand.Cards))
	}

	if game.Phase != PhasePlayerTurn {
		t.Errorf("expected player turn phase, got %s", game.Phase)
	}
}

func TestPlaceBetInvalidAmount(t *testing.T) {
	game := NewGame()

	err := game.PlaceBetNoShuffle(0)
	if err == nil {
		t.Error("expected error for zero bet")
	}

	err = game.PlaceBetNoShuffle(-100)
	if err == nil {
		t.Error("expected error for negative bet")
	}
}

func TestPlaceBetWrongPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhasePlayerTurn

	err := game.PlaceBetNoShuffle(1000)
	if err == nil {
		t.Error("expected error when placing bet in wrong phase")
	}
}

func TestHit(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "5", Suit: "♠", Value: 5}, // P1
		{Rank: "6", Suit: "♠", Value: 6}, // D1
		{Rank: "7", Suit: "♠", Value: 7}, // P2 (12 total)
		{Rank: "8", Suit: "♠", Value: 8}, // D2
		{Rank: "3", Suit: "♠", Value: 3}, // P hits (15 total)
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", game.Phase)
	}

	initialCards := len(game.PlayerHand.Cards)

	err := game.Hit()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(game.PlayerHand.Cards) != initialCards+1 {
		t.Errorf("expected %d cards after hit, got %d", initialCards+1, len(game.PlayerHand.Cards))
	}

	if game.PlayerHand.Value() != 15 {
		t.Errorf("expected hand value 15, got %d", game.PlayerHand.Value())
	}
}

func TestHitWrongPhase(t *testing.T) {
	game := NewGame()

	err := game.Hit()
	if err == nil {
		t.Error("expected error when hitting in wrong phase")
	}
}

func TestHitInGameOverPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseGameOver

	err := game.Hit()
	if err == nil {
		t.Error("expected error when hitting in game over phase")
	}
}

func TestHitInDealerTurnPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseDealerTurn

	err := game.Hit()
	if err == nil {
		t.Error("expected error when hitting in dealer turn phase")
	}
}

func TestStand(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "9", Suit: "♠", Value: 9},  // P1
		{Rank: "6", Suit: "♠", Value: 6},  // D1
		{Rank: "9", Suit: "♥", Value: 9},  // P2 (18 total)
		{Rank: "K", Suit: "♠", Value: 10}, // D2 (16 total)
		{Rank: "5", Suit: "♠", Value: 5},  // D hits to 21
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", game.Phase)
	}

	err := game.Stand()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if game.Phase != PhaseGameOver {
		t.Errorf("expected phase %s after stand, got %s", PhaseGameOver, game.Phase)
	}

	if !game.PlayerStood {
		t.Error("expected PlayerStood to be true")
	}
}

func TestDoubleDown(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "5", Suit: "♠", Value: 5},  // P1
		{Rank: "6", Suit: "♠", Value: 6},  // D1
		{Rank: "6", Suit: "♥", Value: 6},  // P2 (11 total)
		{Rank: "K", Suit: "♠", Value: 10}, // D2 (16 total)
		{Rank: "9", Suit: "♠", Value: 9},  // P doubles (20 total)
		{Rank: "5", Suit: "♥", Value: 5},  // D hits to 21
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", game.Phase)
	}

	err := game.DoubleDown()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if game.Bet != 2000 {
		t.Errorf("expected bet 2000 after double down, got %d", game.Bet)
	}

	if !game.IsDoubled {
		t.Error("expected IsDoubled to be true")
	}

	if len(game.PlayerHand.Cards) != 3 {
		t.Errorf("expected 3 cards after double down, got %d", len(game.PlayerHand.Cards))
	}

	if game.Phase != PhaseGameOver {
		t.Errorf("expected phase %s after double down, got %s", PhaseGameOver, game.Phase)
	}
}

func TestDoubleDownWrongPhase(t *testing.T) {
	game := NewGame()

	err := game.DoubleDown()
	if err == nil {
		t.Error("expected error when doubling down in wrong phase")
	}
}

func TestDoubleDownInGameOverPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseGameOver
	game.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "9", Value: 9})

	err := game.DoubleDown()
	if err == nil {
		t.Error("expected error when doubling down in game over phase")
	}
}

func TestDoubleDownInDealerTurnPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseDealerTurn
	game.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "9", Value: 9})

	err := game.DoubleDown()
	if err == nil {
		t.Error("expected error when doubling down in dealer turn phase")
	}
}

func TestStandInGameOverPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseGameOver

	err := game.Stand()
	if err == nil {
		t.Error("expected error when standing in game over phase")
	}
}

func TestStandInDealerTurnPhase(t *testing.T) {
	game := NewGame()
	game.Phase = PhaseDealerTurn

	err := game.Stand()
	if err == nil {
		t.Error("expected error when standing in dealer turn phase")
	}
}

func TestDoubleDownAfterHit(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "5", Suit: "♠", Value: 5},  // P1
		{Rank: "6", Suit: "♠", Value: 6},  // D1
		{Rank: "6", Suit: "♥", Value: 6},  // P2 (11 total)
		{Rank: "K", Suit: "♠", Value: 10}, // D2
		{Rank: "2", Suit: "♠", Value: 2},  // P hits (13 total)
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", game.Phase)
	}

	game.Hit()

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("should still be player turn after hit, got %s", game.Phase)
	}

	err := game.DoubleDown()
	if err == nil {
		t.Error("expected error when doubling down after hit")
	}
}

func TestCalculatePayout(t *testing.T) {
	tests := []struct {
		name     string
		bet      int64
		result   GameResult
		expected int64
	}{
		{
			name:     "player blackjack",
			bet:      1000,
			result:   ResultPlayerBlackjack,
			expected: 2500, // 1000 + 1500 (3:2)
		},
		{
			name:     "player win",
			bet:      1000,
			result:   ResultPlayerWin,
			expected: 2000, // 1:1
		},
		{
			name:     "push",
			bet:      1000,
			result:   ResultPush,
			expected: 1000,
		},
		{
			name:     "dealer win",
			bet:      1000,
			result:   ResultDealerWin,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			game.Bet = tt.bet
			game.Result = tt.result

			payout := game.CalculatePayout()
			if payout != tt.expected {
				t.Errorf("expected payout %d, got %d", tt.expected, payout)
			}
		})
	}
}

func TestDoubleDownWinPayoutUsesDoubledBet(t *testing.T) {
	// Verify that payout calculation uses the current (doubled) Bet value
	// This guards against future refactors that might use IsDoubled flag instead
	g := NewGame()
	g.Bet = 2000 // Bet was already doubled by DoubleDown()
	g.IsDoubled = true
	g.Result = ResultPlayerWin

	payout := g.CalculatePayout()
	// Should be 2000 * 2 = 4000 (doubled bet pays 1:1)
	if payout != 4000 {
		t.Errorf("payout should be based on current Bet (2000*2), got %d", payout)
	}
}

func TestDealerPlaysCorrectly(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10}, // P1
		{Rank: "6", Suit: "♠", Value: 6},  // D1
		{Rank: "9", Suit: "♠", Value: 9},  // P2 (19 total)
		{Rank: "8", Suit: "♠", Value: 8},  // D2 (14 total)
		{Rank: "5", Suit: "♠", Value: 5},  // D hits (19 total)
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	if game.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", game.Phase)
	}

	game.Stand()

	// Dealer should have hit until 17 or higher
	dealerValue := game.DealerHand.Value()
	if !game.DealerHand.IsBusted() && dealerValue < 17 {
		t.Errorf("dealer stopped at %d, should continue until 17+", dealerValue)
	}

	if dealerValue != 19 {
		t.Errorf("expected dealer to have 19, got %d", dealerValue)
	}
}

func TestGameFlowComplete(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10}, // P1
		{Rank: "7", Suit: "♠", Value: 7},  // D1
		{Rank: "7", Suit: "♥", Value: 7},  // P2 (17 total)
		{Rank: "K", Suit: "♥", Value: 10}, // D2 (17 total)
	}

	game := NewGameWithDeck(deck)

	// Start with bet
	if game.Phase != PhaseWaitingForBet {
		t.Error("game should start waiting for bet")
	}

	game.PlaceBetNoShuffle(1000)

	// Should be player turn
	if game.Phase != PhasePlayerTurn {
		t.Errorf("expected player turn after bet, got %s", game.Phase)
	}

	game.Stand()

	// Should be game over
	if game.Phase != PhaseGameOver {
		t.Errorf("expected game over, got %s", game.Phase)
	}

	// Should have a result
	if game.Result == "" {
		t.Error("game should have a result")
	}

	// Should be a push (both 17)
	if game.Result != ResultPush {
		t.Errorf("expected push, got %s", game.Result)
	}

	// Payout should return bet
	payout := game.CalculatePayout()
	if payout != 1000 {
		t.Errorf("expected payout 1000, got %d", payout)
	}
}

func TestHandString(t *testing.T) {
	hand := NewHand()
	hand.AddCard(Card{Rank: "A", Suit: "♠", Value: 11})
	hand.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})

	str := hand.String()
	if str != "A♠ K♥" {
		t.Errorf("expected 'A♠ K♥', got '%s'", str)
	}
}

func TestGetGameState(t *testing.T) {
	// Use deterministic deck
	deck := []Card{
		{Rank: "7", Suit: "♠", Value: 7}, // P1
		{Rank: "6", Suit: "♠", Value: 6}, // D1
		{Rank: "8", Suit: "♠", Value: 8}, // P2
		{Rank: "9", Suit: "♠", Value: 9}, // D2
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)

	state := game.GetGameState(true)
	if state == "" {
		t.Error("game state should not be empty")
	}

	// Should contain bet information
	if !strings.Contains(state, "Bet:") {
		t.Error("game state should contain bet information")
	}

	// Should contain player hand
	if !strings.Contains(state, "Player Hand:") {
		t.Error("game state should contain player hand")
	}

	// Should contain dealer hand
	if !strings.Contains(state, "Dealer Hand:") {
		t.Error("game state should contain dealer hand")
	}

	// Should hide dealer's second card during player turn
	if !strings.Contains(state, "[Hidden]") {
		t.Error("should hide dealer's second card during player turn")
	}
}

func TestGetGameStateAtGameOver(t *testing.T) {
	// Force a deterministic finish
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10}, // P1
		{Rank: "7", Suit: "♠", Value: 7},  // D1
		{Rank: "9", Suit: "♠", Value: 9},  // P2 (19 total)
		{Rank: "K", Suit: "♥", Value: 10}, // D2 (17 total)
	}

	game := NewGameWithDeck(deck)
	game.PlaceBetNoShuffle(1000)
	game.Stand()

	// Game should be over
	if game.Phase != PhaseGameOver {
		t.Fatalf("expected game over, got %s", game.Phase)
	}

	state := game.GetGameState(false)

	// Should contain result
	if !strings.Contains(state, "Result:") {
		t.Error("game state at game over should contain Result")
	}

	// Should contain payout
	if !strings.Contains(state, "Payout:") {
		t.Error("game state at game over should contain Payout")
	}

	// Should show full dealer hand (not hidden)
	if strings.Contains(state, "[Hidden]") {
		t.Error("game state at game over should not hide dealer cards")
	}
}

// Edge case tests

// Critical bug tests

func TestDeckExhaustionHandled(t *testing.T) {
	g := NewGame()
	g.Bet = 1000
	g.Phase = PhaseDealerTurn

	// Dealer has 16 and needs to hit
	g.DealerHand.AddCard(Card{Rank: "9", Value: 9})
	g.DealerHand.AddCard(Card{Rank: "7", Value: 7})

	// Player has 18
	g.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	g.PlayerHand.AddCard(Card{Rank: "8", Value: 8})

	// Empty deck - should not infinite loop or panic
	g.Deck = &Deck{Cards: nil}

	g.playDealerTurn()

	if g.Phase != PhaseGameOver {
		t.Fatal("dealer turn should end even if deck is empty")
	}

	if g.Result != ResultPush {
		t.Errorf("expected push on deck exhaustion, got %s", g.Result)
	}
}

func TestGetGameStateBeforeDealDoesNotPanic(t *testing.T) {
	g := NewGame()

	// Should not panic with empty hands
	state := g.GetGameState(true)

	if state == "" {
		t.Error("state should not be empty")
	}

	if !strings.Contains(state, "no cards dealt") {
		t.Error("state should indicate no cards dealt")
	}
}

func TestNewDeckCompositionAndValues(t *testing.T) {
	d := NewDeck()

	want := map[string]int{
		"A": 11, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6,
		"7": 7, "8": 8, "9": 9, "10": 10, "J": 10, "Q": 10, "K": 10,
	}
	suits := map[string]bool{"♠": true, "♥": true, "♦": true, "♣": true}
	seen := map[string]bool{}

	for _, c := range d.Cards {
		if !suits[c.Suit] {
			t.Fatalf("unknown suit %q", c.Suit)
		}
		if want[c.Rank] != c.Value {
			t.Fatalf("rank %q has value %d, want %d", c.Rank, c.Value, want[c.Rank])
		}
		seen[c.Rank+c.Suit] = true
	}

	if len(seen) != 52 {
		t.Fatalf("expected 52 unique cards, got %d", len(seen))
	}
}

// Deterministic scenario tests

func TestOpeningBlackjackPayoutDeterministic(t *testing.T) {
	// Deal order: P, D, P, D
	deck := []Card{
		{Rank: "A", Suit: "♠", Value: 11}, // P1
		{Rank: "9", Suit: "♠", Value: 9},  // D1
		{Rank: "K", Suit: "♠", Value: 10}, // P2 -> BJ
		{Rank: "7", Suit: "♠", Value: 7},  // D2
	}

	g := NewGameWithDeck(deck)
	if err := g.PlaceBetNoShuffle(1000); err != nil {
		t.Fatal(err)
	}

	if g.Phase != PhaseGameOver {
		t.Errorf("expected game over after blackjack, got phase %s", g.Phase)
	}

	if g.Result != ResultPlayerBlackjack {
		t.Fatalf("want blackjack, got %v", g.Result)
	}

	payout := g.CalculatePayout()
	if payout != 2500 {
		t.Fatalf("want 2500 (1000 + 1500), got %d", payout)
	}
}

func TestDealerStandsOnSoft17(t *testing.T) {
	g := NewGame()
	g.Bet = 1000
	g.Phase = PhaseDealerTurn

	// Dealer has soft 17 (A + 6)
	g.DealerHand.AddCard(Card{Rank: "A", Value: 11})
	g.DealerHand.AddCard(Card{Rank: "6", Value: 6})

	// Player has 18
	g.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	g.PlayerHand.AddCard(Card{Rank: "8", Value: 8})

	// Mock deck with card that would hit if dealer hits soft 17
	g.Deck = &Deck{Cards: []Card{{Rank: "5", Value: 5}}}

	initialDealerCards := len(g.DealerHand.Cards)
	g.playDealerTurn()

	// Dealer should stand on soft 17 (current policy)
	if len(g.DealerHand.Cards) != initialDealerCards {
		t.Fatal("dealer should stand on soft 17")
	}

	if g.DealerHand.Value() != 17 {
		t.Errorf("dealer should have 17, got %d", g.DealerHand.Value())
	}
}

func TestPushAfterDoubleDown(t *testing.T) {
	// Deal order: P, D, P, D, P(double), D(hits to match)
	deck := []Card{
		{Rank: "5", Suit: "♠", Value: 5},  // P1
		{Rank: "K", Suit: "♠", Value: 10}, // D1
		{Rank: "6", Suit: "♠", Value: 6},  // P2 (11 total)
		{Rank: "7", Suit: "♠", Value: 7},  // D2 (17 total)
		{Rank: "6", Suit: "♥", Value: 6},  // P3 double (17 total)
		// Dealer stands at 17
	}

	g := NewGameWithDeck(deck)
	if err := g.PlaceBetNoShuffle(1000); err != nil {
		t.Fatal(err)
	}

	if g.Phase == PhaseGameOver {
		t.Fatal("should not be game over yet")
	}

	if err := g.DoubleDown(); err != nil {
		t.Fatal(err)
	}

	if g.Bet != 2000 {
		t.Errorf("bet should be doubled to 2000, got %d", g.Bet)
	}

	if g.Result != ResultPush {
		t.Errorf("expected push with both at 17, got %s", g.Result)
	}

	payout := g.CalculatePayout()
	if payout != 2000 {
		t.Errorf("push should return doubled stake of 2000, got %d", payout)
	}
}

func TestStandImmediatelyAfterDealNonBlackjack(t *testing.T) {
	// Deal order: P, D, P, D
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10}, // P1
		{Rank: "6", Suit: "♠", Value: 6},  // D1
		{Rank: "9", Suit: "♠", Value: 9},  // P2 (19 total)
		{Rank: "5", Suit: "♠", Value: 5},  // D2 (11 total)
		{Rank: "7", Suit: "♥", Value: 7},  // D hits to 18
	}

	g := NewGameWithDeck(deck)
	if err := g.PlaceBetNoShuffle(1000); err != nil {
		t.Fatal(err)
	}

	if g.Phase != PhasePlayerTurn {
		t.Fatalf("expected player turn, got %s", g.Phase)
	}

	if err := g.Stand(); err != nil {
		t.Fatal(err)
	}

	if g.Phase != PhaseGameOver {
		t.Errorf("expected game over after stand, got %s", g.Phase)
	}

	if g.Result != ResultPlayerWin {
		t.Errorf("player 19 should beat dealer 18, got %s", g.Result)
	}
}

func TestPlayer21NonBlackjackVsDealer21NonBlackjack(t *testing.T) {
	g := NewGame()
	g.Bet = 1000

	// Player has 21 with 3 cards
	g.PlayerHand.AddCard(Card{Rank: "7", Value: 7})
	g.PlayerHand.AddCard(Card{Rank: "7", Value: 7})
	g.PlayerHand.AddCard(Card{Rank: "7", Value: 7})

	// Dealer has 21 with 3 cards
	g.DealerHand.AddCard(Card{Rank: "6", Value: 6})
	g.DealerHand.AddCard(Card{Rank: "8", Value: 8})
	g.DealerHand.AddCard(Card{Rank: "7", Value: 7})

	g.Phase = PhaseGameOver
	g.determineWinner()

	if g.Result != ResultPush {
		t.Errorf("both 21 (non-BJ) should push, got %s", g.Result)
	}

	payout := g.CalculatePayout()
	if payout != 1000 {
		t.Errorf("push should return bet of 1000, got %d", payout)
	}
}

func TestMultipleAcesInDealerHand(t *testing.T) {
	// Dealer has multiple aces and needs to reduce them properly
	deck := []Card{
		{Rank: "K", Suit: "♠", Value: 10}, // P1
		{Rank: "A", Suit: "♠", Value: 11}, // D1
		{Rank: "8", Suit: "♠", Value: 8},  // P2 (18)
		{Rank: "A", Suit: "♥", Value: 11}, // D2 (12 or 2)
		{Rank: "A", Suit: "♦", Value: 11}, // D3 (13 or 3)
		{Rank: "K", Suit: "♥", Value: 10}, // D4 (23 or 13)
		{Rank: "4", Suit: "♠", Value: 4},  // D5 (17)
	}

	g := NewGameWithDeck(deck)
	if err := g.PlaceBetNoShuffle(1000); err != nil {
		t.Fatal(err)
	}

	if err := g.Stand(); err != nil {
		t.Fatal(err)
	}

	dealerValue := g.DealerHand.Value()
	if dealerValue < 17 || dealerValue > 21 {
		t.Errorf("dealer should end with 17-21, got %d", dealerValue)
	}

	if g.DealerHand.IsBusted() {
		t.Error("dealer should not bust with proper ace handling")
	}
}

func TestBothBlackjack(t *testing.T) {
	game := NewGame()
	game.Bet = 1000

	// Manually set up both blackjacks
	game.PlayerHand.AddCard(Card{Rank: "A", Suit: "♠", Value: 11})
	game.PlayerHand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	game.DealerHand.AddCard(Card{Rank: "Q", Suit: "♥", Value: 10})

	game.Phase = PhaseGameOver
	game.determineWinner()

	if game.Result != ResultPush {
		t.Errorf("expected push when both have blackjack, got %s", game.Result)
	}

	payout := game.CalculatePayout()
	if payout != 1000 {
		t.Errorf("expected bet returned (1000), got %d", payout)
	}
}

func TestBothBlackjackViaDeal(t *testing.T) {
	// Test both-BJ scenario via normal deal path
	deck := []Card{
		{Rank: "A", Suit: "♠", Value: 11}, // P1
		{Rank: "A", Suit: "♥", Value: 11}, // D1
		{Rank: "K", Suit: "♠", Value: 10}, // P2 -> BJ
		{Rank: "Q", Suit: "♥", Value: 10}, // D2 -> BJ
	}

	game := NewGameWithDeck(deck)
	err := game.PlaceBetNoShuffle(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should detect both blackjacks and end game
	if game.Phase != PhaseGameOver {
		t.Errorf("expected game over after both blackjacks, got %s", game.Phase)
	}

	if game.Result != ResultPush {
		t.Errorf("expected push when both have blackjack via deal, got %s", game.Result)
	}

	payout := game.CalculatePayout()
	if payout != 1000 {
		t.Errorf("expected bet returned (1000), got %d", payout)
	}
}

func TestDealerBlackjackPlayerDoesNot(t *testing.T) {
	game := NewGame()
	game.Bet = 1000

	// Player has 20, dealer has blackjack
	game.PlayerHand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "Q", Suit: "♠", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
	game.DealerHand.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})

	game.Phase = PhaseGameOver
	game.determineWinner()

	if game.Result != ResultDealerWin {
		t.Errorf("expected dealer win with blackjack, got %s", game.Result)
	}
}

func TestDealerBlackjackViaDeal(t *testing.T) {
	// Test dealer-only blackjack via normal deal path
	deck := []Card{
		{Rank: "9", Suit: "♠", Value: 9},  // P1
		{Rank: "A", Suit: "♥", Value: 11}, // D1
		{Rank: "9", Suit: "♥", Value: 9},  // P2 (18, not BJ)
		{Rank: "K", Suit: "♥", Value: 10}, // D2 -> dealer BJ
	}

	g := NewGameWithDeck(deck)
	err := g.PlaceBetNoShuffle(1000)
	if err != nil {
		t.Fatal(err)
	}

	// Should detect dealer blackjack and end game immediately
	if g.Phase != PhaseGameOver {
		t.Fatalf("expected game over after dealer blackjack, got %s", g.Phase)
	}

	if g.Result != ResultDealerWin {
		t.Fatalf("expected dealer win with blackjack, got %s", g.Result)
	}
}

func TestDealerBusts(t *testing.T) {
	game := NewGame()
	game.Bet = 1000

	// Player has 18, dealer busts with 22
	game.PlayerHand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "8", Suit: "♠", Value: 8})
	game.DealerHand.AddCard(Card{Rank: "K", Suit: "♥", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "Q", Suit: "♥", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "2", Suit: "♥", Value: 2})

	game.Phase = PhaseGameOver
	game.determineWinner()

	if game.Result != ResultPlayerWin {
		t.Errorf("expected player win when dealer busts, got %s", game.Result)
	}
}

func TestPushWithSameValue(t *testing.T) {
	game := NewGame()
	game.Bet = 1000

	// Both have 19
	game.PlayerHand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	game.DealerHand.AddCard(Card{Rank: "Q", Suit: "♥", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "9", Suit: "♥", Value: 9})

	game.Phase = PhaseGameOver
	game.determineWinner()

	if game.Result != ResultPush {
		t.Errorf("expected push with same values, got %s", game.Result)
	}

	payout := game.CalculatePayout()
	if payout != 1000 {
		t.Errorf("expected bet returned (1000), got %d", payout)
	}
}

func TestMultipleAcesEdgeCase(t *testing.T) {
	hand := NewHand()
	// Three aces
	hand.AddCard(Card{Rank: "A", Value: 11})
	hand.AddCard(Card{Rank: "A", Value: 11})
	hand.AddCard(Card{Rank: "A", Value: 11})

	// Should be 13 (one ace as 11, two as 1)
	if hand.Value() != 13 {
		t.Errorf("expected 13 with three aces, got %d", hand.Value())
	}

	// Add another ace
	hand.AddCard(Card{Rank: "A", Value: 11})
	// Should be 14 (all four aces as 1 except possibility of one as 11)
	if hand.Value() != 14 {
		t.Errorf("expected 14 with four aces, got %d", hand.Value())
	}
}

func TestSoftHandBecomesHard(t *testing.T) {
	hand := NewHand()
	// Soft 18 (A + 7)
	hand.AddCard(Card{Rank: "A", Value: 11})
	hand.AddCard(Card{Rank: "7", Value: 7})

	if hand.Value() != 18 {
		t.Errorf("expected soft 18, got %d", hand.Value())
	}

	// Hit a 10, becomes hard 18 (1 + 7 + 10)
	hand.AddCard(Card{Rank: "K", Value: 10})

	if hand.Value() != 18 {
		t.Errorf("expected hard 18 after ace adjustment, got %d", hand.Value())
	}

	if hand.IsBusted() {
		t.Error("hand should not be busted")
	}
}

func TestDoubleDownAndBust(t *testing.T) {
	game := NewGame()
	game.Bet = 1000
	game.Phase = PhasePlayerTurn

	// Give player a hand that will bust
	game.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "9", Value: 9})

	// Mock deck with a card that will bust
	game.Deck = &Deck{Cards: []Card{{Rank: "K", Value: 10}}}

	err := game.DoubleDown()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if game.Bet != 2000 {
		t.Errorf("expected doubled bet, got %d", game.Bet)
	}

	if !game.PlayerHand.IsBusted() {
		t.Error("player should be busted")
	}

	if game.Result != ResultDealerWin {
		t.Errorf("expected dealer win after bust, got %s", game.Result)
	}
}

func TestDoubleDownAndWin(t *testing.T) {
	game := NewGame()
	game.Bet = 1000
	game.Phase = PhasePlayerTurn

	// Player has 11
	game.PlayerHand.AddCard(Card{Rank: "6", Value: 6})
	game.PlayerHand.AddCard(Card{Rank: "5", Value: 5})

	// Mock deck with good card for player and bad cards for dealer
	game.Deck = &Deck{Cards: []Card{
		{Rank: "K", Value: 10}, // Player draws this (gets 21)
		{Rank: "K", Value: 10}, // Dealer's hidden card
		{Rank: "5", Value: 5},  // Dealer hits (busts at 25)
	}}
	game.DealerHand.AddCard(Card{Rank: "K", Value: 10})

	err := game.DoubleDown()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if game.PlayerHand.Value() != 21 {
		t.Errorf("expected player to have 21, got %d", game.PlayerHand.Value())
	}

	if game.Result != ResultPlayerWin {
		t.Errorf("expected player win, got %s", game.Result)
	}

	payout := game.CalculatePayout()
	if payout != 4000 {
		t.Errorf("expected payout of 4000 (doubled bet x2), got %d", payout)
	}
}

func TestBlackjackPayoutWithOddAmount(t *testing.T) {
	// Test integer division doesn't lose too much precision
	game := NewGame()
	game.Bet = 1500 // $15.00
	game.Result = ResultPlayerBlackjack

	// Should be 1500 + (1500 * 3 / 2) = 1500 + 2250 = 3750
	payout := game.CalculatePayout()
	expected := int64(3750)

	if payout != expected {
		t.Errorf("expected payout %d, got %d", expected, payout)
	}

	// Test with amount that doesn't divide evenly
	game.Bet = 1333 // $13.33
	// Should be 1333 + (1333 * 3 / 2) = 1333 + 1999 = 3332
	payout = game.CalculatePayout()
	expected = 1333 + (1333 * 3 / 2)

	if payout != expected {
		t.Errorf("expected payout %d, got %d", expected, payout)
	}
}

func TestMultipleHits(t *testing.T) {
	game := NewGame()
	game.Bet = 1000
	game.Phase = PhasePlayerTurn

	// Start with low cards
	game.PlayerHand.AddCard(Card{Rank: "2", Value: 2})
	game.PlayerHand.AddCard(Card{Rank: "3", Value: 3})

	// Mock deck with multiple small cards
	game.Deck = &Deck{Cards: []Card{
		{Rank: "2", Value: 2},
		{Rank: "3", Value: 3},
		{Rank: "4", Value: 4},
		{Rank: "K", Value: 10}, // Dealer's second card
	}}
	game.DealerHand.AddCard(Card{Rank: "K", Value: 10})

	// Hit three times
	for i := 0; i < 3; i++ {
		if game.Phase != PhasePlayerTurn {
			break
		}
		err := game.Hit()
		if err != nil {
			t.Errorf("unexpected error on hit %d: %v", i+1, err)
		}
	}

	// Should have 5 cards now (2 + 3 + 2 + 3 + 4 = 14)
	if len(game.PlayerHand.Cards) != 5 {
		t.Errorf("expected 5 cards after 3 hits, got %d", len(game.PlayerHand.Cards))
	}

	if game.PlayerHand.Value() != 14 {
		t.Errorf("expected value 14, got %d", game.PlayerHand.Value())
	}
}

func TestPlayer21WithMultipleCardsVsDealer20(t *testing.T) {
	game := NewGame()
	game.Bet = 1000

	// Player has 21 with 3 cards (not blackjack)
	game.PlayerHand.AddCard(Card{Rank: "7", Value: 7})
	game.PlayerHand.AddCard(Card{Rank: "7", Value: 7})
	game.PlayerHand.AddCard(Card{Rank: "7", Value: 7})

	// Dealer has 20
	game.DealerHand.AddCard(Card{Rank: "K", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "Q", Value: 10})

	game.Phase = PhaseGameOver
	game.determineWinner()

	if game.Result != ResultPlayerWin {
		t.Errorf("expected player win with 21 vs 20, got %s", game.Result)
	}

	// Should pay 1:1, not 3:2 (not a blackjack)
	payout := game.CalculatePayout()
	if payout != 2000 {
		t.Errorf("expected regular win payout 2000, got %d", payout)
	}
}

func TestDealerStopsAtExact17(t *testing.T) {
	game := NewGame()
	game.Bet = 1000
	game.Phase = PhaseDealerTurn

	// Dealer has 17 exactly
	game.DealerHand.AddCard(Card{Rank: "K", Value: 10})
	game.DealerHand.AddCard(Card{Rank: "7", Value: 7})

	// Player has 18
	game.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "8", Value: 8})

	// Mock empty deck (dealer shouldn't draw)
	game.Deck = &Deck{Cards: []Card{}}

	initialDealerCards := len(game.DealerHand.Cards)
	game.playDealerTurn()

	// Dealer should not have drawn any cards
	if len(game.DealerHand.Cards) != initialDealerCards {
		t.Error("dealer should stand at 17")
	}

	if game.DealerHand.Value() != 17 {
		t.Errorf("dealer should have 17, got %d", game.DealerHand.Value())
	}
}

func TestHitUntilBust(t *testing.T) {
	game := NewGame()
	game.Bet = 1000
	game.Phase = PhasePlayerTurn

	// Start with 18
	game.PlayerHand.AddCard(Card{Rank: "K", Value: 10})
	game.PlayerHand.AddCard(Card{Rank: "8", Value: 8})

	// Mock deck with a card that will bust
	game.Deck = &Deck{Cards: []Card{{Rank: "5", Value: 5}}}

	err := game.Hit()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !game.PlayerHand.IsBusted() {
		t.Error("player should be busted")
	}

	if game.Phase != PhaseGameOver {
		t.Errorf("game should be over after bust, phase is %s", game.Phase)
	}

	if game.Result != ResultDealerWin {
		t.Errorf("expected dealer win after player bust, got %s", game.Result)
	}
}
