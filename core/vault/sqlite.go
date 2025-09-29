package vault

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Balance   int64     `json:"balance"` // Balance in cents
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"` // UUID
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserStats struct {
	UserID      int   `json:"user_id"`
	GamesPlayed int64 `json:"games_played"`
	GamesWon    int64 `json:"games_won"`
	GamesLost   int64 `json:"games_lost"`
	TotalBet    int64 `json:"total_bet"`
	TotalWon    int64 `json:"total_won"`
	BiggestWin  int64 `json:"biggest_win"`
	BiggestLoss int64 `json:"biggest_loss"`
}

type DB struct {
	conn *sql.DB
}

func NewDB(filepath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			balance INTEGER NOT NULL DEFAULT 100000,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id INTEGER PRIMARY KEY,
			games_played INTEGER DEFAULT 0,
			games_won INTEGER DEFAULT 0,
			games_lost INTEGER DEFAULT 0,
			total_bet INTEGER DEFAULT 0,
			total_won INTEGER DEFAULT 0,
			biggest_win INTEGER DEFAULT 0,
			biggest_loss INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	return nil
}

func (db *DB) CreateUser(username, hashedPassword string) (*User, error) {
	query := `INSERT INTO users (username, password) VALUES (?, ?)`
	result, err := db.conn.Exec(query, username, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	if err := db.initUserStats(int(id)); err != nil {
		return nil, fmt.Errorf("failed to initialize user stats: %w", err)
	}

	return db.GetUserByID(int(id))
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password, balance, created_at, updated_at FROM users WHERE username = ?`
	row := db.conn.QueryRow(query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (db *DB) GetUserByID(id int) (*User, error) {
	query := `SELECT id, username, password, balance, created_at, updated_at FROM users WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (db *DB) UpdateUserBalance(userID int, newBalance int64) error {
	query := `UPDATE users SET balance = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.conn.Exec(query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}
	return nil
}

func (db *DB) CreateSession(sessionID string, userID int, expiresAt time.Time) error {
	query := `INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, sessionID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (db *DB) GetSession(sessionID string) (*Session, error) {
	query := `SELECT id, user_id, created_at, expires_at FROM sessions WHERE id = ? AND expires_at > CURRENT_TIMESTAMP`
	row := db.conn.QueryRow(query, sessionID)

	var session Session
	err := row.Scan(&session.ID, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (db *DB) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := db.conn.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (db *DB) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

func (db *DB) initUserStats(userID int) error {
	query := `INSERT INTO user_stats (user_id) VALUES (?)`
	_, err := db.conn.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to initialize user stats: %w", err)
	}
	return nil
}

func (db *DB) GetUserStats(userID int) (*UserStats, error) {
	query := `SELECT user_id, games_played, games_won, games_lost, total_bet, total_won, biggest_win, biggest_loss 
			  FROM user_stats WHERE user_id = ?`
	row := db.conn.QueryRow(query, userID)

	var stats UserStats
	err := row.Scan(&stats.UserID, &stats.GamesPlayed, &stats.GamesWon, &stats.GamesLost,
		&stats.TotalBet, &stats.TotalWon, &stats.BiggestWin, &stats.BiggestLoss)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user stats not found")
		}
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

func (db *DB) UpdateUserStats(stats *UserStats) error {
	query := `UPDATE user_stats SET 
			  games_played = ?, games_won = ?, games_lost = ?, 
			  total_bet = ?, total_won = ?, biggest_win = ?, biggest_loss = ?
			  WHERE user_id = ?`
	_, err := db.conn.Exec(query, stats.GamesPlayed, stats.GamesWon, stats.GamesLost,
		stats.TotalBet, stats.TotalWon, stats.BiggestWin, stats.BiggestLoss, stats.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}
	return nil
}
