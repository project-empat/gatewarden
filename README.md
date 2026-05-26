# Gatewarden

Infrastructure security control plane. Monitor, manage, and harden your Linux infrastructure from a single dashboard.

## Architecture

```
gatewarden/
├── apps/
│   ├── api/        # Go REST API (monolith, chi router)
│   └── web/        # React SPA (Vite, TanStack, TailwindCSS, daisyui)
├── agent/          # Go Linux agent (static binary)
├── modules/
│   ├── core/       # Shared core library
│   └── enterprise/ # Enterprise features (stub in OSS)
├── deploy/         # Docker Compose, systemd, install scripts
├── migrations/     # PostgreSQL migrations (Goose)
└── docs/           # Documentation
```

## Quick Start

```bash
# Prerequisites: Go 1.22+, Node.js 20+, pnpm, Docker

# Clone and enter
git clone https://github.com/project-empat/gatewarden.git
cd gatewarden

# Start all services
make dev

# Or with Docker
make docker-up
```

## Development

### Backend

```bash
# Start API with hot reload
make dev-api

# Run tests
make test-api

# Run migrations
make migrate-up
```

### Frontend

```bash
# Start web dev server
make dev-web

# Build for production
make build-web
```

### Agent

```bash
# Build static binary
make build-agent-static

# The binary will be at ./bin/gatewarden-agent
```

### Full build

```bash
make build
```

## Environment Variables

| Variable              | Default            | Description          |
|-----------------------|--------------------|----------------------|
| `GATEWARDEN_PORT`     | `8080`             | API server port      |
| `GATEWARDEN_DB_DSN`   | —                  | PostgreSQL DSN       |
| `GATEWARDEN_JWT_SECRET` | —                | JWT signing secret   |
| `GATEWARDEN_LOG_LEVEL` | `info`            | Log level            |
| `GATEWARDEN_ALLOWED_ORIGINS` | `*`        | CORS origins         |

## Agent Installation

```bash
# On target Ubuntu/Debian server
curl -fsSL https://get.gatewarden.dev/install.sh | bash
```

Or manually:

```bash
sudo ./deploy/install.sh
```

## License

AGPLv3 — see [LICENSE](LICENSE).

## Enterprise

Enterprise features are available under a commercial license. The OSS build includes stubs — drop-in replacements that compile but return `ErrEnterpriseOnly`.
