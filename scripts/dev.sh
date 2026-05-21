#!/bin/bash
set -euo pipefail

echo "Starting Gatewarden development environment..."

# Start PostgreSQL if not running
if ! docker ps --format '{{.Names}}' | grep -q 'gatewarden-postgres'; then
  echo "Starting PostgreSQL..."
  docker run -d \
    --name gatewarden-postgres \
    -e POSTGRES_USER=gatewarden \
    -e POSTGRES_PASSWORD=gatewarden \
    -e POSTGRES_DB=gatewarden \
    -p 5432:5432 \
    postgres:16-alpine

  echo "Waiting for PostgreSQL to be ready..."
  until docker exec gatewarden-postgres pg_isready -U gatewarden &> /dev/null; do
    sleep 1
  done
  echo "PostgreSQL is ready!"
fi

# Run migrations
echo "Running migrations..."
(cd apps/api && go run ./cmd/migrate up)

# Start API with hot reload
echo "Starting API with hot reload..."
(cd apps/api && air) &

# Start web dev server
echo "Starting web dev server..."
(cd apps/web && pnpm dev) &

echo "Development environment started!"
echo "  API:  http://localhost:8080"
echo "  Web:  http://localhost:5173"
echo "  DB:   postgres://gatewarden:gatewarden@localhost:5432/gatewarden"

wait
