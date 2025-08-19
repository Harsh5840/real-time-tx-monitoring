package processor

import (
	"context"
	"encoding/json"
	"time"

	"processing-service/internal/models"
	"processing-service/internal/publisher"
)

// Processor handles validation, enrichment, and risk scoring.
type Processor struct {
	pub *publisher.Publisher
}

// NewProcessor creates a new Processor instance.
func NewProcessor(pub *publisher.Publisher) *Processor {
	return &Processor{pub: pub}
}

// ProcessTransaction validates, enriches, scores, and publishes the result.
func (p *Processor) ProcessTransaction(ctx context.Context, rawMsg []byte) error {
	var tx models.Transaction
	if err := json.Unmarshal(rawMsg, &tx); err != nil {
		return err
	}

	// Basic validation
	isValid, reason := validate(tx)

	// Risk scoring (simple placeholder logic)
	risk := calculateRisk(tx, isValid)

	processed := &models.ProcessedTransaction{
		Transaction:   tx,
		IsValid:       isValid,
		RiskScore:     risk,
		EnrichedAt:    time.Now().UTC(),
		FailureReason: reason,
	}

	return p.pub.Publish(ctx, processed)
}

// validate checks if the transaction is valid.
func validate(tx models.Transaction) (bool, string) {
	if tx.Amount <= 0 {
		return false, "invalid amount"
	}
	if tx.Currency == "" {
		return false, "missing currency"
	}
	if tx.UserID == "" {
		return false, "missing user_id"
	}
	return true, ""
}

// calculateRisk assigns a risk score (0.0 = safe, 1.0 = high risk).
func calculateRisk(tx models.Transaction, isValid bool) float64 {
	if !isValid {
		return 1.0
	}
	if tx.Amount > 10000 {
		return 0.8
	}
	if tx.Type == "transfer" {
		return 0.5
	}
	return 0.1
}
