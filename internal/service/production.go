package service

import (
	"context"
	"fmt"

	"tibber_loader/internal/client"
	"tibber_loader/internal/model"
)

// ProductionService handles production-related operations
type ProductionService struct {
	Client *client.TibberClient
}

// GetProduction fetches production data for a specific home
func (s *ProductionService) GetProduction(ctx context.Context, homeId string, resolution string, lastEntries int) (*model.Home, error) {
	// Set up variables for the query
	variables := map[string]interface{}{
		"homeId":     homeId,
		"resolution": resolution,
		"last":       lastEntries,
	}

	// Execute the query
	resp, err := s.Client.QueryAPI(ctx, model.ProductionQuery, variables)
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

	// Extract production data
	productionData, ok := homeData["production"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no production data in response; check resolution parameter: %s", resolution)
	}

	// Extract nodes
	nodes, ok := productionData["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no production nodes in response")
	}

	// Process production entries
	for _, node := range nodes {
		nodeData, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse time values
		fromStr := client.GetString(nodeData, "from")
		toStr := client.GetString(nodeData, "to")

		// Create production entry
		production := model.Production{
			From:           fromStr,
			To:             toStr,
			Profit:         getFloat64(nodeData, "profit"),
			Currency:       client.GetString(nodeData, "currency"),
			UnitPrice:      getFloat64(nodeData, "unitPrice"),
			UnitPriceVAT:   getFloat64(nodeData, "unitPriceVAT"),
			Production:     getFloat64(nodeData, "production"),
			ProductionUnit: client.GetString(nodeData, "productionUnit"),
		}

		home.Production = append(home.Production, production)
	}

	return home, nil
}

// GetDailySummary provides a daily summary of production
func (s *ProductionService) GetDailySummary(ctx context.Context, homeId string, days int) ([]model.ProductionSummary, error) {
	// Fetch daily data
	home, err := s.GetProduction(ctx, homeId, "DAILY", days)
	if err != nil {
		return nil, err
	}

	// Convert to summary
	summaries := make([]model.ProductionSummary, 0, len(home.Production))
	for _, p := range home.Production {
		summaries = append(summaries, p.ToSummary())
	}

	return summaries, nil
}

// HasProduction checks if a home has production capability
func (s *ProductionService) HasProduction(home model.Home) bool {
	return home.MeteringPointData.ProductionEan != ""
}
