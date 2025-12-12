package game

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestCalculateBustProbability(t *testing.T) {
	tests := []struct {
		name         string
		handValue    int
		setupDeck    func() *Deck
		expectedProb float64
	}{
		{
			name:      "Hand value 20 - only Ace won't bust",
			handValue: 20,
			setupDeck: func() *Deck {
				return NewDeck() // Full deck
			},
			expectedProb: 92.3, // 48 out of 52 cards bust (all except 4 Aces)
		},
		{
			name:      "Hand value 11 - cannot bust",
			handValue: 11,
			setupDeck: func() *Deck {
				return NewDeck()
			},
			expectedProb: 0.0,
		},
		{
			name:      "Hand value 21 - already at max",
			handValue: 21,
			setupDeck: func() *Deck {
				return NewDeck()
			},
			expectedProb: 0.0,
		},
		{
			name:      "Empty deck",
			handValue: 15,
			setupDeck: func() *Deck {
				return &Deck{Cards: []Card{}}
			},
			expectedProb: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := NewHand()
			// Create a hand with the desired value
			if tt.handValue <= 11 {
				hand.AddCard(Card{Rank: "2", Suit: "♠", Value: 2})
				hand.AddCard(Card{Rank: "9", Suit: "♥", Value: 9})
			} else if tt.handValue == 20 {
				hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
				hand.AddCard(Card{Rank: "Q", Suit: "♥", Value: 10})
			} else if tt.handValue == 21 {
				hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
				hand.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
			}

			deck := tt.setupDeck()
			prob := CalculateBustProbability(hand, deck)

			if math.Abs(prob-tt.expectedProb) > 1.0 {
				t.Errorf("Expected bust probability ~%.1f%%, got %.1f%%", tt.expectedProb, prob)
			}
		})
	}
}

func TestCalculateBustProbabilitySpecificScenarios(t *testing.T) {
	// Test with hand value 15 (should bust on 7,8,9,10,J,Q,K)
	hand := NewHand()
	hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
	hand.AddCard(Card{Rank: "6", Suit: "♥", Value: 6})

	deck := NewDeck()
	prob := CalculateBustProbability(hand, deck)

	// In a full deck: 4 sevens, 4 eights, 4 nines, 16 tens (10,J,Q,K) = 28 bust cards
	// 28/52 = 53.8%
	expectedProb := 53.8
	if math.Abs(prob-expectedProb) > 2.0 {
		t.Errorf("For hand value 15, expected ~%.1f%% bust probability, got %.1f%%", expectedProb, prob)
	}
}

func TestGetBustCards(t *testing.T) {
	tests := []struct {
		name          string
		handValue     int
		expectedCards []string
	}{
		{
			name:          "Hand value 15",
			handValue:     15,
			expectedCards: []string{"7", "8", "9", "10", "J", "Q", "K"},
		},
		{
			name:          "Hand value 20",
			handValue:     20,
			expectedCards: []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}, // All except Ace
		},
		{
			name:          "Hand value 11",
			handValue:     11,
			expectedCards: []string{},
		},
		{
			name:          "Hand value 21",
			handValue:     21,
			expectedCards: []string{},
		},
		{
			name:          "Hand value 12",
			handValue:     12,
			expectedCards: []string{"10", "J", "Q", "K"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bustCards := GetBustCards(tt.handValue)

			if len(bustCards) != len(tt.expectedCards) {
				t.Errorf("Expected %d bust cards, got %d: %v", len(tt.expectedCards), len(bustCards), bustCards)
			}

			// Check that all expected cards are present
			for _, expected := range tt.expectedCards {
				found := false
				for _, card := range bustCards {
					if card == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected bust card %s not found in %v", expected, bustCards)
				}
			}
		})
	}
}

