package core

import (
	"testing"

	"github.com/alessandrosisniegas/casino/core/game"
	"github.com/alessandrosisniegas/casino/core/security"
	"github.com/alessandrosisniegas/casino/core/vault"
)

// TestFullGameFlow tests the complete user journey: signup → login → play → stats
func TestFullGameFlow(t *testing.T) {
	// Setup test database
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	// 1. Register a new user
	user, err := authService.RegisterUser("testplayer", "testpass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	if user.Balance != 1000000 {
		t.Errorf("Initial balance = %d, want 1000000", user.Balance)
	}

	// 2. Login
	sessionID, loginUser, err := authService.LoginUser("testplayer", "testpass123")
	if err != nil {
		t.Fatalf("LoginUser() error = %v", err)
	}

	if sessionID == "" {
		t.Error("Expected non-empty sessionID")
	}

	if loginUser.Username != "testplayer" {
		t.Errorf("Username = %s, want testplayer", loginUser.Username)
	}

	// 3. Play a game (simulate winning)
	// Dealer order: P, D, P, D, (dealer hits), (dealer hits)...
	g := game.NewGameWithDeck([]game.Card{
		{Rank: "K", Suit: "♠", Value: 10}, // Player card 1
		{Rank: "5", Suit: "♥", Value: 5},  // Dealer card 1
		{Rank: "9", Suit: "♣", Value: 9},  // Player card 2 -> 19
		{Rank: "6", Suit: "♦", Value: 6},  // Dealer card 2 -> 11
		{Rank: "K", Suit: "♥", Value: 10}, // Dealer hits -> 21
	})

	betAmount := int64(5000) // $50
	if err := g.PlaceBetNoShuffle(betAmount); err != nil {
		t.Fatalf("PlaceBet() error = %v", err)
	}

	// Player stands with 19
	if err := g.Stand(); err != nil {
		t.Fatalf("Stand() error = %v", err)
	}

	// Game should be over, dealer wins with 21 vs player 19
	if g.Phase != game.PhaseGameOver {
		t.Errorf("Phase = %v, want GAME_OVER", g.Phase)
	}

	if g.Result != game.ResultDealerWin {
		t.Errorf("Result = %v, want DEALER_WIN (dealer 21 beats player 19)", g.Result)
	}

	payout := g.CalculatePayout()
	expectedPayout := int64(0) // Loss = no payout
	if payout != expectedPayout {
		t.Errorf("Payout = %d, want %d", payout, expectedPayout)
	}

	// 4. Update balance and stats
	newBalance := loginUser.Balance - betAmount + payout
	if err := authService.UpdateBalance(loginUser.ID, newBalance); err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	stats, err := authService.GetUserStats(loginUser.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	stats.GamesPlayed++
	stats.GamesLost++ // This was a loss
	stats.TotalBet += betAmount
	stats.TotalWon += payout // 0 for a loss
	if betAmount > stats.BiggestLoss {
		stats.BiggestLoss = betAmount
	}

	if err := db.UpdateUserStats(stats); err != nil {
		t.Fatalf("UpdateUserStats() error = %v", err)
	}

	// 5. Verify stats persistence
	retrievedStats, err := authService.GetUserStats(loginUser.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if retrievedStats.GamesPlayed != 1 {
		t.Errorf("GamesPlayed = %d, want 1", retrievedStats.GamesPlayed)
	}

	if retrievedStats.GamesLost != 1 {
		t.Errorf("GamesLost = %d, want 1", retrievedStats.GamesLost)
	}

	if retrievedStats.TotalBet != betAmount {
		t.Errorf("TotalBet = %d, want %d", retrievedStats.TotalBet, betAmount)
	}

	if retrievedStats.TotalWon != payout {
		t.Errorf("TotalWon = %d, want %d", retrievedStats.TotalWon, payout)
	}

	// 6. Verify balance updated correctly
	updatedUser, err := authService.ValidateSession(sessionID)
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}

	expectedBalance := int64(995000) // Started with $10000, lost $50
	if updatedUser.Balance != expectedBalance {
		t.Errorf("Balance = %d, want %d", updatedUser.Balance, expectedBalance)
	}

	// 7. Logout
	if err := authService.LogoutUser(sessionID); err != nil {
		t.Fatalf("LogoutUser() error = %v", err)
	}

	// 8. Verify session is invalid after logout
	_, err = authService.ValidateSession(sessionID)
	if err == nil {
		t.Error("Expected error when validating logged-out session")
	}
}

// TestBalancePersistenceAcrossMultipleGames tests that balance updates correctly over multiple games
func TestBalancePersistenceAcrossMultipleGames(t *testing.T) {
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	user, err := authService.RegisterUser("gambler", "pass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	initialBalance := user.Balance // $10000

	// Simulate 5 games: 3 wins, 2 losses
	type gameOutcome struct {
		bet    int64
		result game.GameResult
	}

	games := []gameOutcome{
		{1000, game.ResultPlayerWin},       // Win $10
		{2000, game.ResultDealerWin},       // Lose $20
		{1500, game.ResultPlayerWin},       // Win $15
		{3000, game.ResultPlayerBlackjack}, // Blackjack win $30
		{1000, game.ResultDealerWin},       // Lose $10
	}

	currentBalance := initialBalance

	for i, outcome := range games {
		// Deduct bet first (simulating server logic)
		currentBalance -= outcome.bet

		// Then add payout based on result
		switch outcome.result {
		case game.ResultPlayerWin:
			currentBalance += outcome.bet * 2 // Get bet back + profit (1:1)
		case game.ResultPlayerBlackjack:
			currentBalance += outcome.bet + (outcome.bet * 3 / 2) // Bet + 3:2 payout
		case game.ResultDealerWin:
			// No payout
		case game.ResultPush:
			currentBalance += outcome.bet // Get bet back only
		}

		if err := authService.UpdateBalance(user.ID, currentBalance); err != nil {
			t.Fatalf("Game %d: UpdateBalance() error = %v", i+1, err)
		}
	}

	// Verify final balance
	finalUser, err := db.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	// Calculate expected:
	// Start: 1000000
	// Bet $10, win -> 1000000 - 1000 + 2000 = 1001000
	// Bet $20, lose -> 1001000 - 2000 + 0 = 999000
	// Bet $15, win -> 999000 - 1500 + 3000 = 1000500
	// Bet $30, BJ -> 1000500 - 3000 + 7500 = 1005000
	// Bet $10, lose -> 1005000 - 1000 + 0 = 1004000
	expectedBalance := int64(1004000)
	if finalUser.Balance != expectedBalance {
		t.Errorf("Final balance = %d, want %d", finalUser.Balance, expectedBalance)
	}
}

// TestStatsAccumulationOverMultipleGames verifies stats are correctly accumulated
func TestStatsAccumulationOverMultipleGames(t *testing.T) {
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	user, err := authService.RegisterUser("statsman", "pass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	// Play 10 games with various outcomes
	type gameOutcome struct {
		bet    int64
		result game.GameResult
	}

	outcomes := []gameOutcome{
		{1000, game.ResultPlayerWin},
		{2000, game.ResultPlayerWin},
		{1500, game.ResultDealerWin},
		{3000, game.ResultPlayerBlackjack},
		{1000, game.ResultDealerWin},
		{2500, game.ResultPlayerWin},
		{2000, game.ResultPush},
		{5000, game.ResultPlayerWin}, // Biggest win
		{4000, game.ResultDealerWin}, // Biggest loss
		{1500, game.ResultPlayerWin},
	}

	stats, err := authService.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	totalWins := 0
	totalLosses := 0

	for _, outcome := range outcomes {
		stats.GamesPlayed++
		stats.TotalBet += outcome.bet

		var payout int64
		switch outcome.result {
		case game.ResultPlayerWin:
			totalWins++
			stats.GamesWon++
			payout = outcome.bet * 2
			stats.TotalWon += payout
			if outcome.bet > stats.BiggestWin {
				stats.BiggestWin = outcome.bet
			}
		case game.ResultPlayerBlackjack:
			totalWins++
			stats.GamesWon++
			payout = outcome.bet + (outcome.bet * 3 / 2)
			stats.TotalWon += payout
			if outcome.bet > stats.BiggestWin {
				stats.BiggestWin = outcome.bet
			}
		case game.ResultDealerWin:
			totalLosses++
			stats.GamesLost++
			if outcome.bet > stats.BiggestLoss {
				stats.BiggestLoss = outcome.bet
			}
		case game.ResultPush:
			// Push doesn't count as win or loss
			payout = outcome.bet
			stats.TotalWon += payout
		}

		if err := db.UpdateUserStats(stats); err != nil {
			t.Fatalf("UpdateUserStats() error = %v", err)
		}
	}

	// Verify final stats
	finalStats, err := authService.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if finalStats.GamesPlayed != 10 {
		t.Errorf("GamesPlayed = %d, want 10", finalStats.GamesPlayed)
	}

	if finalStats.GamesWon != int64(totalWins) {
		t.Errorf("GamesWon = %d, want %d", finalStats.GamesWon, totalWins)
	}

	if finalStats.GamesLost != int64(totalLosses) {
		t.Errorf("GamesLost = %d, want %d", finalStats.GamesLost, totalLosses)
	}

	if finalStats.BiggestWin != 5000 {
		t.Errorf("BiggestWin = %d, want 5000", finalStats.BiggestWin)
	}

	if finalStats.BiggestLoss != 4000 {
		t.Errorf("BiggestLoss = %d, want 4000", finalStats.BiggestLoss)
	}

	// Verify average bet calculation
	// Total: 1000+2000+1500+3000+1000+2500+2000+5000+4000+1500 = 23500
	avgBet := float64(finalStats.TotalBet) / float64(finalStats.GamesPlayed)
	expectedAvg := 23500.0 / 10.0 // Total bet = 23500 cents, 10 games = $2.35 avg
	if avgBet != expectedAvg {
		t.Errorf("Average bet = %.2f, want %.2f", avgBet, expectedAvg)
	}
}

// TestConcurrentUserSessions tests multiple users playing simultaneously
func TestConcurrentUserSessions(t *testing.T) {
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	// Create two users
	user1, err := authService.RegisterUser("player1", "pass1")
	if err != nil {
		t.Fatalf("RegisterUser(player1) error = %v", err)
	}

	user2, err := authService.RegisterUser("player2", "pass2")
	if err != nil {
		t.Fatalf("RegisterUser(player2) error = %v", err)
	}

	// Both login
	session1, _, err := authService.LoginUser("player1", "pass1")
	if err != nil {
		t.Fatalf("LoginUser(player1) error = %v", err)
	}

	session2, _, err := authService.LoginUser("player2", "pass2")
	if err != nil {
		t.Fatalf("LoginUser(player2) error = %v", err)
	}

	// Verify sessions are different
	if session1 == session2 {
		t.Error("Expected different session IDs for different users")
	}

	// User1 plays a game and loses
	user1NewBalance := user1.Balance - 5000 // Lost $50
	if err := authService.UpdateBalance(user1.ID, user1NewBalance); err != nil {
		t.Fatalf("UpdateBalance(user1) error = %v", err)
	}

	// User2 plays a game and wins
	user2NewBalance := user2.Balance + 10000 // Won $100
	if err := authService.UpdateBalance(user2.ID, user2NewBalance); err != nil {
		t.Fatalf("UpdateBalance(user2) error = %v", err)
	}

	// Verify balances are independent
	updatedUser1, err := authService.ValidateSession(session1)
	if err != nil {
		t.Fatalf("ValidateSession(user1) error = %v", err)
	}

	updatedUser2, err := authService.ValidateSession(session2)
	if err != nil {
		t.Fatalf("ValidateSession(user2) error = %v", err)
	}

	if updatedUser1.Balance != user1NewBalance {
		t.Errorf("User1 balance = %d, want %d", updatedUser1.Balance, user1NewBalance)
	}

	if updatedUser2.Balance != user2NewBalance {
		t.Errorf("User2 balance = %d, want %d", updatedUser2.Balance, user2NewBalance)
	}

	// User1 logs out
	if err := authService.LogoutUser(session1); err != nil {
		t.Fatalf("LogoutUser(user1) error = %v", err)
	}

	// User2's session should still be valid
	_, err = authService.ValidateSession(session2)
	if err != nil {
		t.Error("User2 session should still be valid after user1 logout")
	}

	// User1's session should be invalid
	_, err = authService.ValidateSession(session1)
	if err == nil {
		t.Error("User1 session should be invalid after logout")
	}
}

// TestBlackjackPayoutIntegration verifies blackjack payout is calculated correctly in full flow
func TestBlackjackPayoutIntegration(t *testing.T) {
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	user, err := authService.RegisterUser("bjplayer", "pass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	// Create a blackjack scenario
	g := game.NewGameWithDeck([]game.Card{
		{Rank: "A", Suit: "♠", Value: 11}, // Player
		{Rank: "5", Suit: "♥", Value: 5},  // Dealer
		{Rank: "K", Suit: "♣", Value: 10}, // Player -> Blackjack!
		{Rank: "6", Suit: "♦", Value: 6},  // Dealer
	})

	betAmount := int64(10000) // $100
	if err := g.PlaceBetNoShuffle(betAmount); err != nil {
		t.Fatalf("PlaceBet() error = %v", err)
	}

	// Game should automatically resolve with player blackjack
	if g.Phase != game.PhaseGameOver {
		t.Errorf("Phase = %v, want GAME_OVER", g.Phase)
	}

	if g.Result != game.ResultPlayerBlackjack {
		t.Errorf("Result = %v, want PLAYER_BLACKJACK", g.Result)
	}

	payout := g.CalculatePayout()
	expectedPayout := betAmount + (betAmount * 3 / 2) // Bet + 3:2 = $100 + $150 = $250
	if payout != expectedPayout {
		t.Errorf("Payout = %d, want %d", payout, expectedPayout)
	}

	// Update balance
	newBalance := user.Balance - betAmount + payout
	if err := authService.UpdateBalance(user.ID, newBalance); err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	// Verify balance reflects blackjack payout
	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	expectedBalance := int64(1015000) // Started with $10000, won $150 profit
	if updatedUser.Balance != expectedBalance {
		t.Errorf("Balance = %d, want %d", updatedUser.Balance, expectedBalance)
	}
}

// TestDoubleDownIntegration tests double down with balance management
func TestDoubleDownIntegration(t *testing.T) {
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	user, err := authService.RegisterUser("doubler", "pass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	// Create a good doubling scenario: player has 11, dealer shows 6
	g := game.NewGameWithDeck([]game.Card{
		{Rank: "6", Suit: "♠", Value: 6},   // Player
		{Rank: "6", Suit: "♥", Value: 6},   // Dealer
		{Rank: "5", Suit: "♣", Value: 5},   // Player -> 11
		{Rank: "5", Suit: "♦", Value: 5},   // Dealer -> 11
		{Rank: "10", Suit: "♠", Value: 10}, // Player doubles, gets 21
		{Rank: "K", Suit: "♥", Value: 10},  // Dealer hits to 21 (push)
	})

	betAmount := int64(2000) // $20
	if err := g.PlaceBetNoShuffle(betAmount); err != nil {
		t.Fatalf("PlaceBet() error = %v", err)
	}

	initialBalance := user.Balance - betAmount

	// Player doubles down
	if err := g.DoubleDown(); err != nil {
		t.Fatalf("DoubleDown() error = %v", err)
	}

	// Bet should be doubled
	if g.Bet != betAmount*2 {
		t.Errorf("Bet = %d, want %d", g.Bet, betAmount*2)
	}

	// Game should be over after double down
	if g.Phase != game.PhaseGameOver {
		t.Errorf("Phase = %v, want GAME_OVER", g.Phase)
	}

	// This is a push (both 21)
	if g.Result != game.ResultPush {
		t.Errorf("Result = %v, want PUSH", g.Result)
	}

	payout := g.CalculatePayout()

	// Verify payout is correct for a push with doubled bet
	if payout != betAmount*2 {
		t.Errorf("Payout = %d, want %d (doubled bet returned)", payout, betAmount*2)
	}

	// Update balance: deduct additional bet, add payout
	newBalance := initialBalance - betAmount + payout
	if err := authService.UpdateBalance(user.ID, newBalance); err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	// On a push with double down, balance should be back to original
	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if updatedUser.Balance != 1000000 {
		t.Errorf("Balance = %d, want 1000000 (original)", updatedUser.Balance)
	}
}

func TestSurrenderIntegration(t *testing.T) {
	// Setup test database
	db, err := vault.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	authService := security.NewAuthService(db)

	// Register and login
	user, err := authService.RegisterUser("surrenderplayer", "pass123")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	initialBalance := user.Balance // 1000000 cents = $10000

	// Play game 1: Surrender with bet of 100000 cents ($1000)
	// Should lose half: 50000 cents ($500)
	g1 := game.NewGameWithDeck([]game.Card{
		{Rank: "5", Suit: "♠", Value: 5},   // Player card 1
		{Rank: "10", Suit: "♥", Value: 10}, // Dealer card 1
		{Rank: "10", Suit: "♠", Value: 10}, // Player card 2 (15 total - bad hand)
		{Rank: "K", Suit: "♥", Value: 10},  // Dealer card 2
	})

	if err := g1.PlaceBetNoShuffle(100000); err != nil {
		t.Fatalf("PlaceBet() error = %v", err)
	}

	// Surrender
	if err := g1.Surrender(); err != nil {
		t.Fatalf("Surrender() error = %v", err)
	}

	if g1.Result != game.ResultSurrender {
		t.Errorf("Expected ResultSurrender, got %s", g1.Result)
	}

	payout1 := g1.CalculatePayout()
	if payout1 != 50000 {
		t.Errorf("Expected payout of 50000 (half bet), got %d", payout1)
	}

	// Update balance (bet already deducted, add payout)
	newBalance1 := initialBalance - g1.Bet + payout1
	if err := authService.UpdateBalance(user.ID, newBalance1); err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	// Update stats
	stats1, _ := authService.GetUserStats(user.ID)
	stats1.GamesPlayed++
	stats1.GamesLost++ // Surrender counts as loss
	stats1.TotalBet += g1.Bet
	stats1.TotalWon += payout1
	lossAmount1 := g1.Bet / 2
	if lossAmount1 > stats1.BiggestLoss {
		stats1.BiggestLoss = lossAmount1
	}
	if err := db.UpdateUserStats(stats1); err != nil {
		t.Fatalf("UpdateUserStats() error = %v", err)
	}

	// Verify stats after game 1
	stats, err := authService.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if stats.GamesPlayed != 1 {
		t.Errorf("GamesPlayed = %d, want 1", stats.GamesPlayed)
	}

	if stats.GamesWon != 0 {
		t.Errorf("GamesWon = %d, want 0", stats.GamesWon)
	}

	if stats.GamesLost != 1 {
		t.Errorf("GamesLost = %d, want 1 (surrender counts as loss)", stats.GamesLost)
	}

	if stats.TotalBet != 100000 {
		t.Errorf("TotalBet = %d, want 100000", stats.TotalBet)
	}

	if stats.TotalWon != 50000 {
		t.Errorf("TotalWon = %d, want 50000 (half bet returned)", stats.TotalWon)
	}

	expectedNet := stats.TotalWon - stats.TotalBet // 50000 - 100000 = -50000
	if expectedNet != -50000 {
		t.Errorf("Net = %d, want -50000 (lost half bet)", expectedNet)
	}

	if stats.BiggestLoss != 50000 {
		t.Errorf("BiggestLoss = %d, want 50000 (half of bet)", stats.BiggestLoss)
	}

	// Verify balance
	user, _ = db.GetUserByID(user.ID)
	expectedBalance := initialBalance - 50000 // Lost half of 100000
	if user.Balance != expectedBalance {
		t.Errorf("Balance after surrender = %d, want %d", user.Balance, expectedBalance)
	}

	// Play game 2: Another surrender with different bet
	g2 := game.NewGameWithDeck([]game.Card{
		{Rank: "6", Suit: "♠", Value: 6},   // Player card 1
		{Rank: "A", Suit: "♥", Value: 11},  // Dealer card 1
		{Rank: "10", Suit: "♠", Value: 10}, // Player card 2 (16 total vs Ace)
		{Rank: "6", Suit: "♥", Value: 6},   // Dealer card 2 (soft 17)
	})

	if err := g2.PlaceBetNoShuffle(80000); err != nil { // Bet $800
		t.Fatalf("PlaceBet() error = %v", err)
	}

	if err := g2.Surrender(); err != nil {
		t.Fatalf("Surrender() error = %v", err)
	}

	payout2 := g2.CalculatePayout()
	if payout2 != 40000 {
		t.Errorf("Expected payout of 40000 (half of 80000), got %d", payout2)
	}

	// Update balance
	newBalance2 := user.Balance - g2.Bet + payout2
	if err := authService.UpdateBalance(user.ID, newBalance2); err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	// Update stats
	stats2, _ := authService.GetUserStats(user.ID)
	stats2.GamesPlayed++
	stats2.GamesLost++
	stats2.TotalBet += g2.Bet
	stats2.TotalWon += payout2
	lossAmount2 := g2.Bet / 2
	if lossAmount2 > stats2.BiggestLoss {
		stats2.BiggestLoss = lossAmount2
	}
	if err := db.UpdateUserStats(stats2); err != nil {
		t.Fatalf("UpdateUserStats() error = %v", err)
	}

	// Verify final stats
	finalStats, err := authService.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if finalStats.GamesPlayed != 2 {
		t.Errorf("Final GamesPlayed = %d, want 2", finalStats.GamesPlayed)
	}

	if finalStats.GamesLost != 2 {
		t.Errorf("Final GamesLost = %d, want 2", finalStats.GamesLost)
	}

	if finalStats.TotalBet != 180000 {
		t.Errorf("Final TotalBet = %d, want 180000", finalStats.TotalBet)
	}

	if finalStats.TotalWon != 90000 {
		t.Errorf("Final TotalWon = %d, want 90000 (50000 + 40000)", finalStats.TotalWon)
	}

	finalNet := finalStats.TotalWon - finalStats.TotalBet // 90000 - 180000 = -90000
	if finalNet != -90000 {
		t.Errorf("Final Net = %d, want -90000", finalNet)
	}

	if finalStats.BiggestLoss != 50000 {
		t.Errorf("Final BiggestLoss = %d, want 50000", finalStats.BiggestLoss)
	}

	// Verify final balance
	user, _ = db.GetUserByID(user.ID)
	finalExpectedBalance := initialBalance - 90000 // Lost 50000 + 40000
	if user.Balance != finalExpectedBalance {
		t.Errorf("Final balance = %d, want %d", user.Balance, finalExpectedBalance)
	}
}
