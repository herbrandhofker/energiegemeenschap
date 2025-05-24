package collector

import (
	"context"
	"log"
	"os"

	"tibber_loader/internal/client"
	"tibber_loader/internal/db"
	"tibber_loader/internal/model"
	"tibber_loader/internal/service_db"

	"github.com/joho/godotenv"
)

// LoadPrices loads current and future price information
func LoadPrices(ctx context.Context) {
	log.Printf("Loading price data...")

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

	// Haal Tibber API token op
	token := os.Getenv("TIBBER_API_TOKEN")
	if token == "" {
		log.Fatal("TIBBER_API_TOKEN environment variable is not set")
	}

	// Maak clients en services
	apiClient := client.NewClient(token)
	homeService := &service_db.HomeService{
		Client: apiClient,
		DB:     dbConn,
	}
	priceService := &service_db.PriceService{
		Client: apiClient,
		DB:     dbConn,
	}

	// Get homes with production capability
	homes, err := homeService.GetHomesWithProductionCapability(ctx)
	if err != nil {
		log.Printf("Error fetching homes: %v", err)
		return
	}

	// Load prices for each home
	for _, home := range homes {
		if _, err := priceService.GetPrices(ctx, home.Id); err != nil {
			log.Printf("Error loading prices for home %s: %v", home.Id, err)
		}
	}

	log.Printf("Finished loading price data")
}

// LoadConsumptionAndProduction loads historical consumption and production data
func LoadConsumptionAndProduction(ctx context.Context) {
	log.Printf("Loading consumption and production data...")

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

	// Haal Tibber API token op
	token := os.Getenv("TIBBER_API_TOKEN")
	if token == "" {
		log.Fatal("TIBBER_API_TOKEN environment variable is not set")
	}

	// Maak clients en services
	apiClient := client.NewClient(token)
	homeService := &service_db.HomeService{
		Client: apiClient,
		DB:     dbConn,
	}

	// Get homes with production capability
	homes, err := homeService.GetHomesWithProductionCapability(ctx)
	if err != nil {
		log.Printf("Error fetching homes: %v", err)
		return
	}

	// Load data for each home
	for _, home := range homes {
		// Load consumption data
		variables := map[string]interface{}{
			"homeId":     home.Id,
			"resolution": "DAILY",
			"last":       30,
		}
		if _, err := apiClient.QueryAPI(ctx, model.ConsumptionQuery, variables); err != nil {
			log.Printf("Error loading consumption data for home %s: %v", home.Id, err)
		}

		// Load production data
		if _, err := apiClient.QueryAPI(ctx, model.ProductionQuery, variables); err != nil {
			log.Printf("Error loading production data for home %s: %v", home.Id, err)
		}
	}

	log.Printf("Finished loading consumption and production data")
}
