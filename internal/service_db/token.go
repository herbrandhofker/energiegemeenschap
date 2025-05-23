package service_db

import (
	"context"
	"database/sql"
	"fmt"
)

type TokenService struct {
	DB *sql.DB
}

func NewTokenService(db *sql.DB) *TokenService {
	return &TokenService{DB: db}
}

// GetToken retrieves the most recent token from the database
func (s *TokenService) GetToken(ctx context.Context) (string, error) {
	var token string
	err := s.DB.QueryRowContext(ctx, `
		SELECT token 
		FROM tibber.tibber_tokens 
		ORDER BY created_at DESC 
		LIMIT 1
	`).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no token found in database")
		}
		return "", fmt.Errorf("error getting token: %w", err)
	}
	return token, nil
}

// StoreToken stores a new token in the database
func (s *TokenService) StoreToken(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO tibber.tibber_tokens (token)
		VALUES ($1)
	`, token)
	if err != nil {
		return fmt.Errorf("error storing token: %w", err)
	}
	return nil
}
