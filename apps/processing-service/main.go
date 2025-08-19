package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"processing-service/internal/config"
	"processing-service/internal/consumer"
	"processing-service/internal/processor"
	"processing-service/internal/publisher"
)

func main() {
	// load configuration
	cfg := config.LoadConfig()

	// publisher for processed transactions
	pub := publisher.NewPublisher(cfg.KafkaBrokers, cfg.OutputTopic)
	defer pub.Close()

	// processor with business rules
	proc := processor.NewProcessor(pub)

	// consumer for raw transactions
	cons, err := consumer.NewConsumer(cfg.KafkaBrokers, cfg.InputTopic, cfg.ConsumerGroup, proc)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer cons.Close()

	// run consumer
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := cons.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down processing-service...")
	cancel()
}
