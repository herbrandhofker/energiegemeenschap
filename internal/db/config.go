package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

// ParseURL parses a PostgreSQL URL into a Config struct
func ParseURL(url string) (*Config, error) {
	// Expected format: postgresql://username:password@host:port/dbname
	// or: postgresql://username:password@host/dbname (default port 5432)

	// Remove postgresql:// prefix if present
	url = strings.TrimPrefix(url, "postgresql://")

	// Split into credentials and host parts
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid URL format, missing @ separator")
	}

	// Parse credentials
	credentials := strings.Split(parts[0], ":")
	if len(credentials) != 2 {
		return nil, fmt.Errorf("invalid credentials format")
	}
	username := credentials[0]
	password := credentials[1]

	// Parse host and database
	hostParts := strings.Split(parts[1], "/")
	if len(hostParts) != 2 {
		return nil, fmt.Errorf("invalid host/database format")
	}

	// Parse host and port
	host := hostParts[0]
	port := 5432 // default PostgreSQL port
	if strings.Contains(host, ":") {
		hostPort := strings.Split(host, ":")
		host = hostPort[0]
		if p, err := strconv.Atoi(hostPort[1]); err == nil {
			port = p
		}
	}

	config := &Config{
		Host:     host,
		Port:     port,
		User:     username,
		Password: password,
		DBName:   hostParts[1],
	}

	return config, nil
}
