package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DB is the database connection
var DB *sql.DB

// Token represents a Tibber API token
type Token struct {
	ID     int    `json:"id"`
	Token  string `json:"token"`
	Active bool   `json:"active"`
}

// Home represents a home in the database
type Home struct {
	ID                         string    `json:"id"`
	Type                       string    `json:"type"`
	Size                       int       `json:"size"`
	AppNickname                string    `json:"app_nickname"`
	AppAvatar                  string    `json:"app_avatar"`
	MainFuseSize               int       `json:"main_fuse_size"`
	NumberOfResidents          int       `json:"number_of_residents"`
	TimeZone                   string    `json:"time_zone"`
	Address1                   string    `json:"address_1"`
	Address2                   string    `json:"address_2"`
	Address3                   string    `json:"address_3"`
	PostalCode                 string    `json:"postal_code"`
	City                       string    `json:"city"`
	Country                    string    `json:"country"`
	Latitude                   string    `json:"latitude"`
	Longitude                  string    `json:"longitude"`
	ConsumptionEan             string    `json:"consumption_ean"`
	GridCompany                string    `json:"grid_company"`
	GridAreaCode               string    `json:"grid_area_code"`
	PriceAreaCode              string    `json:"price_area_code"`
	ProductionEan              string    `json:"production_ean"`
	EnergyTaxType              string    `json:"energy_tax_type"`
	VatType                    string    `json:"vat_type"`
	EstimatedAnnualConsumption float64   `json:"estimated_annual_consumption"`
	RealTimeConsumptionEnabled bool      `json:"real_time_consumption_enabled"`
	OwnerID                    int       `json:"owner_id"`
	TokenID                    int       `json:"token_id"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

// HomeAPIMapping represents the mapping between API homes and database homes
type HomeAPIMapping struct {
	ID        int    `json:"id"`
	APIHomeID string `json:"api_home_id"`
	HomeID    string `json:"home_id"`
}

// InitDB initializes the database connection and creates tables if they don't exist
func InitDB() {
	log.Println("Starting database initialization...")

	// Load env file
	err := godotenv.Load("env")
	if err != nil {
		log.Printf("Warning: env file not found: %v", err)
	} else {
		log.Println("Successfully loaded env file")
	}

	// Get database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	log.Printf("Found DATABASE_URL: %s", dbURL)

	// Open database connection
	log.Println("Opening database connection...")
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		log.Fatal(err)
	}

	// Test the connection
	log.Println("Testing database connection...")
	err = DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		log.Fatal(err)
	}
	log.Println("Successfully connected to database!")

	// Create tables if they don't exist
	log.Println("Creating tables if they don't exist...")
	createTables()
	log.Println("Database initialization complete!")
}

// createTables creates the necessary tables if they don't exist
func createTables() {
	// Create tibber_tokens table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS tibber_tokens (
			id SERIAL PRIMARY KEY,
			token TEXT NOT NULL UNIQUE,
			active BOOLEAN DEFAULT true
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create homes table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS homes (
			id TEXT PRIMARY KEY,
			type TEXT,
			size INTEGER,
			app_nickname TEXT,
			app_avatar TEXT,
			main_fuse_size INTEGER,
			number_of_residents INTEGER,
			time_zone TEXT,
			address_1 TEXT,
			address_2 TEXT,
			address_3 TEXT,
			postal_code TEXT,
			city TEXT,
			country TEXT,
			latitude TEXT,
			longitude TEXT,
			consumption_ean TEXT,
			grid_company TEXT,
			grid_area_code TEXT,
			price_area_code TEXT,
			production_ean TEXT,
			energy_tax_type TEXT,
			vat_type TEXT,
			estimated_annual_consumption DOUBLE PRECISION,
			real_time_consumption_enabled BOOLEAN,
			owner_id INTEGER,
			token_id INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create home_api_mappings table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS home_api_mappings (
			id SERIAL PRIMARY KEY,
			api_home_id TEXT NOT NULL,
			home_id TEXT NOT NULL REFERENCES homes(id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// GetActiveToken returns the first active token from the database
func GetActiveToken() (string, error) {
	var token string
	err := DB.QueryRow("SELECT token FROM tibber_tokens WHERE active = true LIMIT 1").Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}

// SaveToken saves a new token to the database
func SaveToken(token string) error {
	_, err := DB.Exec("INSERT INTO tibber_tokens (token) VALUES ($1)", token)
	return err
}

// SaveHome saves a home to the database
func SaveHome(home Home) error {
	_, err := DB.Exec(`
		INSERT INTO homes (
			id, type, size, app_nickname, app_avatar, main_fuse_size, number_of_residents,
			time_zone, address_1, address_2, address_3, postal_code, city, country,
			latitude, longitude, consumption_ean, grid_company, grid_area_code,
			price_area_code, production_ean, energy_tax_type, vat_type,
			estimated_annual_consumption, real_time_consumption_enabled, owner_id, token_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
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
			address_3 = EXCLUDED.address_3,
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
			token_id = EXCLUDED.token_id,
			updated_at = CURRENT_TIMESTAMP
	`,
		home.ID, home.Type, home.Size, home.AppNickname, home.AppAvatar,
		home.MainFuseSize, home.NumberOfResidents, home.TimeZone,
		home.Address1, home.Address2, home.Address3, home.PostalCode,
		home.City, home.Country, home.Latitude, home.Longitude,
		home.ConsumptionEan, home.GridCompany, home.GridAreaCode,
		home.PriceAreaCode, home.ProductionEan, home.EnergyTaxType,
		home.VatType, home.EstimatedAnnualConsumption,
		home.RealTimeConsumptionEnabled, home.OwnerID, home.TokenID)
	return err
}

// SaveHomeAPIMapping saves a mapping between an API home and a database home
func SaveHomeAPIMapping(mapping HomeAPIMapping) error {
	_, err := DB.Exec(`
		INSERT INTO home_api_mappings (api_home_id, home_id)
		VALUES ($1, $2)
		ON CONFLICT (api_home_id) DO UPDATE SET
			home_id = EXCLUDED.home_id
	`,
		mapping.APIHomeID, mapping.HomeID)
	return err
}

// GetHomeByAPIID returns a home by its API ID
func GetHomeByAPIID(apiHomeID string) (Home, error) {
	var home Home
	err := DB.QueryRow(`
		SELECT h.* FROM homes h
		JOIN home_api_mappings m ON h.id = m.home_id
		WHERE m.api_home_id = $1
	`, apiHomeID).Scan(
		&home.ID, &home.Type, &home.Size, &home.AppNickname, &home.AppAvatar,
		&home.MainFuseSize, &home.NumberOfResidents, &home.TimeZone,
		&home.Address1, &home.Address2, &home.Address3, &home.PostalCode,
		&home.City, &home.Country, &home.Latitude, &home.Longitude,
		&home.ConsumptionEan, &home.GridCompany, &home.GridAreaCode,
		&home.PriceAreaCode, &home.ProductionEan, &home.EnergyTaxType,
		&home.VatType, &home.EstimatedAnnualConsumption,
		&home.RealTimeConsumptionEnabled, &home.OwnerID, &home.TokenID,
		&home.CreatedAt, &home.UpdatedAt)
	return home, err
}

// GetAllHomes returns all homes from the database
func GetAllHomes() ([]Home, error) {
	rows, err := DB.Query(`
		SELECT * FROM homes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var homes []Home
	for rows.Next() {
		var home Home
		err := rows.Scan(
			&home.ID, &home.Type, &home.Size, &home.AppNickname, &home.AppAvatar,
			&home.MainFuseSize, &home.NumberOfResidents, &home.TimeZone,
			&home.Address1, &home.Address2, &home.Address3, &home.PostalCode,
			&home.City, &home.Country, &home.Latitude, &home.Longitude,
			&home.ConsumptionEan, &home.GridCompany, &home.GridAreaCode,
			&home.PriceAreaCode, &home.ProductionEan, &home.EnergyTaxType,
			&home.VatType, &home.EstimatedAnnualConsumption,
			&home.RealTimeConsumptionEnabled, &home.OwnerID, &home.TokenID,
			&home.CreatedAt, &home.UpdatedAt)
		if err != nil {
			return nil, err
		}
		homes = append(homes, home)
	}
	return homes, nil
}

// GetAllTokens returns all tokens from the database
func GetAllTokens() ([]Token, error) {
	rows, err := DB.Query("SELECT id, token, active FROM tibber_tokens")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var token Token
		err := rows.Scan(&token.ID, &token.Token, &token.Active)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// UpdateTokenActive updates the active status of a token
func UpdateTokenActive(id int, active bool) error {
	_, err := DB.Exec("UPDATE tibber_tokens SET active = $1 WHERE id = $2", active, id)
	return err
}

// DeleteToken deletes a token from the database
func DeleteToken(id int) error {
	_, err := DB.Exec("DELETE FROM tibber_tokens WHERE id = $1", id)
	return err
}

func InitDatabase(db *sql.DB) error {
	// Create schema if it doesn't exist
	_, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS tibber`)
	if err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	// Create tables if they don't exist
	createQueries := []string{
		`CREATE TABLE IF NOT EXISTS tibber.tibber_tokens (
			id SERIAL PRIMARY KEY,
			token VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.owners (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255),
			first_name VARCHAR(255),
			middle_name VARCHAR(255),
			last_name VARCHAR(255),
			address_1 TEXT,
			address_2 TEXT,
			address_3 TEXT,
			city VARCHAR(100),
			postal_code VARCHAR(20),
			country VARCHAR(100),
			latitude FLOAT,
			longitude FLOAT,
			email VARCHAR(255) UNIQUE,
			mobile VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.homes (
			id VARCHAR(255) PRIMARY KEY,
			type VARCHAR(50),
			size INTEGER,
			app_nickname VARCHAR(255),
			app_avatar VARCHAR(255),
			main_fuse_size INTEGER,
			number_of_residents INTEGER,
			time_zone VARCHAR(50),
			address_1 TEXT,
			address_2 TEXT,
			address_3 TEXT,
			postal_code VARCHAR(20),
			city VARCHAR(100),
			country VARCHAR(100),
			latitude VARCHAR(50),
			longitude VARCHAR(50),
			consumption_ean VARCHAR(255),
			grid_company VARCHAR(255),
			grid_area_code VARCHAR(50),
			price_area_code VARCHAR(50),
			production_ean VARCHAR(255),
			energy_tax_type VARCHAR(50),
			vat_type VARCHAR(50),
			estimated_annual_consumption FLOAT,
			real_time_consumption_enabled BOOLEAN DEFAULT false,
			owner_id INTEGER REFERENCES tibber.owners(id),
			token_id INTEGER REFERENCES tibber.tibber_tokens(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.prices (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			price_date DATE NOT NULL,
			hour_of_day INTEGER NOT NULL,
			total DECIMAL(10,4) NOT NULL,
			energy DECIMAL(10,4) NOT NULL,
			tax DECIMAL(10,4) NOT NULL,
			currency VARCHAR(10) NOT NULL,
			level VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(home_id, price_date, hour_of_day)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.consumption (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			consumption DECIMAL(10,4) NOT NULL,
			cost DECIMAL(10,4),
			currency VARCHAR(10),
			unit VARCHAR(10) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(home_id, from_time)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.production (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			production DECIMAL(10,4) NOT NULL,
			profit DECIMAL(10,4),
			currency VARCHAR(10),
			unit VARCHAR(10),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(home_id, from_time)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.real_time_measurements (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			power DECIMAL(10,4) NOT NULL,
			min_power DECIMAL(10,4),
			average_power DECIMAL(10,4),
			max_power DECIMAL(10,4),
			power_production DECIMAL(10,4),
			max_power_production DECIMAL(10,4),
			accumulated_consumption DECIMAL(10,4),
			accumulated_production DECIMAL(10,4),
			last_meter_consumption DECIMAL(10,4),
			last_meter_production DECIMAL(10,4),
			current_l1 DECIMAL(10,4),
			current_l2 DECIMAL(10,4),
			current_l3 DECIMAL(10,4),
			voltage_phase1 DECIMAL(10,4),
			voltage_phase2 DECIMAL(10,4),
			voltage_phase3 DECIMAL(10,4),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(home_id, timestamp)
		)`,
	}

	// Execute create queries
	for _, query := range createQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %w", err)
		}
	}

	// Create indexes
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_prices_home_id ON tibber.prices(home_id)`,
		`CREATE INDEX IF NOT EXISTS idx_prices_date ON tibber.prices(price_date)`,
		`CREATE INDEX IF NOT EXISTS idx_consumption_home_id ON tibber.consumption(home_id)`,
		`CREATE INDEX IF NOT EXISTS idx_consumption_from_time ON tibber.consumption(from_time)`,
		`CREATE INDEX IF NOT EXISTS idx_production_home_id ON tibber.production(home_id)`,
		`CREATE INDEX IF NOT EXISTS idx_production_from_time ON tibber.production(from_time)`,
		`CREATE INDEX IF NOT EXISTS idx_real_time_measurements_home_id ON tibber.real_time_measurements(home_id)`,
		`CREATE INDEX IF NOT EXISTS idx_real_time_measurements_timestamp ON tibber.real_time_measurements(timestamp)`,
	}

	// Execute index queries
	for _, query := range indexQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
	}

	// Create trigger function for token changes
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION tibber.notify_token_changes()
		RETURNS TRIGGER AS $$
		BEGIN
			PERFORM pg_notify(
				'token_changes',
				json_build_object(
					'action', TG_OP,
					'token_id', NEW.id,
					'active', NEW.active
				)::text
			);
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		return fmt.Errorf("error creating trigger function: %w", err)
	}

	// Create trigger for token changes
	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS token_changes_trigger ON tibber.tibber_tokens;
		CREATE TRIGGER token_changes_trigger
			AFTER INSERT OR UPDATE ON tibber.tibber_tokens
			FOR EACH ROW
			EXECUTE FUNCTION tibber.notify_token_changes();
	`)
	if err != nil {
		return fmt.Errorf("error creating trigger: %w", err)
	}

	return nil
}
