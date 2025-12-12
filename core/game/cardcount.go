package game

// CardCounter implements the Hi-Lo card counting system
type CardCounter struct {
	RunningCount int
	CardsDealt   int
}

// NewCardCounter creates a new card counter
func NewCardCounter() *CardCounter {
	return &CardCounter{
		RunningCount: 0,
		CardsDealt:   0,
	}
}

// UpdateCount updates the running count based on the dealt card
// Hi-Lo System:
// - Cards 2-6: +1 (low cards favor dealer, good for player when gone)
// - Cards 7-9: 0 (neutral)
// - Cards 10-A: -1 (high cards favor player, bad for player when gone)
func (cc *CardCounter) UpdateCount(card Card) {
	cc.CardsDealt++

	switch card.Rank {
	case "2", "3", "4", "5", "6":
		cc.RunningCount++
	case "7", "8", "9":
		// Neutral - no change
	case "10", "J", "Q", "K", "A":
		cc.RunningCount--
	}
}

// GetRunningCount returns the current running count
func (cc *CardCounter) GetRunningCount() int {
	return cc.RunningCount
}

// GetTrueCount calculates the true count based on decks remaining
// True Count = Running Count / Decks Remaining
// This normalizes the count for the amount of cards left to be dealt
func (cc *CardCounter) GetTrueCount(decksRemaining float64) float64 {
	if decksRemaining <= 0 {
		return 0
	}
	return float64(cc.RunningCount) / decksRemaining
}

// GetAdvantage estimates the player's advantage based on the true count
// Rule of thumb: Each +1 true count gives approximately 0.5% player advantage
// Negative true counts give the house an advantage
func (cc *CardCounter) GetAdvantage(decksRemaining float64) float64 {
	trueCount := cc.GetTrueCount(decksRemaining)
	return trueCount * 0.5
}

// Reset resets the counter (typically done when deck is shuffled)
func (cc *CardCounter) Reset() {
	cc.RunningCount = 0
	cc.CardsDealt = 0
}

// GetCardsDealt returns the number of cards dealt since last reset
func (cc *CardCounter) GetCardsDealt() int {
	return cc.CardsDealt
}

// CalculateDecksRemaining calculates how many decks remain based on cards dealt
// Assumes a standard 52-card deck
func CalculateDecksRemaining(totalDecks int, cardsDealt int) float64 {
	totalCards := totalDecks * 52
	cardsRemaining := totalCards - cardsDealt
	if cardsRemaining <= 0 {
		return 0
	}
	return float64(cardsRemaining) / 52.0
}
