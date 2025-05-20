package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// HTTP response helpers
func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("Error response: %d - %s", code, message)
	http.Error(w, message, code)
}

func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding JSON response")
	}
}

// handleHome toont de homepage met dashboard
func (wd *WebDashboard) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": wd.Title,
			"Homes": wd.Homes,
		}

		if err := wd.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error rendering template")
			return
		}
	}
}

// handlePricePartial toont het price partial
func (wd *WebDashboard) handlePricePartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if _, err := wd.findHomeByID(homeID); err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		data := map[string]interface{}{
			"IsActive": false,
			"HomeId":   homeID,
			"Message":  "Dit huis is nog niet aangesloten",
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		homeWithPrices, err := wd.PriceSvc.GetPrices(ctx, homeID)
		if err == nil && homeWithPrices != nil &&
			homeWithPrices.CurrentSubscription != nil &&
			homeWithPrices.CurrentSubscription.PriceInfo.Current.StartTime != "" {

			endTime, err := calculateEndTime(homeWithPrices.CurrentSubscription.PriceInfo.Current.StartTime)
			if err == nil {
				homeWithPrices.CurrentSubscription.PriceInfo.Current.EndTime = endTime
			}

			data["PriceInfo"] = homeWithPrices.CurrentSubscription.PriceInfo
			data["IsActive"] = true
		}

		if err := wd.Templates.ExecuteTemplate(w, "price.html", data); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error rendering template")
		}
	}
}

// handleConsumptionPartial toont het consumption partial
func (wd *WebDashboard) handleConsumptionPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if _, err := wd.findHomeByID(homeID); err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		data := map[string]interface{}{
			"IsActive": false,
			"Message":  "Dit huis is nog niet aangesloten",
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Try daily consumption first
		homeWithConsumption, err := wd.ConsumptionSvc.GetConsumption(ctx, homeID, "DAILY", 7)
		if err != nil || homeWithConsumption == nil || len(homeWithConsumption.Consumption) == 0 {
			// Fallback to hourly data
			homeWithConsumption, err = wd.ConsumptionSvc.GetConsumption(ctx, homeID, "HOURLY", 1)
		}

		if err == nil && homeWithConsumption != nil && len(homeWithConsumption.Consumption) > 0 {
			sortedConsumption := sortConsumptionByDate(homeWithConsumption.Consumption)

			var totalConsumption, totalCost float64
			for _, item := range sortedConsumption {
				totalConsumption += item.Consumption
				totalCost += item.Cost
			}

			data = map[string]interface{}{
				"HomeId":           homeID,
				"Consumption":      sortedConsumption,
				"IsActive":         true,
				"TotalConsumption": totalConsumption,
				"TotalCost":        totalCost,
			}
		}

		if err := wd.Templates.ExecuteTemplate(w, "consumption.html", data); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error rendering template")
		}
	}
}

// handleProductionPartial toont het production partial
func (wd *WebDashboard) handleProductionPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		selectedHome, err := wd.findHomeByID(homeID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		data := map[string]interface{}{
			"IsActive": false,
			"Message":  "Dit huis is nog niet aangesloten",
		}

		if selectedHome.MeteringPointData.ProductionEan == "" {
			if err := wd.Templates.ExecuteTemplate(w, "production.html", data); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error rendering template")
			}
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		homeWithProduction, err := wd.ProductionSvc.GetProduction(ctx, homeID, "DAILY", 7)
		if err == nil && homeWithProduction != nil && len(homeWithProduction.Production) > 0 {
			sortedProduction := sortProductionByDate(homeWithProduction.Production)

			var totalProduction, totalProfit float64
			for _, item := range sortedProduction {
				totalProduction += item.Production
				totalProfit += item.Profit
			}

			data = map[string]interface{}{
				"HomeId":          homeID,
				"Production":      sortedProduction,
				"IsActive":        true,
				"TotalProduction": totalProduction,
				"TotalProfit":     totalProfit,
			}
		}

		if err := wd.Templates.ExecuteTemplate(w, "production.html", data); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error rendering template")
		}
	}
}

