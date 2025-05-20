package service

import (
	"context"
	"fmt"

	"ws/internal/client"
	"ws/internal/model"
)

// HomeService handles home-related operations
type HomeService struct {
	Client *client.TibberClient
}

// GetHomes fetches basic information about all homes
func (s *HomeService) GetHomes(ctx context.Context) ([]model.Home, error) {
	// Simple query to get all home IDs, names, etc.
	query := `
        query {
            viewer {
                homes {
                    id
                    appNickname
                    address {
                        address1
                        city
                    }
                    meteringPointData {
                        consumptionEan
                        productionEan
                    }
                }
            }
        }
    `

	resp, err := s.Client.QueryAPI(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	// Extract viewer data
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no viewer data in response")
	}

	// Extract homes data
	homesData, ok := viewerData["homes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no homes data in response")
	}

	// Parse homes
	homes := make([]model.Home, 0, len(homesData))
	for _, homeRaw := range homesData {
		homeData, ok := homeRaw.(map[string]interface{})
		if !ok {
			continue
		}

		home := model.Home{
			Id:          client.GetString(homeData, "id"),
			AppNickname: client.GetString(homeData, "appNickname"),
		}

		// Parse address if available
		if addressData, ok := homeData["address"].(map[string]interface{}); ok {
			home.Address = model.Address{
				Address1: client.GetString(addressData, "address1"),
				City:     client.GetString(addressData, "city"),
			}
		}

		// Parse meteringPointData if available
		if mpData, ok := homeData["meteringPointData"].(map[string]interface{}); ok {
			home.MeteringPointData = model.MeteringPointData{
				ConsumptionEan: client.GetString(mpData, "consumptionEan"),
				ProductionEan:  client.GetString(mpData, "productionEan"),
			}
		}

		homes = append(homes, home)
	}

	return homes, nil
}

// GetHomeDetails fetches detailed information about homes
func (s *HomeService) GetHomeDetails(ctx context.Context) ([]model.Home, error) {
	resp, err := s.Client.QueryAPI(ctx, model.HomeDetailsQuery, nil)
	if err != nil {
		return nil, err
	}

	// Extract viewer data
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no viewer data in response")
	}

	// Extract homes data
	homesData, ok := viewerData["homes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no homes data in response")
	}

	// Parse homes
	homes := make([]model.Home, 0, len(homesData))
	for _, homeRaw := range homesData {
		homeData, ok := homeRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Create basic home info
		home := model.Home{
			Id:           client.GetString(homeData, "id"),
			Type:         client.GetString(homeData, "type"),
			Size:         client.GetInt(homeData, "size"),
			AppNickname:  client.GetString(homeData, "appNickname"),
			AppAvatar:    client.GetString(homeData, "appAvatar"),
			MainFuseSize: client.GetInt(homeData, "mainFuseSize"),
		}

		// Parse address if available
		if addressData, ok := homeData["address"].(map[string]interface{}); ok {
			home.Address = model.Address{
				Address1:   client.GetString(addressData, "address1"),
				Address2:   client.GetString(addressData, "address2"),
				PostalCode: client.GetString(addressData, "postalCode"),
				City:       client.GetString(addressData, "city"),
				Country:    client.GetString(addressData, "country"),
				Latitude:   client.GetString(addressData, "latitude"),
				Longitude:  client.GetString(addressData, "longitude"),
			}
		}

		// Parse meteringPointData if available
		if mpData, ok := homeData["meteringPointData"].(map[string]interface{}); ok {
			home.MeteringPointData = model.MeteringPointData{
				ConsumptionEan:             client.GetString(mpData, "consumptionEan"),
				GridCompany:                client.GetString(mpData, "gridCompany"),
				GridAreaCode:               client.GetString(mpData, "gridAreaCode"),
				PriceAreaCode:              client.GetString(mpData, "priceAreaCode"),
				ProductionEan:              client.GetString(mpData, "productionEan"),
				EnergyTaxType:              client.GetString(mpData, "energyTaxType"),
				VatType:                    client.GetString(mpData, "vatType"),
				EstimatedAnnualConsumption: float64(client.GetInt(mpData, "estimatedAnnualConsumption")),
			}
		}

		// Parse features if available
		if featuresData, ok := homeData["features"].(map[string]interface{}); ok {
			home.Features = model.HomeFeatures{
				RealTimeConsumptionEnabled: featuresData["realTimeConsumptionEnabled"] == true,
			}
		}

		homes = append(homes, home)
	}

	return homes, nil
}

// GetHomesWithProductionCapability returns only homes that have production capability
func (s *HomeService) GetHomesWithProductionCapability(ctx context.Context) ([]model.Home, error) {
	homes, err := s.GetHomes(ctx)
	if err != nil {
		return nil, err
	}

	// Filter homes with production capability
	productionHomes := make([]model.Home, 0)
	for _, home := range homes {
		if home.MeteringPointData.ProductionEan != "" {
			productionHomes = append(productionHomes, home)
		}
	}

	return productionHomes, nil
}
