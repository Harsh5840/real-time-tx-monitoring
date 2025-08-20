package models

import (
	"time"
)

// StoredTransaction represents a transaction stored in the database
type StoredTransaction struct {
	ID             string            `json:"id" db:"id"`
	IdempotencyKey string            `json:"idempotency_key" db:"idempotency_key"`
	AccountID      string            `json:"account_id" db:"account_id"`
	UserID         string            `json:"user_id" db:"user_id"`
	Amount         float64           `json:"amount" db:"amount"`
	Currency       string            `json:"currency" db:"currency"`
	Type           string            `json:"type" db:"type"`
	Category       string            `json:"category" db:"category"`
	Merchant       string            `json:"merchant" db:"merchant"`
	Reference      string            `json:"reference" db:"reference"`
	Status         string            `json:"status" db:"status"`
	Timestamp      time.Time         `json:"timestamp" db:"timestamp"`
	Metadata       map[string]string `json:"metadata" db:"metadata"`

	// Processing results
	RiskScore       float64 `json:"risk_score" db:"risk_score"`
	RiskLevel       string  `json:"risk_level" db:"risk_level"`
	IsApproved      bool    `json:"is_approved" db:"is_approved"`
	RejectionReason string  `json:"rejection_reason" db:"rejection_reason"`

	// Business validation results
	IsValid          bool     `json:"is_valid" db:"is_valid"`
	ValidationErrors []string `json:"validation_errors" db:"validation_errors"`

	// Enrichment data
	Country    string `json:"country" db:"country"`
	IPAddress  string `json:"ip_address" db:"ip_address"`
	DeviceInfo string `json:"device_info" db:"device_info"`

	// Processing metadata
	ProcessedAt    time.Time     `json:"processed_at" db:"processed_at"`
	ProcessingTime time.Duration `json:"processing_time" db:"processing_time"`
	ProcessorID    string        `json:"processor_id" db:"processor_id"`

	// Storage metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Account represents a bank account
type Account struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	AccountType string    `json:"account_type" db:"account_type"`
	Balance     float64   `json:"balance" db:"balance"`
	Currency    string    `json:"currency" db:"currency"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	AccountID         string    `json:"account_id" db:"account_id"`
	TotalTransactions int64     `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64   `json:"total_amount" db:"total_amount"`
	AverageAmount     float64   `json:"average_amount" db:"average_amount"`
	LastTransaction   time.Time `json:"last_transaction" db:"last_transaction"`
	RiskLevel         string    `json:"risk_level" db:"risk_level"`
}

// RiskMetrics represents risk-related metrics
type RiskMetrics struct {
	AccountID     string    `json:"account_id" db:"account_id"`
	RiskScore     float64   `json:"risk_score" db:"risk_score"`
	RiskLevel     string    `json:"risk_level" db:"risk_level"`
	TotalFlagged  int64     `json:"total_flagged" db:"total_flagged"`
	TotalRejected int64     `json:"total_rejected" db:"total_rejected"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
}

// Database schema constants
const (
	// Table names
	TableTransactions = "transactions"
	TableAccounts     = "accounts"
	TableRiskMetrics  = "risk_metrics"

	// Index names
	IndexTransactionsAccountID = "idx_transactions_account_id"
	IndexTransactionsUserID    = "idx_transactions_user_id"
	IndexTransactionsStatus    = "idx_transactions_status"
	IndexTransactionsTimestamp = "idx_transactions_timestamp"
	IndexTransactionsRiskLevel = "idx_transactions_risk_level"

	// Status values
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusFlagged  = "flagged"
	StatusFailed   = "failed"

	// Risk levels
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"

	// Account types
	AccountTypeChecking = "checking"
	AccountTypeSavings  = "savings"
	AccountTypeCredit   = "credit"
	AccountTypeBusiness = "business"
)

// CreateTablesSQL returns the SQL to create the necessary tables
func CreateTablesSQL() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			account_type VARCHAR(50) NOT NULL,
			balance DECIMAL(15,2) DEFAULT 0.00,
			currency VARCHAR(3) NOT NULL,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(255) PRIMARY KEY,
			idempotency_key VARCHAR(255) UNIQUE NOT NULL,
			account_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(100),
			merchant VARCHAR(255),
			reference VARCHAR(255),
			status VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			metadata JSONB,
			risk_score DECIMAL(3,2),
			risk_level VARCHAR(20),
			is_approved BOOLEAN DEFAULT false,
			rejection_reason TEXT,
			is_valid BOOLEAN DEFAULT true,
			validation_errors TEXT[],
			country VARCHAR(3),
			ip_address INET,
			device_info TEXT,
			processed_at TIMESTAMP,
			processing_time INTERVAL,
			processor_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS risk_metrics (
			account_id VARCHAR(255) PRIMARY KEY,
			risk_score DECIMAL(3,2) DEFAULT 0.00,
			risk_level VARCHAR(20) DEFAULT 'low',
			total_flagged BIGINT DEFAULT 0,
			total_rejected BIGINT DEFAULT 0,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
}

// CreateIndexesSQL returns the SQL to create the necessary indexes
func CreateIndexesSQL() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_risk_level ON transactions(risk_level)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_idempotency_key ON transactions(idempotency_key)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status)`,
	}
}
