package game

import (
	"fmt"
	"strings"
)

// CalculateBustProbability calculates the probability of busting on the next hit
// given the current hand value and the remaining cards in the deck
func CalculateBustProbability(hand *Hand, deck *Deck) float64 {
	handValue := hand.Value()

	// If already busted or at 21, probability is 0 or 100%
	if handValue >= 21 {
		return 0.0
	}

	// Count how many cards would bust
	bustCards := 0
	totalCards := len(deck.Cards)

	if totalCards == 0 {
		return 0.0
	}

	for _, card := range deck.Cards {
		// Check if this card would bust the hand
		// Need to account for Aces being 11 or 1
		testValue := handValue + card.Value

		// If the card is an Ace and would bust, try counting it as 1
		if card.Rank == "A" && testValue > 21 {
			testValue = handValue + 1
		}

		if testValue > 21 {
			bustCards++
		}
	}

	return float64(bustCards) / float64(totalCards) * 100.0
}

// GetBustCards returns a list of card ranks that would bust the hand
func GetBustCards(handValue int) []string {
	if handValue >= 21 {
		return []string{}
	}

	bustThreshold := 21 - handValue
	bustRanks := []string{}

	// Ordered list of card ranks and their values
	cardRanks := []struct {
		rank  string
		value int
	}{
		{"2", 2}, {"3", 3}, {"4", 4}, {"5", 5}, {"6", 6},
		{"7", 7}, {"8", 8}, {"9", 9}, {"10", 10},
		{"J", 10}, {"Q", 10}, {"K", 10}, {"A", 11},
	}

	for _, card := range cardRanks {
		// Aces can be 1 or 11, so they never bust (can always be counted as 1)
		if card.rank == "A" {
			continue
		}

		if card.value > bustThreshold {
			bustRanks = append(bustRanks, card.rank)
		}
	}

	return bustRanks
}

// GetOptimalAction suggests the optimal action based on basic strategy
// This is a simplified version of basic strategy
func GetOptimalAction(playerHand *Hand, dealerUpcard Card, canDoubleDown bool, canSurrender bool) string {
	playerValue := playerHand.Value()
	dealerValue := dealerUpcard.Value

	// Handle soft hands (hands with an Ace counted as 11)
	isSoft := false
	for _, card := range playerHand.Cards {
		if card.Rank == "A" && playerValue <= 21 {
			// Check if the Ace is being counted as 11
			valueWithoutAce := 0
			for _, c := range playerHand.Cards {
				if c.Rank != "A" {
					valueWithoutAce += c.Value
				}
			}
			if valueWithoutAce+11 == playerValue {
				isSoft = true
				break
			}
		}
	}

	// Pair splitting (only on initial 2-card hand)
	if len(playerHand.Cards) == 2 && playerHand.Cards[0].Rank == playerHand.Cards[1].Rank {
		// Note: Current implementation doesn't support splitting, so we skip this
	}

	// Surrender strategy (only on initial 2-card hand)
	// Check surrender before other actions for hard 15-16
	if len(playerHand.Cards) == 2 && canSurrender && !isSoft {
		if playerValue == 16 && dealerValue >= 9 {
			return "SURRENDER"
		}
		if playerValue == 15 && dealerValue >= 10 {
			return "SURRENDER"
		}
	}

	// Soft hand strategy
	if isSoft && playerValue < 21 {
		if playerValue >= 19 {
			return "STAND"
		}
		if playerValue == 18 {
			if dealerValue >= 9 {
				return "HIT"
			}
			return "STAND"
		}
		// Soft 17 or less
		return "HIT"
	}

	// Hard hand strategy
	if playerValue >= 17 {
		return "STAND"
	}

	if playerValue >= 13 && playerValue <= 16 {
		if dealerValue >= 7 {
			return "HIT"
		}
		return "STAND"
	}

	if playerValue == 12 {
		if dealerValue >= 4 && dealerValue <= 6 {
			return "STAND"
		}
		return "HIT"
	}

	if playerValue == 11 && canDoubleDown {
		return "DOUBLEDOWN"
	}

	if playerValue == 10 && canDoubleDown {
		if dealerValue <= 9 {
			return "DOUBLEDOWN"
		}
		return "HIT"
	}

	if playerValue == 9 && canDoubleDown {
		if dealerValue >= 3 && dealerValue <= 6 {
			return "DOUBLEDOWN"
		}
		return "HIT"
	}

	// Default: hit on anything 11 or less
	if playerValue <= 11 {
		return "HIT"
	}

	return "STAND"
}

// FormatBustProbability formats the bust probability with bust cards
func FormatBustProbability(probability float64, bustCards []string) string {
	if probability == 0 {
		return "Bust probability: 0% (safe to hit)"
	}

	// Sort bust cards for consistent display
	bustCardsStr := strings.Join(bustCards, ",")

	return fmt.Sprintf("Bust probability: %.1f%% (busts on: %s)", probability, bustCardsStr)
}
