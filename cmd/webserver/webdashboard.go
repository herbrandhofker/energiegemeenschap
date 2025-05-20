package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"ws/internal/client"
	"ws/internal/model"
	"ws/internal/service"
	"ws/internal/tibber"

	"github.com/go-chi/chi/v5"
)

// LiveData represents the data structure for SSE updates
type LiveData struct {
	Timestamp              time.Time `json:"timestamp"`
	Power                  float64   `json:"power"`
	PowerProduction        float64   `json:"powerProduction"`
	AccumulatedConsumption float64   `json:"accumulatedConsumption"`
	AccumulatedProduction  float64   `json:"accumulatedProduction"`
}

// WebDashboard represents the web dashboard
type WebDashboard struct {
	Port   int
	Router *chi.Mux
	Title  string

	// Services
	Client         *client.TibberClient
	HomeService    *service.HomeService
	ConsumptionSvc *service.ConsumptionService
	ProductionSvc  *service.ProductionService
	PriceSvc       *service.PriceService

	// State
	Homes    []model.Home
	AllHomes []model.Home

	Templates *template.Template

	// Price refresh
	priceRefreshTicker  *time.Ticker
	priceRefreshStopCh  chan struct{}
	priceRefreshWg      sync.WaitGroup
	priceUpdateChannels sync.Map // Maps homeID+clientAddr to notification channel

	// Nightly energy data refresh
	energyRefreshTicker *time.Ticker
	energyRefreshStopCh chan struct{}
	energyRefreshWg     sync.WaitGroup

	ApiEndPoint  string
	ApiToken     string
	TibberClient *tibber.Client
	Wg           *sync.WaitGroup

	// Live data
	liveDataChannels sync.Map // Maps client ID to channel
	ctx              context.Context
}

// de constructor voor WebDashboard
func NewWebDashboard(title string, port int, apiEndPoint, apiToken string) (*WebDashboard, error) {
	// Create GraphQL client for regular API calls
	graphqlClient := client.NewClient(apiToken)

	// Create services
	homeService := &service.HomeService{Client: graphqlClient}
	consumptionSvc := &service.ConsumptionService{Client: graphqlClient}
	productionSvc := &service.ProductionService{Client: graphqlClient}
	priceSvc := &service.PriceService{Client: graphqlClient}

	// Create websocket client
	wsClient := tibber.NewClient(apiToken, os.Getenv("TIBBER_HOUSE_ID"))

	wd := &WebDashboard{
		// Server configuration
		Port:   port,
		Title:  title,
		Router: chi.NewRouter(),
		// Services
		Client:         graphqlClient,
		HomeService:    homeService,
		ConsumptionSvc: consumptionSvc,
		ProductionSvc:  productionSvc,
		PriceSvc:       priceSvc,
		TibberClient:   wsClient,
		Wg:             &sync.WaitGroup{},
		ApiEndPoint:    apiEndPoint,
		ApiToken:       apiToken,
	}

	// Laad templates
	templates, err := wd.loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}
	wd.Templates = templates

	// Setup routes
	wd.setupRoutes()

	return wd, nil
}

// loadTemplates laadt HTML templates en geeft deze terug
func (wd *WebDashboard) loadTemplates() (*template.Template, error) {
	// Pad naar templates
	templatesPath := "internal/web/templates"
	layoutPath := filepath.Join(templatesPath, "layout.html")
	partialsPath := filepath.Join(templatesPath, "partials")

	// Debug: bekijk welke partials beschikbaar zijn
	partialFiles, err := filepath.Glob(filepath.Join(partialsPath, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("error bij zoeken naar partials: %w", err)
	}

	// FuncMap voor template functies
	funcMap := template.FuncMap{
		"now": time.Now,
		"formatCents": func(price float64) string {
			// Vermenigvuldig met 100 om naar centen te converteren
			// Gebruik strconv om komma als decimaalteken te krijgen
			return strings.Replace(fmt.Sprintf("%.1f", price*100), ".", ",", 1)
		},
		"formatTime": func(timeStr string) string {
			// Parse de ISO 8601 timestamp
			t, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				return "ongeldige tijd"
			}

			// Formatteer naar alleen uur
			return fmt.Sprintf("%d uur", t.Hour())
		},
		// Eenvoudigere versie voor directe formatting
		"replaceDate": func(dateStr string) string {
			// Vervang Engelse maandnamen door Nederlandse
			replacements := map[string]string{
				"January":   "januari",
				"February":  "februari",
				"March":     "maart",
				"April":     "april",
				"May":       "mei",
				"June":      "juni",
				"July":      "juli",
				"August":    "augustus",
				"September": "september",
				"October":   "oktober",
				"November":  "november",
				"December":  "december",
			}

			result := dateStr
			for eng, nl := range replacements {
				result = strings.Replace(result, eng, nl, 1)
			}
			return result
		},
	}

	// Begin met een lege template en voeg de layout toe
	t, err := template.New("").Funcs(funcMap).ParseFiles(layoutPath)
	if err != nil {
		return nil, fmt.Errorf("error bij parsen van layout: %w", err)
	}

	// Voeg alle partials toe
	if len(partialFiles) > 0 {
		t, err = t.ParseFiles(partialFiles...)
		if err != nil {
			return nil, fmt.Errorf("error bij parsen van partials: %w", err)
		}
	}

	return t, nil
}

