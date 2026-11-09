package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/vivandesaii/trackd/internal/config"
	"github.com/vivandesaii/trackd/internal/database"
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

	pool, err := database.Connect(context.Background(), cfg.DatabaseURL)
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

	log.Printf("listening on :%s", cfg.Port)
	return http.ListenAndServe(":"+cfg.Port, mux)
}
