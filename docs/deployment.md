# Deployment

## Agent Installation

One-liner:

```bash
curl -sSL https://install.gatewarden.com | sudo bash
```

Manual:

```bash
sudo ./deploy/install.sh
```

### Install Script Flow

1. Detect OS / architecture (Ubuntu/Debian, amd64/arm64)
2. Download latest agent binary from releases
3. Verify checksum
4. Copy binary to `/usr/local/bin/gatewarden-agent`
5. Install systemd service from `deploy/systemd/gatewarden-agent.service`
6. Prompt for API server URL + registration token
7. Start and enable service
8. Output agent status

### Systemd Service

- User: `gatewarden` (created if not exists)
- ExecStart: `/usr/local/bin/gatewarden-agent`
- Restart: always
- Hardened with `ProtectSystem=strict`, `NoNewPrivileges=true`, etc.
- Logs to journald

## Backend / Control Plane

### Docker Compose (recommended for MVP)

```
deploy/docker-compose.yml
```

**Services:**

- `api` — Go API server
- `db` — PostgreSQL 16

**Volumes:**

- `pgdata` — PostgreSQL data

**Environment:**

- Injected via `.env` file or environment
- `GATEWARDEN_DB_DSN` — auto-configured for compose
- `GATEWARDEN_JWT_SECRET` — required

### Production Considerations

- Reverse proxy (Caddy / nginx) in front of API for TLS termination
- PostgreSQL connection pooling (PgBouncer optional for low scale)
- Secrets management (consider Doppler, 1Password, or env files)
- Backups: `pg_dump` cron or docker volume snapshots
- Agent API key generation at registration time

## Agent Connection Model

- Agents connect **outbound-only** to the API server
- No inbound ports required on agent side
- Agent polls or maintains persistent SSE/WS connection for commands
- API server never initiates connection to agent
- Works behind NAT, firewalls, and restrictive networks

## Installation Flow

```
1️⃣ Deploy backend (Docker Compose)
    ├── Start PostgreSQL
    ├── Run migrations
    └── Start API server

2️⃣ Configure API URL + JWT secret

3️⃣ Install agent on target machine
    ├── curl install script
    ├── Agent registers with API
    └── Agent begins reporting

4️⃣ Open dashboard
    ├── See node appear
    ├── Review security status
    └── Respond to incidents
```

## Build Artefacts

### Agent

```bash
make build-agent-static
# Output: ./bin/gatewarden-agent
```

Static binary, compatible with Ubuntu 20.04+, Debian 11+, amd64 and arm64.

### API

```bash
make build-api
# Output: ./bin/gatewarden-api
```

### Web

```bash
make build-web
# Output: ./apps/web/dist/
```

Served by API as static files or by nginx in production.

## Development

```bash
# Start everything with hot reload
make dev

# Or individual services
make dev-api     # API with air hot reload
make dev-web     # Vite dev server
```

### Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm
- Docker (for Postgres)

## Makefile Targets

| Target | Description |
|--------|-------------|
| `dev` | Start all dev services |
| `dev-api` | API with hot reload (air) |
| `dev-web` | Vite dev server |
| `build` | Build all |
| `build-api` | Build API binary |
| `build-web` | Build web static files |
| `build-agent-static` | Build agent static binary |
| `test` | Run all tests |
| `test-api` | Run API tests |
| `lint` | Lint all |
| `docker-up` | Start Docker Compose |
| `docker-down` | Stop Docker Compose |
| `migrate-up` | Run pending migrations |
| `migrate-down` | Rollback last migration |
| `clean` | Clean build artifacts |
