package models

import (
	"time"
)

// RawTransaction represents the incoming transaction from ingestion service
type RawTransaction struct {
	ID             string            `json:"id"`
	IdempotencyKey string            `json:"idempotency_key"`
	AccountID      string            `json:"account_id"`
	UserID         string            `json:"user_id"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	Type           string            `json:"type"`
	Category       string            `json:"category"`
	Merchant       string            `json:"merchant,omitempty"`
	Reference      string            `json:"reference,omitempty"`
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// ProcessedTransaction represents the transaction after business logic processing
type ProcessedTransaction struct {
	RawTransaction
	// Processing results
	RiskScore       float64 `json:"risk_score"`
	RiskLevel       string  `json:"risk_level"`
	IsApproved      bool    `json:"is_approved"`
	RejectionReason string  `json:"rejection_reason,omitempty"`

	// Business validation results
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`

	// Enrichment data
	Country    string `json:"country,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	DeviceInfo string `json:"device_info,omitempty"`

	// Processing metadata
	ProcessedAt    time.Time     `json:"processed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	ProcessorID    string        `json:"processor_id"`
}

// TransactionValidation represents validation rules and results
type TransactionValidation struct {
	IsValid  bool                `json:"is_valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RiskAssessment represents the risk analysis of a transaction
type RiskAssessment struct {
	RiskScore      float64      `json:"risk_score"`
	RiskLevel      string       `json:"risk_level"`
	RiskFactors    []RiskFactor `json:"risk_factors"`
	Recommendation string       `json:"recommendation"`
}

// RiskFactor represents a specific risk factor
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
}

// ProcessingResult represents the final result of transaction processing
type ProcessingResult struct {
	TransactionID   string        `json:"transaction_id"`
	Status          string        `json:"status"`
	RiskScore       float64       `json:"risk_score"`
	IsApproved      bool          `json:"is_approved"`
	RejectionReason string        `json:"rejection_reason,omitempty"`
	ProcessingTime  time.Duration `json:"processing_time"`
	Timestamp       time.Time     `json:"timestamp"`
}

// Constants for risk levels
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

// Constants for transaction statuses
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusFlagged  = "flagged"
	StatusFailed   = "failed"
)

// Constants for validation codes
const (
	ValidationCodeRequiredField   = "REQUIRED_FIELD"
	ValidationCodeInvalidAmount   = "INVALID_AMOUNT"
	ValidationCodeInvalidCurrency = "INVALID_CURRENCY"
	ValidationCodeBlockedCountry  = "BLOCKED_COUNTRY"
	ValidationCodeBlockedMerchant = "BLOCKED_MERCHANT"
	ValidationCodeExceedsLimit    = "EXCEEDS_LIMIT"
	ValidationCodeInvalidType     = "INVALID_TYPE"
)
