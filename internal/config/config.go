package config

import (
	"errors"
	"os"
)

// Config holds every setting the app needs. Only main.go calls Load,
// so the rest of the code never touches os.Getenv directly.
type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	return cfg, nil
}

// getEnv is lowercase, so it's private to this package.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
