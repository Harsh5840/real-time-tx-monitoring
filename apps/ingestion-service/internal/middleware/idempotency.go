package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ingestion-service/internal/models"
	"ingestion-service/internal/redis"
)

// IdempotencyMiddleware ensures idempotent operations
type IdempotencyMiddleware struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// NewIdempotencyMiddleware creates a new idempotency middleware
func NewIdempotencyMiddleware(redisClient *redis.Client, ttl time.Duration) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// Wrap wraps an HTTP handler with idempotency checks
func (i *IdempotencyMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract idempotency key from header
		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}

		// Check if we've seen this request before
		cachedResponse, err := i.redisClient.GetIdempotencyKey(r.Context(), idempotencyKey)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Redis error during idempotency check: %v\n", err)
		}

		if cachedResponse != nil {
			// Return cached response
			var response models.TransactionResponse
			if err := json.Unmarshal(cachedResponse, &response); err != nil {
				http.Error(w, "Invalid cached response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotency-Cache", "true")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Create a response recorder to capture the response
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           make([]byte, 0),
		}

		// Process the request
		next.ServeHTTP(recorder, r)

		// If successful, cache the response
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			response := models.TransactionResponse{
				Status:    "success",
				Message:   "Transaction processed",
				Timestamp: time.Now(),
			}

			// Try to extract transaction ID from response body if available
			var txnResponse map[string]interface{}
			if json.Unmarshal(recorder.body, &txnResponse) == nil {
				if id, ok := txnResponse["id"].(string); ok {
					response.ID = id
				}
			}

			// Cache the response
			if err := i.redisClient.SetIdempotencyKey(r.Context(), idempotencyKey, response, i.ttl); err != nil {
				fmt.Printf("Failed to cache idempotency response: %v\n", err)
			}
		}
	}
}

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)
	return r.ResponseWriter.Write(data)
}
