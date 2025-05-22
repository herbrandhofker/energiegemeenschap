package db

import (
	"database/sql"
	"fmt"
)


// InitSchema initializes the database schema
func InitSchema(db *sql.DB) error {
	// Drop existing tables in reverse dependency order


	// Create tables if they don't exist
	createQueries := []string{
		`CREATE TABLE IF NOT EXISTS tibber.owners (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			first_name VARCHAR(255),
			middle_name VARCHAR(255),
			last_name VARCHAR(255),
			-- Address fields
			address_1 VARCHAR(255),
			address_2 VARCHAR(255),
			address_3 VARCHAR(255),
			city VARCHAR(100),
			postal_code VARCHAR(20),
			country VARCHAR(50),
			latitude VARCHAR(20),
			longitude VARCHAR(20),
			-- Contact info
			email VARCHAR(255) NOT NULL UNIQUE,
			mobile VARCHAR(50),
			-- Timestamps
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.homes (
			id VARCHAR(50) PRIMARY KEY,
			type VARCHAR(20),
			size INTEGER,
			app_nickname VARCHAR(100),
			app_avatar VARCHAR(255),
			main_fuse_size INTEGER,
			number_of_residents INTEGER,
			time_zone VARCHAR(50),
			-- Address fields
			address_1 VARCHAR(255),
			address_2 VARCHAR(255),
			postal_code VARCHAR(20),
			city VARCHAR(100),
			country VARCHAR(50),
			latitude VARCHAR(20),
			longitude VARCHAR(20),
			-- Metering point data
			consumption_ean VARCHAR(50),
			grid_company VARCHAR(100),
			grid_area_code VARCHAR(50),
			price_area_code VARCHAR(50),
			production_ean VARCHAR(50),
			energy_tax_type VARCHAR(50),
			vat_type VARCHAR(20),
			estimated_annual_consumption DECIMAL(10,2),
			-- Features
			real_time_consumption_enabled BOOLEAN,
			-- Owner reference
			owner_id INTEGER REFERENCES owners(id),
			-- Timestamps
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.consumption (
			home_id VARCHAR(50),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			consumption DECIMAL(10,2),
			cost DECIMAL(10,2),
			currency TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (home_id, from_time),
			FOREIGN KEY (home_id) REFERENCES homes(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.production (
			home_id VARCHAR(50),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			production DECIMAL(10,2),
			profit DECIMAL(10,2),
			currency TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (home_id, from_time),
			FOREIGN KEY (home_id) REFERENCES homes(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.prices (
			home_id VARCHAR(50),
			price_date DATE,
			hour_of_day INTEGER,
			total DECIMAL(10,4),
			energy DECIMAL(10,4),
			tax DECIMAL(10,4),
			currency TEXT,
			level TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (home_id, price_date, hour_of_day),
			FOREIGN KEY (home_id) REFERENCES homes(id),
			CHECK (hour_of_day >= 0 AND hour_of_day < 24)
		)`,
		`CREATE TABLE IF NOT EXISTS tibber.real_time_measurements (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			power DECIMAL(10,2) NOT NULL,
			power_production DECIMAL(10,2) NOT NULL,
			min_power DECIMAL(10,2),
			average_power DECIMAL(10,2),
			max_power DECIMAL(10,2),
			max_power_production DECIMAL(10,2),
			accumulated_consumption DECIMAL(10,2) NOT NULL,
			accumulated_production DECIMAL(10,2) NOT NULL,
			last_meter_consumption DECIMAL(10,2),
			last_meter_production DECIMAL(10,2),
			current_l1 DECIMAL(10,2),
			current_l2 DECIMAL(10,2),
			current_l3 DECIMAL(10,2),
			voltage_phase1 DECIMAL(10,2),
			voltage_phase2 DECIMAL(10,2),
			voltage_phase3 DECIMAL(10,2),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (home_id) REFERENCES homes(id),
			UNIQUE (home_id, timestamp)
		)`,
	}

	// Execute create queries
	for _, query := range createQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating schema: %w", err)
		}
	}

	return nil
}
