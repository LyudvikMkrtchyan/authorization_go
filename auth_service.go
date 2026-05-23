package main

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type LoginResult struct {
	UserID       uint64 `json:"user_id"`
	JWTToken     string `json:"jwt_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(login string, password string) error
	Login(login string, password string) (*LoginResult, error)
	RefreshToken(refreshToken string) (*LoginResult, error)
}

type AuthServiceImpl struct {
	db           Database
	tokenManager TokenManager
	cfg          *Config
}

func NewAuthService(db Database, tokenManager TokenManager, cfg *Config) *AuthServiceImpl {
	return &AuthServiceImpl{db: db, tokenManager: tokenManager, cfg: cfg}
}

func (s *AuthServiceImpl) Register(login string, password string) error {
	exists, err := s.db.UserExists(login)
	if err != nil {
		return fmt.Errorf("Register: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Register: hash password: %w", err)
	}

	if _, err := s.db.CreateUser(login, string(hash)); err != nil {
		return fmt.Errorf("Register: %w", err)
	}
	return nil
}

func (s *AuthServiceImpl) Login(login string, password string) (*LoginResult, error) {
	user, err := s.db.GetUserByLogin(login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	jwtToken, err := GenerateJWT(user.ID, s.cfg.JWTSecret, s.cfg.JWTTTLSeconds)
	if err != nil {
		return nil, fmt.Errorf("Login: %w", err)
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Login: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.RefreshTTLSeconds) * time.Second)
	if err := s.tokenManager.AddSession(user.ID, jwtToken, refreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("Login: %w", err)
	}

	return &LoginResult{UserID: user.ID, JWTToken: jwtToken, RefreshToken: refreshToken}, nil
}

func (s *AuthServiceImpl) RefreshToken(refreshToken string) (*LoginResult, error) {
	session, err := s.tokenManager.GetSessionByRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := s.tokenManager.DeleteSession(refreshToken); err != nil {
		return nil, fmt.Errorf("RefreshToken: delete old session: %w", err)
	}

	newJWT, err := GenerateJWT(session.UserID, s.cfg.JWTSecret, s.cfg.JWTTTLSeconds)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: %w", err)
	}

	newRefresh, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.RefreshTTLSeconds) * time.Second)
	if err := s.tokenManager.AddSession(session.UserID, newJWT, newRefresh, expiresAt); err != nil {
		return nil, fmt.Errorf("RefreshToken: %w", err)
	}

	return &LoginResult{UserID: session.UserID, JWTToken: newJWT, RefreshToken: newRefresh}, nil
}
