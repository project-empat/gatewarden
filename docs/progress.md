# Progress

> Master tracking document. Cross-references all feature areas.
> Mark `[x]` as items are completed.

## Phase 0: Foundation

### Agent Scaffold

- [ ] `agent/cmd/agent/main.go` — entry point
- [ ] `agent/internal/` — package structure (cloudflare, tailscale, ssh, crowdsec, fail2ban, docker, firewall, journald, reporter)
- [ ] `agent/pkg/proto/` — shared report types
- [ ] Static binary build target (`make build-agent-static`)
- [ ] Systemd service file

See [Agent Features](agent-features.md).

### Backend Scaffold

- [ ] `apps/api/cmd/` — entry point
- [ ] Chi router + middleware stack
- [ ] PostgreSQL connection + Goose migrations runner
- [ ] sqlc generated queries
- [ ] Auth (JWT, registration, login, refresh)
- [ ] Agent registration + API key auth
- [ ] Report ingest endpoint
- [ ] Node CRUD
- [ ] Incident CRUD
- [ ] Event feed + SSE streaming
- [ ] Policy CRUD (MVP)
- [ ] Settings CRUD
- [ ] Environment-based config
- [ ] Structured logging

See [Backend API](backend-api.md).

### Database

- [ ] Migration: `001_initial_schema.sql` — users, nodes, agents, incidents, events, agent_reports, policies, settings

### Frontend Scaffold

- [ ] Vite + React + TypeScript project
- [ ] TailwindCSS + daisyUI config
- [ ] TanStack Router setup
- [ ] TanStack Query setup
- [ ] Zustand store (auth, theme)
- [ ] Dark mode
- [ ] API client layer
- [ ] Login page
- [ ] Dashboard page
- [ ] Nodes list + detail pages
- [ ] Incidents list + detail pages
- [ ] Policies page
- [ ] Settings page

See [Frontend Pages](frontend-pages.md).

### Deployment

- [ ] Docker Compose (api + db)
- [ ] `deploy/install.sh` agent installer
- [ ] `deploy/systemd/gatewarden-agent.service`
- [ ] Makefile targets

See [Deployment](deployment.md).

---

## Phase 1: Agent Integrations

### Port Scanning

- [ ] Listening port detection (`ss -tlnp` equivalent)
- [ ] Process-to-port mapping
- [ ] Public exposure detection (0.0.0.0 bindings)

### Docker Discovery

- [ ] Running container listing
- [ ] Published port detection
- [ ] Docker socket exposure check
- [ ] Container-to-image mapping

### Firewall Inspection

- [ ] UFW status and rules parsing
- [ ] nftables ruleset inspection
- [ ] Rule-to-port mapping
- [ ] Active/inactive detection

### SSH Hardening Check — [full spec](ssh-hardening-features.md)

- [ ] SSH config file parsing
- [ ] Password auth detection
- [ ] Root login setting
- [ ] Port configuration
- [ ] Public exposure flag
### auth.log / journald Parsing — [full spec](authlog-features.md)

- [ ] Failed SSH attempt detection
- [ ] Source IP aggregation
- [ ] Root login attempt detection
- [ ] sudo usage tracking
- [ ] Time-windowed counters (last hour, last 24h)

### CrowdSec Status — [full spec](crowdsec-features.md)

- [ ] CrowdSec installation detection
- [ ] Service running check
- [ ] Active decisions count
- [ ] Bouncer status
- [ ] Alert count (last hour)

### Tailscale Status — [full spec](tailscale-features.md)

- [ ] Tailscale installation detection
- [ ] Node name and IP
- [ ] Online status
- [ ] ACL version check

### Cloudflare Tunnel Status — [full spec](cloudflare-features.md)

- [ ] cloudflared installation detection
- [ ] Tunnel ID and name
- [ ] Ingress rules
- [ ] Running status

### Reporter Client

- [ ] HTTP POST report to API
- [ ] Heartbeat endpoint
- [ ] Retry with backoff
- [ ] Agent ID persistence

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

- [ ] Error handling throughout
- [ ] Input validation
- [ ] Structured logging coverage
- [ ] Agent graceful shutdown
- [ ] API rate limiting
- [ ] CORS configuration
- [ ] Database connection pooling
- [ ] Agent reconnection on network loss
- [ ] Frontend error boundaries
- [ ] Loading states on all pages
- [ ] Empty states on all pages
- [ ] Responsive layout
