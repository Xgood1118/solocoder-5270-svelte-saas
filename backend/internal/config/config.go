package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          int
	DBPath        string
	JWTSecret     string
	WebhookSecret string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	return &Config{
		Port:          port,
		DBPath:        getEnv("DB_PATH", "saas.db"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		WebhookSecret: getEnv("WEBHOOK_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
