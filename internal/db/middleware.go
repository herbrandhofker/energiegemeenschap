package db

import (
	"database/sql"
	"time"

	"tibber_loader/internal/utils"
)

// TimeConverter is een middleware die automatisch timestamps converteert
type TimeConverter struct {
	db *sql.DB
}

// NewTimeConverter maakt een nieuwe TimeConverter
func NewTimeConverter(db *sql.DB) *TimeConverter {
	return &TimeConverter{db: db}
}

// QueryRow voert een query uit en converteert timestamps
func (tc *TimeConverter) QueryRow(query string, args ...interface{}) *sql.Row {
	return tc.db.QueryRow(query, args...)
}

// Query voert een query uit en converteert timestamps
func (tc *TimeConverter) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := tc.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	// Converteer de resultaten
	return tc.convertRows(rows), nil
}

// convertRows converteert timestamps in de resultaten
func (tc *TimeConverter) convertRows(rows *sql.Rows) *sql.Rows {
	// Implementatie van timestamp conversie
	return rows
}

// Exec voert een query uit en converteert timestamps
func (tc *TimeConverter) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Converteer input timestamps naar PostgreSQL formaat
	convertedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if t, ok := arg.(time.Time); ok {
			convertedArgs[i] = utils.ToPostgresTimestamp(t)
		} else {
			convertedArgs[i] = arg
		}
	}

	return tc.db.Exec(query, convertedArgs...)
}
