package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ws/internal/client"
	"ws/internal/model"
)

// PriceService handles price information operations
type PriceService struct {
	Client *client.TibberClient
	DB     *sql.DB
}

// Update GetPrices to handle the homes array response and store in database
func (s *PriceService) GetPrices(ctx context.Context, homeId string) (*model.Home, error) {
	// Note: No variables needed for the query now
	resp, err := s.Client.QueryAPI(ctx, model.PriceQuery, nil)
	if err != nil {
		return nil, err
	}

	// Extract viewer data
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no viewer data in response")
	}

	// Extract homes array
	homesData, ok := viewerData["homes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no homes data in response")
	}

	// Find the home with the matching ID
	var targetHome *model.Home
	for _, homeRaw := range homesData {
		homeData, ok := homeRaw.(map[string]interface{})
		if !ok {
			continue
		}

		id := client.GetString(homeData, "id")
		if id == homeId {
			// Found the right home
			targetHome = &model.Home{
				Id: id,
			}

			// Extract subscription data if available
			if subscriptionData, ok := homeData["currentSubscription"].(map[string]interface{}); ok {
				targetHome.CurrentSubscription = &model.Subscription{}

				// Extract priceInfo data if available
				if priceInfoData, ok := subscriptionData["priceInfo"].(map[string]interface{}); ok {
					priceInfo := model.PriceInfo{}

					// Begin a transaction for batch inserts
					tx, err := s.DB.BeginTx(ctx, nil)
					if err != nil {
						return nil, fmt.Errorf("failed to begin transaction: %w", err)
					}
					defer tx.Rollback() // Rollback if not committed

					// Prepare the insert statement
					stmt, err := tx.PrepareContext(ctx, `
						INSERT INTO prices (home_id, price_date, hour_of_day, total, energy, tax, currency, level)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
						ON CONFLICT (home_id, price_date, hour_of_day) DO NOTHING
					`)
					if err != nil {
						return nil, fmt.Errorf("failed to prepare statement: %w", err)
					}
					defer stmt.Close()

					// Parse current price if available
					if currentData, ok := priceInfoData["current"].(map[string]interface{}); ok {
						current := parsePriceEntry(currentData)
						priceInfo.Current = current

						// Store current price in database
						if err := storePriceInDB(ctx, stmt, homeId, current); err != nil {
							return nil, fmt.Errorf("failed to store current price: %w", err)
						}
					}

					// Parse today prices if available
					if todayData, ok := priceInfoData["today"].([]interface{}); ok {
						today := make([]model.Price, 0, len(todayData))
						for _, entry := range todayData {
							if entryMap, ok := entry.(map[string]interface{}); ok {
								price := parsePriceEntry(entryMap)
								today = append(today, price)

								// Store today's price in database
								if err := storePriceInDB(ctx, stmt, homeId, price); err != nil {
									return nil, fmt.Errorf("failed to store today's price: %w", err)
								}
							}
						}
						priceInfo.Today = today
					}

					// Parse tomorrow prices if available
					if tomorrowData, ok := priceInfoData["tomorrow"].([]interface{}); ok {
						tomorrow := make([]model.Price, 0, len(tomorrowData))
						for _, entry := range tomorrowData {
							if entryMap, ok := entry.(map[string]interface{}); ok {
								price := parsePriceEntry(entryMap)
								tomorrow = append(tomorrow, price)

								// Store tomorrow's price in database
								if err := storePriceInDB(ctx, stmt, homeId, price); err != nil {
									return nil, fmt.Errorf("failed to store tomorrow's price: %w", err)
								}
							}
						}
						priceInfo.Tomorrow = tomorrow
					}

					// Commit the transaction
					if err := tx.Commit(); err != nil {
						return nil, fmt.Errorf("failed to commit transaction: %w", err)
					}

					targetHome.CurrentSubscription.PriceInfo = priceInfo
				}
			}

			break // Found the home we're looking for
		}
	}

	if targetHome == nil {
		return nil, fmt.Errorf("home with ID %s not found in response", homeId)
	}

	return targetHome, nil
}

