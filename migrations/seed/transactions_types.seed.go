package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedTransactionTypes(ctx context.Context, db *pgxpool.Pool) error {
	query := `
		INSERT INTO transaction_types (type_name)
		VALUES 
			('Top-Up'),
			('Payment'),
			('Transfer');
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		log.Printf("Failed to seed transaction types: %v", err)
		return err
	}

	log.Println("Seeded transaction types successfully.")
	return nil
}
