package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/your-org/go-template-project/internal/api"
	"github.com/your-org/go-template-project/internal/db"
	"github.com/your-org/go-template-project/internal/db/sqlc"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := db.RunMigrations(ctx, pool, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
	server := api.NewServer(sqlc.New(pool))
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Routes()))
}
