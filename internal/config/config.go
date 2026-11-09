package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds every setting the app needs. Only main.go calls Load,
// so the rest of the code never touches os.Getenv directly.
type Config struct {
	Port          string
	DatabaseURL   string
	DBMaxAttempts int
}

func Load() (Config, error) {
	maxAttempts := 5

	if val, ok := os.LookupEnv("DB_MAX_ATTEMPTS"); ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxAttempts = parsed
		}
	}

	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		DBMaxAttempts: maxAttempts,
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
