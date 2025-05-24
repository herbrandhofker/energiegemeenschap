package service

import (
	"context"
	"fmt"

	"tibber_loader/internal/client"
	"tibber_loader/internal/model"
)

// ConsumptionService handles consumption-related operations
type ConsumptionService struct {
	Client *client.TibberClient
}

// GetConsumption fetches consumption data for a specific home
func (s *ConsumptionService) GetConsumption(ctx context.Context, homeId string, resolution string, lastEntries int) (*model.Home, error) {
	// Set up variables for the query
	variables := map[string]interface{}{
		"homeId":     homeId,
		"resolution": resolution,
		"last":       lastEntries,
	}

	// Execute the query
	resp, err := s.Client.QueryAPI(ctx, model.ConsumptionQuery, variables)
	if err != nil {
		return nil, fmt.Errorf("API query failed: %w", err)
	}

	// Extract viewer data
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no viewer data in response")
	}

	// Extract home data
	homeData, ok := viewerData["home"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no home data in response; check if homeId is correct: %s", homeId)
	}

	// Create home with ID
	home := &model.Home{
		Id: homeId,
	}

	// Extract consumption data
	consumptionData, ok := homeData["consumption"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no consumption data in response; check resolution parameter: %s", resolution)
	}

	// Extract nodes
	nodes, ok := consumptionData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no consumption nodes in response")
	}

	// Process consumption entries
	for _, node := range nodes {
		nodeData, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse time values
		fromStr := client.GetString(nodeData, "from")
		toStr := client.GetString(nodeData, "to")

		// Parse tijdelijk naar time.Time om de formatted date te maken

		// Create consumption entry (met string datumvelden)
		consumption := model.Consumption{
			From:            fromStr, // Behoud als string
			To:              toStr,   // Behoud als string
			Cost:            getFloat64(nodeData, "cost"),
			Currency:        client.GetString(nodeData, "currency"),
			UnitPrice:       getFloat64(nodeData, "unitPrice"),
			UnitPriceVat:    getFloat64(nodeData, "unitPriceVAT"),
			Consumption:     getFloat64(nodeData, "consumption"),
			ConsumptionUnit: client.GetString(nodeData, "consumptionUnit"),
		}

		home.Consumption = append(home.Consumption, consumption)
	}

	return home, nil
}

// GetDailySummary provides a daily summary of consumption
func (s *ConsumptionService) GetDailySummary(ctx context.Context, homeId string, days int) ([]model.ConsumptionSummary, error) {
	// Fetch daily data
	home, err := s.GetConsumption(ctx, homeId, "DAILY", days)
	if err != nil {
		return nil, err
	}

	// Convert to summary
	summaries := make([]model.ConsumptionSummary, 0, len(home.Consumption))
	for _, c := range home.Consumption {
		summaries = append(summaries, c.ToSummary())
	}

	return summaries, nil
}

// Helper function to extract float64 values
func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}
