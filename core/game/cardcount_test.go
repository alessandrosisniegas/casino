package game

import (
	"math"
	"testing"
)

func TestNewCardCounter(t *testing.T) {
	counter := NewCardCounter()

	if counter.RunningCount != 0 {
		t.Errorf("Expected RunningCount = 0, got %d", counter.RunningCount)
	}

	if counter.CardsDealt != 0 {
		t.Errorf("Expected CardsDealt = 0, got %d", counter.CardsDealt)
	}
}

func TestUpdateCountLowCards(t *testing.T) {
	counter := NewCardCounter()

	// Low cards (2-6) should increase count by +1
	lowCards := []Card{
		{Rank: "2", Suit: "♠", Value: 2},
		{Rank: "3", Suit: "♥", Value: 3},
		{Rank: "4", Suit: "♦", Value: 4},
		{Rank: "5", Suit: "♣", Value: 5},
		{Rank: "6", Suit: "♠", Value: 6},
	}

	for i, card := range lowCards {
		counter.UpdateCount(card)
		expectedCount := i + 1
		if counter.RunningCount != expectedCount {
			t.Errorf("After card %s, expected count = %d, got %d", card.Rank, expectedCount, counter.RunningCount)
		}
	}

	if counter.CardsDealt != 5 {
		t.Errorf("Expected CardsDealt = 5, got %d", counter.CardsDealt)
	}
}

func TestUpdateCountNeutralCards(t *testing.T) {
	counter := NewCardCounter()

	// Neutral cards (7-9) should not change count
	neutralCards := []Card{
		{Rank: "7", Suit: "♠", Value: 7},
		{Rank: "8", Suit: "♥", Value: 8},
		{Rank: "9", Suit: "♦", Value: 9},
	}

	for _, card := range neutralCards {
		counter.UpdateCount(card)
		if counter.RunningCount != 0 {
			t.Errorf("After card %s, expected count = 0, got %d", card.Rank, counter.RunningCount)
		}
	}

	if counter.CardsDealt != 3 {
		t.Errorf("Expected CardsDealt = 3, got %d", counter.CardsDealt)
	}
}

func TestUpdateCountHighCards(t *testing.T) {
	counter := NewCardCounter()

	// High cards (10-A) should decrease count by -1
	highCards := []Card{
		{Rank: "10", Suit: "♠", Value: 10},
		{Rank: "J", Suit: "♥", Value: 10},
		{Rank: "Q", Suit: "♦", Value: 10},
		{Rank: "K", Suit: "♣", Value: 10},
		{Rank: "A", Suit: "♠", Value: 11},
	}

	for i, card := range highCards {
		counter.UpdateCount(card)
		expectedCount := -(i + 1)
		if counter.RunningCount != expectedCount {
			t.Errorf("After card %s, expected count = %d, got %d", card.Rank, expectedCount, counter.RunningCount)
		}
	}

	if counter.CardsDealt != 5 {
		t.Errorf("Expected CardsDealt = 5, got %d", counter.CardsDealt)
	}
}

func TestUpdateCountMixed(t *testing.T) {
	counter := NewCardCounter()

	// Simulate a realistic sequence
	cards := []Card{
		{Rank: "K", Suit: "♠", Value: 10},  // -1
		{Rank: "5", Suit: "♥", Value: 5},   // +1, total: 0
		{Rank: "8", Suit: "♦", Value: 8},   // 0, total: 0
		{Rank: "2", Suit: "♣", Value: 2},   // +1, total: +1
		{Rank: "A", Suit: "♠", Value: 11},  // -1, total: 0
		{Rank: "6", Suit: "♥", Value: 6},   // +1, total: +1
		{Rank: "10", Suit: "♦", Value: 10}, // -1, total: 0
	}

	expectedCounts := []int{-1, 0, 0, 1, 0, 1, 0}

	for i, card := range cards {
		counter.UpdateCount(card)
		if counter.RunningCount != expectedCounts[i] {
			t.Errorf("After card %d (%s), expected count = %d, got %d", i+1, card.Rank, expectedCounts[i], counter.RunningCount)
		}
	}

	if counter.CardsDealt != 7 {
		t.Errorf("Expected CardsDealt = 7, got %d", counter.CardsDealt)
	}
}

func TestGetRunningCount(t *testing.T) {
	counter := NewCardCounter()
	counter.RunningCount = 5

	if counter.GetRunningCount() != 5 {
		t.Errorf("Expected GetRunningCount() = 5, got %d", counter.GetRunningCount())
	}
}

