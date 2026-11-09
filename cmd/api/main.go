package main

import (
	"context"
	"fmt"
	"github.com/vivandesaii/trackd/internal/config"
	"github.com/vivandesaii/trackd/internal/database"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run does the real work and returns errors instead of exiting.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxAttempts)
	if err != nil {
		return err
	}
	defer pool.Close() // runs when run() returns, whatever the reason

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	// runs the server in a goroutine which is a background thread
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// listen for either a server startup error OR a Ctrl+C signal (ctx.Done())
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		log.Println("shutting down server gracefully...")

		// give ongoing requests 5 seconds to complete before forcing shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			srv.Close() // Force close if graceful shutdown fails
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		log.Println("server stopped cleanly.")
	}

	return nil
}
