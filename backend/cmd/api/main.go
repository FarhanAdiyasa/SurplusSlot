package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"surplusslot/backend/internal/httpapi"
	"surplusslot/backend/internal/store"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/surplusslot?sslmode=disable")
	addr := getEnv("API_ADDR", "127.0.0.1:8080")

	ctx := context.Background()
	var st httpapi.Store
	db, err := store.New(ctx, dsn)
	if err == nil {
		err = db.Ping(ctx)
	}

	if err != nil {
		log.Printf("database connection failed, falling back to memory store: %v", err)
		st = store.NewMemoryStore()
	} else {
		defer db.Close()
		st = store.NewPostgresStore(db)
	}

	handler := httpapi.NewServer(st)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
