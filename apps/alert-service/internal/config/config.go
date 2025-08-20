package config

import (
	"log"
	"os"
)

// Config holds environment variables for the alert service
type Config struct {
	KafkaBrokers  string
	ConsumerGroup string
	InputTopic    string
	SlackWebhook  string
}

// LoadConfig reads configuration from environment variables with fallbacks
func LoadConfig() *Config {
	cfg := &Config{
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "alert-service"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "alerts"),
		SlackWebhook:  getEnv("SLACK_WEBHOOK", ""),
	}

	log.Printf("loaded config: KafkaBrokers=%s, GroupID=%s, InputTopic=%s",
		cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.InputTopic)

	return cfg
}

// getEnv returns the environment variable value or a default value if not set
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
