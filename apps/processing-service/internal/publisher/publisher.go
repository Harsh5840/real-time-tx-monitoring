package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"processing-service/internal/models"

	"github.com/segmentio/kafka-go"
)

// Publisher handles publishing processed transactions to Kafka
type Publisher struct {
	writer *kafka.Writer
	topic  string
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(brokers, topic string) *Publisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokers},
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // Use hash balancer for partitioning
		Async:        true,          // Enable async publishing for better performance
		RequiredAcks: 1,             // Require acknowledgment for reliability
	})

	return &Publisher{
		writer: writer,
		topic:  topic,
	}
}

// PublishProcessedTransaction publishes a processed transaction to Kafka
func (p *Publisher) PublishProcessedTransaction(ctx context.Context, transaction *models.ProcessedTransaction) error {
	start := time.Now()

	// Serialize the transaction
	message, err := json.Marshal(transaction)
	if err != nil {
		log.Printf("Failed to serialize processed transaction: %v", err)
		return err
	}

	// Create Kafka message with account-based partitioning
	kafkaMessage := kafka.Message{
		Topic: p.topic,
		Key:   []byte(transaction.AccountID), // Partition by account ID
		Value: message,
		Headers: []kafka.Header{
			{Key: "idempotency_key", Value: []byte(transaction.IdempotencyKey)},
			{Key: "user_id", Value: []byte(transaction.UserID)},
			{Key: "risk_level", Value: []byte(transaction.RiskLevel)},
			{Key: "status", Value: []byte(transaction.Status)},
			{Key: "processed_at", Value: []byte(transaction.ProcessedAt.Format(time.RFC3339))},
		},
	}

	// Publish message
	err = p.writer.WriteMessages(ctx, kafkaMessage)

	// Log the result
	if err != nil {
		log.Printf("Failed to publish processed transaction %s to topic %s: %v",
			transaction.ID, p.topic, err)
	} else {
		log.Printf("Published processed transaction %s to topic %s in %v",
			transaction.ID, p.topic, time.Since(start))
	}

	return err
}

// PublishBatch publishes multiple processed transactions in a batch
func (p *Publisher) PublishBatch(ctx context.Context, transactions []*models.ProcessedTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	start := time.Now()
	messages := make([]kafka.Message, len(transactions))

	for i, txn := range transactions {
		message, err := json.Marshal(txn)
		if err != nil {
			log.Printf("Failed to serialize transaction %d: %v", i, err)
			continue
		}

		messages[i] = kafka.Message{
			Topic: p.topic,
			Key:   []byte(txn.AccountID),
			Value: message,
			Headers: []kafka.Header{
				{Key: "idempotency_key", Value: []byte(txn.IdempotencyKey)},
				{Key: "user_id", Value: []byte(txn.UserID)},
				{Key: "risk_level", Value: []byte(txn.RiskLevel)},
				{Key: "status", Value: []byte(txn.Status)},
				{Key: "processed_at", Value: []byte(txn.ProcessedAt.Format(time.RFC3339))},
			},
		}
	}

	// Publish batch
	err := p.writer.WriteMessages(ctx, messages...)

	// Log the result
	if err != nil {
		log.Printf("Failed to publish batch of %d transactions to topic %s: %v",
			len(transactions), p.topic, err)
	} else {
		log.Printf("Published batch of %d transactions to topic %s in %v",
			len(transactions), p.topic, time.Since(start))
	}

	return err
}

// Close shuts down the Kafka writer
func (p *Publisher) Close() error {
	return p.writer.Close()
}
