package collector

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"ws/internal/client"
	"ws/internal/db"
	"ws/internal/model"
	"ws/internal/service_db"
	"ws/internal/tibber"

	"github.com/joho/godotenv"
)

// RunRealTimeCollector is de real-time data collector
func RunRealTimeCollector(ctx context.Context) {
	log.Printf("Starting real-time collector...")

	// Laad .env bestand
	if err := godotenv.Load("./.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Printf("Loaded .env file")

	// Haal database URL op
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

	// Initialiseer database schema
	if err := db.InitSchema(dbConn); err != nil {
		log.Fatalf("Error initializing database schema: %v", err)
	}
	log.Printf("Initialized database schema")

	// Haal Tibber API token en huis ID op
	token := os.Getenv("TIBBER_API_TOKEN")
	if token == "" {
		log.Fatal("TIBBER_API_TOKEN environment variable is not set")
	}

	houseID := os.Getenv("TIBBER_HOUSE_ID")
	if houseID == "" {
		log.Fatal("TIBBER_HOUSE_ID environment variable is not set")
	}
	log.Printf("Found Tibber credentials for house ID: %s", houseID)

	// Maak Tibber clients
	wsClient := tibber.NewClient(token, houseID)
	apiClient := client.NewClient(token)
	log.Printf("Created Tibber clients")

	// Verifieer toegang tot Tibber API
	if err := wsClient.VerifyAccess(); err != nil {
		log.Fatalf("Error verifying Tibber access: %v", err)
	}
	log.Printf("Verified Tibber API access")

	// Maak services
	homeService := &service_db.HomeService{
		Client: apiClient,
		DB:     dbConn,
	}

	// Start real-time collection for each home
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Get homes with production capability
			homes, err := homeService.GetHomesWithProductionCapability(ctx)
			if err != nil {
				log.Printf("Error fetching homes: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Start real-time collection for each home
			for _, home := range homes {
				go func(h model.Home) {
					if err := collectRealTimeData(ctx, wsClient, dbConn, h); err != nil {
						log.Printf("Error collecting real-time data for home %s: %v", h.Id, err)
					}
				}(home)
			}

			// Wait before next update
			time.Sleep(5 * time.Minute)
		}
	}
}

// collectRealTimeData collects real-time data for a specific home
func collectRealTimeData(ctx context.Context, client *tibber.Client, db *sql.DB, home model.Home) error {
	// ... rest of the code ...
	return nil
}
