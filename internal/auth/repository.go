package auth

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, passwordHash, name string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`INSERT INTO users (email, password_hash, name) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, email, name`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.Name)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(email string) (*User, string, error) {
	var user User
	var passwordHash string

	err := r.db.QueryRow(
		`SELECT id, email, name, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &passwordHash)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	return &user, passwordHash, nil
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	).Scan(&exists)

	return exists, err
}

