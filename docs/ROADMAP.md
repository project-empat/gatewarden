# Gatewarden Roadmap

## Phase 0: Foundation (MVP)

**Goal:** Working agent + API + basic UI that proves the concept.

### Agent
- [x] Go-based Linux agent (static binary)
- [x] Ubuntu/Debian support, amd64 + arm64
- [x] Docker discovery (list containers, inspect ports)
- [x] Open port scanning
- [x] SSH hardening checks
- [x] Firewall inspection (UFW/nftables)
- [x] CrowdSec status reporting
- [x] Tailscale status reporting
- [x] Cloudflare Tunnel status reporting
- [x] journald/auth log parsing
- [x] Outbound-only agent connection to API
- [x] Agent graceful shutdown
- [x] Agent reconnection on network loss

### Backend
- [x] Golang monolith (Chi router)
- [x] PostgreSQL database (sqlc + Goose migrations)
- [x] Agent registration and authentication
- [x] Agent heartbeat / status ingestion
- [x] SSE for real-time updates
- [x] Environment-based config
- [x] Structured logging
- [x] Initial schema: nodes, incidents, events, policies, settings

### Frontend
- [x] Vite + React + TypeScript scaffold
- [x] TanStack Router + TanStack Query
- [x] TailwindCSS + daisyUI + lucide icons
- [x] Zustand state management
- [x] Dark mode
- [x] Login page
- [x] Dashboard page
- [x] Nodes list + detail pages
- [x] Incidents list + detail pages
- [x] Policies page (list + create/edit modal)
- [x] Settings page
- [x] Error boundaries
- [x] Loading/empty states across pages

### Deployment
- [x] Docker Compose (API + DB + web proxy)
- [x] `install.sh` one-liner for agent (supports amd64/arm64)
- [x] systemd service for agent
- [x] Makefile targets for dev, build, test

## Phase 1: Secure Connectivity

**Goal:** Users can manage tunnels and access policies from one place.

- [ ] Cloudflare Zero Trust integration (list/manage tunnels) — [full spec](cloudflare-features.md)
- [ ] Tailscale integration (list nodes, check ACLs) — [full spec](tailscale-features.md)
- [ ] MFA enforcement visibility
- [ ] Identity-aware access mapping
- [ ] UI for "Expose privately", "Require MFA", "Restrict to team"

## Phase 2: Host Security

**Goal:** Hardened hosts with automatic attack detection and remediation.

- [ ] CrowdSec integration (alerts, decisions, metrics) — [full spec](crowdsec-features.md)
- [ ] Fail2Ban management (jails, bans, whitelist, config) — [full spec](fail2ban-features.md)
- [ ] UFW/nftables rule management from UI — [full spec](firewall-features.md)
- [ ] SSH hardening automation — [full spec](ssh-hardening-features.md)
- [ ] Docker exposure scanning and alerts — [full spec](docker-security-features.md)
- [ ] Brute-force blocking awareness
- [ ] Geo-blocking configuration
- [ ] Suspicious IP detection
- [ ] Automatic remediation actions

## Phase 3: Operational Visibility

**Goal:** At-a-glance security status across all nodes.

- [ ] Exposed services view
- [ ] Incident timeline and details
- [ ] Attack attempt feeds
- [ ] Authentication activity log — [full spec](authlog-features.md)
- [ ] Unhealthy systems alerting
- [ ] Active tunnels map
- [ ] Blocked IPs list
- [ ] Connected nodes inventory
- [ ] Infrastructure Security Graph POC — [full spec](security-graph.md)

## Phase 4: Enterprise & Scale

**Goal:** Multi-team, multi-node management with governance features.

- [ ] RBAC (roles, permissions)
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
- [ ] Hosted cloud control plane
- [ ] License key enforcement

## Explicit Out-of-Scope (for now)

- Kubernetes support
- Windows agent
- Full SIEM
- Generic observability / metrics dashboards
- Compliance reporting frameworks
- Plugin marketplace
- Custom query language
- Enterprise procurement features (quote-to-cash)

## See Also

- [Architecture](architecture.md)
- [Agent Features](agent-features.md)
- [Backend API](backend-api.md)
- [Frontend Pages](frontend-pages.md)
- [Deployment](deployment.md)
- [Licensing](licensing.md)
- [Progress](progress.md)
- [CrowdSec](crowdsec-features.md)
- [Cloudflare](cloudflare-features.md)
- [Tailscale](tailscale-features.md)
- [Firewall](firewall-features.md)
- [SSH Hardening](ssh-hardening-features.md)
- [Docker Security](docker-security-features.md)
- [Auth Log](authlog-features.md)
- [Fail2Ban](fail2ban-features.md)
- [Security Graph](security-graph.md)
- [Policies](policies-features.md)
