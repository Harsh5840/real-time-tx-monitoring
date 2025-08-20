package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"alert-service/internal/config"
	"alert-service/internal/consumer"
	"alert-service/internal/handler"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Initialize handler
	alertHandler := handler.NewAlertHandler(cfg.SlackWebhook)

	// Setup Kafka consumer
	cons := consumer.NewConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.InputTopic, alertHandler)
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

	log.Println("Shutting down alert-service...")
	cancel()
}
