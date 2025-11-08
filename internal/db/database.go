package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &DB{DB: db}

	if err := database.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := database.Seed(); err != nil {
		return nil, fmt.Errorf("failed to seed database: %w", err)
	}

	return database, nil
}

func (db *DB) Migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			balance NUMERIC(10,2) NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			transaction_id TEXT UNIQUE,
			state TEXT,
			amount NUMERIC(10,2),
			source_type TEXT,
			applied BOOLEAN,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_transaction_id ON transactions(transaction_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (db *DB) Seed() error {
	// Insert users with ON CONFLICT to handle existing users gracefully
	queries := []string{
		`INSERT INTO users (id, balance) VALUES (1, 100.00) ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO users (id, balance) VALUES (2, 50.00) ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO users (id, balance) VALUES (3, 0.00) ON CONFLICT (id) DO NOTHING`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to seed users: %w", err)
		}
	}
	log.Println("Database seeded with initial users")

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

