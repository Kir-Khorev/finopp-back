package common

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/Kir-Khorev/finopp-back/pkg/config"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	// For production (Neon) - use SSL
	sslMode := "disable"
	if cfg.Environment == "production" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("✅ Database connected")
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	// Users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Profiles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS profiles (
			id SERIAL PRIMARY KEY,
			user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			monthly_income DECIMAL(10,2) DEFAULT 0,
			monthly_expenses DECIMAL(10,2) DEFAULT 0,
			savings_goal DECIMAL(10,2) DEFAULT 0,
			debt_amount DECIMAL(10,2) DEFAULT 0,
			currency VARCHAR(3) DEFAULT 'USD',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create profiles table: %w", err)
	}

	// Advice sessions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS advice_sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255),
			context_snapshot JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create advice_sessions table: %w", err)
	}

	// Advice messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS advice_messages (
			id SERIAL PRIMARY KEY,
			session_id INTEGER REFERENCES advice_sessions(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create advice_messages table: %w", err)
	}

	log.Println("✅ Migrations completed")
	return nil
}

