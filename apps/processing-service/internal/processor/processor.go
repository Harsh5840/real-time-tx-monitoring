package processor

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"processing-service/internal/models"
)

// Processor handles transaction processing with business logic
type Processor struct {
	publisher Publisher
}

// Publisher interface for publishing processed transactions
type Publisher interface {
	PublishProcessedTransaction(ctx context.Context, transaction *models.ProcessedTransaction) error
}

// NewProcessor creates a new transaction processor
func NewProcessor(publisher Publisher) *Processor {
	return &Processor{
		publisher: publisher,
	}
}

// ProcessTransaction processes a raw transaction through business logic
func (p *Processor) ProcessTransaction(ctx context.Context, rawTxn *models.RawTransaction) error {
	startTime := time.Now()

	log.Printf("Processing transaction %s for account %s", rawTxn.ID, rawTxn.AccountID)

	// Create processed transaction
	processedTxn := &models.ProcessedTransaction{
		RawTransaction: *rawTxn,
		ProcessedAt:    time.Now(),
		ProcessorID:    "processor-001",
	}

	// Step 1: Validate transaction
	validation := p.validateTransaction(rawTxn)
	processedTxn.IsValid = validation.IsValid

	if !validation.IsValid {
		processedTxn.Status = models.StatusRejected
		processedTxn.RejectionReason = p.formatValidationErrors(validation.Errors)
		processedTxn.ProcessingTime = time.Since(startTime)

		// Publish rejected transaction
		return p.publisher.PublishProcessedTransaction(ctx, processedTxn)
	}

	// Step 2: Enrich transaction data
	p.enrichTransaction(processedTxn)

	// Step 3: Assess risk
	riskAssessment := p.assessRisk(processedTxn)
	processedTxn.RiskScore = riskAssessment.RiskScore
	processedTxn.RiskLevel = riskAssessment.RiskLevel

	// Step 4: Apply business rules
	p.applyBusinessRules(processedTxn)

	// Step 5: Set final status
	p.setFinalStatus(processedTxn)

	// Calculate processing time
	processedTxn.ProcessingTime = time.Since(startTime)

	log.Printf("Transaction %s processed: Risk=%s, Status=%s, Time=%v",
		processedTxn.ID, processedTxn.RiskLevel, processedTxn.Status, processedTxn.ProcessingTime)

	// Publish processed transaction
	return p.publisher.PublishProcessedTransaction(ctx, processedTxn)
}

// validateTransaction validates the transaction against business rules
func (p *Processor) validateTransaction(txn *models.RawTransaction) *models.TransactionValidation {
	validation := &models.TransactionValidation{
		IsValid:  true,
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
	}

	// Required field validation
	if txn.AccountID == "" {
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "account_id",
			Code:    models.ValidationCodeRequiredField,
			Message: "Account ID is required",
		})
		validation.IsValid = false
	}

	if txn.UserID == "" {
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "user_id",
			Code:    models.ValidationCodeRequiredField,
			Message: "User ID is required",
		})
		validation.IsValid = false
	}

	if txn.Amount <= 0 {
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "amount",
			Code:    models.ValidationCodeInvalidAmount,
			Message: "Amount must be greater than 0",
		})
		validation.IsValid = false
	}

	// Amount limit validation
	if txn.Amount > 100000.0 { // Configurable limit
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "amount",
			Code:    models.ValidationCodeExceedsLimit,
			Message: "Amount exceeds maximum allowed limit",
		})
		validation.IsValid = false
	}

	// Currency validation
	validCurrencies := []string{"USD", "EUR", "GBP", "INR", "CAD", "AUD"}
	currencyValid := false
	for _, curr := range validCurrencies {
		if curr == txn.Currency {
			currencyValid = true
			break
		}
	}

	if !currencyValid {
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "currency",
			Code:    models.ValidationCodeInvalidCurrency,
			Message: "Invalid currency code",
		})
		validation.IsValid = false
	}

	// Transaction type validation
	validTypes := []string{"purchase", "transfer", "withdrawal", "deposit", "refund"}
	typeValid := false
	for _, t := range validTypes {
		if t == txn.Type {
			typeValid = true
			break
		}
	}

	if !typeValid {
		validation.Errors = append(validation.Errors, models.ValidationError{
			Field:   "type",
			Code:    models.ValidationCodeInvalidType,
			Message: "Invalid transaction type",
		})
		validation.IsValid = false
	}

	return validation
}

// enrichTransaction adds additional data to the transaction
func (p *Processor) enrichTransaction(txn *models.ProcessedTransaction) {
	// Simulate data enrichment
	if txn.Metadata != nil {
		if country, exists := txn.Metadata["country"]; exists {
			txn.Country = country
		}
		if ip, exists := txn.Metadata["ip_address"]; exists {
			txn.IPAddress = ip
		}
		if device, exists := txn.Metadata["device_info"]; exists {
			txn.DeviceInfo = device
		}
	}

	// Set default values if not present
	if txn.Country == "" {
		txn.Country = "US" // Default country
	}
	if txn.IPAddress == "" {
		txn.IPAddress = "192.168.1.1" // Default IP
	}
}

