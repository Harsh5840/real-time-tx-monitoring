package models

import (
	"time"
)

// Alert represents a fraud or operational alert
type Alert struct {
	ID              string            `json:"id"`
	TransactionID   string            `json:"transaction_id"`
	AccountID       string            `json:"account_id"`
	UserID          string            `json:"user_id"`
	AlertType       string            `json:"alert_type"`
	Severity        string            `json:"severity"`
	RiskScore       float64           `json:"risk_score"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Description     string            `json:"description"`
	RuleTriggered   string            `json:"rule_triggered"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty"`
	ResolvedBy      string            `json:"resolved_by,omitempty"`
	ResolutionNotes string            `json:"resolution_notes,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// AlertRule represents a rule that can trigger alerts
type AlertRule struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Conditions  []Condition `json:"conditions"`
	Actions     []Action    `json:"actions"`
	Enabled     bool        `json:"enabled"`
	Priority    int         `json:"priority"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Condition represents a condition that must be met for an alert rule
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Action represents an action to take when an alert rule is triggered
type Action struct {
	Type    string            `json:"type"`
	Config  map[string]string `json:"config"`
	Enabled bool              `json:"enabled"`
}

// Notification represents a notification sent for an alert
type Notification struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
	Error     string    `json:"error,omitempty"`
}

// AlertSummary represents aggregated alert data
type AlertSummary struct {
	TotalAlerts       int64   `json:"total_alerts"`
	OpenAlerts        int64   `json:"open_alerts"`
	ResolvedAlerts    int64   `json:"resolved_alerts"`
	HighSeverity      int64   `json:"high_severity"`
	MediumSeverity    int64   `json:"medium_severity"`
	LowSeverity       int64   `json:"low_severity"`
	FraudAlerts       int64   `json:"fraud_alerts"`
	OperationalAlerts int64   `json:"operational_alerts"`
	AverageRiskScore  float64 `json:"average_risk_score"`
}

// Constants for alert types
const (
	AlertTypeFraud       = "fraud"
	AlertTypeOperational = "operational"
	AlertTypeCompliance  = "compliance"
	AlertTypeRisk        = "risk"
)

// Constants for alert severity
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Constants for alert status
const (
	StatusOpen          = "open"
	StatusInvestigating = "investigating"
	StatusResolved      = "resolved"
	StatusFalsePositive = "false_positive"
	StatusClosed        = "closed"
)

// Constants for notification channels
const (
	ChannelSlack   = "slack"
	ChannelEmail   = "email"
	ChannelWebhook = "webhook"
	ChannelSMS     = "sms"
)

// Constants for notification status
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

// Constants for rule types
const (
	RuleTypeRiskScore = "risk_score"
	RuleTypeAmount    = "amount"
	RuleTypeFrequency = "frequency"
	RuleTypeLocation  = "location"
	RuleTypeMerchant  = "merchant"
	RuleTypeTime      = "time"
	RuleTypePattern   = "pattern"
)

// Constants for condition operators
const (
	OperatorEquals      = "equals"
	OperatorNotEquals   = "not_equals"
	OperatorGreaterThan = "greater_than"
	OperatorLessThan    = "less_than"
	OperatorContains    = "contains"
	OperatorNotContains = "not_contains"
	OperatorIn          = "in"
	OperatorNotIn       = "not_in"
	OperatorBetween     = "between"
	OperatorRegex       = "regex"
)

// CreateTablesSQL returns the SQL to create the alert-related tables
func CreateTablesSQL() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS alerts (
			id VARCHAR(255) PRIMARY KEY,
			transaction_id VARCHAR(255) NOT NULL,
			account_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			alert_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			risk_score DECIMAL(3,2),
			amount DECIMAL(15,2),
			currency VARCHAR(3),
			description TEXT,
			rule_triggered VARCHAR(255),
			status VARCHAR(50) DEFAULT 'open',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			resolved_at TIMESTAMP,
			resolved_by VARCHAR(255),
			resolution_notes TEXT,
			metadata JSONB
		)`,

		`CREATE TABLE IF NOT EXISTS alert_rules (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			conditions JSONB,
			actions JSONB,
			enabled BOOLEAN DEFAULT true,
			priority INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS notifications (
			id VARCHAR(255) PRIMARY KEY,
			alert_id VARCHAR(255) NOT NULL,
			channel VARCHAR(50) NOT NULL,
			recipient VARCHAR(255),
			subject VARCHAR(500),
			message TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			sent_at TIMESTAMP,
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
}

// CreateIndexesSQL returns the SQL to create the necessary indexes
func CreateIndexesSQL() []string {
	return []string{
		`CREATE INDEX IF NOT EXISTS idx_alerts_account_id ON alerts(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_alert_type ON alerts(alert_type)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_alert_id ON notifications(alert_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_rules_priority ON alert_rules(priority)`,
	}
}
