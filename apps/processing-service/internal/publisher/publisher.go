package publisher

import (
	"context"
	"encoding/json"
	"strings"

	"processing-service/internal/models"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new Kafka publisher for processed transactions.
func NewPublisher(brokers string, topic string) *Publisher {
	parts := strings.Split(brokers, ",")
	addrs := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			addrs = append(addrs, s)
		}
	}
	if len(addrs) == 0 {
		addrs = []string{brokers}
	}

	return &Publisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(addrs...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Publish sends a processed transaction to Kafka.
func (p *Publisher) Publish(ctx context.Context, tx *models.ProcessedTransaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(tx.ID),
		Value: data,
	})
}

// Close shuts down the publisher safely.
func (p *Publisher) Close() error {
	return p.writer.Close()
}
