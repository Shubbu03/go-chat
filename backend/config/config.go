package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port   string
	DB_URL string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := &Config{
		Port:   getEnvOrDefault("PORT", "8080"),
		DB_URL: getEnvOrDefault("DB_URL", ""),
	}

	return cfg, nil
}

func ConnectToDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DB_URL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
