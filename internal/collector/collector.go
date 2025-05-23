package collector

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"ws/internal/client"
	"ws/internal/db"
	"ws/internal/model"
	"ws/internal/service_db"
	"ws/internal/tibber"

	"github.com/lib/pq"
)

// activeCollectors keeps track of active collectors
var (
	activeCollectors = make(map[int]context.CancelFunc)
	collectorMutex   sync.RWMutex
)

// TokenChangeEvent represents a token change notification
type TokenChangeEvent struct {
	Action  string `json:"action"`
	TokenID int    `json:"token_id"`
	Active  bool   `json:"active"`
}

// listenForTokenChanges listens for changes in the tibber_tokens table
func listenForTokenChanges(ctx context.Context, dbConn *sql.DB) {
	// Create a new listener
	listener := pq.NewListener(os.Getenv("DATABASE_URL"), 10*time.Second, time.Minute, nil)
	err := listener.Listen("token_changes")
	if err != nil {
		log.Fatalf("Error listening for token changes: %v", err)
	}
	defer listener.Close()

	log.Printf("Listening for token changes...")

	for {
		select {
		case notification := <-listener.Notify:
			var event TokenChangeEvent
			if err := json.Unmarshal([]byte(notification.Extra), &event); err != nil {
				log.Printf("Error parsing token change event: %v", err)
				continue
			}

			log.Printf("Received token change event: %+v", event)

			switch event.Action {
			case "INSERT", "UPDATE":
				if event.Active {
					// Start collector for this token
					go startCollectorForToken(ctx, dbConn, event.TokenID)
				} else {
					// Stop collector for this token
					stopCollectorForToken(event.TokenID)
				}
			}
		case <-time.After(90 * time.Second):
			// Check connection
			if err := listener.Ping(); err != nil {
				log.Printf("Listener connection lost: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// startCollectorForToken starts the collector for a specific token
func startCollectorForToken(ctx context.Context, dbConn *sql.DB, tokenID int) {
	// Check if collector is already running
	collectorMutex.RLock()
	if _, exists := activeCollectors[tokenID]; exists {
		collectorMutex.RUnlock()
		log.Printf("Collector for token %d is already running", tokenID)
		return
	}
	collectorMutex.RUnlock()

	// Create a new context for this collector
	collectorCtx, cancel := context.WithCancel(ctx)

	// Store the cancel function
	collectorMutex.Lock()
	activeCollectors[tokenID] = cancel
	collectorMutex.Unlock()

	// Get token from database
	var token string
	err := dbConn.QueryRowContext(collectorCtx, `
		SELECT token 
		FROM tibber.tibber_tokens 
		WHERE id = $1
	`, tokenID).Scan(&token)
	if err != nil {
		log.Printf("Error getting token: %v", err)
		stopCollectorForToken(tokenID)
		return
	}

	// Create Tibber clients
	wsClient := tibber.NewClient(token, "")
	apiClient := client.NewClient(token)

	// Verify access to Tibber API
	if err := wsClient.VerifyAccess(); err != nil {
		log.Printf("Error verifying Tibber access: %v", err)
		stopCollectorForToken(tokenID)
		return
	}

	// Create services
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

	// Start collector for each home
	homes, err := homeService.GetHomes(collectorCtx)
	if err != nil {
		log.Printf("Error getting homes: %v", err)
		stopCollectorForToken(tokenID)
		return
	}

	for _, home := range homes {
		go startHomeCollector(collectorCtx, home.Id, wsClient, apiClient, priceService, consumptionService, productionService, realTimeService)
	}
}

// stopCollectorForToken stops the collector for a specific token
func stopCollectorForToken(tokenID int) {
	collectorMutex.Lock()
	defer collectorMutex.Unlock()

	if cancel, exists := activeCollectors[tokenID]; exists {
		cancel()
		delete(activeCollectors, tokenID)
		log.Printf("Stopped collector for token %d", tokenID)
	} else {
		log.Printf("No active collector found for token %d", tokenID)
	}
}

// Collector is de real-time data collector
func Collector(ctx context.Context, dbConn *sql.DB) {
	log.Printf("Starting collector...")

	// Initialiseer database schema
	if err := db.InitDatabase(dbConn); err != nil {
		log.Fatalf("Error initializing database schema: %v", err)
	}
	log.Printf("Initialized database schema")

	// Start cleanup goroutine
	go cleanupOldMeasurements(ctx, dbConn)

	// Start listening for token changes
	go listenForTokenChanges(ctx, dbConn)

	// Keep the main goroutine running
	<-ctx.Done()
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

// startHomeCollector starts the collector for a specific home
func startHomeCollector(
	ctx context.Context,
	homeID string,
	wsClient *tibber.Client,
	apiClient *client.Client,
	priceService *service_db.PriceService,
	consumptionService *service_db.ConsumptionService,
	productionService *service_db.ProductionService,
	realTimeService *service_db.RealTimeService,
) {
	// Load initial data
	if _, err := priceService.GetPrices(ctx, homeID); err != nil {
		log.Printf("Error loading initial prices for home %s: %v", homeID, err)
	}

	if _, err := consumptionService.GetConsumption(ctx, homeID, "DAILY", 30); err != nil {
		log.Printf("Error loading initial consumption data for home %s: %v", homeID, err)
	}

	if _, err := productionService.GetProduction(ctx, homeID, "DAILY", 30); err != nil {
		log.Printf("Error loading initial production data for home %s: %v", homeID, err)
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
			if err := realTimeService.StoreMeasurement(ctx, homeID, measurement); err != nil {
				log.Printf("Error storing measurement for home %s: %v", homeID, err)
			}
		}
	}
}
