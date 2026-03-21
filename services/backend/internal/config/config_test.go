package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	keys := []string{"ENV", "PORT", "MONGO_URI", "CORS_ORIGINS"}

	// Save original environment variables and restore them after tests
	originalEnv := make(map[string]string)
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			originalEnv[key] = val
		}
	}
	defer func() {
		for _, key := range keys {
			if val, ok := originalEnv[key]; ok {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("Default values (development mode)", func(t *testing.T) {
		for _, key := range keys {
			os.Unsetenv(key)
		}

		cfg := Load()

		if cfg.Port != "8081" {
			t.Errorf("Expected default Port '8081', got %q", cfg.Port)
		}
		if cfg.MongoURI != "mongodb://devroot:devpassword@localhost:27017" {
			t.Errorf("Expected default MongoURI 'mongodb://devroot:devpassword@localhost:27017', got %q", cfg.MongoURI)
		}
		if cfg.CORSOrigins != "http://localhost:5174" {
			t.Errorf("Expected default CORSOrigins 'http://localhost:5174', got %q", cfg.CORSOrigins)
		}
		if cfg.SecureCookie != false {
			t.Errorf("Expected SecureCookie to be false by default, got true")
		}
	})

	t.Run("Custom values (production mode)", func(t *testing.T) {
		os.Setenv("ENV", "production")
		os.Setenv("PORT", "9090")
		os.Setenv("MONGO_URI", "mongodb://prod:pass@mongo:27017")
		os.Setenv("CORS_ORIGINS", "https://example.com")

		cfg := Load()

		if cfg.Port != "9090" {
			t.Errorf("Expected custom Port '9090', got %q", cfg.Port)
		}
		if cfg.MongoURI != "mongodb://prod:pass@mongo:27017" {
			t.Errorf("Expected custom MongoURI 'mongodb://prod:pass@mongo:27017', got %q", cfg.MongoURI)
		}
		if cfg.CORSOrigins != "https://example.com" {
			t.Errorf("Expected custom CORSOrigins 'https://example.com', got %q", cfg.CORSOrigins)
		}
		if cfg.SecureCookie != true {
			t.Errorf("Expected SecureCookie to be true in production, got false")
		}
	})
}
