#!/bin/sh
set -e

echo "Running DB migration..."

migrate -path ./migrations -database "$DB_URL" -verbose up

echo "Seeding data..."
go run ./cmd/seeder/seed.main.go

echo "Starting the app..."
exec ./main