// handleHomeDetailsPartial toont gedetailleerde huisgegevens
func (wd *WebDashboard) handleHomeDetailsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		selectedHome, err := wd.findHomeByID(homeID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		if err := wd.Templates.ExecuteTemplate(w, "home_details.html", selectedHome); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error rendering template")
		}
	}
}

// API Handlers

// handlePriceData returns price data for the chart
func (wd *WebDashboard) handlePriceData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if _, err := wd.findHomeByID(homeID); err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		homeWithPrices, err := wd.PriceSvc.GetPrices(ctx, homeID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error fetching price data")
			return
		}

		allPrices := append(
			homeWithPrices.CurrentSubscription.PriceInfo.Today,
			homeWithPrices.CurrentSubscription.PriceInfo.Tomorrow...,
		)

		times := make([]string, len(allPrices))
		prices := make([]float64, len(allPrices))

		for i, price := range allPrices {
			times[i] = price.StartTime
			prices[i] = price.Total
		}

		respondWithJSON(w, map[string]interface{}{
			"times":    times,
			"prices":   prices,
			"currency": homeWithPrices.CurrentSubscription.PriceInfo.Current.Currency,
		})
	}
}

// handleConsumptionData returns consumption data for the chart
func (wd *WebDashboard) handleConsumptionData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		selectedHome, err := wd.findHomeByID(homeID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		if selectedHome.MeteringPointData.ConsumptionEan == "" {
			respondWithError(w, http.StatusNotFound, "No consumption data available")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		homeWithConsumption, err := wd.ConsumptionSvc.GetConsumption(ctx, homeID, "DAILY", 7)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error fetching consumption data")
			return
		}

		if len(homeWithConsumption.Consumption) == 0 {
			respondWithJSON(w, map[string]interface{}{
				"dates":       []string{},
				"consumption": []float64{},
				"cost":        []float64{},
				"currency":    "EUR",
			})
			return
		}

		sortedConsumption := sortConsumptionByDateOldestFirst(homeWithConsumption.Consumption)

		dates := make([]string, len(sortedConsumption))
		consumption := make([]float64, len(sortedConsumption))
		cost := make([]float64, len(sortedConsumption))

		for i, item := range sortedConsumption {
			dates[i] = item.From
			consumption[i] = item.Consumption
			cost[i] = item.Cost
		}

		respondWithJSON(w, map[string]interface{}{
			"dates":       dates,
			"consumption": consumption,
			"cost":        cost,
			"currency":    sortedConsumption[0].Currency,
		})
	}
}

// handleProductionData returns production data for the chart
func (wd *WebDashboard) handleProductionData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		selectedHome, err := wd.findHomeByID(homeID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		if selectedHome.MeteringPointData.ProductionEan == "" {
			respondWithError(w, http.StatusNotFound, "No production data available")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		homeWithProduction, err := wd.ProductionSvc.GetProduction(ctx, homeID, "DAILY", 7)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error fetching production data")
			return
		}

		sortedProduction := sortProductionByDateOldestFirst(homeWithProduction.Production)

		dates := make([]string, len(sortedProduction))
		production := make([]float64, len(sortedProduction))
		profit := make([]float64, len(sortedProduction))

		for i, item := range sortedProduction {
			dates[i] = item.From
			production[i] = item.Production
			profit[i] = item.Profit
		}

		respondWithJSON(w, map[string]interface{}{
			"dates":      dates,
			"production": production,
			"profit":     profit,
		})
	}
}

// ServeLiveData serves live data via Server-Sent Events
func (wd *WebDashboard) ServeLiveData(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a unique ID for this client
	clientID := uuid.New().String()

	// Create a channel for this client
	clientChan := make(chan LiveData)
	wd.liveDataChannels.Store(clientID, clientChan)
	defer func() {
		wd.liveDataChannels.Delete(clientID)
	}()

	// Send initial empty state
	fmt.Fprintf(w, "data: %s\n\n", "{}")
	w.(http.Flusher).Flush()

	// Get the client's context
	ctx := r.Context()

	// Start a goroutine to handle the client's connection
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case measurement := <-clientChan:
				// Convert measurement to JSON
				jsonData, err := json.Marshal(measurement)
				if err != nil {
					continue
				}

				// Send the data to the client
				fmt.Fprintf(w, "data: %s\n\n", jsonData)
				w.(http.Flusher).Flush()
			}
		}
	}()

	// Keep the connection alive
	<-ctx.Done()
}

