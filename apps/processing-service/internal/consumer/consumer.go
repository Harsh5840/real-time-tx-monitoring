package consumer

import (
	"context"
	"log"
	"strings"

	"processing-service/internal/processor"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	proc   *processor.Processor
}

func NewConsumer(brokers string, topic string, groupID string, proc *processor.Processor) (*Consumer, error) {
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
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: r, proc: proc}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	defer c.reader.Close()
	log.Println("processing-service consumer started...")

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("read error: %v", err)
			continue
		}
		if err := c.proc.ProcessTransaction(ctx, m.Value); err != nil {
			log.Printf("process error: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
