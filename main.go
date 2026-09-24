package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"candidate-app/internal/product"
	"candidate-app/internal/web"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	repository := product.NewPostgresRepository(pool)
	service := product.NewService(repository)
	handler, err := web.NewHandler(service, "web/templates")
	if err != nil {
		log.Fatalf("create web handler: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:              getenv("HTTP_ADDR", ":8080"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
