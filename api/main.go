package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"capacity/api/config"
	"capacity/api/handlers"
	"capacity/api/middleware"
	"capacity/api/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	var pingErr error
	for i := 0; i < cfg.DBPingRetries; i++ {
		if pingErr = db.Ping(ctx); pingErr == nil {
			break
		}
		log.Printf("waiting for database (%d/%d): %v", i+1, cfg.DBPingRetries, pingErr)
		time.Sleep(cfg.DBPingWait)
	}
	if pingErr != nil {
		log.Fatalf("ping: %v", pingErr)
	}

	api := &handlers.API{
		Capacity: services.NewCapacityService(db),
		People:   services.NewPeopleService(db),
		Health:   services.NewHealthService(db),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", api.HandleHealth)
	mux.HandleFunc("GET /api/capacity", api.HandleCapacity)
	mux.HandleFunc("PATCH /api/people/{id}", api.HandleUpdatePerson)

	handler := middleware.Chain(
		mux,
		middleware.Recover,
		middleware.RequestLog,
		middleware.SecurityHeaders,
		middleware.MaxBody(1<<20), // 1 MiB
	)

	log.Printf("listening on %s", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, handler))
}
