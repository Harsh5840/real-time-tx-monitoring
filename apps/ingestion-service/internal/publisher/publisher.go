package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ingestion-service/internal/middleware"
	"ingestion-service/internal/models"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a Kafka writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer initializes a new Kafka producer
func NewProducer(brokers string) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokers},
		Balancer:     &kafka.Hash{}, // Use hash balancer for partitioning
		Async:        true,          // Enable async publishing for better performance
		RequiredAcks: 1,             // Require acknowledgment for reliability
	})
	return &Producer{writer: writer}, nil
}

// Publish sends a message to the given Kafka topic with account-based partitioning
func (p *Producer) Publish(topic string, transaction models.Transaction) error {
	start := time.Now()

	// Serialize the transaction
	message, err := json.Marshal(transaction)
	if err != nil {
		middleware.RecordKafkaMessagePublished(topic, "failed")
		log.Printf("failed to serialize transaction: %v", err)
		return err
	}

	// Create Kafka message with account-based partitioning
	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(transaction.AccountID), // Partition by account ID
		Value: message,
		Headers: []kafka.Header{
			{Key: "idempotency_key", Value: []byte(transaction.IdempotencyKey)},
			{Key: "user_id", Value: []byte(transaction.UserID)},
			{Key: "currency", Value: []byte(transaction.Currency)},
			{Key: "type", Value: []byte(transaction.Type)},
		},
	}

	// Publish message
	err = p.writer.WriteMessages(context.Background(), kafkaMessage)

	// Record metrics
	duration := time.Since(start)
	if err != nil {
		middleware.RecordKafkaMessagePublished(topic, "failed")
		log.Printf("failed to publish message to topic %s: %v", topic, err)
	} else {
		middleware.RecordKafkaMessagePublished(topic, "success")
	}

	middleware.RecordKafkaPublishDuration(topic, duration)
	return err
}

// PublishBatch publishes multiple messages in a batch for better throughput
func (p *Producer) PublishBatch(topic string, transactions []models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	start := time.Now()
	messages := make([]kafka.Message, len(transactions))

	for i, txn := range transactions {
		message, err := json.Marshal(txn)
		if err != nil {
			log.Printf("failed to serialize transaction %d: %v", i, err)
			continue
		}

		messages[i] = kafka.Message{
			Topic: topic,
			Key:   []byte(txn.AccountID),
			Value: message,
			Headers: []kafka.Header{
				{Key: "idempotency_key", Value: []byte(txn.IdempotencyKey)},
				{Key: "user_id", Value: []byte(txn.UserID)},
				{Key: "currency", Value: []byte(txn.Currency)},
				{Key: "type", Value: []byte(txn.Type)},
			},
		}
	}

	// Publish batch
	err := p.writer.WriteMessages(context.Background(), messages...)

	// Record metrics
	duration := time.Since(start)
	if err != nil {
		middleware.RecordKafkaMessagePublished(topic, "failed")
		log.Printf("failed to publish batch to topic %s: %v", topic, err)
	} else {
		middleware.RecordKafkaMessagePublished(topic, "success")
	}

	middleware.RecordKafkaPublishDuration(topic, duration)
	return err
}

// Close shuts down the Kafka writer
func (p *Producer) Close() error {
	return p.writer.Close()
}