// setupRoutes configures all HTTP routes
func (wd *WebDashboard) setupRoutes() {
	// Alleen essentiële middleware
	wd.Router.Use(middleware.Recoverer) // Herstel van panics
	wd.Router.Use(middleware.RealIP)    // Real IP detectie
	wd.Router.Use(middleware.CleanPath) // URL path cleaning

	// Static files - direct en simpel
	wd.Router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))

	// Main routes
	wd.Router.Get("/", wd.handleHome())

	// Combineer gerelateerde routes in subrouters
	wd.Router.Route("/partials", func(r chi.Router) {
		r.Get("/{type}/{homeID}", wd.handlePartial()) // Gecombineerde partial handler
	})

	// API endpoints
	wd.Router.Route("/api", func(r chi.Router) {
		r.Get("/{type}/{homeID}", wd.handleData()) // Gecombineerde data handler
	})

	// Server-Sent Events
	wd.Router.Get("/live-data", wd.ServeLiveData)
	wd.Router.Get("/events/price/{homeID}", wd.ServePriceEvents)
}

// Gecombineerde partial handler
func (wd *WebDashboard) handlePartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		partialType := chi.URLParam(r, "type")

		switch partialType {
		case "price":
			wd.handlePricePartial().ServeHTTP(w, r)
		case "consumption":
			wd.handleConsumptionPartial().ServeHTTP(w, r)
		case "production":
			wd.handleProductionPartial().ServeHTTP(w, r)
		case "home-details":
			wd.handleHomeDetailsPartial().ServeHTTP(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "Unknown partial type")
		}
	}
}

// Gecombineerde data handler
func (wd *WebDashboard) handleData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dataType := chi.URLParam(r, "type")

		switch dataType {
		case "price":
			wd.handlePriceData().ServeHTTP(w, r)
		case "consumption":
			wd.handleConsumptionData().ServeHTTP(w, r)
		case "production":
			wd.handleProductionData().ServeHTTP(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "Unknown data type")
		}
	}
}

// ServePriceEvents serves price updates via Server-Sent Events
func (wd *WebDashboard) ServePriceEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	homeID := chi.URLParam(r, "homeID")
	if _, err := wd.findHomeByID(homeID); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	// Create a channel for this client
	clientChan := make(chan struct{})
	key := fmt.Sprintf("%s+%s", homeID, r.RemoteAddr)
	wd.priceUpdateChannels.Store(key, clientChan)
	defer func() {
		wd.priceUpdateChannels.Delete(key)
	}()

	// Send initial price data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	homeWithPrices, err := wd.PriceSvc.GetPrices(ctx, homeID)
	cancel()

	if err == nil && homeWithPrices != nil {
		jsonData, err := json.Marshal(homeWithPrices.CurrentSubscription.PriceInfo)
		if err == nil {
			fmt.Fprintf(w, "event: price-update\ndata: %s\n\n", jsonData)
			w.(http.Flusher).Flush()
		}
	}

	// Get the client's context
	ctx = r.Context()

	// Start a goroutine to handle the client's connection
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-clientChan:
				// Fetch latest price data
				priceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				homeWithPrices, err := wd.PriceSvc.GetPrices(priceCtx, homeID)
				cancel()

				if err != nil {
					continue
				}

				// Convert to JSON
				jsonData, err := json.Marshal(homeWithPrices.CurrentSubscription.PriceInfo)
				if err != nil {
					continue
				}

				// Send the data to the client
				fmt.Fprintf(w, "event: price-update\ndata: %s\n\n", jsonData)
				w.(http.Flusher).Flush()
			}
		}
	}()

	// Keep the connection alive
	<-ctx.Done()
}
