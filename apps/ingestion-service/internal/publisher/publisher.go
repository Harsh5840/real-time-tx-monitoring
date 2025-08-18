package publisher

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a Kafka writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer initializes a new Kafka producer
func NewProducer(brokers string) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokers},
		Balancer: &kafka.LeastBytes{},
	})
	return &Producer{writer: writer}, nil
}

// Publish sends a message to the given Kafka topic
func (p *Producer) Publish(topic string, message []byte) error {
	// Kafka WriteMessages requires a context.Context as the first argument
	err := p.writer.WriteMessages(
		context.Background(), // required context
		kafka.Message{
			Topic: topic,
			Value: message,
		},
	)
	if err != nil {
		log.Printf("failed to publish message to topic %s: %v", topic, err)
	}
	return err
}

// Close shuts down the Kafka writer
func (p *Producer) Close() error {
	return p.writer.Close()
}