func TestGetOptimalActionHardHands(t *testing.T) {
	tests := []struct {
		name           string
		playerValue    int
		dealerUpcard   int
		canDoubleDown  bool
		expectedAction string
	}{
		// Hard 17+
		{"Hard 17 vs 10", 17, 10, false, "STAND"},
		{"Hard 18 vs 10", 18, 10, false, "STAND"},
		{"Hard 20 vs 10", 20, 10, false, "STAND"},

		// Hard 13-16
		{"Hard 16 vs 7", 16, 7, false, "HIT"},
		{"Hard 15 vs 10", 15, 10, false, "HIT"},
		{"Hard 14 vs 6", 14, 6, false, "STAND"},
		{"Hard 13 vs 5", 13, 5, false, "STAND"},

		// Hard 12
		{"Hard 12 vs 4", 12, 4, false, "STAND"},
		{"Hard 12 vs 3", 12, 3, false, "HIT"},
		{"Hard 12 vs 7", 12, 7, false, "HIT"},

		// Hard 11
		{"Hard 11 vs 6 (can double)", 11, 6, true, "DOUBLEDOWN"},
		{"Hard 11 vs 10 (can double)", 11, 10, true, "DOUBLEDOWN"},

		// Hard 10
		{"Hard 10 vs 9 (can double)", 10, 9, true, "DOUBLEDOWN"},
		{"Hard 10 vs 10 (can double)", 10, 10, true, "HIT"},

		// Hard 9
		{"Hard 9 vs 6 (can double)", 9, 6, true, "DOUBLEDOWN"},
		{"Hard 9 vs 2 (can double)", 9, 2, true, "HIT"},

		// Low hands
		{"Hard 8 vs 10", 8, 10, false, "HIT"},
		{"Hard 5 vs 10", 5, 10, false, "HIT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := createHandWithValue(tt.playerValue, false)
			dealerCard := Card{Rank: "10", Suit: "♠", Value: tt.dealerUpcard}
			if tt.dealerUpcard < 10 {
				dealerCard.Rank = string(rune('0' + tt.dealerUpcard))
			}

			action := GetOptimalAction(hand, dealerCard, tt.canDoubleDown, false)
			if action != tt.expectedAction {
				t.Errorf("Expected %s, got %s", tt.expectedAction, action)
			}
		})
	}
}

func TestGetOptimalActionSoftHands(t *testing.T) {
	tests := []struct {
		name           string
		playerValue    int
		dealerUpcard   int
		expectedAction string
	}{
		{"Soft 19 vs 6", 19, 6, "STAND"},
		{"Soft 18 vs 9", 18, 9, "HIT"},
		{"Soft 18 vs 7", 18, 7, "STAND"},
		{"Soft 17 vs 6", 17, 6, "HIT"},
		{"Soft 13 vs 5", 13, 5, "HIT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := createHandWithValue(tt.playerValue, true)
			dealerCard := Card{Rank: "10", Suit: "♠", Value: tt.dealerUpcard}
			if tt.dealerUpcard < 10 {
				dealerCard.Rank = string(rune('0' + tt.dealerUpcard))
			}

			action := GetOptimalAction(hand, dealerCard, false, false)
			if action != tt.expectedAction {
				t.Errorf("Expected %s, got %s", tt.expectedAction, action)
			}
		})
	}
}

func TestGetOptimalActionSurrender(t *testing.T) {
	tests := []struct {
		name           string
		playerValue    int
		dealerUpcard   int
		expectedAction string
	}{
		{"16 vs 9 (should surrender)", 16, 9, "SURRENDER"},
		{"16 vs 10 (should surrender)", 16, 10, "SURRENDER"},
		{"15 vs 10 (should surrender)", 15, 10, "SURRENDER"},
		{"14 vs 10 (should not surrender)", 14, 10, "HIT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := createHandWithValue(tt.playerValue, false)
			dealerCard := Card{Rank: "10", Suit: "♠", Value: tt.dealerUpcard}
			if tt.dealerUpcard < 10 {
				dealerCard.Rank = string(rune('0' + tt.dealerUpcard))
			}

			action := GetOptimalAction(hand, dealerCard, false, true)
			if action != tt.expectedAction {
				t.Errorf("Expected %s, got %s", tt.expectedAction, action)
			}
		})
	}
}

