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

	"ingestion-service/internal/config"
	"ingestion-service/internal/models"
	"ingestion-service/internal/publisher"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Setup Kafka producer
	producer, err := publisher.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}
	defer producer.Close() // Make sure Close() is exported in publisher

	// Setup HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", IngestHandler(producer, cfg.KafkaTopic))

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPORT,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Ingestion service running on port %s", cfg.HTTPPORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down ingestion...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// IngestHandler accepts a JSON transaction and publishes it to Kafka
func IngestHandler(p *publisher.Producer, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var txn models.Transaction
		if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		payload, err := json.Marshal(txn)
		if err != nil {
			http.Error(w, "failed to serialize payload", http.StatusInternalServerError)
			return
		}

		if err := p.Publish(topic, payload); err != nil {
			http.Error(w, "failed to enqueue event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("enqueued"))
	}
}
