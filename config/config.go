package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

// GetEnv gets an environment variable or returns a default value if not present
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
