package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	MongoURI     string
	CORSOrigins  string
	SecureCookie bool
}

func Load() *Config {
	return &Config{
		Port:         getEnvOrDefault("PORT", "8081"),
		MongoURI:     getEnvOrDefault("MONGO_URI", "mongodb://devroot:devpassword@localhost:27017"),
		CORSOrigins:  getEnvOrDefault("CORS_ORIGINS", "http://localhost:5174"),
		SecureCookie: strings.ToLower(os.Getenv("ENV")) == "production",
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
