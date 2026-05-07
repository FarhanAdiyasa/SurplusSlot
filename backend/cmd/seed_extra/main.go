package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:123@localhost:5432/surpluslot?sslmode=disable"
	}

	fmt.Printf("Connecting to %s...\n", dsn)
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	now := time.Now()
	start := now.Add(1 * time.Hour)
	end := now.Add(4 * time.Hour)

	offers := []struct {
		Merchant    string
		Title       string
		Description string
		PriceCents  int
		Stock       int
	}{
		{"Artisan Bakery", "Sourdough Surplus Bag", "3 assorted sourdough loaves and 2 baguettes.", 1200, 5},
		{"Green Garden", "Salad Bowl Combo", "Mixed organic salad with house dressing.", 750, 8},
		{"Sushi Stop", "Evening Sushi Pack", "12 pieces of assorted nigiri and rolls.", 1500, 4},
		{"The Donut Hole", "Dozen Mystery Box", "12 assorted fresh donuts from today's batch.", 1000, 10},
	}

	for _, o := range offers {
		fmt.Printf("Adding offer from %s...\n", o.Merchant)
		id := fmt.Sprintf("offer-%d", time.Now().UnixNano())
		_, err := db.Exec(ctx, `
			INSERT INTO offers (id, merchant, title, description, price_cents, stock, pickup_start, pickup_end, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active')
		`, id, o.Merchant, o.Title, o.Description, o.PriceCents, o.Stock, start, end)
		if err != nil {
			log.Printf("Failed to add offer: %v\n", err)
		}
	}

	fmt.Println("Seed data added successfully!")
}
