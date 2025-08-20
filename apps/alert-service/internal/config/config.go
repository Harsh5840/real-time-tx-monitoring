package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the alert service
type Config struct {
	// Kafka configuration
	KafkaBrokers  string
	InputTopic    string
	ConsumerGroup string

	// Notification configuration
	SlackWebhook  string
	EmailSMTP     string
	EmailFrom     string
	EmailPassword string
	EmailTo       []string

	// Alert rules configuration
	RiskThreshold      float64
	AmountThreshold    float64
	FrequencyThreshold int // alerts per hour

	// Service configuration
	BatchSize      int
	MaxRetries     int
	ProcessTimeout int // in seconds

	// Monitoring configuration
	MetricsEnabled bool
	MetricsPort    string

	// Alert channels
	EnableSlack   bool
	EnableEmail   bool
	EnableWebhook bool
	WebhookURL    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		// Kafka configuration
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "transactions.processed"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "alert-service"),

		// Notification configuration
		SlackWebhook:  getEnv("SLACK_WEBHOOK", ""),
		EmailSMTP:     getEnv("EMAIL_SMTP", "smtp.gmail.com:587"),
		EmailFrom:     getEnv("EMAIL_FROM", "alerts@barclays.com"),
		EmailPassword: getEnv("EMAIL_PASSWORD", ""),
		EmailTo:       getEnvAsSlice("EMAIL_TO", []string{"fraud@barclays.com"}),

		// Alert rules configuration
		RiskThreshold:      getEnvAsFloat("RISK_THRESHOLD", 0.7),
		AmountThreshold:    getEnvAsFloat("AMOUNT_THRESHOLD", 10000.0),
		FrequencyThreshold: getEnvAsInt("FREQUENCY_THRESHOLD", 5),

		// Service configuration
		BatchSize:      getEnvAsInt("BATCH_SIZE", 100),
		MaxRetries:     getEnvAsInt("MAX_RETRIES", 3),
		ProcessTimeout: getEnvAsInt("PROCESS_TIMEOUT", 30),

		// Monitoring configuration
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort:    getEnv("METRICS_PORT", "9093"),

		// Alert channels
		EnableSlack:   getEnvAsBool("ENABLE_SLACK", true),
		EnableEmail:   getEnvAsBool("ENABLE_EMAIL", false),
		EnableWebhook: getEnvAsBool("ENABLE_WEBHOOK", false),
		WebhookURL:    getEnv("WEBHOOK_URL", ""),
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
