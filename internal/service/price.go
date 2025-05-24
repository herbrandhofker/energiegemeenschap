package service

import (
	"context"
	"fmt"

	"tibber_loader/internal/client"
	"tibber_loader/internal/model"
)

// PriceService handles price information operations
type PriceService struct {
	Client *client.TibberClient
}

// Update GetPrices to handle the homes array response
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

					// Parse current price if available
					if currentData, ok := priceInfoData["current"].(map[string]interface{}); ok {
						current := parsePriceEntry(currentData)
						priceInfo.Current = current
					}

					// Parse today prices if available
					if todayData, ok := priceInfoData["today"].([]interface{}); ok {
						today := make([]model.Price, 0, len(todayData))
						for _, entry := range todayData {
							if entryMap, ok := entry.(map[string]interface{}); ok {
								price := parsePriceEntry(entryMap)
								today = append(today, price)
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
							}
						}
						priceInfo.Tomorrow = tomorrow
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

// GetCurrentPrice provides just the current price information
func (s *PriceService) GetCurrentPrice(ctx context.Context, homeId string) (*model.Price, error) {
	home, err := s.GetPrices(ctx, homeId)
	if err != nil {
		return nil, err
	}

	if home.CurrentSubscription == nil || home.CurrentSubscription.PriceInfo.Current.StartTime == "" {
		return nil, fmt.Errorf("no current price information available")
	}

	return &home.CurrentSubscription.PriceInfo.Current, nil
}

// FindLowestPriceHour returns the timestamp with the lowest price today or tomorrow
func (s *PriceService) FindLowestPriceHour(ctx context.Context, homeId string, includeTomorrow bool) (*model.Price, error) {
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
