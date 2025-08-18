package config

import (
	"log"
	"os"
)

// Config holds application configuration
type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	HTTPPORT     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "events"),
		HTTPPORT:     getEnv("HTTP_PORT", "8080"),
	}

	log.Printf("Loaded config: %+v\n", cfg)
	return cfg
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
