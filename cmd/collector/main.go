package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ws/internal/client"
	"ws/internal/db"

	"ws/internal/service_db"

	"github.com/joho/godotenv"
)

type DataCollector struct {
	// Services
	Client         *client.TibberClient
	HomeService    *service_db.HomeService
	ConsumptionSvc *service_db.ConsumptionService
	ProductionSvc  *service_db.ProductionService
	PriceSvc       *service_db.PriceService

	// State
	Homes []string // Only store home IDs

	// Refresh control
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewDataCollector(apiToken string, db *sql.DB) *DataCollector {
	client := client.NewClient(apiToken)
	return &DataCollector{
		Client:         client,
		HomeService:    &service_db.HomeService{Client: client, DB: db},
		ConsumptionSvc: &service_db.ConsumptionService{Client: client, DB: db},
		ProductionSvc:  &service_db.ProductionService{Client: client, DB: db},
		PriceSvc:       &service_db.PriceService{Client: client, DB: db},
		stopChan:       make(chan struct{}),
	}
}

func (dc *DataCollector) Start() error {
	// Get initial homes
	if err := dc.updateHomesList(); err != nil {
		return fmt.Errorf("failed to get initial homes list: %w", err)
	}

	// Start collection routines
	dc.wg.Add(2)
	go dc.runPriceCollection()
	go dc.runEnergyCollection()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down collector...")
	close(dc.stopChan)
	dc.wg.Wait()
	log.Println("Collector stopped")

	return nil
}

func (dc *DataCollector) updateHomesList() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get homes from service (will fetch from DB or API as needed)
	homes, err := dc.HomeService.GetHomeDetails(ctx)
	if err != nil {
		return err
	}

	// Store only home IDs
	dc.Homes = make([]string, len(homes))
	for i, home := range homes {
		dc.Homes[i] = home.Id
	}

	return nil
}

func (dc *DataCollector) runPriceCollection() {
	defer dc.wg.Done()
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	dc.collectPrices()

	for {
		select {
		case <-dc.stopChan:
			return
		case <-ticker.C:
			dc.collectPrices()
		}
	}
}

func (dc *DataCollector) runEnergyCollection() {
	defer dc.wg.Done()

	// Run immediately on start
	dc.collectEnergyData()

	// Calculate time until next 2 AM
	now := time.Now()
	next2AM := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(next2AM) {
		next2AM = next2AM.Add(24 * time.Hour)
	}
	timer := time.NewTimer(time.Until(next2AM))
	defer timer.Stop()

	for {
		select {
		case <-dc.stopChan:
			return
		case <-timer.C:
			dc.collectEnergyData()
			// Reset timer for next day
			timer.Reset(24 * time.Hour)
		}
	}
}

func (dc *DataCollector) collectPrices() {
	log.Println("Collecting prices for all homes...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, homeID := range dc.Homes {
		if _, err := dc.PriceSvc.GetPrices(ctx, homeID); err != nil {
			log.Printf("Error collecting prices for home %s: %v", homeID, err)
		} else {
			log.Printf("Successfully collected prices for home %s", homeID)
		}
		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

func (dc *DataCollector) collectEnergyData() {
	log.Println("Collecting energy data for all homes...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, homeID := range dc.Homes {
		// Collect consumption data
		if _, err := dc.ConsumptionSvc.GetConsumption(ctx, homeID, "DAILY", 7); err != nil {
			log.Printf("Error collecting consumption for home %s: %v", homeID, err)
		} else {
			log.Printf("Successfully collected consumption for home %s", homeID)
		}

		// Collect production data
		if _, err := dc.ProductionSvc.GetProduction(ctx, homeID, "DAILY", 7); err != nil {
			log.Printf("Error collecting production for home %s: %v", homeID, err)
		} else {
			log.Printf("Successfully collected production for home %s", homeID)
		}

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

func main() {
	// Load .env file from root directory
	if err := godotenv.Load("./.env"); err != nil {
		fmt.Printf("⚠️ Error: Could not load .env file: %v\n", err)
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Printf("DATABASE_URL environment variable is required\n")
		os.Exit(1)
	}

	// Initialize database
	dbConfig, err := db.ParseURL(dbURL)
	if err != nil {
		fmt.Printf("❌ Error parsing database URL: %v\n", err)
		os.Exit(1)
	}

	dbConn, err := db.NewConnection(dbConfig)
	if err != nil {
		fmt.Printf("❌ Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	// Initialize schema
	if err := db.InitSchema(dbConn); err != nil {
		fmt.Printf("❌ Error initializing database schema: %v\n", err)
		os.Exit(1)
	}

	// Continue with normal collector operation
	apiToken := os.Getenv("TIBBER_API_TOKEN")
	if apiToken == "" {
		fmt.Printf("TIBBER_API_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	// Create and start collector
	collector := NewDataCollector(apiToken, dbConn)
	if err := collector.Start(); err != nil {
		fmt.Printf("❌ Error running collector: %v\n", err)
		os.Exit(1)
	}
}
