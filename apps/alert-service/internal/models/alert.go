package models

import "time"

// Alert represents an alert message that will be processed by the alert service
type Alert struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`     // e.g., "high_risk", "fraud", "threshold_exceeded"
	Severity      string            `json:"severity"` // e.g., "low", "medium", "high", "critical"
	Message       string            `json:"message"`
	TransactionID string            `json:"transaction_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

