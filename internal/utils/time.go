time.go

// internal/utils/time.go
package utils

import (
    "time"
)

// ToISOTimestamp converteert een PostgreSQL timestamp naar ISO 8601
func ToISOTimestamp(pgTime time.Time) string {
    return pgTime.UTC().Format(time.RFC3339Nano)
}

// FromISOTimestamp converteert een ISO 8601 string naar time.Time
func FromISOTimestamp(isoTime string) (time.Time, error) {
    return time.Parse(time.RFC3339Nano, isoTime)
}

// ToPostgresTimestamp converteert een time.Time naar PostgreSQL timestamp
func ToPostgresTimestamp(t time.Time) time.Time {
    return t.UTC()
}