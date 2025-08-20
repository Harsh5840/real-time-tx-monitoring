package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for our service
type Config struct {
	// HTTP server configuration
	HTTPPORT string
	HTTPHOST string

	// Kafka configuration
	KafkaBrokers string
	KafkaTopic   string

	// Redis configuration for idempotency and caching
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT configuration
	JWTSecret     string
	JWTExpiration int // in hours

	// Security configuration
	RateLimitPerSecond int
	MaxRequestSize     int64 // in bytes

	// Monitoring configuration
	MetricsEnabled bool
	MetricsPort    string
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_PER_SECOND", "1000"))
	maxRequestSize, _ := strconv.ParseInt(getEnv("MAX_REQUEST_SIZE", "1048576"), 10, 64) // 1MB default
	jwtExpiration, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	metricsEnabled, _ := strconv.ParseBool(getEnv("METRICS_ENABLED", "true"))

	return &Config{
		HTTPPORT:           getEnv("HTTP_PORT", "8080"),
		HTTPHOST:           getEnv("HTTP_HOST", "0.0.0.0"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:         getEnv("KAFKA_TOPIC", "transactions.raw"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration:      jwtExpiration,
		RateLimitPerSecond: rateLimit,
		MaxRequestSize:     maxRequestSize,
		MetricsEnabled:     metricsEnabled,
		MetricsPort:        getEnv("METRICS_PORT", "9090"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