func TestFormatBustProbability(t *testing.T) {
	tests := []struct {
		name        string
		probability float64
		bustCards   []string
		expected    string
	}{
		{
			name:        "Zero probability",
			probability: 0.0,
			bustCards:   []string{},
			expected:    "Bust probability: 0% (safe to hit)",
		},
		{
			name:        "High probability",
			probability: 92.3,
			bustCards:   []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"},
			expected:    "Bust probability: 92.3% (busts on: 2,3,4,5,6,7,8,9,10,J,Q,K)",
		},
		{
			name:        "Medium probability",
			probability: 53.8,
			bustCards:   []string{"7", "8", "9", "10", "J", "Q", "K"},
			expected:    "Bust probability: 53.8% (busts on: 7,8,9,10,J,Q,K)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBustProbability(tt.probability, tt.bustCards)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestCalculateBustProbabilityWithAces(t *testing.T) {
	// Test that Aces are handled correctly (can be 1 or 11)
	hand := NewHand()
	hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
	hand.AddCard(Card{Rank: "9", Suit: "♥", Value: 9}) // Hand value: 19

	// Create a deck with only Aces
	deck := &Deck{
		Cards: []Card{
			{Rank: "A", Suit: "♠", Value: 11},
			{Rank: "A", Suit: "♥", Value: 11},
			{Rank: "A", Suit: "♦", Value: 11},
			{Rank: "A", Suit: "♣", Value: 11},
		},
	}

	prob := CalculateBustProbability(hand, deck)

	// Aces can be counted as 1, so 19 + 1 = 20, which doesn't bust
	// Therefore, bust probability should be 0%
	if prob != 0.0 {
		t.Errorf("Expected 0%% bust probability with only Aces remaining, got %.1f%%", prob)
	}
}

// Helper function to create a hand with a specific value
func createHandWithValue(value int, soft bool) *Hand {
	hand := NewHand()

	if soft {
		// Soft hand: Ace + another card
		hand.AddCard(Card{Rank: "A", Suit: "♠", Value: 11})
		remaining := value - 11
		if remaining > 0 {
			if remaining <= 10 {
				hand.AddCard(Card{Rank: string(rune('0' + remaining)), Suit: "♥", Value: remaining})
			} else {
				hand.AddCard(Card{Rank: "10", Suit: "♥", Value: 10})
			}
		}
	} else {
		// Hard hand: no Ace or Ace counted as 1
		if value <= 10 {
			hand.AddCard(Card{Rank: string(rune('0' + value)), Suit: "♠", Value: value})
		} else if value == 11 {
			hand.AddCard(Card{Rank: "9", Suit: "♠", Value: 9})
			hand.AddCard(Card{Rank: "2", Suit: "♥", Value: 2})
		} else if value <= 20 {
			hand.AddCard(Card{Rank: "10", Suit: "♠", Value: 10})
			remaining := value - 10
			if remaining <= 10 {
				rankStr := string(rune('0' + remaining))
				if remaining == 10 {
					rankStr = "10"
				}
				hand.AddCard(Card{Rank: rankStr, Suit: "♥", Value: remaining})
			}
		} else {
			hand.AddCard(Card{Rank: "K", Suit: "♠", Value: 10})
			hand.AddCard(Card{Rank: "A", Suit: "♥", Value: 11})
		}
	}

	return hand
}

func TestGetBustCardsComprehensive(t *testing.T) {
	// Test all hand values from 12 to 20
	tests := []struct {
		handValue    int
		minBustCards int
		maxBustCards int
	}{
		{12, 4, 4},   // 10, J, Q, K
		{13, 5, 5},   // 9, 10, J, Q, K
		{14, 6, 6},   // 8, 9, 10, J, Q, K
		{15, 7, 7},   // 7, 8, 9, 10, J, Q, K
		{16, 8, 8},   // 6, 7, 8, 9, 10, J, Q, K
		{17, 9, 9},   // 5, 6, 7, 8, 9, 10, J, Q, K
		{18, 10, 10}, // 4, 5, 6, 7, 8, 9, 10, J, Q, K
		{19, 11, 11}, // 3, 4, 5, 6, 7, 8, 9, 10, J, Q, K
		{20, 12, 12}, // All except Ace (Ace can be 1)
		{21, 0, 0},   // Already at 21
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Hand value %d", tt.handValue), func(t *testing.T) {
			bustCards := GetBustCards(tt.handValue)
			if len(bustCards) < tt.minBustCards || len(bustCards) > tt.maxBustCards {
				t.Errorf("For hand value %d, expected %d-%d bust cards, got %d: %v",
					tt.handValue, tt.minBustCards, tt.maxBustCards, len(bustCards), bustCards)
			}
		})
	}
}

func TestFormatBustProbabilityContainsKeyElements(t *testing.T) {
	prob := 58.3
	bustCards := []string{"7", "8", "9", "10", "J", "Q", "K"}

	result := FormatBustProbability(prob, bustCards)

	// Check that result contains key elements
	if !strings.Contains(result, "58.3%") {
		t.Error("Result should contain probability percentage")
	}

	if !strings.Contains(result, "busts on:") {
		t.Error("Result should contain 'busts on:' label")
	}

	for _, card := range bustCards {
		if !strings.Contains(result, card) {
			t.Errorf("Result should contain bust card %s", card)
		}
	}
}
