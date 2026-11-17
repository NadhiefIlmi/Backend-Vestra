package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/config"
)

func NewConnection(cfg *config.Config) (*sql.DB, error) {
	// Note: We use standard database/sql interface. Drivers like pgx/pq can register via _ import
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

