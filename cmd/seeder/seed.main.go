package main

import (
	"context"
	"log"

	"github.com/redha28/foomlet/internal/config"
	"github.com/redha28/foomlet/migrations/seed"
	"github.com/redha28/foomlet/pkg"
)

func main() {
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	pg, err := pkg.Posql()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer pg.Close()
	log.Println("DB connected successfully")

	ctx := context.Background()

	// 1. Seed transaction types first
	if err := seed.SeedTransactionTypes(ctx, pg); err != nil {
		log.Fatalf("Failed to seed transaction types: %v", err)
	}

	// 2. Seed initial users and sample transactions
	if err := seed.SeedInitialData(ctx, pg); err != nil {
		log.Fatalf("Failed to seed initial data: %v", err)
	}

	log.Println("All seeding completed successfully.")
}
