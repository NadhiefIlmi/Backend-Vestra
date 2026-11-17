package database

import (
	"testing"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/config"
)

func TestNewConnection_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL:        "invalid-url://unknown-host:5432/db",
		DBMaxOpenConns:     10,
		DBMaxIdleConns:     5,
		DBConnMaxLifetime: 5 * time.Minute,
	}

	db, err := NewConnection(cfg)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatal("expected error with unreachable / invalid database driver, got nil")
	}
}

