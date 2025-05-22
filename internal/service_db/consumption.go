package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ws/internal/client"
	"ws/internal/model"
)

// ConsumptionService handles consumption-related operations
type ConsumptionService struct {
	Client *client.TibberClient
	DB     *sql.DB
}

// GetConsumption fetches consumption data for a specific home and stores new data in the database
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

	// Begin a transaction for batch inserts
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare the insert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO consumption (home_id, from_time, to_time, consumption, cost, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (home_id, from_time) DO NOTHING
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Process consumption entries
	for _, node := range nodes {
		nodeData, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse time values
		fromStr := client.GetString(nodeData, "from")
		toStr := client.GetString(nodeData, "to")

		// Parse time strings to time.Time for database
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			continue // Skip invalid times
		}
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			continue // Skip invalid times
		}

		// Create consumption entry
		consumption := model.Consumption{
			From:            fromStr,
			To:              toStr,
			Cost:            getFloat64(nodeData, "cost"),
			Currency:        client.GetString(nodeData, "currency"),
			UnitPrice:       getFloat64(nodeData, "unitPrice"),
			UnitPriceVat:    getFloat64(nodeData, "unitPriceVAT"),
			Consumption:     getFloat64(nodeData, "consumption"),
			ConsumptionUnit: client.GetString(nodeData, "consumptionUnit"),
		}

		// Add to home's consumption list
		home.Consumption = append(home.Consumption, consumption)

		// Store in database
		_, err = stmt.ExecContext(ctx,
			homeId,
			fromTime,
			toTime,
			consumption.Consumption,
			consumption.Cost,
			consumption.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert consumption data: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return home, nil
}

// GetDailySummary provides a daily summary of consumption
func (s *ConsumptionService) GetDailySummary(ctx context.Context, homeId string, days int) ([]model.ConsumptionSummary, error) {
	// First try to get from database
	summaries, err := s.getDailySummaryFromDB(ctx, homeId, days)
	if err == nil && len(summaries) > 0 {
		return summaries, nil
	}

	// If not in database or error, fetch from API
	home, err := s.GetConsumption(ctx, homeId, "DAILY", days)
	if err != nil {
		return nil, err
	}

	// Convert to summary
	summaries = make([]model.ConsumptionSummary, 0, len(home.Consumption))
	for _, c := range home.Consumption {
		summaries = append(summaries, c.ToSummary())
	}

	return summaries, nil
}

// getDailySummaryFromDB retrieves consumption summary from the database
func (s *ConsumptionService) getDailySummaryFromDB(ctx context.Context, homeId string, days int) ([]model.ConsumptionSummary, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT from_time, to_time, consumption, cost, currency
		FROM consumption
		WHERE home_id = $1
		AND from_time>= CURRENT_DATE - INTERVAL '1 day' * $2
		ORDER BY from_time DESC
	`, homeId, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []model.ConsumptionSummary
	for rows.Next() {
		var summary model.ConsumptionSummary
		var fromTime time.Time
		var toTime time.Time
		err := rows.Scan(&fromTime, &toTime, &summary.Consumption, &summary.Cost, &summary.Currency)
		if err != nil {
			return nil, err
		}
		summary.From = fromTime.Format(time.RFC3339)
		summary.To = toTime.Format(time.RFC3339)
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// Helper function to extract float64 values
func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}
