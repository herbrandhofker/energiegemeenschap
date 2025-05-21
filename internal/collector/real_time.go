package collector

import (
	"context"
	"log"
	"os"
	"time"

	"ws/internal/client"
	"ws/internal/db"
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
	realTimeService := &service_db.RealTimeService{DB: dbConn}
	userService := &service_db.UserService{
		Client: apiClient,
		DB:     dbConn,
	}
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
	log.Printf("Created all services")

	// Haal initiële data op
	log.Printf("Fetching initial data...")

	// Haal users op
	user, err := userService.GetUserData(ctx)
	if err != nil {
		log.Printf("Error fetching user data: %v", err)
	} else {
		if err := userService.StoreUserData(ctx, user); err != nil {
			log.Printf("Error storing user data: %v", err)
		} else {
			log.Printf("Successfully stored user data for %s", user.Email)
		}
	}

	// Haal homes op
	homes, err := homeService.GetHomeDetails(ctx)
	if err != nil {
		log.Printf("Error fetching homes: %v", err)
	} else {
		log.Printf("Successfully fetched %d homes", len(homes))
	}

	// Voor elk huis, haal prijzen, consumptie en productie op
	for _, home := range homes {
		// Haal prijzen op
		_, err := priceService.GetPrices(ctx, home.Id)
		if err != nil {
			log.Printf("Error fetching prices for home %s: %v", home.Id, err)
		} else {
			log.Printf("Successfully fetched prices for home %s", home.Id)
		}

		// Haal consumptie op (laatste 7 dagen)
		_, err = consumptionService.GetConsumption(ctx, home.Id, "DAILY", 7)
		if err != nil {
			log.Printf("Error fetching consumption for home %s: %v", home.Id, err)
		} else {
			log.Printf("Successfully fetched consumption for home %s", home.Id)
		}

		// Haal productie op als het huis productie-mogelijkheden heeft
		if home.MeteringPointData.ProductionEan != "" {
			_, err = productionService.GetProduction(ctx, home.Id, "DAILY", 7)
			if err != nil {
				log.Printf("Error fetching production for home %s: %v", home.Id, err)
			} else {
				log.Printf("Successfully fetched production for home %s", home.Id)
			}
		}
	}

	// Start websocket subscription
	wsClient.Wg.Add(1)
	go wsClient.Subscribe(ctx)
	log.Printf("Started WebSocket subscription")

	// Verwerk metingen
	log.Printf("Starting measurement processing loop")
	measurementCount := 0
	lastCleanup := time.Now()
	for {
		select {
		case measurement := <-wsClient.WebsocketClient.Data:
			// Store real-time measurement
			if err := realTimeService.StoreMeasurement(ctx, houseID, measurement); err != nil {
				log.Printf("Error storing measurement: %v", err)
			} else {
				measurementCount++
				if measurementCount <= 3 {
					log.Printf("Stored measurement %d: timestamp=%s, power=%.2fW, production=%.2fW", 
						measurementCount,
						measurement.Timestamp.Format(time.RFC3339),
						measurement.Power,
						measurement.PowerProduction)
				} else if measurementCount == 4 {
					log.Printf("Continuing to store measurements silently...")
				}
			}

			// Controleer of het tijd is voor periodieke taken (elk uur)
			now := time.Now()
			if now.Hour() != lastCleanup.Hour() {
				log.Printf("Starting hourly maintenance tasks")

				// Verwijder oude realtime data (ouder dan 24 uur)
				if err := realTimeService.CleanupOldMeasurements(ctx, 24*time.Hour); err != nil {
					log.Printf("Error cleaning up old measurements: %v", err)
				} else {
					log.Printf("Successfully cleaned up measurements older than 24 hours")
				}

				// Update user data
				user, err := userService.GetUserData(ctx)
				if err != nil {
					log.Printf("Error fetching user data: %v", err)
				} else {
					if err := userService.StoreUserData(ctx, user); err != nil {
						log.Printf("Error storing user data: %v", err)
					} else {
						log.Printf("Successfully updated user data for %s", user.Email)
					}
				}

				// Update homes en gerelateerde data
				homes, err := homeService.GetHomeDetails(ctx)
				if err != nil {
					log.Printf("Error updating homes: %v", err)
				} else {
					for _, home := range homes {
						// Update prijzen
						_, err := priceService.GetPrices(ctx, home.Id)
						if err != nil {
							log.Printf("Error updating prices for home %s: %v", home.Id, err)
						}

						// Update consumptie
						_, err = consumptionService.GetConsumption(ctx, home.Id, "DAILY", 7)
						if err != nil {
							log.Printf("Error updating consumption for home %s: %v", home.Id, err)
						}

						// Update productie als beschikbaar
						if home.MeteringPointData.ProductionEan != "" {
							_, err = productionService.GetProduction(ctx, home.Id, "DAILY", 7)
							if err != nil {
								log.Printf("Error updating production for home %s: %v", home.Id, err)
							}
						}
					}
					log.Printf("Successfully updated all homes and related data")
				}

				lastCleanup = now
			}

		case <-ctx.Done():
			log.Printf("Received shutdown signal, stopping collector")
			return
		}
	}
}
