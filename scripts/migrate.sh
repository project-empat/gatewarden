#!/bin/bash
set -euo pipefail

CMD="${1:-up}"

case "$CMD" in
  up)
    echo "Running migrations up..."
    (cd apps/api && go run ./cmd/migrate up)
    ;;
  down)
    echo "Running migrations down..."
    (cd apps/api && go run ./cmd/migrate down)
    ;;
  create)
    NAME="${2:-}"
    if [ -z "$NAME" ]; then
      echo "Usage: $0 create <migration_name>"
      exit 1
    fi
    echo "Creating migration: $NAME"
    (cd apps/api && go run ./cmd/migrate create "$NAME")
    ;;
  *)
    echo "Usage: $0 [up|down|create <name>]"
    exit 1
    ;;
esac
