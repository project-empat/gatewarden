#!/bin/bash
set -euo pipefail

echo "=== Gatewarden Development Environment ==="

# Find an existing PostgreSQL container (handles Docker Swarm naming like dev-postgre_db.1.xxx)
PG_CONTAINER=$(docker ps --filter 'name=dev-postgre_db' --format '{{.Names}}' | head -1)

if [ -z "$PG_CONTAINER" ]; then
  echo "Starting PostgreSQL..."
  docker run -d \
    --name dev-postgres_db \
    -e POSTGRES_USER=gatewarden \
    -e POSTGRES_PASSWORD=gatewarden \
    -e POSTGRES_DB=gatewarden \
    -p 5432:5432 \
    postgres:16-alpine

  echo "Waiting for PostgreSQL to be ready..."
  until docker exec dev-postgres_db pg_isready -U gatewarden &> /dev/null; do
    sleep 1
  done
  echo "PostgreSQL is ready!"
  PG_CONTAINER="dev-postgres_db"
else
  echo "Found existing PostgreSQL container: ${PG_CONTAINER}"

  # Wait for it to accept connections if it's still starting
  until docker exec "$PG_CONTAINER" pg_isready -U gatewarden &> /dev/null; do
    echo "Waiting for PostgreSQL to be ready..."
    sleep 1
  done
fi

# Discover PostgreSQL host port
DB_PORT=$(docker inspect "$PG_CONTAINER" --format='{{(index (index .NetworkSettings.Ports "5432/tcp") 0).HostPort}}' 2>/dev/null || echo "5432")
echo "PostgreSQL ready on port ${DB_PORT}"

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
echo "  API:  http://localhost:8085"
echo "  Web:  http://localhost:5174"
echo "  DB:   postgres://gatewarden:gatewarden@localhost:${DB_PORT}/gatewarden"

wait