// Helper function to store a price in the database
func storePriceInDB(ctx context.Context, stmt *sql.Stmt, homeId string, price model.Price) error {
	startTime, err := time.Parse(time.RFC3339, price.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	_, err = stmt.ExecContext(ctx,
		homeId,
		startTime.Truncate(24*time.Hour),
		startTime.Hour(),
		price.Total,
		price.Energy,
		price.Tax,
		price.Currency,
		price.Level,
	)
	return err
}

// GetCurrentPrice provides just the current price information
func (s *PriceService) GetCurrentPrice(ctx context.Context, homeId string) (*model.Price, error) {
	// First try to get from database
	price, err := s.getCurrentPriceFromDB(ctx, homeId)
	if err == nil {
		return price, nil
	}

	// If not in database or error, fetch from API
	home, err := s.GetPrices(ctx, homeId)
	if err != nil {
		return nil, err
	}

	if home.CurrentSubscription == nil || home.CurrentSubscription.PriceInfo.Current.StartTime == "" {
		return nil, fmt.Errorf("no current price information available")
	}

	return &home.CurrentSubscription.PriceInfo.Current, nil
}

// getCurrentPriceFromDB retrieves the current price from the database
func (s *PriceService) getCurrentPriceFromDB(ctx context.Context, homeId string) (*model.Price, error) {
	var price model.Price
	var priceDate time.Time
	var hourOfDay int

	err := s.DB.QueryRowContext(ctx, `
		SELECT price_date, hour_of_day, total, energy, tax, currency, level
		FROM prices
		WHERE home_id = $1
		AND price_date = CURRENT_DATE
		AND hour_of_day = EXTRACT(HOUR FROM CURRENT_TIMESTAMP)::INTEGER
		LIMIT 1
	`, homeId).Scan(&priceDate, &hourOfDay, &price.Total, &price.Energy, &price.Tax, &price.Currency, &price.Level)

	if err != nil {
		return nil, err
	}

	// Reconstruct the start time from date and hour
	startTime := time.Date(priceDate.Year(), priceDate.Month(), priceDate.Day(), hourOfDay, 0, 0, 0, priceDate.Location())
	price.StartTime = startTime.Format(time.RFC3339)
	return &price, nil
}

// FindLowestPriceHour returns the timestamp with the lowest price today or tomorrow
func (s *PriceService) FindLowestPriceHour(ctx context.Context, homeId string, includeTomorrow bool) (*model.Price, error) {
	// First try to get from database
	price, err := s.findLowestPriceFromDB(ctx, homeId, includeTomorrow)
	if err == nil {
		return price, nil
	}

	// If not in database or error, fetch from API
	home, err := s.GetPrices(ctx, homeId)
	if err != nil {
		return nil, err
	}

	if home.CurrentSubscription == nil {
		return nil, fmt.Errorf("no subscription information available")
	}

	var lowestPrice *model.Price
	var pricesToCheck []model.Price

	// Add today's prices
	pricesToCheck = append(pricesToCheck, home.CurrentSubscription.PriceInfo.Today...)

	// Add tomorrow's prices if requested and available
	if includeTomorrow && len(home.CurrentSubscription.PriceInfo.Tomorrow) > 0 {
		pricesToCheck = append(pricesToCheck, home.CurrentSubscription.PriceInfo.Tomorrow...)
	}

	// Find lowest price
	for i, price := range pricesToCheck {
		if i == 0 || price.Total < lowestPrice.Total {
			priceCopy := price // Create a copy to avoid reference issues
			lowestPrice = &priceCopy
		}
	}

	if lowestPrice == nil {
		return nil, fmt.Errorf("no price information available")
	}

	return lowestPrice, nil
}

// findLowestPriceFromDB finds the lowest price in the database
func (s *PriceService) findLowestPriceFromDB(ctx context.Context, homeId string, includeTomorrow bool) (*model.Price, error) {
	query := `
		SELECT price_date, hour_of_day, total, energy, tax, currency, level
		FROM prices
		WHERE home_id = $1
		AND price_date = CURRENT_DATE
	`
	if includeTomorrow {
		query = `
			SELECT price_date, hour_of_day, total, energy, tax, currency, level
			FROM prices
			WHERE home_id = $1
			AND price_date IN (CURRENT_DATE, CURRENT_DATE + INTERVAL '1 day')
		`
	}
	query += ` ORDER BY total ASC LIMIT 1`

	var price model.Price
	var priceDate time.Time
	var hourOfDay int

	err := s.DB.QueryRowContext(ctx, query, homeId).Scan(
		&priceDate,
		&hourOfDay,
		&price.Total,
		&price.Energy,
		&price.Tax,
		&price.Currency,
		&price.Level,
	)

	if err != nil {
		return nil, err
	}

	// Reconstruct the start time from date and hour
	startTime := time.Date(priceDate.Year(), priceDate.Month(), priceDate.Day(), hourOfDay, 0, 0, 0, priceDate.Location())
	price.StartTime = startTime.Format(time.RFC3339)
	return &price, nil
}

// Helper function to parse a single price entry
func parsePriceEntry(entryData map[string]interface{}) model.Price {
	// Parse time values
	startsAtStr := client.GetString(entryData, "startsAt")

	return model.Price{
		Total:     getFloat64(entryData, "total"),
		Energy:    getFloat64(entryData, "energy"),
		Tax:       getFloat64(entryData, "tax"),
		StartTime: startsAtStr, // Gebruik de originele string
		Currency:  client.GetString(entryData, "currency"),
		Level:     client.GetString(entryData, "level"),
	}
}