// Start de web server
func (wd *WebDashboard) Start() error {
	// Haal homes op
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	homes, err := wd.HomeService.GetHomeDetails(ctx)
	if err != nil {
		return fmt.Errorf("error fetching homes: %w", err)
	}

	if len(homes) == 0 {
		return fmt.Errorf("no homes found in your Tibber account")
	}

	// Filter homes met consumptiegegevens
	homesWithConsumption := make([]model.Home, 0)

	for _, home := range homes {
		consumptionCtx, cancelConsumption := context.WithTimeout(context.Background(), 5*time.Second)
		// Gebruik de juiste parameters voor GetConsumption
		homeWithConsumption, err := wd.ConsumptionSvc.GetConsumption(consumptionCtx, home.Id, "DAILY", 1)
		cancelConsumption()

		// Controleer op fouten en of er consumptiegegevens zijn
		if err == nil && homeWithConsumption != nil && len(homeWithConsumption.Consumption) > 0 {
			homesWithConsumption = append(homesWithConsumption, home)
		}
	}

	wd.Homes = homesWithConsumption
	wd.AllHomes = homes

	// Start prijsverversing
	wd.startPriceRefresh()

	// Start nachtelijke data verversing
	wd.startEnergyRefresh()

	// Start de server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", wd.Port),
		Handler:      wd.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Setup clean shutdown
	wd.setupSignalHandler(server)

	// Create context for graceful shutdown
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Start the Tibber websocket connection
	wd.StartTibberWebsocket(ctx)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// setupSignalHandler configureert graceful shutdown
func (wd *WebDashboard) setupSignalHandler(server *http.Server) {
	// Signaalkanaal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Wacht op signaal
		<-stop
		log.Println("\nAfsluiten... even geduld aub")

		// Stop verversingsmechanismen
		wd.stopPriceRefresh()
		wd.stopEnergyRefresh()

		// Server shutdown
		log.Println("HTTP server afsluiten...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.Shutdown(ctx)
		log.Println("Server netjes afgesloten!")
		os.Exit(0)
	}()
}

// setupRoutes is now defined in handlers.go

// startPriceRefresh start een uurlijkse prijsverversing
func (wd *WebDashboard) startPriceRefresh() {
	wd.priceRefreshTicker = time.NewTicker(1 * time.Hour)
	wd.priceRefreshStopCh = make(chan struct{})
	wd.priceRefreshWg.Add(1)

	go func() {
		defer wd.priceRefreshWg.Done()
		log.Println("Uurlijkse prijsverversing gestart")

		for {
			select {
			case <-wd.priceRefreshTicker.C:
				// Ververs prijzen voor alle huizen
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				for _, home := range wd.Homes {
					homeID := home.Id
					log.Printf("Verversing prijzen voor home %s", homeID)

					_, err := wd.PriceSvc.GetPrices(ctx, homeID)
					if err != nil {
						log.Printf("Error bij verversen prijzen voor %s: %v", homeID, err)
					} else {
						log.Printf("✅ Prijzen voor %s succesvol ververst", homeID)

						// Notify all connected clients for this home
						wd.priceUpdateChannels.Range(func(key, value interface{}) bool {
							if strings.HasPrefix(key.(string), homeID) {
								if ch, ok := value.(chan struct{}); ok {
									// Non-blocking send
									select {
									case ch <- struct{}{}:
									default:
									}
								}
							}
							return true
						})
					}
				}
				cancel()

			case <-wd.priceRefreshStopCh:
				wd.priceRefreshTicker.Stop()
				return
			}
		}
	}()
}

// stopPriceRefresh stopt de uurlijkse prijsverversing
func (wd *WebDashboard) stopPriceRefresh() {
	if wd.priceRefreshTicker != nil {
		close(wd.priceRefreshStopCh)
		wd.priceRefreshWg.Wait()
		log.Println("Uurlijkse prijsverversing gestopt")
	}
}

// startEnergyRefresh start een nachtelijke energiedata verversing
func (wd *WebDashboard) startEnergyRefresh() {
	// Bereken tijd tot 2 uur 's nachts
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	initialDelay := time.Until(nextRun)

	// Start een timer voor de eerste run
	initialTimer := time.NewTimer(initialDelay)
	wd.energyRefreshStopCh = make(chan struct{})
	wd.energyRefreshWg.Add(1)

	go func() {
		defer wd.energyRefreshWg.Done()

		// Wacht op de eerste timer
		select {
		case <-initialTimer.C:
			// Start de refresh en dan de periodieke timer
			wd.refreshAllEnergyData()
			wd.energyRefreshTicker = time.NewTicker(24 * time.Hour)

		case <-wd.energyRefreshStopCh:
			initialTimer.Stop()
			return
		}

		// Begin de dagelijkse cyclus
		for {
			select {
			case <-wd.energyRefreshTicker.C:
				wd.refreshAllEnergyData()

			case <-wd.energyRefreshStopCh:
				if wd.energyRefreshTicker != nil {
					wd.energyRefreshTicker.Stop()
				}
				return
			}
		}
	}()
}

// refreshAllEnergyData ververs alle consumptie- en productiegegevens
func (wd *WebDashboard) refreshAllEnergyData() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Ververs consumptie voor alle homes
	for _, home := range wd.Homes {
		homeID := home.Id

		// Haal consumptiegegevens op (7 dagen)
		_, err := wd.ConsumptionSvc.GetConsumption(ctx, homeID, "DAILY", 7)
		if err != nil {
			continue
		}

		// Controleer of het huis productie-mogelijkheden heeft
		if home.MeteringPointData.ProductionEan != "" {
			// Haal productiegegevens op (7 dagen)
			_, err := wd.ProductionSvc.GetProduction(ctx, homeID, "DAILY", 7)
			if err != nil {
				continue
			}
		}

		// Kort pauze om de API niet te overbelasten
		time.Sleep(500 * time.Millisecond)
	}
}

