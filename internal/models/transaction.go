package models

import "time"

type Transaction struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	State         string    `json:"state"`
	Amount        string    `json:"amount"`
	SourceType    string    `json:"source_type"`
	Applied       bool      `json:"applied"`
	CreatedAt     time.Time `json:"created_at"`
}

type TransactionRequest struct {
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}

type TransactionResponse struct {
	UserID        int64  `json:"userId"`
	TransactionID string `json:"transactionId"`
	Balance       string `json:"balance"`
	Message       string `json:"message"`
}

type BalanceResponse struct {
	UserID  int64  `json:"userId"`
	Balance string `json:"balance"`
}

