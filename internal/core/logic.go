package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"assignment/internal/models"
	"assignment/internal/utils"
)

type TransactionService struct {
	db *sql.DB
}

func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) ProcessTransaction(userID int64, req models.TransactionRequest, sourceType string) (*models.TransactionResponse, error) {
	// Validate inputs
	if err := utils.ValidateSourceType(sourceType); err != nil {
		return nil, err
	}
	if err := utils.ValidateState(req.State); err != nil {
		return nil, err
	}
	if err := utils.ValidateAmount(req.Amount); err != nil {
		return nil, err
	}

	amount, err := utils.ParseAmount(req.Amount)
	if err != nil {
		return nil, err
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if transaction already exists
	var existingTransaction models.Transaction
	var existingBalance string
	err = tx.QueryRow(
		`SELECT id, user_id, transaction_id, state, amount, source_type, applied, created_at 
		 FROM transactions WHERE transaction_id = $1`,
		req.TransactionID,
	).Scan(
		&existingTransaction.ID,
		&existingTransaction.UserID,
		&existingTransaction.TransactionID,
		&existingTransaction.State,
		&existingTransaction.Amount,
		&existingTransaction.SourceType,
		&existingTransaction.Applied,
		&existingTransaction.CreatedAt,
	)

	if err == nil {
		// Transaction already exists - return duplicate response
		err = tx.QueryRow(
			`SELECT balance FROM users WHERE id = $1 FOR UPDATE`,
			existingTransaction.UserID,
		).Scan(&existingBalance)
		if err != nil {
			return nil, fmt.Errorf("failed to get user balance: %w", err)
		}

		tx.Commit()
		return &models.TransactionResponse{
			UserID:        existingTransaction.UserID,
			TransactionID: existingTransaction.TransactionID,
			Balance:       existingBalance,
			Message:       "Duplicate transaction ignored",
		}, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing transaction: %w", err)
	}

	// Get user with lock for update
	var currentBalance string
	err = tx.QueryRow(
		`SELECT balance FROM users WHERE id = $1 FOR UPDATE`,
		userID,
	).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Parse current balance
	currentBalanceFloat, err := utils.ParseAmount(currentBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current balance: %w", err)
	}

	// Calculate new balance
	var newBalance float64
	if req.State == "win" {
		newBalance = currentBalanceFloat + amount
	} else {
		newBalance = currentBalanceFloat - amount
	}

	// Check if balance would go negative
	if newBalance < 0 {
		tx.Commit()
		return &models.TransactionResponse{
			UserID:        userID,
			TransactionID: req.TransactionID,
			Balance:       utils.FormatBalance(currentBalanceFloat),
			Message:       "Insufficient funds",
		}, nil
	}

	// Update user balance
	newBalanceStr := utils.FormatBalance(newBalance)
	_, err = tx.Exec(
		`UPDATE users SET balance = $1, updated_at = NOW() WHERE id = $2`,
		newBalanceStr,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	// Insert transaction record
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, transaction_id, state, amount, source_type, applied) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID,
		req.TransactionID,
		req.State,
		req.Amount,
		sourceType,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Transaction processed: userID=%d, transactionID=%s, state=%s, amount=%s, newBalance=%s",
		userID, req.TransactionID, req.State, req.Amount, newBalanceStr)

	return &models.TransactionResponse{
		UserID:        userID,
		TransactionID: req.TransactionID,
		Balance:       newBalanceStr,
		Message:       "Transaction applied successfully",
	}, nil
}

func (s *TransactionService) GetBalance(userID int64) (*models.BalanceResponse, error) {
	var balance string
	err := s.db.QueryRow(
		`SELECT balance FROM users WHERE id = $1`,
		userID,
	).Scan(&balance)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return &models.BalanceResponse{
		UserID:  userID,
		Balance: balance,
	}, nil
}

