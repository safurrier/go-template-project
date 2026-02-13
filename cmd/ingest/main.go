package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/your-org/go-template-project/internal/db"
	"github.com/your-org/go-template-project/internal/db/sqlc"
	"github.com/your-org/go-template-project/internal/ingest"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	csvPath := os.Getenv("INGEST_CSV")
	if csvPath == "" {
		csvPath = "seed/arizona_games.csv"
	}
	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := db.RunMigrations(ctx, pool, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
	service := ingest.NewService(sqlc.New(pool))
	windows := []int32{5, 10}
	if w := os.Getenv("WINDOWS"); w != "" {
		windows = []int32{}
		for _, x := range strings.Split(w, ",") {
			if x == "5" {
				windows = append(windows, 5)
			}
			if x == "10" {
				windows = append(windows, 10)
			}
		}
	}
	if err := service.IngestCSV(ctx, csvPath, "arizona-mbb", "Arizona Wildcats", "Big 12", windows); err != nil {
		log.Fatal(err)
	}
	log.Println("ingest completed")
}
