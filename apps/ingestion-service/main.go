package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ingestion-service/internal/auth"
	"ingestion-service/internal/config"
	"ingestion-service/internal/middleware"
	"ingestion-service/internal/models"
	"ingestion-service/internal/publisher"
	"ingestion-service/internal/redis"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Setup Redis client
	redisClient, err := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Setup JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)

	// Setup Kafka producer
	producer, err := publisher.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Setup middleware
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(redisClient, 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	metricsMiddleware := middleware.NewMetricsMiddleware()

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// Metrics endpoint for Prometheus
	if cfg.MetricsEnabled {
		router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	}

	// Protected API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Transaction ingestion endpoint with all middleware
	apiRouter.HandleFunc("/transactions",
		metricsMiddleware.Wrap(
			idempotencyMiddleware.Wrap(
				authMiddleware.RequireAuth(
					authMiddleware.RequireAnyRole("teller", "admin")(
						IngestTransactionHandler(producer, cfg.KafkaTopic),
					),
				),
			),
		),
	).Methods("POST")

	// Batch transaction ingestion endpoint
	apiRouter.HandleFunc("/transactions/batch",
		metricsMiddleware.Wrap(
			idempotencyMiddleware.Wrap(
				authMiddleware.RequireAuth(
					authMiddleware.RequireRole("admin")(
						IngestBatchTransactionHandler(producer, cfg.KafkaTopic),
					),
				),
			),
		),
	).Methods("POST")

	// JWT token generation endpoint (for testing)
	apiRouter.HandleFunc("/auth/token",
		metricsMiddleware.Wrap(
			GenerateTokenHandler(jwtManager),
		),
	).Methods("POST")

	// Start HTTP server
	server := &http.Server{
		Addr:           cfg.HTTPHOST + ":" + cfg.HTTPPORT,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: int(cfg.MaxRequestSize),
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Ingestion service running on %s:%s", cfg.HTTPHOST, cfg.HTTPPORT)
		if cfg.MetricsEnabled {
			log.Printf("Metrics endpoint available at /metrics")
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down ingestion service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// IngestTransactionHandler accepts a JSON transaction and publishes it to Kafka
func IngestTransactionHandler(p *publisher.Producer, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RecordTransactionFailed("invalid_json")
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.IdempotencyKey == "" || req.AccountID == "" || req.UserID == "" {
			middleware.RecordTransactionFailed("missing_required_fields")
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}

		// Create transaction with generated ID and timestamp
		txn := models.Transaction{
			ID:             generateTransactionID(),
			IdempotencyKey: req.IdempotencyKey,
			AccountID:      req.AccountID,
			UserID:         req.UserID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Type:           req.Type,
			Category:       req.Category,
			Merchant:       req.Merchant,
			Reference:      req.Reference,
			Status:         "pending",
			Timestamp:      time.Now(),
			Metadata:       req.Metadata,
		}

		// Publish to Kafka
		if err := p.Publish(topic, txn); err != nil {
			middleware.RecordTransactionFailed("kafka_publish_failed")
			http.Error(w, "failed to enqueue transaction", http.StatusInternalServerError)
			return
		}

		// Record success metrics
		middleware.RecordTransactionIngested(txn.Currency, txn.Type, "success")

		// Return success response
		response := models.TransactionResponse{
			ID:        txn.ID,
			Status:    "accepted",
			Message:   "Transaction queued for processing",
			Timestamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}

// IngestBatchTransactionHandler accepts multiple transactions and publishes them in batch
func IngestBatchTransactionHandler(p *publisher.Producer, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqs []models.TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		if len(reqs) == 0 {
			http.Error(w, "empty batch", http.StatusBadRequest)
			return
		}

		// Convert requests to transactions
		transactions := make([]models.Transaction, len(reqs))
		for i, req := range reqs {
			transactions[i] = models.Transaction{
				ID:             generateTransactionID(),
				IdempotencyKey: req.IdempotencyKey,
				AccountID:      req.AccountID,
				UserID:         req.UserID,
				Amount:         req.Amount,
				Currency:       req.Currency,
				Type:           req.Type,
				Category:       req.Category,
				Merchant:       req.Merchant,
				Reference:      req.Reference,
				Status:         "pending",
				Timestamp:      time.Now(),
				Metadata:       req.Metadata,
			}
		}

		// Publish batch to Kafka
		if err := p.PublishBatch(topic, transactions); err != nil {
			http.Error(w, "failed to enqueue batch", http.StatusInternalServerError)
			return
		}

		// Return success response
		response := map[string]interface{}{
			"status":    "accepted",
			"message":   "Batch queued for processing",
			"count":     len(transactions),
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}

// GenerateTokenHandler generates JWT tokens for testing
func GenerateTokenHandler(jwtManager *auth.JWTManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID    string   `json:"user_id"`
			AccountID string   `json:"account_id"`
			Roles     []string `json:"roles"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Generate token
		token, err := jwtManager.GenerateToken(req.UserID, req.AccountID, req.Roles)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"token": token,
			"type":  "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	return "txn_" + time.Now().Format("20060102150405.000000000")
}
