package models

import "time"

// Transaction represents a financial transaction that will be ingested.
// This is the core data structure we're moving through the pipeline.
type Transaction struct {
	ID             string            `json:"id"`                  // unique identifier for the transaction
	IdempotencyKey string            `json:"idempotency_key"`     // idempotency key for deduplication
	AccountID      string            `json:"account_id"`          // account identifier for Kafka partitioning
	UserID         string            `json:"user_id"`             // the user who initiated it
	Amount         float64           `json:"amount"`              // how much money is involved
	Currency       string            `json:"currency"`            // currency code (e.g., USD, INR)
	Type           string            `json:"type"`                // transaction type (deposit, withdrawal, transfer, etc.)
	Category       string            `json:"category"`            // transaction category (e.g., "groceries", "utilities")
	Merchant       string            `json:"merchant,omitempty"`  // merchant name for card transactions
	Reference      string            `json:"reference,omitempty"` // external reference number
	Status         string            `json:"status"`              // transaction status (pending, completed, failed)
	Timestamp      time.Time         `json:"timestamp"`           // when the transaction happened
	Metadata       map[string]string `json:"metadata,omitempty"`  // optional extra info (tags, source, notes)
}

// TransactionRequest represents the incoming HTTP request
type TransactionRequest struct {
	IdempotencyKey string            `json:"idempotency_key" binding:"required"`
	AccountID      string            `json:"account_id" binding:"required"`
	UserID         string            `json:"user_id" binding:"required"`
	Amount         float64           `json:"amount" binding:"required,gt=0"`
	Currency       string            `json:"currency" binding:"required"`
	Type           string            `json:"type" binding:"required"`
	Category       string            `json:"category" binding:"required"`
	Merchant       string            `json:"merchant,omitempty"`
	Reference      string            `json:"reference,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// TransactionResponse represents the API response
type TransactionResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
