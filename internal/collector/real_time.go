package collector

import (
	"context"
	"log"
	"os"
	"time"

	"ws/internal/db"
	"ws/internal/service_db"
	"ws/internal/tibber"

	"github.com/joho/godotenv"
)

// RunRealTimeCollector is de real-time data collector
func RunRealTimeCollector(ctx context.Context) {
	// Laad .env bestand
	if err := godotenv.Load("./.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Haal database URL op
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

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

	// Initialiseer database schema
	if err := db.InitSchema(dbConn); err != nil {
		log.Fatalf("Error initializing database schema: %v", err)
	}

	// Haal Tibber API token en huis ID op
	token := os.Getenv("TIBBER_API_TOKEN")
	if token == "" {
		log.Fatal("TIBBER_API_TOKEN environment variable is not set")
	}

	houseID := os.Getenv("TIBBER_HOUSE_ID")
	if houseID == "" {
		log.Fatal("TIBBER_HOUSE_ID environment variable is not set")
	}

	// Maak Tibber client
	client := tibber.NewClient(token, houseID)

	// Verifieer toegang tot Tibber API
	if err := client.VerifyAccess(); err != nil {
		log.Fatalf("Error verifying Tibber access: %v", err)
	}

	// Maak real-time service
	realTimeService := &service_db.RealTimeService{DB: dbConn}

	// Start websocket subscription
	client.Wg.Add(1)
	go client.Subscribe(ctx)

	// Verwerk metingen
	for {
		select {
		case measurement := <-client.WebsocketClient.Data:
			if err := realTimeService.StoreMeasurement(ctx, houseID, measurement); err != nil {
				log.Printf("Error storing measurement: %v", err)
			}

			// Cleanup oude metingen (ouder dan 24 uur)
			if err := realTimeService.CleanupOldMeasurements(ctx, 24*time.Hour); err != nil {
				log.Printf("Error cleaning up old measurements: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
