package db

import (
	"context"
	"database/sql"
	"embed"
	"log"

	_ "github.com/lib/pq"
)

//go:embed migrations.sql migrations_drop.sql
var migrationsFS embed.FS

// RunMigrations executes the database migrations
func RunMigrations(ctx context.Context, dbConn *sql.DB) error {
	// Read drop migrations file
	dropMigrations, err := migrationsFS.ReadFile("migrations_drop.sql")
	if err != nil {
		return err
	}

	// Execute drop migrations
	_, err = dbConn.ExecContext(ctx, string(dropMigrations))
	if err != nil {
		return err
	}

	// Read create migrations file
	createMigrations, err := migrationsFS.ReadFile("migrations.sql")
	if err != nil {
		return err
	}

	// Execute create migrations
	_, err = dbConn.ExecContext(ctx, string(createMigrations))
	if err != nil {
		return err
	}

	log.Printf("Database migrations completed successfully")
	return nil
}
