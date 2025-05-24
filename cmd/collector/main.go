package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tibber_loader/internal/collector"
	"tibber_loader/internal/db"

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

	// Start rea
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	log.Printf("Found DATABASE_URL")

	// Parse database URL en maak verbinding
	dbConfig, err := db.ParseURL(dbURL)
	if err != nil {
		log.Fatalf("Error parsing database URL: %v", err)
	}

	dbConn, err := db.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbConn.Close()
	log.Printf("Connected to database")
	go collector.Collector(ctx, dbConn)
	// Wait for shutdown signal
	<-sigChan
	cancel()
}
