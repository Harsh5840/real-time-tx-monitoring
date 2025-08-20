package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Business metrics
	transactionsIngested = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_ingested_total",
			Help: "Total number of transactions ingested",
		},
		[]string{"currency", "type", "status"},
	)

	transactionsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_failed_total",
			Help: "Total number of failed transactions",
		},
		[]string{"reason"},
	)

	// Kafka metrics
	kafkaMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_published_total",
			Help: "Total number of Kafka messages published",
		},
		[]string{"topic", "status"},
	)

	kafkaPublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_publish_duration_seconds",
			Help:    "Duration of Kafka publish operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)

	// Redis metrics
	redisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	redisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Duration of Redis operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// MetricsMiddleware wraps HTTP handlers with Prometheus metrics
type MetricsMiddleware struct{}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Wrap wraps an HTTP handler with metrics collection
func (m *MetricsMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(recorder, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		endpoint := r.URL.Path
		if endpoint == "" {
			endpoint = "/"
		}

		httpRequestsTotal.WithLabelValues(r.Method, endpoint, strconv.Itoa(recorder.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
	}
}

// RecordTransactionIngested records a successful transaction ingestion
func RecordTransactionIngested(currency, txnType, status string) {
	transactionsIngested.WithLabelValues(currency, txnType, status).Inc()
}

// RecordTransactionFailed records a failed transaction
func RecordTransactionFailed(reason string) {
	transactionsFailed.WithLabelValues(reason).Inc()
}

// RecordKafkaMessagePublished records a Kafka message publication
func RecordKafkaMessagePublished(topic, status string) {
	kafkaMessagesPublished.WithLabelValues(topic, status).Inc()
}

// RecordKafkaPublishDuration records Kafka publish duration
func RecordKafkaPublishDuration(topic string, duration time.Duration) {
	kafkaPublishDuration.WithLabelValues(topic).Observe(duration.Seconds())
}

// RecordRedisOperation records a Redis operation
func RecordRedisOperation(operation, status string) {
	redisOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordRedisOperationDuration records Redis operation duration
func RecordRedisOperationDuration(operation string, duration time.Duration) {
	redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// statusRecorder captures the HTTP status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
