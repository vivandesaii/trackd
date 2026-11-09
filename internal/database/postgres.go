package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a connection pool and checks the database is reachable.
func Connect(ctx context.Context, databaseURL string, maxAttempts int) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	delay := 1 * time.Second
	var lastErr error

	for i := 1; i <= maxAttempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		lastErr = pool.Ping(pingCtx)
		cancel()

		if lastErr == nil {
			return pool, nil
		}

		log.Printf("db not ready, attempt %d/%d: %v", i, maxAttempts, lastErr)

		if i == maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("waiting for database: %w", ctx.Err())
		case <-time.After(delay):
			delay *= 2
		}
	}

	pool.Close()
	return nil, fmt.Errorf("database failed to become ready after %d attempts: %w", maxAttempts, lastErr)
}