// assessRisk calculates the risk score for the transaction
func (p *Processor) assessRisk(txn *models.ProcessedTransaction) *models.RiskAssessment {
	riskScore := 0.0
	var riskFactors []models.RiskFactor

	// Amount-based risk
	if txn.Amount > 10000 {
		riskScore += 0.3
		riskFactors = append(riskFactors, models.RiskFactor{
			Factor:      "high_amount",
			Weight:      0.3,
			Description: "Transaction amount exceeds $10,000",
			Severity:    "medium",
		})
	}

	// Time-based risk (late night transactions)
	hour := txn.Timestamp.Hour()
	if hour >= 22 || hour <= 6 {
		riskScore += 0.2
		riskFactors = append(riskFactors, models.RiskFactor{
			Factor:      "late_night",
			Weight:      0.2,
			Description: "Transaction during late night hours",
			Severity:    "low",
		})
	}

	// Country-based risk
	if txn.Country == "XX" || txn.Country == "YY" {
		riskScore += 0.5
		riskFactors = append(riskFactors, models.RiskFactor{
			Factor:      "blocked_country",
			Weight:      0.5,
			Description: "Transaction from blocked country",
			Severity:    "high",
		})
	}

	// Merchant-based risk
	if strings.Contains(strings.ToLower(txn.Merchant), "gambling") ||
		strings.Contains(strings.ToLower(txn.Merchant), "crypto") {
		riskScore += 0.4
		riskFactors = append(riskFactors, models.RiskFactor{
			Factor:      "risky_merchant",
			Weight:      0.4,
			Description: "Transaction with risky merchant category",
			Severity:    "medium",
		})
	}

	// Random factor for demonstration (in real system, this would be ML-based)
	rand.Seed(time.Now().UnixNano())
	randomRisk := rand.Float64() * 0.1
	riskScore += randomRisk

	// Cap risk score at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	// Determine risk level
	var riskLevel string
	var recommendation string

	switch {
	case riskScore < 0.3:
		riskLevel = models.RiskLevelLow
		recommendation = "Approve automatically"
	case riskScore < 0.6:
		riskLevel = models.RiskLevelMedium
		recommendation = "Review manually"
	case riskScore < 0.8:
		riskLevel = models.RiskLevelHigh
		recommendation = "Flag for investigation"
	default:
		riskLevel = models.RiskLevelCritical
		recommendation = "Block immediately"
	}

	return &models.RiskAssessment{
		RiskScore:      riskScore,
		RiskLevel:      riskLevel,
		RiskFactors:    riskFactors,
		Recommendation: recommendation,
	}
}

// applyBusinessRules applies business logic to the transaction
func (p *Processor) applyBusinessRules(txn *models.ProcessedTransaction) {
	// Auto-approve low-risk transactions
	if txn.RiskScore < 0.3 {
		txn.IsApproved = true
		return
	}

	// Auto-reject high-risk transactions
	if txn.RiskScore > 0.8 {
		txn.IsApproved = false
		txn.RejectionReason = "High risk score - automatic rejection"
		return
	}

	// For medium risk, apply additional rules
	if txn.RiskScore >= 0.3 && txn.RiskScore <= 0.8 {
		// Check for specific risk factors
		hasBlockedCountry := false
		hasBlockedMerchant := false

		for _, factor := range []string{"XX", "YY"} {
			if txn.Country == factor {
				hasBlockedCountry = true
				break
			}
		}

		for _, merchant := range []string{"blocked_merchant_1", "blocked_merchant_2"} {
			if txn.Merchant == merchant {
				hasBlockedMerchant = true
				break
			}
		}

		if hasBlockedCountry || hasBlockedMerchant {
			txn.IsApproved = false
			txn.RejectionReason = "Blocked country or merchant"
		} else {
			txn.IsApproved = true
		}
	}
}

// setFinalStatus sets the final status based on processing results
func (p *Processor) setFinalStatus(txn *models.ProcessedTransaction) {
	if !txn.IsValid {
		txn.Status = models.StatusRejected
		return
	}

	if txn.IsApproved {
		txn.Status = models.StatusApproved
	} else {
		txn.Status = models.StatusRejected
	}

	// Flag high-risk approved transactions for review
	if txn.IsApproved && txn.RiskScore > 0.6 {
		txn.Status = models.StatusFlagged
	}
}

// formatValidationErrors formats validation errors into a readable string
func (p *Processor) formatValidationErrors(errors []models.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return strings.Join(messages, "; ")
}

// ProcessBatch processes multiple transactions in batch
func (p *Processor) ProcessBatch(ctx context.Context, transactions []*models.RawTransaction) error {
	log.Printf("Processing batch of %d transactions", len(transactions))

	for _, txn := range transactions {
		if err := p.ProcessTransaction(ctx, txn); err != nil {
			log.Printf("Failed to process transaction %s: %v", txn.ID, err)
			// Continue processing other transactions
		}
	}

	log.Printf("Batch processing completed")
	return nil
}
