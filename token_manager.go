package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	queryCreateSessionsTable = `CREATE TABLE IF NOT EXISTS sessions (` +
		`id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,` +
		`user_id BIGINT UNSIGNED NOT NULL,` +
		`jwt_token TEXT NOT NULL,` +
		`refresh_token VARCHAR(512) NOT NULL UNIQUE,` +
		`created_at DATETIME DEFAULT CURRENT_TIMESTAMP,` +
		`expires_at DATETIME NOT NULL)`

	queryAddSession          = `INSERT INTO sessions (user_id, jwt_token, refresh_token, expires_at) VALUES (?, ?, ?, ?)`
	queryGetSessionByRefresh = `SELECT id, user_id, jwt_token, refresh_token, created_at, expires_at FROM sessions WHERE refresh_token = ? AND expires_at > NOW()`
	queryDeleteSession       = `DELETE FROM sessions WHERE refresh_token = ?`
	queryCleanExpiredByUser  = `DELETE FROM sessions WHERE user_id = ? AND expires_at <= NOW()`
	queryCountSessionsByUser = `SELECT COUNT(*) FROM sessions WHERE user_id = ?`
)

type Session struct {
	ID           uint64
	UserID       uint64
	JWTToken     string
	RefreshToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type TokenManager interface {
	AddSession(userID uint64, jwtToken string, refreshToken string, expiresAt time.Time) error
	GetSessionByRefresh(refreshToken string) (*Session, error)
	DeleteSession(refreshToken string) error
	CleanExpiredByUser(userID uint64) error
	CountSessionsByUser(userID uint64) (int, error)
}

type MySQLTokenManager struct {
	db          *sql.DB
	maxSessions int
	logger      WarningLogger
}

func NewMySQLTokenManager(cfg *Config, db *sql.DB, logger WarningLogger) (*MySQLTokenManager, error) {
	if _, err := db.Exec(queryCreateSessionsTable); err != nil {
		return nil, fmt.Errorf("NewMySQLTokenManager: create table: %w", err)
	}
	return &MySQLTokenManager{
		db:          db,
		maxSessions: cfg.TokenMapMax,
		logger:      logger,
	}, nil
}

func (m *MySQLTokenManager) AddSession(userID uint64, jwtToken string, refreshToken string, expiresAt time.Time) error {
	count, err := m.CountSessionsByUser(userID)
	if err != nil {
		return fmt.Errorf("AddSession: %w", err)
	}
	if count >= m.maxSessions {
		_ = m.logger.NoSpace(fmt.Sprintf("user %d has %d sessions, limit is %d — cleaning expired", userID, count, m.maxSessions))
		if err := m.CleanExpiredByUser(userID); err != nil {
			return fmt.Errorf("AddSession: %w", err)
		}
	}
	if _, err := m.db.Exec(queryAddSession, userID, jwtToken, refreshToken, expiresAt); err != nil {
		return fmt.Errorf("AddSession: %w", err)
	}
	return nil
}

func (m *MySQLTokenManager) GetSessionByRefresh(refreshToken string) (*Session, error) {
	row := m.db.QueryRow(queryGetSessionByRefresh, refreshToken)
	var s Session
	if err := row.Scan(&s.ID, &s.UserID, &s.JWTToken, &s.RefreshToken, &s.CreatedAt, &s.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetSessionByRefresh: session not found or expired")
		}
		return nil, fmt.Errorf("GetSessionByRefresh: %w", err)
	}
	return &s, nil
}

func (m *MySQLTokenManager) DeleteSession(refreshToken string) error {
	if _, err := m.db.Exec(queryDeleteSession, refreshToken); err != nil {
		return fmt.Errorf("DeleteSession: %w", err)
	}
	return nil
}

func (m *MySQLTokenManager) CleanExpiredByUser(userID uint64) error {
	if _, err := m.db.Exec(queryCleanExpiredByUser, userID); err != nil {
		return fmt.Errorf("CleanExpiredByUser: %w", err)
	}
	return nil
}

func (m *MySQLTokenManager) CountSessionsByUser(userID uint64) (int, error) {
	var count int
	if err := m.db.QueryRow(queryCountSessionsByUser, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("CountSessionsByUser: %w", err)
	}
	return count, nil
}

type jwtClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uint64, secret string, ttlSeconds int) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlSeconds) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("GenerateJWT: %w", err)
	}
	return signed, nil
}

func GenerateRefreshToken() (string, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("GenerateRefreshToken: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
