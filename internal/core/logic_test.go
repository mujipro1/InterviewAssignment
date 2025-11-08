package core

import (
	"database/sql"
	"testing"

	"assignment/internal/models"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := "host=localhost user=postgres password=postgres dbname=assignment_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}

	// Clean up and setup test database
	db.Exec("DROP TABLE IF EXISTS transactions CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	db.Exec(`CREATE TABLE users (
		id BIGSERIAL PRIMARY KEY,
		balance NUMERIC(10,2) NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	)`)
	db.Exec(`CREATE TABLE transactions (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT REFERENCES users(id),
		transaction_id TEXT UNIQUE,
		state TEXT,
		amount NUMERIC(10,2),
		source_type TEXT,
		applied BOOLEAN,
		created_at TIMESTAMP DEFAULT NOW()
	)`)
	db.Exec("INSERT INTO users (id, balance) VALUES (1, 100.00), (2, 50.00), (3, 0.00)")

	return db
}

func TestProcessTransaction_Win(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewTransactionService(db)

	req := models.TransactionRequest{
		State:         "win",
		Amount:        "10.50",
		TransactionID: "test-win-1",
	}

	resp, err := service.ProcessTransaction(1, req, "game")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Message != "Transaction applied successfully" {
		t.Errorf("Expected success message, got: %s", resp.Message)
	}

	if resp.Balance != "110.50" {
		t.Errorf("Expected balance 110.50, got: %s", resp.Balance)
	}
}

func TestProcessTransaction_Lose(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewTransactionService(db)

	req := models.TransactionRequest{
		State:         "lose",
		Amount:        "25.00",
		TransactionID: "test-lose-1",
	}

	resp, err := service.ProcessTransaction(1, req, "server")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Balance != "75.00" {
		t.Errorf("Expected balance 75.00, got: %s", resp.Balance)
	}
}

func TestProcessTransaction_InsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewTransactionService(db)

	req := models.TransactionRequest{
		State:         "lose",
		Amount:        "100.00",
		TransactionID: "test-insufficient-1",
	}

	resp, err := service.ProcessTransaction(3, req, "payment")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Message != "Insufficient funds" {
		t.Errorf("Expected 'Insufficient funds' message, got: %s", resp.Message)
	}

	if resp.Balance != "0.00" {
		t.Errorf("Expected balance 0.00, got: %s", resp.Balance)
	}
}

func TestProcessTransaction_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewTransactionService(db)

	req := models.TransactionRequest{
		State:         "win",
		Amount:        "10.00",
		TransactionID: "test-dup-1",
	}

	// First transaction
	resp1, err := service.ProcessTransaction(1, req, "game")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Duplicate transaction
	resp2, err := service.ProcessTransaction(1, req, "game")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp2.Message != "Duplicate transaction ignored" {
		t.Errorf("Expected duplicate message, got: %s", resp2.Message)
	}

	if resp1.Balance != resp2.Balance {
		t.Errorf("Duplicate transaction should return same balance")
	}
}

func TestGetBalance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewTransactionService(db)

	resp, err := service.GetBalance(1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Balance != "100.00" {
		t.Errorf("Expected balance 100.00, got: %s", resp.Balance)
	}
}

