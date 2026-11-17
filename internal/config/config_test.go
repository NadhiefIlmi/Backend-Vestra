package config

import (
	"os"
	"testing"
)

func TestConfigLoadAndValidate(t *testing.T) {
	// Test Default Load
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid default config, got err: %v", err)
	}

	if cfg.HTTPPort != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.HTTPPort)
	}

	// Test Invalid Port Validation
	cfg.HTTPPort = 999999
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error for invalid port, got nil")
	}

	// Test Redacted String does not leak credentials
	cfg.StorageSecretKey = "super-secret-password-12345"
	redacted := cfg.RedactedString()
	if contains(redacted, "super-secret-password-12345") {
		t.Errorf("redacted config contains raw secret key: %s", redacted)
	}
}

func TestConfigProductionValidation(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	cfg := &Config{
		AppEnv:           "production",
		HTTPPort:         8080,
		DatabaseURL:      "",
		StorageSecretKey: "minioadmin",
	}

	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error for empty DATABASE_URL in production, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

