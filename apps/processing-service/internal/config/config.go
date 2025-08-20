package config

import (
	"log"
	"os"
)

type Config struct {
	KafkaBrokers  string
	ConsumerGroup string
	InputTopic    string
	OutputTopic   string
}

// getEnv returns the value of an env variable or a fallback if not set
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfig() *Config {
	cfg := &Config{
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "processing-service"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "events"),
		OutputTopic:   getEnv("KAFKA_OUTPUT_TOPIC", "transactions_processed"),
	}

	log.Printf(" Loaded config: %+v", cfg)
	return cfg
}
