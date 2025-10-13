package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Card struct {
	Suit  string
	Rank  string
	Value int
}

type Deck struct {
	Cards []Card
}

type Hand struct {
	Cards []Card
}

type GamePhase string

const (
	PhaseWaitingForBet GamePhase = "WAITING_FOR_BET"
	PhasePlayerTurn    GamePhase = "PLAYER_TURN"
	PhaseDealerTurn    GamePhase = "DEALER_TURN"
	PhaseGameOver      GamePhase = "GAME_OVER"
)

type GameResult string

const (
	ResultPlayerWin       GameResult = "PLAYER_WIN"
	ResultDealerWin       GameResult = "DEALER_WIN"
	ResultPush            GameResult = "PUSH"
	ResultPlayerBlackjack GameResult = "PLAYER_BLACKJACK"
)

type Game struct {
	Deck        *Deck
	PlayerHand  *Hand
	DealerHand  *Hand
	Phase       GamePhase
	Bet         int64 // in cents
	Result      GameResult
	IsDoubled   bool
	PlayerStood bool
}

func NewDeck() *Deck {
	suits := []string{"♠", "♥", "♦", "♣"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

	// Map ranks to their values
	rankValues := map[string]int{
		"A": 11, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7,
		"8": 8, "9": 9, "10": 10, "J": 10, "Q": 10, "K": 10,
	}

	deck := &Deck{Cards: make([]Card, 0, 52)}

	for _, suit := range suits {
		for _, rank := range ranks {
			deck.Cards = append(deck.Cards, Card{
				Suit:  suit,
				Rank:  rank,
				Value: rankValues[rank],
			})
		}
	}

	return deck
}

func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) Draw() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, fmt.Errorf("deck is empty")
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, nil
}

func NewHand() *Hand {
	return &Hand{Cards: make([]Card, 0)}
}

func (h *Hand) AddCard(card Card) {
	h.Cards = append(h.Cards, card)
}

// Value calculates the best value of the hand (handling Aces)
func (h *Hand) Value() int {
	value := 0
	aces := 0

	for _, card := range h.Cards {
		if card.Rank == "A" {
			aces++
			value += 11
		} else {
			value += card.Value
		}
	}

	for aces > 0 && value > 21 {
		value -= 10
		aces--
	}

	return value
}

func (h *Hand) IsBusted() bool {
	return h.Value() > 21
}

func (h *Hand) IsBlackjack() bool {
	return len(h.Cards) == 2 && h.Value() == 21
}

// String representation of the hand
func (h *Hand) String() string {
	var cards []string
	for _, card := range h.Cards {
		cards = append(cards, fmt.Sprintf("[%s%s]", card.Rank, card.Suit))
	}
	return strings.Join(cards, " ")
}

func NewGame() *Game {
	return &Game{
		Deck:        NewDeck(),
		PlayerHand:  NewHand(),
		DealerHand:  NewHand(),
		Phase:       PhaseWaitingForBet,
		Bet:         0,
		IsDoubled:   false,
		PlayerStood: false,
	}
}

func NewGameWithDeck(cards []Card) *Game {
	// Copy cards to avoid modifying the original slice
	deckCopy := make([]Card, len(cards))
	copy(deckCopy, cards)

	return &Game{
		Deck:        &Deck{Cards: deckCopy},
		PlayerHand:  NewHand(),
		DealerHand:  NewHand(),
		Phase:       PhaseWaitingForBet,
		Bet:         0,
		IsDoubled:   false,
		PlayerStood: false,
	}
}

// Places a bet without shuffling (for testing with deterministic decks)
func (g *Game) PlaceBetNoShuffle(amount int64) error {
	if g.Phase != PhaseWaitingForBet {
		return fmt.Errorf("cannot place bet in current phase")
	}
	if amount <= 0 {
		return fmt.Errorf("bet must be positive")
	}

	g.Bet = amount

	// Deal initial cards - check for deck exhaustion
	card1, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.PlayerHand.AddCard(card1)

	card2, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.DealerHand.AddCard(card2)

	card3, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.PlayerHand.AddCard(card3)

	card4, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.DealerHand.AddCard(card4)

	pBJ := g.PlayerHand.IsBlackjack()
	dBJ := g.DealerHand.IsBlackjack()

	switch {
	case pBJ && dBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultPush
	case pBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultPlayerBlackjack
	case dBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultDealerWin
	default:
		g.Phase = PhasePlayerTurn
	}

	return nil
}

// Places a bet and starts the game
func (g *Game) PlaceBet(amount int64) error {
	if g.Phase != PhaseWaitingForBet {
		return fmt.Errorf("cannot place bet in current phase")
	}
	if amount <= 0 {
		return fmt.Errorf("bet must be positive")
	}

	g.Bet = amount
	g.Deck.Shuffle()

	// Deal initial cards - check for deck exhaustion
	card1, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.PlayerHand.AddCard(card1)

	card2, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.DealerHand.AddCard(card2)

	card3, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.PlayerHand.AddCard(card3)

	card4, err := g.Deck.Draw()
	if err != nil {
		return fmt.Errorf("failed to deal: %w", err)
	}
	g.DealerHand.AddCard(card4)

	pBJ := g.PlayerHand.IsBlackjack()
	dBJ := g.DealerHand.IsBlackjack()

	switch {
	case pBJ && dBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultPush
	case pBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultPlayerBlackjack
	case dBJ:
		g.Phase = PhaseGameOver
		g.Result = ResultDealerWin
	default:
		g.Phase = PhasePlayerTurn
	}

	return nil
}

