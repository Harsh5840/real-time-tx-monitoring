package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the processing service
type Config struct {
	// Kafka configuration
	KafkaBrokers   string
	InputTopic     string
	OutputTopic    string
	ConsumerGroup  string

	// Processing configuration
	MaxRetries     int
	BatchSize      int
	ProcessTimeout int // in seconds

	// Monitoring configuration
	MetricsEnabled bool
	MetricsPort    string

	// Business rules configuration
	RiskThreshold float64
	MaxAmount     float64
	BlockedCountries []string
	BlockedMerchants []string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		// Kafka configuration
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "transactions.raw"),
		OutputTopic:   getEnv("KAFKA_OUTPUT_TOPIC", "transactions.processed"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "processing-service"),

		// Processing configuration
		MaxRetries:     getEnvAsInt("MAX_RETRIES", 3),
		BatchSize:      getEnvAsInt("BATCH_SIZE", 100),
		ProcessTimeout: getEnvAsInt("PROCESS_TIMEOUT", 30),

		// Monitoring configuration
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort:    getEnv("METRICS_PORT", "9091"),

		// Business rules configuration
		RiskThreshold:    getEnvAsFloat("RISK_THRESHOLD", 0.7),
		MaxAmount:        getEnvAsFloat("MAX_AMOUNT", 100000.0),
		BlockedCountries: getEnvAsSlice("BLOCKED_COUNTRIES", []string{"XX", "YY"}),
		BlockedMerchants: getEnvAsSlice("BLOCKED_MERCHANTS", []string{"blocked_merchant_1", "blocked_merchant_2"}),
	}

	return cfg
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated values
		return []string{value}
	}
	return defaultValue
}
