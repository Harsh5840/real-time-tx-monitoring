package models

import "time"

// Transaction mirrors the ingestion/processing model
type Transaction struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Amount    float64           `json:"amount"`
	Currency  string            `json:"currency"`
	Type      string            `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ProcessedTransaction mirrors processing-service's output
type ProcessedTransaction struct {
	Transaction
	IsValid       bool      `json:"is_valid"`
	RiskScore     float64   `json:"risk_score"`
	EnrichedAt    time.Time `json:"enriched_at"`
	FailureReason string    `json:"failure_reason,omitempty"`
}
