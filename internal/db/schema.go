package db

import (
	"database/sql"
	"fmt"
)

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

	return nil
}
