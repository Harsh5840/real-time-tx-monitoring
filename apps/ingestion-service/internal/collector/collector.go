package collector

import (
	"encoding/json"
	"log"
	"time"
)

// Transaction represents a normalized financial transaction event
type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Type      string    `json:"type"` // "credit" or "debit"
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // e.g. "api", "kafka", "file"
}

// Collector defines the interface for transaction collectors
type Collector interface {
	Collect(chan<- Transaction) error
}

// MockCollector is a simple implementation that generates fake transactions
type MockCollector struct {
	Count int // number of fake transactions to generate
}

// Collect generates Count fake transactions and sends them into the channel
func (m *MockCollector) Collect(out chan<- Transaction) error {
	log.Printf("[collector] Generating %d fake transactions...\n", m.Count)

	for i := 0; i < m.Count; i++ {
		txn := Transaction{
			ID:        generateID(),
			UserID:    generateUserID(),
			Amount:    generateAmount(),
			Currency:  "USD",
			Type:      pickType(),
			Timestamp: time.Now(),
			Source:    "mock",
		}

		// log transaction in JSON format
		bytes, _ := json.Marshal(txn)
		log.Printf("[collector] New Transaction: %s\n", string(bytes))

		out <- txn
		time.Sleep(500 * time.Millisecond) // simulate delay
	}

	return nil
}

// helper functions — these can be replaced with real data later
func generateID() string {
	return "txn_" + time.Now().Format("150405.000")
}

func generateUserID() string {
	return "user_" + time.Now().Format("05")
}

func generateAmount() float64 {
	return float64(time.Now().UnixNano()%10000) / 100
}

func pickType() string {
	if time.Now().UnixNano()%2 == 0 {
		return "credit"
	}
	return "debit"
}