func TestGetTrueCount(t *testing.T) {
	tests := []struct {
		name           string
		runningCount   int
		decksRemaining float64
		expectedTrue   float64
	}{
		{"Positive count, 2 decks", 6, 2.0, 3.0},
		{"Positive count, 1 deck", 4, 1.0, 4.0},
		{"Negative count, 2 decks", -8, 2.0, -4.0},
		{"Zero count", 0, 1.5, 0.0},
		{"Fractional deck", 5, 0.5, 10.0},
		{"Zero decks remaining", 5, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCardCounter()
			counter.RunningCount = tt.runningCount

			trueCount := counter.GetTrueCount(tt.decksRemaining)
			if math.Abs(trueCount-tt.expectedTrue) > 0.01 {
				t.Errorf("Expected true count = %.2f, got %.2f", tt.expectedTrue, trueCount)
			}
		})
	}
}

func TestGetAdvantage(t *testing.T) {
	tests := []struct {
		name              string
		runningCount      int
		decksRemaining    float64
		expectedAdvantage float64
	}{
		{"Positive advantage", 4, 2.0, 1.0},   // True count +2 → 1% advantage
		{"Negative advantage", -6, 2.0, -1.5}, // True count -3 → -1.5% advantage
		{"Zero advantage", 0, 1.0, 0.0},
		{"High advantage", 10, 2.0, 2.5}, // True count +5 → 2.5% advantage
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCardCounter()
			counter.RunningCount = tt.runningCount

			advantage := counter.GetAdvantage(tt.decksRemaining)
			if math.Abs(advantage-tt.expectedAdvantage) > 0.01 {
				t.Errorf("Expected advantage = %.2f%%, got %.2f%%", tt.expectedAdvantage, advantage)
			}
		})
	}
}

func TestReset(t *testing.T) {
	counter := NewCardCounter()
	counter.RunningCount = 10
	counter.CardsDealt = 20

	counter.Reset()

	if counter.RunningCount != 0 {
		t.Errorf("After Reset(), expected RunningCount = 0, got %d", counter.RunningCount)
	}

	if counter.CardsDealt != 0 {
		t.Errorf("After Reset(), expected CardsDealt = 0, got %d", counter.CardsDealt)
	}
}

func TestGetCardsDealt(t *testing.T) {
	counter := NewCardCounter()

	cards := []Card{
		{Rank: "K", Suit: "♠", Value: 10},
		{Rank: "5", Suit: "♥", Value: 5},
		{Rank: "8", Suit: "♦", Value: 8},
	}

	for i, card := range cards {
		counter.UpdateCount(card)
		if counter.GetCardsDealt() != i+1 {
			t.Errorf("After %d cards, expected GetCardsDealt() = %d, got %d", i+1, i+1, counter.GetCardsDealt())
		}
	}
}

func TestCalculateDecksRemaining(t *testing.T) {
	tests := []struct {
		name          string
		totalDecks    int
		cardsDealt    int
		expectedDecks float64
	}{
		{"Full single deck", 1, 0, 1.0},
		{"Half single deck", 1, 26, 0.5},
		{"Full double deck", 2, 0, 2.0},
		{"One deck dealt from two", 2, 52, 1.0},
		{"Almost empty", 1, 51, 1.0 / 52.0},
		{"All cards dealt", 1, 52, 0.0},
		{"More than dealt (edge case)", 1, 60, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decksRemaining := CalculateDecksRemaining(tt.totalDecks, tt.cardsDealt)
			if math.Abs(decksRemaining-tt.expectedDecks) > 0.01 {
				t.Errorf("Expected %.2f decks remaining, got %.2f", tt.expectedDecks, decksRemaining)
			}
		})
	}
}

func TestFullDeckCount(t *testing.T) {
	// Test that a full deck results in a count of 0 (balanced deck)
	counter := NewCardCounter()
	deck := NewDeck()

	for _, card := range deck.Cards {
		counter.UpdateCount(card)
	}

	if counter.RunningCount != 0 {
		t.Errorf("Expected full deck count = 0 (balanced), got %d", counter.RunningCount)
	}

	if counter.CardsDealt != 52 {
		t.Errorf("Expected 52 cards dealt, got %d", counter.CardsDealt)
	}
}

func TestMultipleDeckCount(t *testing.T) {
	// Test that multiple full decks also result in count of 0
	counter := NewCardCounter()

	for i := 0; i < 6; i++ { // 6 decks
		deck := NewDeck()
		for _, card := range deck.Cards {
			counter.UpdateCount(card)
		}
	}

	if counter.RunningCount != 0 {
		t.Errorf("Expected 6 full decks count = 0 (balanced), got %d", counter.RunningCount)
	}

	if counter.CardsDealt != 312 {
		t.Errorf("Expected 312 cards dealt (6 decks), got %d", counter.CardsDealt)
	}
}
