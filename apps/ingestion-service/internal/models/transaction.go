package models

import "time"

// Transaction represents a financial transaction that will be ingested.
// This is the core data structure we’re moving through the pipeline.
type Transaction struct {
	ID        string            `json:"id"`                 // unique identifier for the transaction
	UserID    string            `json:"user_id"`            // the user who initiated it
	Amount    float64           `json:"amount"`             // how much money is involved
	Currency  string            `json:"currency"`           // currency code (e.g., USD, INR)
	Type      string            `json:"type"`               // transaction type (deposit, withdrawal, transfer, etc.)
	Timestamp time.Time         `json:"timestamp"`          // when the transaction happened
	Metadata  map[string]string `json:"metadata,omitempty"` // optional extra info (tags, source, notes)
}
