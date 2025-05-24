package pkg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Posql() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DB_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return dbpool, nil
}
