package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"processing-service/internal/config"
	"processing-service/internal/consumer"
	"processing-service/internal/processor"
	"processing-service/internal/publisher"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting processing service with config: %+v", cfg)

	// Initialize Prometheus metrics
	initMetrics()

	// Create publisher for processed transactions
	pub := publisher.NewPublisher(cfg.KafkaBrokers, cfg.OutputTopic)
	defer pub.Close()

	// Create processor with business rules
	proc := processor.NewProcessor(pub)

	// Create consumer for raw transactions
	cons, err := consumer.NewConsumer(cfg.KafkaBrokers, cfg.InputTopic, cfg.ConsumerGroup, proc)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer cons.Close()

	// Start metrics server if enabled
	if cfg.MetricsEnabled {
		go startMetricsServer(cfg.MetricsPort)
	}

	// Run consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := cons.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down processing-service...")
	cancel()

	// Give some time for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout, forcing exit")
	case <-time.After(5 * time.Second):
		log.Println("Graceful shutdown completed")
	}
}

// Prometheus metrics
var (
	transactionsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_processed_total",
			Help: "Total number of transactions processed",
		},
		[]string{"status", "risk_level"},
	)

	processingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_processing_duration_seconds",
			Help:    "Duration of transaction processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	processingErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_processing_errors_total",
			Help: "Total number of processing errors",
		},
		[]string{"error_type"},
	)
)

// initMetrics initializes Prometheus metrics
func initMetrics() {
	prometheus.MustRegister(transactionsProcessed)
	prometheus.MustRegister(processingDuration)
	prometheus.MustRegister(processingErrors)
}

// startMetricsServer starts the Prometheus metrics server
func startMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting metrics server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}
