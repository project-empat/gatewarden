# Progress

> Master tracking document. Cross-references all feature areas.
> Mark `[x]` as items are completed.

## Phase 0: Foundation

### Agent Scaffold

- [x] `agent/cmd/agent/main.go` — entry point
- [x] `agent/internal/` — package structure (cloudflare, tailscale, ssh, crowdsec, fail2ban, docker, firewall, journald, reporter)
- [x] `agent/pkg/proto/` — shared report types
- [x] Static binary build target (`make build-agent-static`)
- [x] Systemd service file
- [x] Agent reconnection on network loss (re-register on 401)

See [Agent Features](agent-features.md).

### Backend Scaffold

- [x] `apps/api/cmd/` — entry point
- [x] Chi router + middleware stack
- [x] PostgreSQL connection + Goose migrations runner
- [ ] sqlc generated queries (raw SQL via pgx used instead)
- [x] Auth (JWT, registration, login, refresh)
- [x] Agent registration + API key auth
- [x] Report ingest endpoint
- [x] Node CRUD
- [x] Incident CRUD
- [x] Event feed + SSE streaming
- [x] Policy CRUD (MVP)
- [x] Settings CRUD
- [x] Environment-based config
- [x] Structured logging

See [Backend API](backend-api.md).

### Database

- [x] Migration: `001_initial_schema.sql` + `002_agents_and_reports.sql` — users, nodes, agents, incidents, events, agent_reports, policies, settings

### Frontend Scaffold

- [x] Vite + React + TypeScript project
- [x] TailwindCSS + daisyUI config
- [x] TanStack Router setup
- [x] TanStack Query setup
- [x] Zustand store (auth, theme)
- [x] Dark mode
- [x] API client layer
- [x] Login page
- [x] Dashboard page
- [x] Nodes list + detail pages
- [x] Incidents list + detail pages
- [x] Policies page (list + create/edit modal)
- [x] Settings page
- [x] Error boundary wrapping root app

See [Frontend Pages](frontend-pages.md).

### Deployment

- [x] Docker Compose (api + db + web + migrate)
- [x] `deploy/install.sh` agent installer (supports amd64/arm64 + systemd)
- [x] `deploy/systemd/gatewarden-agent.service`
- [x] Makefile targets

See [Deployment](deployment.md).

---

## Phase 1: Agent Integrations

### Port Scanning

- [x] Listening port detection (`ss -tlnp` equivalent)
- [x] Process-to-port mapping
- [x] Public exposure detection (0.0.0.0 bindings)

### Docker Discovery

- [x] Running container listing
- [x] Published port detection
- [x] Docker socket exposure check
- [x] Container-to-image mapping

### Firewall Inspection

- [x] UFW status and rules parsing
- [x] nftables ruleset inspection
- [x] Rule-to-port mapping
- [x] Active/inactive detection

### SSH Hardening Check — [full spec](ssh-hardening-features.md)

- [x] SSH config file parsing
- [x] Password auth detection
- [x] Root login setting
- [x] Port configuration
- [x] Public exposure flag
### auth.log / journald Parsing — [full spec](authlog-features.md)

- [x] Failed SSH attempt detection
- [x] Source IP aggregation
- [x] Root login attempt detection
- [x] sudo usage tracking
- [x] Time-windowed counters (last hour, last 24h)

### CrowdSec Status — [full spec](crowdsec-features.md)

- [x] CrowdSec installation detection
- [x] Service running check
- [x] Active decisions count
- [x] Bouncer status
- [x] Alert count (last hour)

### Tailscale Status — [full spec](tailscale-features.md)

- [x] Tailscale installation detection
- [x] Node name and IP
- [x] Online status
- [ ] ACL version check

### Cloudflare Tunnel Status — [full spec](cloudflare-features.md)

- [x] cloudflared installation detection
- [x] Tunnel ID and name
- [x] Ingress rules
- [x] Running status

### Reporter Client

- [x] HTTP POST report to API
- [x] Heartbeat endpoint
- [x] Retry with backoff
- [x] Agent ID persistence (registration)

---

## Phase 2: Secure Connectivity

### Cloudflare Integration

- [ ] API token configuration in settings
- [ ] Tunnel listing from dashboard
- [ ] Tunnel health monitoring
- [ ] "Expose privately" action

### Tailscale Integration

- [ ] API key configuration
- [ ] Node listing from dashboard
- [ ] ACL inspection alerts
- [ ] "Restrict to team" visibility
- [ ] "Require MFA" check

---

## Phase 3: Host Security

### CrowdSec Dashboard Integration — [full spec](crowdsec-features.md)

- [ ] Alert feed in incidents
- [ ] Decision count per node
- [ ] Bouncer status per node
- [ ] Geo-IP / suspicious IP highlighting

### Fail2Ban Dashboard Integration — [full spec](fail2ban-features.md)

- [ ] Jail list per node — [full spec](fail2ban-features.md)
- [ ] Per-jail detail view (banned IPs, config, whitelist)
- [ ] Unban IP action from UI
- [ ] Jail enable/disable controls
- [ ] Whitelist management UI
- [ ] Incident generation from Fail2Ban events
- [ ] Policy integration (auto-enable jails, adjust thresholds)

### Firewall Rule Management — [full spec](firewall-features.md)

- [ ] Rule listing from UI
- [ ] Add/remove rules
- [ ] Rule suggestions based on exposure

### Docker Security — [full spec](docker-security-features.md)

- [ ] Exposure warnings in UI
- [ ] Socket exposure alert
- [ ] Container port audit

### Automated Remediation — [full spec](policies-features.md)

- [ ] Block IP via firewall policy
- [ ] Restrict Docker port policy
- [ ] SSH hardening automation

---

## Phase 4: Operational Visibility

### Infrastructure Security Graph

- [ ] Relationship mapping between entities
- [ ] Visual graph (internet → tunnel → host → container → service)
- [ ] Click-through from graph nodes to details
- [ ] Risk path highlighting

### Dashboard Enhancements

- [ ] Security score calculation
- [ ] Exposed services aggregate view
- [ ] Attack timeline
- [ ] Connectivity map

### Reporting

- [ ] Security posture reports
- [ ] Incident summary
- [ ] Node health overview

---

## Phase 5: Enterprise

- [ ] RBAC
- [ ] Audit logs
- [ ] SSO/SAML/OIDC
- [ ] Advanced policy engine
- [ ] Team management
- [ ] Approval workflows
- [ ] Advanced automation
- [ ] Compliance exports
- [ ] Premium alert routing
- [ ] MSP multi-tenancy
- [ ] Client isolation
- [ ] Delegated access
- [ ] Advanced reporting
- [ ] License key enforcement
- [ ] Hosted cloud control plane

---

## Cross-Cutting Concerns

- [x] Error handling throughout
- [x] Input validation
- [x] Structured logging coverage
- [x] Agent graceful shutdown
- [x] API rate limiting
- [x] CORS configuration
- [x] Database connection pooling
- [x] Agent reconnection on network loss
- [x] Frontend error boundaries
- [x] Loading states on all pages
- [x] Empty states on all pages
- [x] Responsive layout
