package service_db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// InitializeSchema creates the necessary database tables if they don't exist
func InitializeSchema(db *sql.DB) error {
	ctx := context.Background()

	// First check if schema exists
	var schemaExists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM pg_namespace 
			WHERE nspname = 'tibber'
		)
	`).Scan(&schemaExists)
	if err != nil {
		return fmt.Errorf("error checking schema existence: %w", err)
	}

	// Create schema if it doesn't exist
	if !schemaExists {
		log.Println("Creating tibber schema...")
		_, err = db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS tibber`)
		if err != nil {
			return fmt.Errorf("error creating schema: %w", err)
		}
		log.Println("Schema created successfully")
	}

	// Create homes table
	log.Println("Creating homes table...")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tibber.homes (
			id VARCHAR(255) PRIMARY KEY,
			address TEXT,
			postal_code VARCHAR(20),
			city VARCHAR(100),
			country VARCHAR(100),
			latitude FLOAT,
			longitude FLOAT,
			timezone VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating homes table: %w", err)
	}
	log.Println("Homes table created successfully")

	// Create production table
	log.Println("Creating production table...")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tibber.production (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			production FLOAT,
			profit FLOAT,
			currency VARCHAR(10),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating production table: %w", err)
	}
	log.Println("Production table created successfully")

	// Create consumption table
	log.Println("Creating consumption table...")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tibber.consumption (
			id SERIAL PRIMARY KEY,
			home_id VARCHAR(255) REFERENCES tibber.homes(id),
			from_time TIMESTAMP WITH TIME ZONE,
			to_time TIMESTAMP WITH TIME ZONE,
			consumption FLOAT,
			cost FLOAT,
			currency VARCHAR(10),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating consumption table: %w", err)
	}
	log.Println("Consumption table created successfully")

	log.Println("Database schema initialized successfully")
	return nil
}
