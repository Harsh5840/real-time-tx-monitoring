package consumer

import (
	"context"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(ctx context.Context, payload []byte) error
}

// Consumer wraps the kafka.Reader
type Consumer struct {
	reader *kafka.Reader
	h      Handler
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers string, groupID, topic string, h Handler) *Consumer {
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

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  addrs,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{reader: r, h: h}
}

// Start begins consuming messages and forwarding to the handler
func (c *Consumer) Start(ctx context.Context) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("read error: %v", err)
			continue
		}
		if err := c.h.Handle(ctx, m.Value); err != nil {
			log.Printf("handler error: %v", err)
		}
	}
}

// Close shuts down the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
