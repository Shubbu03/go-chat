package config

import (
	"os"
)

type Config struct {
	Port  string
	DB_URL string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:  getEnvOrDefault("PORT", "8080"),
		DB_URL: getEnvOrDefault("DB_URL", "postgres://postgres:password@localhost:5432/chat_db?sslmode=disable"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
