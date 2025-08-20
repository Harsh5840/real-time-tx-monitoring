package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"processing-service/internal/models"

	"github.com/segmentio/kafka-go"
)

// Consumer handles consuming raw transactions from Kafka
type Consumer struct {
	reader    *kafka.Reader
	processor Processor
}

// Processor interface for processing transactions
type Processor interface {
	ProcessTransaction(ctx context.Context, transaction *models.RawTransaction) error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers, topic, consumerGroup string, processor Processor) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         []string{brokers},
		Topic:           topic,
		GroupID:         consumerGroup,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
		CommitInterval:  1 * time.Second,
	})

	return &Consumer{
		reader:    reader,
		processor: processor,
	}, nil
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer context cancelled, stopping...")
			return nil
		default:
			// Read message with timeout
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			message, err := c.reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue // Timeout, continue to next iteration
				}
				log.Printf("Error reading message: %v", err)
				continue
			}

			// Process message in goroutine for better performance
			go func(msg kafka.Message) {
				if err := c.processMessage(ctx, msg); err != nil {
					log.Printf("Failed to process message: %v", err)
				}
			}(message)
		}
	}
}

// processMessage processes a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, message kafka.Message) error {
	start := time.Now()

	log.Printf("Processing message: Topic=%s, Partition=%d, Offset=%d, Key=%s",
		message.Topic, message.Partition, message.Offset, string(message.Key))

	// Deserialize the raw transaction
	var rawTxn models.RawTransaction
	if err := json.Unmarshal(message.Value, &rawTxn); err != nil {
		log.Printf("Failed to deserialize message: %v", err)
		return err
	}

	// Validate basic message structure
	if rawTxn.ID == "" {
		log.Printf("Message missing transaction ID, skipping")
		return nil
	}

	// Process the transaction
	if err := c.processor.ProcessTransaction(ctx, &rawTxn); err != nil {
		log.Printf("Failed to process transaction %s: %v", rawTxn.ID, err)
		return err
	}

	// Log successful processing
	log.Printf("Successfully processed transaction %s in %v",
		rawTxn.ID, time.Since(start))

	return nil
}

// Close shuts down the consumer safely
func (c *Consumer) Close() error {
	log.Println("Closing consumer...")
	return c.reader.Close()
}

// GetStats returns consumer statistics
func (c *Consumer) GetStats() kafka.ReaderStats {
	return c.reader.Stats()
}
