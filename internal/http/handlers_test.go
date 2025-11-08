package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"assignment/internal/core"
	"assignment/internal/models"
	_ "github.com/lib/pq"
)

func setupTestHandlers(t *testing.T) (*Handlers, *sql.DB) {
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

	service := core.NewTransactionService(db)
	handlers := NewHandlers(service)

	return handlers, db
}

func TestHandleTransaction_Success(t *testing.T) {
	handlers, db := setupTestHandlers(t)
	defer db.Close()

	reqBody := models.TransactionRequest{
		State:         "win",
		Amount:        "10.50",
		TransactionID: "test-api-1",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	handlers.HandleTransaction(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var resp models.TransactionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Message != "Transaction applied successfully" {
		t.Errorf("Expected success message, got: %s", resp.Message)
	}
}

func TestHandleTransaction_InvalidSourceType(t *testing.T) {
	handlers, db := setupTestHandlers(t)
	defer db.Close()

	reqBody := models.TransactionRequest{
		State:         "win",
		Amount:        "10.50",
		TransactionID: "test-api-2",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "invalid")

	w := httptest.NewRecorder()
	handlers.HandleTransaction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", w.Code)
	}
}

func TestHandleGetBalance_Success(t *testing.T) {
	handlers, db := setupTestHandlers(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/user/1/balance", nil)

	w := httptest.NewRecorder()
	handlers.HandleGetBalance(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var resp models.BalanceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Balance != "100.00" {
		t.Errorf("Expected balance 100.00, got: %s", resp.Balance)
	}
}

func TestHandleGetBalance_UserNotFound(t *testing.T) {
	handlers, db := setupTestHandlers(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/user/999/balance", nil)

	w := httptest.NewRecorder()
	handlers.HandleGetBalance(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

