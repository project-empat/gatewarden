# Backend API

## Stack

- **Language:** Go 1.22+
- **Router:** Chi
- **Database:** PostgreSQL
- **Query layer:** sqlc (type-safe generated queries)
- **Migrations:** Goose
- **Real-time:** SSE or WebSocket
- **Config:** Environment-based
- **Logging:** Structured (zerolog or `slog`)

## API Endpoints (MVP)

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | Login, returns JWT |
| POST | `/api/v1/auth/register` | First-user registration |
| POST | `/api/v1/auth/refresh` | Refresh JWT |

### Agents

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/agents/register` | Agent registration (returns API key) |
| POST | `/api/v1/agents/:id/report` | Agent status report ingest |
| POST | `/api/v1/agents/:id/heartbeat` | Agent heartbeat |
| GET | `/api/v1/agents` | List all agents |
| GET | `/api/v1/agents/:id` | Get agent details |
| DELETE | `/api/v1/agents/:id` | Remove agent |

### Incidents

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/incidents` | List incidents |
| GET | `/api/v1/incidents/:id` | Incident details |
| PUT | `/api/v1/incidents/:id/status` | Update status (acknowledge, resolve) |

### Events

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/events` | Event feed (paginated, filterable) |
| GET | `/api/v1/events/stream` | SSE/WS real-time event stream |

### Nodes / Hosts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes` | List nodes (aggregated from agent reports) |
| GET | `/api/v1/nodes/:id` | Node details |
| GET | `/api/v1/nodes/:id/security-snapshot` | Current security posture |
| GET | `/api/v1/nodes/:id/fail2ban/status` | Fail2Ban status — [full spec](fail2ban-features.md) |
| GET | `/api/v1/nodes/:id/fail2ban/jails` | List Fail2Ban jails |
| GET | `/api/v1/nodes/:id/fail2ban/jails/:name` | Jail detail + banned IPs |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/unban` | Unban IP from jail |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/ban` | Manually ban IP |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/start` | Enable jail |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/stop` | Disable jail |
| PUT | `/api/v1/nodes/:id/fail2ban/jails/:name/config` | Update jail config |
| GET | `/api/v1/nodes/:id/fail2ban/whitelist` | View whitelist |
| POST | `/api/v1/nodes/:id/fail2ban/whitelist/add` | Add to whitelist |
| POST | `/api/v1/nodes/:id/fail2ban/whitelist/remove` | Remove from whitelist |
| GET | `/api/v1/nodes/:id/crowdsec/status` | CrowdSec status — [full spec](crowdsec-features.md) |
| GET | `/api/v1/nodes/:id/crowdsec/alerts` | CrowdSec alerts |
| GET | `/api/v1/nodes/:id/crowdsec/decisions` | CrowdSec decisions |
| POST | `/api/v1/nodes/:id/crowdsec/decisions` | Add ban decision |
| DELETE | `/api/v1/nodes/:id/crowdsec/decisions/:id` | Remove decision |
| GET | `/api/v1/nodes/:id/cloudflare/status` | Cloudflare tunnel status — [full spec](cloudflare-features.md) |
| GET | `/api/v1/nodes/:id/cloudflare/tunnels` | List tunnels |
| GET | `/api/v1/nodes/:id/cloudflare/ingress` | Ingress rules |
| GET | `/api/v1/nodes/:id/tailscale/status` | Tailscale status — [full spec](tailscale-features.md) |
| GET | `/api/v1/nodes/:id/tailscale/peers` | Tailscale peers |
| GET | `/api/v1/nodes/:id/firewall/status` | Firewall status — [full spec](firewall-features.md) |
| GET | `/api/v1/nodes/:id/firewall/rules` | Firewall rules |
| POST | `/api/v1/nodes/:id/firewall/rules` | Add firewall rule |
| DELETE | `/api/v1/nodes/:id/firewall/rules/:ruleId` | Delete firewall rule |
| POST | `/api/v1/nodes/:id/firewall/enable` | Enable firewall |
| POST | `/api/v1/nodes/:id/firewall/disable` | Disable firewall |
| GET | `/api/v1/nodes/:id/firewall/exposure-audit` | Port vs rule audit |
| GET | `/api/v1/nodes/:id/ssh/status` | SSH hardening status — [full spec](ssh-hardening-features.md) |
| POST | `/api/v1/nodes/:id/ssh/harden` | Apply SSH hardening |
| POST | `/api/v1/nodes/:id/ssh/disable-password-auth` | Disable SSH password auth |
| POST | `/api/v1/nodes/:id/ssh/disable-root-login` | Disable SSH root login |
| GET | `/api/v1/nodes/:id/docker/status` | Docker status — [full spec](docker-security-features.md) |
| GET | `/api/v1/nodes/:id/docker/containers` | List containers with security analysis |
| GET | `/api/v1/nodes/:id/docker/issues` | Aggregated security issues |
| GET | `/api/v1/nodes/:id/authlog/status` | Auth log status — [full spec](authlog-features.md) |
| GET | `/api/v1/nodes/:id/authlog/summary` | Auth log aggregated metrics |
| GET | `/api/v1/nodes/:id/authlog/failed-attempts` | Failed attempts + top IPs |
| GET | `/api/v1/nodes/:id/authlog/top-ips` | Top offending source IPs |
| GET | `/api/v1/graph` | Security graph — [full spec](security-graph.md) |
| GET | `/api/v1/graph/nodes/:id` | Sub-graph for a node |
| GET | `/api/v1/graph/risk-paths` | Risk path analysis |

