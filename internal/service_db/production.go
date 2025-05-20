package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ws/internal/client"
	"ws/internal/model"
)

// ProductionService handles production-related operations
type ProductionService struct {
	Client *client.TibberClient
	DB     *sql.DB
}

// GetProduction fetches production data for a specific home and stores new data in the database
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

	// Begin a transaction for batch inserts
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare the insert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO production (home_id, from_date, to_time, production, profit, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (home_id, from_date) DO NOTHING
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Process production entries
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

		// Add to home's production list
		home.Production = append(home.Production, production)

		// Store in database - note we now use fromTime.Truncate(24*time.Hour) to get just the date
		_, err = stmt.ExecContext(ctx,
			homeId,
			fromTime.Truncate(24*time.Hour),
			toTime,
			production.Production,
			production.Profit,
			production.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert production data: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return home, nil
}

// GetDailySummary provides a daily summary of production
func (s *ProductionService) GetDailySummary(ctx context.Context, homeId string, days int) ([]model.ProductionSummary, error) {
	// First try to get from database
	summaries, err := s.getDailySummaryFromDB(ctx, homeId, days)
	if err == nil && len(summaries) > 0 {
		return summaries, nil
	}

	// If not in database or error, fetch from API
	home, err := s.GetProduction(ctx, homeId, "DAILY", days)
	if err != nil {
		return nil, err
	}

	// Convert to summary
	summaries = make([]model.ProductionSummary, 0, len(home.Production))
	for _, p := range home.Production {
		summaries = append(summaries, p.ToSummary())
	}

	return summaries, nil
}

// getDailySummaryFromDB retrieves production summary from the database
func (s *ProductionService) getDailySummaryFromDB(ctx context.Context, homeId string, days int) ([]model.ProductionSummary, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT from_date, to_time, production, profit, currency
		FROM production
		WHERE home_id = $1
		AND from_date >= CURRENT_DATE - INTERVAL '1 day' * $2
		ORDER BY from_date DESC
	`, homeId, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []model.ProductionSummary
	for rows.Next() {
		var summary model.ProductionSummary
		var fromDate time.Time
		var toTime time.Time
		err := rows.Scan(&fromDate, &toTime, &summary.Production, &summary.Profit, &summary.Currency)
		if err != nil {
			return nil, err
		}
		summary.From = fromDate.Format(time.RFC3339)
		summary.To = toTime.Format(time.RFC3339)
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// HasProduction checks if a home has production capability
func (s *ProductionService) HasProduction(home model.Home) bool {
	return home.MeteringPointData.ProductionEan != ""
}
