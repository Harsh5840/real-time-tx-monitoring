package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the storage service
type Config struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBUrl      string

	// Kafka configuration
	KafkaBrokers  string
	InputTopic    string
	ConsumerGroup string

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Service configuration
	BatchSize      int
	MaxRetries     int
	ProcessTimeout int // in seconds

	// Monitoring configuration
	MetricsEnabled bool
	MetricsPort    string

	// Storage configuration
	MaxConnections int
	IdleTimeout    int // in seconds
	QueryTimeout   int // in seconds
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "barclays_tx"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Kafka configuration
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		InputTopic:    getEnv("KAFKA_INPUT_TOPIC", "transactions.processed"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "storage-service"),

		// Redis configuration
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Service configuration
		BatchSize:      getEnvAsInt("BATCH_SIZE", 100),
		MaxRetries:     getEnvAsInt("MAX_RETRIES", 3),
		ProcessTimeout: getEnvAsInt("PROCESS_TIMEOUT", 30),

		// Monitoring configuration
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort:    getEnv("METRICS_PORT", "9092"),

		// Storage configuration
		MaxConnections: getEnvAsInt("MAX_CONNECTIONS", 10),
		IdleTimeout:    getEnvAsInt("IDLE_TIMEOUT", 300),
		QueryTimeout:   getEnvAsInt("QUERY_TIMEOUT", 30),
	}

	// Build database URL
	cfg.DBUrl = buildDatabaseURL(cfg)

	return cfg
}

// buildDatabaseURL constructs the PostgreSQL connection string
func buildDatabaseURL(cfg *Config) string {
	if dbUrl := os.Getenv("DATABASE_URL"); dbUrl != "" {
		return dbUrl
	}

	return "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=" + cfg.DBSSLMode
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
