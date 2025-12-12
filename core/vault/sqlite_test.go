package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestNewDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("NewDB() did not create database file")
	}
}

func TestCreateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	username := "testuser"
	password := "hashedpassword123"

	user, err := db.CreateUser(username, password)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if user.Username != username {
		t.Errorf("CreateUser() username = %v, want %v", user.Username, username)
	}

	if user.Password != password {
		t.Errorf("CreateUser() password = %v, want %v", user.Password, password)
	}

	if user.Balance != 1000000 { // Default $10000.00 in cents
		t.Errorf("CreateUser() balance = %v, want %v", user.Balance, 1000000)
	}

	if user.ID == 0 {
		t.Error("CreateUser() did not set user ID")
	}
}

func TestGetUserByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	username := "testuser"
	password := "hashedpassword123"
	createdUser, err := db.CreateUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	user, err := db.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("GetUserByUsername() ID = %v, want %v", user.ID, createdUser.ID)
	}

	_, err = db.GetUserByUsername("nonexistent")
	if err == nil {
		t.Error("GetUserByUsername() should return error for non-existent user")
	}
}

func TestCreateAndGetSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	sessionID := "test-session-uuid"
	expiresAt := time.Now().Add(24 * time.Hour)

	err = db.CreateSession(sessionID, user.ID, expiresAt)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	session, err := db.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("GetSession() ID = %v, want %v", session.ID, sessionID)
	}

	if session.UserID != user.ID {
		t.Errorf("GetSession() UserID = %v, want %v", session.UserID, user.ID)
	}
}

func TestGetExpiredSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	sessionID := "expired-session"
	expiresAt := time.Now().Add(-time.Hour) // Expired 1 hour ago

	err = db.CreateSession(sessionID, user.ID, expiresAt)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	_, err = db.GetSession(sessionID)
	if err == nil {
		t.Error("GetSession() should return error for expired session")
	}
}

func TestUpdateUserBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	newBalance := int64(50000) // $500.00
	err = db.UpdateUserBalance(user.ID, newBalance)
	if err != nil {
		t.Fatalf("UpdateUserBalance() error = %v", err)
	}

	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if updatedUser.Balance != newBalance {
		t.Errorf("UpdateUserBalance() balance = %v, want %v", updatedUser.Balance, newBalance)
	}
}

func TestUserStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	stats, err := db.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if stats.GamesPlayed != 0 {
		t.Errorf("Initial GamesPlayed = %v, want 0", stats.GamesPlayed)
	}

	if stats.GamesWon != 0 {
		t.Errorf("Initial GamesWon = %v, want 0", stats.GamesWon)
	}

	stats.GamesPlayed = 5
	stats.GamesWon = 3
	stats.TotalBet = 10000 // $100.00

	err = db.UpdateUserStats(stats)
	if err != nil {
		t.Fatalf("UpdateUserStats() error = %v", err)
	}

	updatedStats, err := db.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats() error = %v", err)
	}

	if updatedStats.GamesPlayed != 5 {
		t.Errorf("Updated GamesPlayed = %v, want 5", updatedStats.GamesPlayed)
	}

	if updatedStats.GamesWon != 3 {
		t.Errorf("Updated GamesWon = %v, want 3", updatedStats.GamesWon)
	}
}

func TestInitUserPreferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("prefsuser", "hashedpass123")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	prefs, err := db.GetUserPreferences(user.ID)
	if err != nil {
		t.Fatalf("GetUserPreferences() error = %v", err)
	}

	if prefs.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", prefs.UserID, user.ID)
	}

	if prefs.ShowStats != false {
		t.Errorf("ShowStats = %v, want false (default)", prefs.ShowStats)
	}
}

func TestGetUserPreferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("prefsuser2", "hashedpass123")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	prefs, err := db.GetUserPreferences(user.ID)
	if err != nil {
		t.Fatalf("GetUserPreferences() error = %v", err)
	}

	if prefs == nil {
		t.Fatal("Expected preferences, got nil")
	}

	if prefs.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", prefs.UserID, user.ID)
	}
}

func TestGetUserPreferencesNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.GetUserPreferences(99999)
	if err == nil {
		t.Error("Expected error for non-existent user preferences, got nil")
	}
}

func TestUpdateUserPreferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := db.CreateUser("prefsuser3", "hashedpass123")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	// Initially should be false
	prefs, err := db.GetUserPreferences(user.ID)
	if err != nil {
		t.Fatalf("GetUserPreferences() error = %v", err)
	}
	if prefs.ShowStats != false {
		t.Errorf("Initial ShowStats = %v, want false", prefs.ShowStats)
	}

	// Update to true
	err = db.UpdateUserPreferences(user.ID, true)
	if err != nil {
		t.Fatalf("UpdateUserPreferences() error = %v", err)
	}

	// Verify update
	prefs, err = db.GetUserPreferences(user.ID)
	if err != nil {
		t.Fatalf("GetUserPreferences() error = %v", err)
	}
	if prefs.ShowStats != true {
		t.Errorf("Updated ShowStats = %v, want true", prefs.ShowStats)
	}

	// Update back to false
	err = db.UpdateUserPreferences(user.ID, false)
	if err != nil {
		t.Fatalf("UpdateUserPreferences() error = %v", err)
	}

	// Verify update
	prefs, err = db.GetUserPreferences(user.ID)
	if err != nil {
		t.Fatalf("GetUserPreferences() error = %v", err)
	}
	if prefs.ShowStats != false {
		t.Errorf("Updated ShowStats = %v, want false", prefs.ShowStats)
	}
}

func TestUpdateUserPreferencesNonExistent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.UpdateUserPreferences(99999, true)
	if err != nil {
		// This is acceptable - updating non-existent preferences should fail or no-op
		// SQLite will just not update any rows
		return
	}

	// If no error, verify that no preferences were created
	_, err = db.GetUserPreferences(99999)
	if err == nil {
		t.Error("Should not have created preferences for non-existent user")
	}
}