// stopEnergyRefresh stopt de nachtelijke energiedata verversing
func (wd *WebDashboard) stopEnergyRefresh() {
	if wd.energyRefreshTicker != nil || wd.energyRefreshStopCh != nil {
		close(wd.energyRefreshStopCh)
		wd.energyRefreshWg.Wait()
		log.Println("Dagelijkse energiedata verversing gestopt")
	}
}

func (wd *WebDashboard) StartTibberWebsocket(ctx context.Context) {
	// Start subscription
	wd.TibberClient.Wg.Add(1)
	go wd.TibberClient.Subscribe(ctx)

	// Handle incoming measurements
	lastMeasurement := time.Time{}
	go func() {
		for {
			select {
			case measurement := <-wd.TibberClient.WebsocketClient.Data:
				if !measurement.Timestamp.Equal(lastMeasurement) {
					// Convert measurement to LiveData
					liveData := LiveData{
						Timestamp:              measurement.Timestamp,
						Power:                  measurement.Power,
						PowerProduction:        measurement.PowerProduction,
						AccumulatedConsumption: measurement.AccumulatedConsumption,
						AccumulatedProduction:  measurement.AccumulatedProduction,
					}

					// Broadcast to all connected SSE clients
					wd.liveDataChannels.Range(func(key, value interface{}) bool {
						if ch, ok := value.(chan LiveData); ok {
							select {
							case ch <- liveData:
							default:
								// Channel is full or closed, remove it
								wd.liveDataChannels.Delete(key)
							}
						}
						return true
					})

					lastMeasurement = measurement.Timestamp
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (wd *WebDashboard) handleTibberWebsocket() {
	for {
		select {
		case measurement := <-wd.TibberClient.WebsocketClient.Data:
			// Convert measurement to LiveData
			liveData := LiveData{
				Timestamp:              measurement.Timestamp,
				Power:                  measurement.Power,
				PowerProduction:        measurement.PowerProduction,
				AccumulatedConsumption: measurement.AccumulatedConsumption,
				AccumulatedProduction:  measurement.AccumulatedProduction,
			}

			// Send to all connected clients
			wd.liveDataChannels.Range(func(key, value interface{}) bool {
				if ch, ok := value.(chan LiveData); ok {
					select {
					case ch <- liveData:
					default:
						// Channel is full or closed, remove it
						wd.liveDataChannels.Delete(key)
					}
				}
				return true
			})
		case <-wd.ctx.Done():
			return
		}
	}
}
