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
 [x] ACL version check (`tailscale debug acls`)

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

### File Integrity & Vulnerability (lightweight, not a SIEM)

- [x] FIM: periodic hashing of critical files (sshd, passwd, shadow, cron, systemd, ufw)
- [x] FIM: per-node baseline, change detection -> incidents (agent + API)
- [x] Packages: dpkg inventory + pending security-updates count (apt-check)
- [x] Vulnerability: OSV CVE matching, cached per package@version, background enrichment
- [x] Endpoints: /api/vulnerabilities, /api/nodes/{id}/vulnerabilities, /api/fim, /api/nodes/{id}/fim
- [x] Node detail UI: File Integrity + Patch & Vulnerabilities sections
- [x] Dedicated Vulnerabilities page (fleet CVEs, FIM changes, security updates)
- [x] Security summary: fleet-wide fim_changes / vulnerable_packages / security_updates_pending
- [ ] FIM realtime (fanotify) — enterprise tier

---

## Phase 2: Secure Connectivity

### Cloudflare Integration

 [x] API token configuration in settings
 [x] Tunnel listing from dashboard
 [x] Tunnel health monitoring
 [x] "Expose privately" action

### Tailscale Integration

 [x] API key configuration
 [x] Node listing from dashboard
 [x] ACL inspection alerts
 [x] "Restrict to team" visibility
 [x] "Require MFA" check

---

## Phase 3: Host Security

### Dashboard Security Summary

- [x] `GET /api/dashboard/security-summary` aggregation endpoint
- [x] Exposed services widget (SSH public, Docker socket, password auth)
- [x] CrowdSec widget (nodes with CrowdSec, decisions, alerts)
- [x] Fail2Ban widget (jails, current bans)
- [x] Incidents widget (open, high/critical, total)

### Agent Action System (Remediation Framework)

- [x] `agent_actions` table (migration 004)
- [x] Action creation API endpoint
- [x] Agent action polling (every 30s)
- [x] Action executors: fail2ban ban/unban, UFW allow/deny, restart service
- [x] Quick Actions UI on node detail page

### CrowdSec Dashboard Integration — [full spec](crowdsec-features.md)

- [x] Alert feed in incidents (via agent report analysis)
- [x] Decision count per node (dashboard + node detail)
 [x] Bouncer status per node
- [x] Geo-IP / suspicious IP highlighting

### Fail2Ban Dashboard Integration — [full spec](fail2ban-features.md)

- [x] Jail list per node (node detail page)
- [x] Per-jail detail view with ban/unban actions
- [x] Unban IP action from UI
- [x] Whitelist management UI
- [x] Incident generation from Fail2Ban events (via report analysis)
- [x] Policy integration (auto-create actions for matched incidents)

### Firewall Rule Management — [full spec](firewall-features.md)

- [x] Rule listing from UI (node detail page)
- [x] Add/remove rules (via action system)
- [x] Exposure-based suggestions (dashboard exposed services widget)

### Docker Security — [full spec](docker-security-features.md)

- [x] Exposure warnings in UI (node detail page)
- [x] Socket exposure alert (node detail + dashboard summary)
- [x] Container port audit (node detail page)

### Automated Remediation — [full spec](policies-features.md)

- [x] Block IP via firewall policy
- [x] Restrict Docker port policy
- [x] SSH hardening automation (via policy engine)

---

## Phase 4: Operational Visibility

### Infrastructure Security Graph

- [x] Relationship mapping between entities
- [x] Visual graph (internet → tunnel → host → container → service)
- [x] Click-through from graph nodes to details
- [x] Risk path highlighting

### Dashboard Enhancements

- [x] Security score calculation
- [x] Exposed services aggregate view
- [x] Attack timeline
- [x] Connectivity map

### Reporting

 [x] Security posture reports
 [x] Incident summary
 [x] Node health overview

---

## Phase 5: Enterprise

- [x] RBAC (basic admin/viewer roles, JWT role claims, admin gate)
- [x] Audit logs (audit trail, basic querying)
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
- [x] License key enforcement
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
