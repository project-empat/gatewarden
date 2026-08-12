# Architecture

## Repository Layout

```
gatewarden/
├── apps/
│   ├── api/                 # Go REST API (monolith)
│   │   ├── cmd/             # Entry point
│   │   ├── internal/        # Handlers, middleware, services, db
│   │   ├── migrations/      # Goose migrations
│   │   └── Dockerfile
│   └── web/                 # React SPA
│       ├── src/
│       │   ├── routes/      # TanStack Router route tree
│       │   ├── pages/       # Page components
│       │   ├── components/  # Shared UI components
│       │   ├── stores/      # Zustand stores
│       │   ├── hooks/       # TanStack Query hooks
│       │   ├── lib/         # API client, utilities
│       │   └── types/       # TypeScript types
│       ├── vite.config.ts
│       └── Dockerfile
│
├── agent/                   # Go Linux agent (static binary)
│   ├── cmd/agent/           # Entry point
│   ├── internal/
│   │   ├── cloudflare/      # Cloudflare Tunnel — [detail](cloudflare-features.md)
│   │   ├── tailscale/       # Tailscale — [detail](tailscale-features.md)
│   │   ├── crowdsec/        # CrowdSec — [detail](crowdsec-features.md)
│   │   ├── fail2ban/        # Fail2Ban — [detail](fail2ban-features.md)
│   │   ├── docker/          # Docker security — [detail](docker-security-features.md)
│   │   ├── firewall/        # Firewall management — [detail](firewall-features.md)
│   │   ├── ssh/             # SSH hardening — [detail](ssh-hardening-features.md)
│   │   ├── journald/        # Auth log parsing — [detail](authlog-features.md)
│   │   └── reporter/        # Status reporting to API
│   └── pkg/
│       └── proto/           # Shared protocol types
│
├── modules/
│   ├── core/                # Shared Go core library
│   └── enterprise/          # Enterprise features (stub in OSS)
│
├── deploy/                  # Deployment artifacts
│   ├── docker-compose.yml
│   ├── install.sh
│   └── systemd/
│
├── migrations/              # PostgreSQL migrations (Goose)
├── scripts/                 # Dev scripts
├── docs/                    # Documentation
├── go.work                  # Go workspace
└── Makefile                 # Build targets
```

## Technology Decisions

### Why Go (agent and backend)

- Static binaries (single-file deployment)
- Easy cross-compilation (amd64, arm64)
- Linux process interaction (signals, exec)
- Docker API integration (client SDK)
- Firewall/system interaction (nftables, UFW)
- Low-resource agents (ideal for VPSes)
- Easy deployment on remote machines

### Why Vite + React (frontend)

- Fast development (HMR)
- Simple architecture (no SSR needed)
- Ideal for dashboard applications

### Why TanStack Router

- Type-safe routing
- Built-in search param handling
- Better DX than React Router for this scale

## Data Flow

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│  Agent      │────▶│  API Server  │────▶│ PostgreSQL │
│  (Go binary)│     │  (Go monolith)│     │            │
│             │     │              │     │            │
│ - reports   │     │ - /api/v1/* │     │ - nodes    │
│   status    │     │ - SSE/WS    │     │ - incidents│
│ - heartbeats│     │ - auth      │     │ - events   │
│ - logs      │     │ - webhooks  │     │ - users    │
└─────────────┘     └──────┬───────┘     └────────────┘
                           │
                    ┌──────▼───────┐
                    │  Web App     │
                    │  (React SPA) │
                    │              │
                    │ - Dashboard  │
                    │ - Nodes      │
                    │ - Incidents  │
                    │ - Policies   │
                    │ - Settings   │
                    └──────────────┘
```

- Agent connection is **outbound-only** (agent connects to API, not the other way)
- Real-time updates via SSE (server-sent events) or WebSocket
- API is a monolith — no microservices in MVP

## Enterprise Architecture

```
Public repo:  gatewarden/           (OSS, AGPLv3)
              ├── modules/core/
              └── modules/enterprise_stub/  (OSS stub — returns ErrEnterpriseOnly; real impls in gatewarden-enterprise)

Private repo: gatewarden-enterprise/  (commercial license)
              └── enterprise/         (real implementation)
```

Go build tags switch between stub and real implementation:

```go
//go:build enterprise
```

- OSS build: `go build ./...` (uses stubs)
- Enterprise build: `go build -tags enterprise ./...` (real features)
- Local dev: `replace` directive in go.mod for the private repo
