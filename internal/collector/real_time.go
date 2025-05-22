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

	// Verifieer toegang tot Tibber API
	if err := wsClient.VerifyAccess(); err != nil {
		log.Fatalf("Error verifying Tibber access: %v", err)
	}

	// Maak services
	homeService := &service_db.HomeService{
		Client: apiClient,
		DB:     dbConn,
	}
	priceService := &service_db.PriceService{
		Client: apiClient,
		DB:     dbConn,
	}
	consumptionService := &service_db.ConsumptionService{
		Client: apiClient,
		DB:     dbConn,
	}
	productionService := &service_db.ProductionService{
		Client: apiClient,
		DB:     dbConn,
	}
	realTimeService := &service_db.RealTimeService{
		DB: dbConn,
	}

	// Start cleanup goroutine
	go cleanupOldMeasurements(ctx, dbConn)

	// Start a goroutine for each home
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
				home := home // Create new variable for goroutine
				go func() {
					// Load initial data
					if _, err := priceService.GetPrices(ctx, home.Id); err != nil {
						log.Printf("Error loading initial prices for home %s: %v", home.Id, err)
					}

					if _, err := consumptionService.GetConsumption(ctx, home.Id, "DAILY", 30); err != nil {
						log.Printf("Error loading initial consumption data for home %s: %v", home.Id, err)
					}

					if _, err := productionService.GetProduction(ctx, home.Id, "DAILY", 30); err != nil {
						log.Printf("Error loading initial production data for home %s: %v", home.Id, err)
					}

					// Start WebSocket subscription
					wsClient.Wg.Add(1)
					go wsClient.Subscribe(ctx)

					// Process measurements
					for {
						select {
						case <-ctx.Done():
							return
						case measurement := <-wsClient.WebsocketClient.Data:
							if err := realTimeService.StoreMeasurement(ctx, home.Id, measurement); err != nil {
								log.Printf("Error storing measurement for home %s: %v", home.Id, err)
							}
						}
					}
				}()
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

// cleanupOldMeasurements removes measurements older than 24 hours
func cleanupOldMeasurements(ctx context.Context, dbConn *sql.DB) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Wait until 3 AM
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(next.Sub(now))

			// Delete measurements older than 24 hours
			query := `
				DELETE FROM real_time_measurements 
				WHERE timestamp < NOW() - INTERVAL '24 hours'
			`
			if _, err := dbConn.ExecContext(ctx, query); err != nil {
				log.Printf("Error cleaning up old measurements: %v", err)
			} else {
				log.Printf("Cleaned up measurements older than 24 hours")
			}
		}
	}
}
