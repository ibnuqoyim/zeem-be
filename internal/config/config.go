package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	AllowedOrigins []string
	Environment    string
	Host           string
}

func New() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	allowedOrigins := strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ",")
	environment := getEnv("ENV", "development")
	host := getEnv("HOST", "0.0.0.0")

	return &Config{
		Port:           port,
		AllowedOrigins: allowedOrigins,
		Environment:    environment,
		Host:           host,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
