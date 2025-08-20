package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"storage-service/internal/config"
	"storage-service/internal/consumer"
	"storage-service/internal/handler"
	"storage-service/internal/storage"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect DB
	store, err := storage.NewStorage(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer store.Close()

	// Initialize handler
	txHandler := handler.NewTransactionHandler(store)

	// Setup Kafka consumer
	cons := consumer.NewConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.InputTopic, txHandler)
	defer cons.Close()

	// Run consumer
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := cons.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down storage-service...")
	cancel()
}
