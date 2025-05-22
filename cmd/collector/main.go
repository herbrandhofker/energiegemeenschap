package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ws/internal/collector"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from root directory
	if err := godotenv.Load("./.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start real-time collector in a goroutine
	go collector.Collector(ctx)

	// Wait for shutdown signal
	<-sigChan
	cancel()
}
