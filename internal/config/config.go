package config

import (
	"errors"
	"fmt"
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
		// 2. If it IS set, strictly parse it. If it's a typo (like "abc"), fail immediately!
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return Config{}, fmt.Errorf("invalid DB_MAX_ATTEMPTS %q: %w", val, err)
		}
		// 3. Prevent 0 or negative numbers which would break the retry loop logic
		if parsed <= 0 {
			return Config{}, errors.New("DB_MAX_ATTEMPTS must be greater than 0")
		}
		maxAttempts = parsed
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
