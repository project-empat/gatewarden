.PHONY: dev build test lint migrate clean

# ─── Development ──────────────────────────────────────────────────────────────

dev: dev-api dev-web

dev-api:
	cd apps/api && air

dev-web:
	cd apps/web && pnpm dev

# ─── Build ────────────────────────────────────────────────────────────────────

build: build-api build-web build-agent

build-api:
	cd apps/api && go build -o ../../bin/api ./cmd/api

build-web:
	cd apps/web && pnpm build

build-agent:
	cd agent && go build -o ../bin/gatewarden-agent ./cmd/agent

build-agent-static:
	cd agent && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/gatewarden-agent ./cmd/agent

build-enterprise:
	cd apps/api && go build -tags enterprise -o ../../bin/api-enterprise ./cmd/api

# ─── Test ─────────────────────────────────────────────────────────────────────

test: test-api test-agent

test-api:
	cd apps/api && go test ./... -v -count=1

test-agent:
	cd agent && go test ./... -v -count=1

test-web:
	cd apps/web && pnpm typecheck

test-all: test test-web

# ─── Database ─────────────────────────────────────────────────────────────────

migrate-up:
	cd apps/api && go run ./cmd/migrate up

migrate-down:
	cd apps/api && go run ./cmd/migrate down

migrate-create:
	cd apps/api && go run ./cmd/migrate create $(name)

# ─── Lint ─────────────────────────────────────────────────────────────────────

lint: lint-api lint-agent lint-web

lint-api:
	cd apps/api && golangci-lint run ./...

lint-agent:
	cd agent && golangci-lint run ./...

lint-web:
	cd apps/web && pnpm lint

# ─── Docker ───────────────────────────────────────────────────────────────────

docker-build:
	docker compose -f deploy/docker-compose.yml build

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

docker-logs:
	docker compose -f deploy/docker-compose.yml logs -f

# ─── Clean ────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/
	cd apps/api && go clean
	cd agent && go clean
	cd apps/web && rm -rf dist

# ─── Tools ────────────────────────────────────────────────────────────────────

tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