### Policies (MVP basic)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/policies` | List policies |
| POST | `/api/v1/policies` | Create policy |
| PUT | `/api/v1/policies/:id` | Update policy |
| DELETE | `/api/v1/policies/:id` | Delete policy |

### Settings

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/settings` | Get settings |
| PUT | `/api/v1/settings` | Update settings |

## Database Schema (MVP)

### `users`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| email | TEXT | Unique |
| password_hash | TEXT | bcrypt |
| role | TEXT | `admin` / `viewer` |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### `nodes`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| agent_id | UUID | FK to agents |
| hostname | TEXT | |
| os | TEXT | Ubuntu 22.04 etc |
| kernel | TEXT | |
| ip_address | INET | |
| last_seen | TIMESTAMPTZ | |
| status | TEXT | `online` / `offline` / `degraded` |
| created_at | TIMESTAMPTZ | |

### `agents`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| api_key_hash | TEXT | |
| node_id | UUID | FK to nodes |
| version | TEXT | Agent version string |
| last_heartbeat | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |

### `incidents`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| node_id | UUID | FK |
| severity | TEXT | `info` / `warning` / `critical` |
| category | TEXT | `exposure` / `attack` / `misconfig` |
| title | TEXT | Short description |
| details | JSONB | Machine-readable payload |
| status | TEXT | `open` / `acknowledged` / `resolved` |
| created_at | TIMESTAMPTZ | |
| resolved_at | TIMESTAMPTZ | Nullable |

### `events`

| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL | PK |
| node_id | UUID | FK |
| type | TEXT | `auth_failure` / `port_exposed` / `docker_issue` / etc |
| severity | TEXT | |
| payload | JSONB | |
| created_at | TIMESTAMPTZ | |

### `agent_reports`

Stores the last N raw reports for debugging and audit.

| Column | Type | Notes |
|--------|------|-------|
| id | BIGSERIAL | PK |
| agent_id | UUID | FK |
| report | JSONB | Full report body |
| received_at | TIMESTAMPTZ | |

### `policies` (MVP — basic rule store)

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| name | TEXT | |
| description | TEXT | |
| enabled | BOOLEAN | |
| conditions | JSONB | Rule conditions |
| actions | JSONB | Remediation actions |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### `settings`

| Column | Type | Notes |
|--------|------|-------|
| key | TEXT | PK |
| value | JSONB | |

## Security

- Agent API keys are hashed with bcrypt on storage
- JWT for user sessions (short-lived access + refresh token)
- CORS restricted to allowed origins
- Input validation on all endpoints
- Rate limiting on auth endpoints

## Real-Time Updates

- SSE endpoint `/api/v1/events/stream` for real-time dashboard updates
- Events pushed: new incidents, status changes, agent online/offline, attack alerts
- WebSocket alternative if bidirectional communication needed later

## Config

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWARDEN_PORT` | `8080` | API server port |
| `GATEWARDEN_DB_DSN` | — | PostgreSQL DSN |
| `GATEWARDEN_JWT_SECRET` | — | JWT signing secret |
| `GATEWARDEN_LOG_LEVEL` | `info` | Log level |
| `GATEWARDEN_ALLOWED_ORIGINS` | `*` | CORS origins |

## Implementation Order

1. Scaffold Chi router + middleware (logging, CORS, recovery)
2. Database connection + Goose migrations
3. sqlc query generation setup
4. Auth: register, login, JWT middleware
5. Agent registration and API key auth
6. Agent report ingest endpoint
7. Node listing/details
8. Incident model + CRUD
9. Event model + feed
10. SSE streaming endpoint
11. Policy CRUD (MVP basic)
12. Settings CRUD
