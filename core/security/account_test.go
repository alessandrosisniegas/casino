package security

import (
	"strings"
	"testing"
	"time"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "testuser123", false},
		{"valid with underscore", "test_user_123", false},
		{"valid short", "ab", false},
		{"too short", "a", true},
		{"too long", strings.Repeat("a", 31), true},
		{"invalid chars", "test@user", true},
		{"empty", "", true},
		{"spaces", "test user", true},
		{"tab character", "test\tuser", true},
		{"newline character", "test\nuser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "password123", false},
		{"valid with underscore", "my_password_123", false},
		{"valid short", "ab", false},
		{"too short", "a", true},
		{"too long", strings.Repeat("a", 31), true},
		{"invalid chars", "pass@word", true},
		{"with spaces", "pass word", true},
		{"with tab", "pass\tword", true},
		{"with newline", "pass\nword", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordWithUsername(t *testing.T) {
	tests := []struct {
		name     string
		password string
		username string
		wantErr  bool
	}{
		{"valid different", "password123", "username456", false},
		{"password same as username", "testuser", "testuser", true},
		{"invalid password chars", "test@pass", "username", true},
		{"password too short", "a", "username", true},
		{"password with space", "pass word", "username", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordWithUsername(tt.password, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordWithUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() returned plaintext password")
	}

	if !strings.HasPrefix(hash, "$2a$") {
		t.Error("HashPassword() doesn't appear to be bcrypt hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("VerifyPassword() failed for correct password: %v", err)
	}

	err = VerifyPassword("wrongpassword123", hash)
	if err == nil {
		t.Error("VerifyPassword() should fail for incorrect password")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := GenerateSessionID()
	id2 := GenerateSessionID()

	if id1 == "" {
		t.Error("GenerateSessionID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateSessionID() returned duplicate IDs")
	}

	if len(id1) != 36 {
		t.Errorf("GenerateSessionID() returned wrong length: got %d, want 36", len(id1))
	}
}

func TestGetSessionExpiry(t *testing.T) {
	now := time.Now()
	expiry := GetSessionExpiry()

	if expiry.Before(now) {
		t.Error("GetSessionExpiry() returned past time")
	}

	expectedExpiry := now.Add(SessionDuration)
	diff := expiry.Sub(expectedExpiry)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("GetSessionExpiry() time difference too large: %v", diff)
	}
}

func TestIsSessionExpired(t *testing.T) {
	now := time.Now()

	expired := now.Add(-time.Hour)
	if !IsSessionExpired(expired) {
		t.Error("IsSessionExpired() should return true for past time")
	}

	future := now.Add(time.Hour)
	if IsSessionExpired(future) {
		t.Error("IsSessionExpired() should return false for future time")
	}
}