// Draws a card for the player
func (g *Game) Hit() error {
	if g.Phase != PhasePlayerTurn {
		return fmt.Errorf("cannot hit in current phase")
	}

	card, err := g.Deck.Draw()
	if err != nil {
		return err
	}

	g.PlayerHand.AddCard(card)

	if g.PlayerHand.IsBusted() {
		g.Phase = PhaseGameOver
		g.Result = ResultDealerWin
	}

	return nil
}

// Ends the player's turn and starts dealer's turn
func (g *Game) Stand() error {
	if g.Phase != PhasePlayerTurn {
		return fmt.Errorf("cannot stand in current phase")
	}

	g.PlayerStood = true
	g.Phase = PhaseDealerTurn
	g.playDealerTurn()

	return nil
}

// Doubles the bet, draws one card, and ends player's turn
func (g *Game) DoubleDown() error {
	if g.Phase != PhasePlayerTurn {
		return fmt.Errorf("cannot double down in current phase")
	}
	if len(g.PlayerHand.Cards) != 2 {
		return fmt.Errorf("can only double down on initial hand")
	}

	g.Bet *= 2
	g.IsDoubled = true

	card, err := g.Deck.Draw()
	if err != nil {
		return err
	}

	g.PlayerHand.AddCard(card)

	if g.PlayerHand.IsBusted() {
		g.Phase = PhaseGameOver
		g.Result = ResultDealerWin
	} else {
		g.Phase = PhaseDealerTurn
		g.playDealerTurn()
	}

	return nil
}

// Plays the dealer's turn according to standard rules
func (g *Game) playDealerTurn() {
	// Dealer must hit on 16 or less, stand on 17 or more
	for g.DealerHand.Value() < 17 {
		card, err := g.Deck.Draw()
		if err != nil {
			// Deck exhausted - treat as a push to avoid corruption
			g.Phase = PhaseGameOver
			g.Result = ResultPush
			return
		}
		g.DealerHand.AddCard(card)
	}

	g.Phase = PhaseGameOver
	g.determineWinner()
}

func (g *Game) determineWinner() {
	playerValue := g.PlayerHand.Value()
	dealerValue := g.DealerHand.Value()

	if g.PlayerHand.IsBusted() {
		g.Result = ResultDealerWin
	} else if g.DealerHand.IsBusted() {
		g.Result = ResultPlayerWin
	} else if playerValue > dealerValue {
		g.Result = ResultPlayerWin
	} else if dealerValue > playerValue {
		g.Result = ResultDealerWin
	} else {
		g.Result = ResultPush
	}
}

func (g *Game) CalculatePayout() int64 {
	switch g.Result {
	case ResultPlayerBlackjack:
		return g.Bet + (g.Bet * 3 / 2) // Blackjack pays 3:2
	case ResultPlayerWin:
		return g.Bet * 2 // Regular win pays 1:1
	case ResultPush:
		return g.Bet // Push returns bet
	case ResultDealerWin:
		return 0 // Loss returns nothing
	default:
		return 0
	}
}

func (g *Game) GetGameState(hideDealer bool) string {
	state := fmt.Sprintf("Bet: $%.2f\n", float64(g.Bet)/100)

	if len(g.PlayerHand.Cards) == 0 {
		state += "Player Hand: (no cards dealt)\n"
	} else {
		state += fmt.Sprintf("Player Hand: %s (Value: %d)\n", g.PlayerHand.String(), g.PlayerHand.Value())
	}

	if len(g.DealerHand.Cards) == 0 {
		state += "Dealer Hand: (no cards dealt)\n"
	} else if hideDealer && g.Phase == PhasePlayerTurn {
		// Hide dealer's second card during player turn
		firstCard := g.DealerHand.Cards[0]
		state += fmt.Sprintf("Dealer Hand: [%s%s] [Hidden]\n", firstCard.Rank, firstCard.Suit)
	} else {
		state += fmt.Sprintf("Dealer Hand: %s (Value: %d)\n", g.DealerHand.String(), g.DealerHand.Value())
	}

	if g.Phase == PhaseGameOver {
		state += fmt.Sprintf("\nResult: %s\n", g.getResultMessage())
		state += fmt.Sprintf("Payout: $%.2f\n", float64(g.CalculatePayout())/100)
	}

	return state
}

func (g *Game) getResultMessage() string {
	switch g.Result {
	case ResultPlayerBlackjack:
		return "Blackjack! You win!"
	case ResultPlayerWin:
		return "You win!"
	case ResultDealerWin:
		if g.PlayerHand.IsBusted() {
			return "Bust! Dealer wins."
		}
		return "Dealer wins."
	case ResultPush:
		return "Push - tie game."
	default:
		return ""
	}
}

// GetValidActions returns a list of valid actions based on the current game state
func (g *Game) GetValidActions() []string {
	if g.Phase != PhasePlayerTurn {
		return []string{}
	}

	actions := []string{"HIT", "STAND"}

	// DOUBLEDOWN is only valid on the first action (2 cards)
	if len(g.PlayerHand.Cards) == 2 && !g.IsDoubled {
		actions = append(actions, "DOUBLEDOWN")
	}

	return actions
}
