package config

import (
	"log"
	"os"
)

type Config struct {
	// Database
	DBUrl string

	// Kafka
	KafkaBrokers  string
	ConsumerGroup string
	InputTopic    string
}

func LoadConfig() *Config {
	cfg := &Config{
		DBUrl:         getEnv("DB_URL", "postgresql://user:password@localhost:5432/postgres?sslmode=disable"),
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "storage-service"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "transactions_processed"),
	}

	log.Printf("loaded config: DB=%s, KafkaBrokers=%s, GroupID=%s, InputTopic=%s",
		cfg.DBUrl, cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.InputTopic)

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
