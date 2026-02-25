package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_DSN string
	Port   string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DB_DSN: os.Getenv("DB_DSN"),
		Port:   getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
