package security

import (
	"path/filepath"
	"testing"

	"github.com/alessandrosisniegas/casino/core/vault"
)

func setupTestAuthService(t *testing.T) (*AuthService, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := vault.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	authService := NewAuthService(db)

	cleanup := func() {
		db.Close()
	}

	return authService, cleanup
}

func TestRegisterUser(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	user, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}

	if user.Username != username {
		t.Errorf("RegisterUser() username = %v, want %v", user.Username, username)
	}

	if user.Balance != 100000 { // $1000.00 in cents
		t.Errorf("RegisterUser() balance = %v, want 100000", user.Balance)
	}

	if user.Password == password {
		t.Error("RegisterUser() stored plaintext password")
	}
}

func TestRegisterUserInvalidUsername(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	invalidUsernames := []string{"a", "test@user", "", "test user"}

	for _, username := range invalidUsernames {
		_, err := auth.RegisterUser(username, "validpassword123")
		if err == nil {
			t.Errorf("RegisterUser() should fail for invalid username: %v", username)
		}
	}
}

func TestRegisterUserInvalidPassword(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	invalidPasswords := []string{"a", "pass@word", "pass word", "validuser"}

	for _, password := range invalidPasswords {
		_, err := auth.RegisterUser("validuser", password)
		if err == nil {
			t.Errorf("RegisterUser() should fail for invalid password: %v", password)
		}
	}
}

func TestRegisterDuplicateUser(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	_, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("First RegisterUser() error = %v", err)
	}

	_, err = auth.RegisterUser(username, password)
	if err == nil {
		t.Error("RegisterUser() should fail for duplicate username")
	}
}

func TestLoginUser(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	_, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	sessionID, user, err := auth.LoginUser(username, password)
	if err != nil {
		t.Fatalf("LoginUser() error = %v", err)
	}

	if sessionID == "" {
		t.Error("LoginUser() returned empty session ID")
	}

	if user.Username != username {
		t.Errorf("LoginUser() username = %v, want %v", user.Username, username)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	_, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	_, _, err = auth.LoginUser(username, "wrongpassword123")
	if err == nil {
		t.Error("LoginUser() should fail for wrong password")
	}

	_, _, err = auth.LoginUser("wronguser", password)
	if err == nil {
		t.Error("LoginUser() should fail for wrong username")
	}
}

func TestValidateSession(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	_, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	sessionID, originalUser, err := auth.LoginUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	user, err := auth.ValidateSession(sessionID)
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}

	if user.ID != originalUser.ID {
		t.Errorf("ValidateSession() user ID = %v, want %v", user.ID, originalUser.ID)
	}
}

func TestValidateInvalidSession(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	_, err := auth.ValidateSession("invalid-session-id")
	if err == nil {
		t.Error("ValidateSession() should fail for invalid session")
	}
}

func TestLogoutUser(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	_, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	sessionID, _, err := auth.LoginUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = auth.LogoutUser(sessionID)
	if err != nil {
		t.Fatalf("LogoutUser() error = %v", err)
	}

	_, err = auth.ValidateSession(sessionID)
	if err == nil {
		t.Error("Session should be invalid after logout")
	}
}

func TestUpdateBalance(t *testing.T) {
	auth, cleanup := setupTestAuthService(t)
	defer cleanup()

	username := "testuser123"
	password := "testpassword456"

	user, err := auth.RegisterUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	newBalance := int64(50000) // $500.00
	err = auth.UpdateBalance(user.ID, newBalance)
	if err != nil {
		t.Fatalf("UpdateBalance() error = %v", err)
	}

	sessionID, _, err := auth.LoginUser(username, password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	updatedUser, err := auth.ValidateSession(sessionID)
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}

	if updatedUser.Balance != newBalance {
		t.Errorf("UpdateBalance() balance = %v, want %v", updatedUser.Balance, newBalance)
	}
}
