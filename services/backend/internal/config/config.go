package config

import (
	"log"
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
	env := strings.ToLower(os.Getenv("ENV"))

	if env != "production" {
		log.Print("\n\n" +
			"\033[31m*****************************************************************\033[0m\n" +
			"\033[31m* WARNING: You are running the application in DEVELOPMENT mode! *\033[0m\n" +
			"\033[31m* This setup is NOT SECURE and should NOT be used in production.*\033[0m\n" +
			"\033[31m* To run in production, set the ENV environment variable to     *\033[0m\n" +
			"\033[31m* 'production' and ensure all other environment variables       *\033[0m\n" +
			"\033[31m* (e.g. MONGO_URI, CORS_ORIGINS, PORT) are configured securely. *\033[0m\n" +
			"\033[31m*****************************************************************\033[0m\n" +
			"\n")
	}

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
