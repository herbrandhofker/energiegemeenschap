package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ws/internal/client"
	"ws/internal/model"
)

// HomeService handles home-related operations
type HomeService struct {
	Client *client.TibberClient
	DB     *sql.DB
}

// GetHomes fetches basic information about all homes
func (s *HomeService) GetHomes(ctx context.Context) ([]model.Home, error) {
	// First try to get from database
	homes, err := s.getHomesFromDB(ctx)
	if err == nil && len(homes) > 0 {
		return homes, nil
	}

	// If not in database or error, fetch from API
	return s.fetchAndStoreHomes(ctx)
}

// GetHomeDetails fetches detailed information about homes
func (s *HomeService) GetHomeDetails(ctx context.Context) ([]model.Home, error) {
	// First try to get from database
	homes, err := s.getHomesFromDB(ctx)
	if err == nil && len(homes) > 0 {
		return homes, nil
	}

	// If not in database or error, fetch from API
	return s.fetchAndStoreHomes(ctx)
}

// fetchAndStoreHomes fetches homes from API and stores them in database
func (s *HomeService) fetchAndStoreHomes(ctx context.Context) ([]model.Home, error) {
	// First get user data to get the owner information
	userSvc := &UserService{Client: s.Client}
	user, err := userSvc.GetUserData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// Begin a transaction for batch inserts
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// First store or update the owner and get the ID
	var ownerID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO owners (email, name, tibber_id, account_type, last_login, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			tibber_id = EXCLUDED.tibber_id,
			account_type = EXCLUDED.account_type,
			last_login = EXCLUDED.last_login,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`,
		user.Email,
		user.Name,
		user.ID,
		user.AccountType,
		user.LastLogin,
		time.Now(),
	).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to store owner: %w", err)
	}

	// Execute the homes query
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

	// Prepare the homes insert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO homes (
			id, type, size, app_nickname, app_avatar, main_fuse_size, number_of_residents,
			time_zone, address_1, address_2, postal_code, city, country, latitude, longitude,
			consumption_ean, grid_company, grid_area_code, price_area_code, production_ean,
			energy_tax_type, vat_type, estimated_annual_consumption, real_time_consumption_enabled,
			owner_id, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			size = EXCLUDED.size,
			app_nickname = EXCLUDED.app_nickname,
			app_avatar = EXCLUDED.app_avatar,
			main_fuse_size = EXCLUDED.main_fuse_size,
			number_of_residents = EXCLUDED.number_of_residents,
			time_zone = EXCLUDED.time_zone,
			address_1 = EXCLUDED.address_1,
			address_2 = EXCLUDED.address_2,
			postal_code = EXCLUDED.postal_code,
			city = EXCLUDED.city,
			country = EXCLUDED.country,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			consumption_ean = EXCLUDED.consumption_ean,
			grid_company = EXCLUDED.grid_company,
			grid_area_code = EXCLUDED.grid_area_code,
			price_area_code = EXCLUDED.price_area_code,
			production_ean = EXCLUDED.production_ean,
			energy_tax_type = EXCLUDED.energy_tax_type,
			vat_type = EXCLUDED.vat_type,
			estimated_annual_consumption = EXCLUDED.estimated_annual_consumption,
			real_time_consumption_enabled = EXCLUDED.real_time_consumption_enabled,
			owner_id = EXCLUDED.owner_id,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

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

		// Store in database with owner_id
		_, err = stmt.ExecContext(ctx,
			home.Id, home.Type, home.Size, home.AppNickname, home.AppAvatar,
			home.MainFuseSize, home.NumberOfResidents, home.TimeZone,
			home.Address.Address1, home.Address.Address2, home.Address.PostalCode,
			home.Address.City, home.Address.Country, home.Address.Latitude,
			home.Address.Longitude, home.MeteringPointData.ConsumptionEan,
			home.MeteringPointData.GridCompany, home.MeteringPointData.GridAreaCode,
			home.MeteringPointData.PriceAreaCode, home.MeteringPointData.ProductionEan,
			home.MeteringPointData.EnergyTaxType, home.MeteringPointData.VatType,
			home.MeteringPointData.EstimatedAnnualConsumption,
			home.Features.RealTimeConsumptionEnabled,
			ownerID, // Use the owner ID we got from the insert/update
			time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to store home %s: %w", home.Id, err)
		}

		homes = append(homes, home)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return homes, nil
}

// getHomesFromDB retrieves homes from the database
func (s *HomeService) getHomesFromDB(ctx context.Context) ([]model.Home, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT 
			h.id, h.type, h.size, h.app_nickname, h.app_avatar, h.main_fuse_size,
			h.number_of_residents, h.time_zone, h.address_1, h.address_2,
			h.postal_code, h.city, h.country, h.latitude, h.longitude,
			h.consumption_ean, h.grid_company, h.grid_area_code,
			h.price_area_code, h.production_ean, h.energy_tax_type,
			h.vat_type, h.estimated_annual_consumption,
			h.real_time_consumption_enabled,
			o.id as owner_id, o.email as owner_email, o.name as owner_name, 
			o.tibber_id as owner_tibber_id, o.account_type as owner_account_type
		FROM homes h
		LEFT JOIN owners o ON h.owner_id = o.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var homes []model.Home
	for rows.Next() {
		var home model.Home
		var owner model.User
		err := rows.Scan(
			&home.Id, &home.Type, &home.Size, &home.AppNickname, &home.AppAvatar,
			&home.MainFuseSize, &home.NumberOfResidents, &home.TimeZone,
			&home.Address.Address1, &home.Address.Address2, &home.Address.PostalCode,
			&home.Address.City, &home.Address.Country, &home.Address.Latitude,
			&home.Address.Longitude, &home.MeteringPointData.ConsumptionEan,
			&home.MeteringPointData.GridCompany, &home.MeteringPointData.GridAreaCode,
			&home.MeteringPointData.PriceAreaCode, &home.MeteringPointData.ProductionEan,
			&home.MeteringPointData.EnergyTaxType, &home.MeteringPointData.VatType,
			&home.MeteringPointData.EstimatedAnnualConsumption,
			&home.Features.RealTimeConsumptionEnabled,
			&owner.ID, &owner.Email, &owner.Name, &owner.ID, &owner.AccountType,
		)
		if err != nil {
			return nil, err
		}
		home.Owner = &owner
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
