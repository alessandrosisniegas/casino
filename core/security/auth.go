package security

import (
	"fmt"

	"github.com/alessandrosisniegas/casino/core/vault"
)

type AuthService struct {
	db *vault.DB
}

func NewAuthService(db *vault.DB) *AuthService {
	return &AuthService{db: db}
}

func (as *AuthService) RegisterUser(username, password string) (*vault.User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	if err := ValidatePasswordWithUsername(password, username); err != nil {
		return nil, err
	}

	if _, err := as.db.GetUserByUsername(username); err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := as.db.CreateUser(username, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (as *AuthService) LoginUser(username, password string) (string, *vault.User, error) {
	user, err := as.db.GetUserByUsername(username)
	if err != nil {
		return "", nil, fmt.Errorf("invalid username or password")
	}

	if err := VerifyPassword(password, user.Password); err != nil {
		return "", nil, fmt.Errorf("invalid username or password")
	}

	sessionID := GenerateSessionID()
	expiresAt := GetSessionExpiry()

	if err := as.db.CreateSession(sessionID, user.ID, expiresAt); err != nil {
		return "", nil, fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, user, nil
}

func (as *AuthService) ValidateSession(sessionID string) (*vault.User, error) {
	session, err := as.db.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	user, err := as.db.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (as *AuthService) LogoutUser(sessionID string) error {
	return as.db.DeleteSession(sessionID)
}

func (as *AuthService) GetUserStats(userID int) (*vault.UserStats, error) {
	return as.db.GetUserStats(userID)
}

func (as *AuthService) UpdateBalance(userID int, newBalance int64) error {
	return as.db.UpdateUserBalance(userID, newBalance)
}
