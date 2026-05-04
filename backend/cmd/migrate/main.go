package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"io/ioutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:123@localhost:5432/surplusslot?sslmode=disable"
	}

	fmt.Printf("Connecting to %s...\n", dsn)
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Ping failed: %v\n", err)
	}
	fmt.Println("Connected!")

	files := []string{
		"migrations/001_init.up.sql",
		"migrations/002_seed.up.sql",
	}

	for _, f := range files {
		fmt.Printf("Running %s...\n", f)
		content, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalf("Read file error: %v\n", err)
		}

		_, err = db.Exec(ctx, string(content))
		if err != nil {
			log.Fatalf("Exec error in %s: %v\n", f, err)
		}
	}

	fmt.Println("Migrations completed successfully!")
}
