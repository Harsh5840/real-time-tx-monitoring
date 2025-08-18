package collector

import (
	"encoding/json"
	"log"
	"time"
)

// Transaction represents a single financial transaction
type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Type      string    `json:"type"` // e.g. "credit" or "debit"
	Timestamp time.Time `json:"timestamp"`
}

// Collector is responsible for collecting transaction data
type Collector struct {
	input chan []byte        // raw JSON messages
	out   chan<- Transaction // parsed transaction structs
	done  chan struct{}      // signals shutdown
}

// NewCollector initializes a new Collector
func NewCollector(input chan []byte, out chan<- Transaction) *Collector {
	return &Collector{
		input: input,
		out:   out,
		done:  make(chan struct{}),
	}
}

// Start begins the collection process
func (c *Collector) Start() {
	go func() {
		for {
			select {
			case msg := <-c.input:
				var tx Transaction
				if err := json.Unmarshal(msg, &tx); err != nil {
					log.Printf("❌ failed to parse transaction: %v", err)
					continue
				}
				// send parsed transaction to pipeline
				c.out <- tx

			case <-c.done:
				log.Println("🛑 Collector shutting down...")
				return
			}
		}
	}()
}

// Stop signals the collector to stop
func (c *Collector) Stop() {
	close(c.done)
}
