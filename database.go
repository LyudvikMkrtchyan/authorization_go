package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryCreateTable   = `CREATE TABLE IF NOT EXISTS users (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, login VARCHAR(255) NOT NULL UNIQUE, password_hash VARCHAR(255) NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`
	queryCreateUser    = `INSERT INTO users (login, password_hash) VALUES (?, ?)`
	queryGetByLogin    = `SELECT id, login, password_hash, created_at FROM users WHERE login = ?`
	queryUserExists    = `SELECT COUNT(*) FROM users WHERE login = ?`
)

type User struct {
	ID           uint64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Database interface {
	CreateUser(login string, passwordHash string) (userID uint64, err error)
	GetUserByLogin(login string) (*User, error)
	UserExists(login string) (bool, error)
	Close() error
}

type MySQLDatabase struct {
	db *sql.DB
}

func NewMySQLDatabase(cfg *Config) (*MySQLDatabase, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("NewMySQLDatabase: open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("NewMySQLDatabase: ping: %w", err)
	}

	if _, err := db.Exec(queryCreateTable); err != nil {
		return nil, fmt.Errorf("NewMySQLDatabase: create table: %w", err)
	}

	return &MySQLDatabase{db: db}, nil
}

func (m *MySQLDatabase) CreateUser(login string, passwordHash string) (uint64, error) {
	res, err := m.db.Exec(queryCreateUser, login, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("CreateUser: last insert id: %w", err)
	}
	return uint64(id), nil
}

func (m *MySQLDatabase) GetUserByLogin(login string) (*User, error) {
	row := m.db.QueryRow(queryGetByLogin, login)
	var u User
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetUserByLogin: user %q not found", login)
		}
		return nil, fmt.Errorf("GetUserByLogin: %w", err)
	}
	return &u, nil
}

func (m *MySQLDatabase) UserExists(login string) (bool, error) {
	var count int
	if err := m.db.QueryRow(queryUserExists, login).Scan(&count); err != nil {
		return false, fmt.Errorf("UserExists: %w", err)
	}
	return count > 0, nil
}

func (m *MySQLDatabase) Close() error {
	return m.db.Close()
}
