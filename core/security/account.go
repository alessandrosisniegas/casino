package security

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinUsernameAndPasswordLength = 2
	MaxUsernameAndPasswordLength = 30
	SessionDuration              = 24 * time.Hour
)

func ValidateUsername(username string) error {
	if len(username) < MinUsernameAndPasswordLength {
		return fmt.Errorf("username must be at least %d characters long", MinUsernameAndPasswordLength)
	}
	if len(username) > MaxUsernameAndPasswordLength {
		return fmt.Errorf("username must be no more than %d characters long", MaxUsernameAndPasswordLength)
	}

	if strings.Contains(username, " ") || strings.Contains(username, "\t") || strings.Contains(username, "\n") {
		return fmt.Errorf("username cannot contain whitespace")
	}

	matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if err != nil {
		return fmt.Errorf("error validating username: %w", err)
	}
	if !matched {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < MinUsernameAndPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinUsernameAndPasswordLength)
	}
	if len(password) > MaxUsernameAndPasswordLength {
		return fmt.Errorf("password must be no more than %d characters long", MaxUsernameAndPasswordLength)
	}

	if strings.Contains(password, " ") || strings.Contains(password, "\t") || strings.Contains(password, "\n") {
		return fmt.Errorf("password cannot contain whitespace")
	}

	matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", password)
	if err != nil {
		return fmt.Errorf("error validating password: %w", err)
	}
	if !matched {
		return fmt.Errorf("password can only contain letters, numbers, and underscores")
	}

	return nil
}

func ValidatePasswordWithUsername(password, username string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	if password == username {
		return fmt.Errorf("password cannot be the same as username")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func GenerateSessionID() string {
	return uuid.New().String()
}

func GetSessionExpiry() time.Time {
	return time.Now().Add(SessionDuration)
}

func IsSessionExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}
