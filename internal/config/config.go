package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv               string
	HTTPPort             int
	DatabaseURL          string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetime    time.Duration
	StorageEndpoint      string
	StorageBucket        string
	StorageAccessKey     string
	StorageSecretKey     string
	StorageUseSSL        bool
	CORSAllowedOrigins   []string
	RequestTimeout       time.Duration
	DecartAPIKey         string
}

func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:           getEnvAsInt("HTTP_PORT", 8080),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ar_tryon_dev?sslmode=disable"),
		DBMaxOpenConns:     getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime:  getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		StorageEndpoint:    getEnv("STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageBucket:      getEnv("STORAGE_BUCKET", "catalog"),
		StorageAccessKey:   getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey:   getEnv("STORAGE_SECRET_KEY", "minioadmin"),
		StorageUseSSL:      getEnvAsBool("STORAGE_USE_SSL", false),
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		RequestTimeout:     getEnvAsDuration("REQUEST_TIMEOUT_SECONDS", 15*time.Second),
		DecartAPIKey:       getEnv("DECART_API_KEY", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP_PORT: %d", c.HTTPPort)
	}
	if c.AppEnv == "production" {
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required in production")
		}
		if c.StorageSecretKey == "minioadmin" {
			return fmt.Errorf("default STORAGE_SECRET_KEY is forbidden in production")
		}
	}
	return nil
}

func (c *Config) RedactedString() string {
	return fmt.Sprintf(
		"AppEnv: %s, HTTPPort: %d, DBMaxOpenConns: %d, DBMaxIdleConns: %d, StorageBucket: %s, StorageEndpoint: %s, CORS: %v",
		c.AppEnv, c.HTTPPort, c.DBMaxOpenConns, c.DBMaxIdleConns, c.StorageBucket, c.StorageEndpoint, c.CORSAllowedOrigins,
	)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	// Try parsing as raw integer seconds
	if seconds, err := strconv.Atoi(valStr); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string) []string {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	parts := strings.Split(valStr, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
}

