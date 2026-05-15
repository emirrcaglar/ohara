package db

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	IsApproved   bool
}

func (db *DB) EnsureAdmin(username, password string) error {
	if username == "" || password == "" {
		return nil // Skip if not provided
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO user (username, password_hash, role, is_approved)
		VALUES (?, ?, 'admin', 1)
		ON CONFLICT(username) DO UPDATE SET
			password_hash = excluded.password_hash,
			role = 'admin',
			is_approved = 1
	`, username, string(hash))

	if err != nil {
		return fmt.Errorf("failed to ensure admin user: %w", err)
	}

	return nil
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, password_hash, role, is_approved
		FROM user WHERE username = ?
	`, username)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.IsApproved)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
