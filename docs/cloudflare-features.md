# Cloudflare Zero Trust Integration

## Overview

Cloudflare Zero Trust (CFZT) provides secure connectivity through Argo Tunnels (`cloudflared`) and identity-aware access policies. Gatewarden integrates to detect tunnel status, report ingress rules, check Access Application enforcement, and manage tunnel lifecycle.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Installed | Check `cloudflared` binary | P0 |
| Running | Check `cloudflared` process or systemd service | P0 |
| Tunnels list | `cloudflared tunnel list` | P0 |
| Tunnel status | `cloudflared tunnel info <id>` | P0 |
| Ingress rules | Parse `config.yml` for tunnel ingress | P0 |
| Access Applications | Cloudflare API — list Access apps for the domain | P1 |
| Cert status | Check `~/.cloudflared/cert.pem` validity | P1 |
| Version | `cloudflared version` | P0 |
| Quick tunnels | Detect quick/random tunnel URLs | P1 |

### Agent Checks

- Is `cloudflared` installed?
- Is the cloudflared tunnel service running?
- What tunnels are configured? What is their status (active/inactive)?
- What are the ingress rules? Which local services are exposed?
- Is there an active Cloudflare Access policy protecting each exposed service?
- Are there any unauthenticated services routed through the tunnel?
- Is a quick/trycloudflare tunnel active (security risk)?
- What version of cloudflared is running?

### Agent Report Format

```jsonc
{
  "cloudflare_tunnel": {
    "installed": true,
    "running": true,
    "version": "2024.12.1",
    "tunnels": [
      {
        "id": "a1b2c3d4-e5f6-...",
        "name": "production-tunnel",
        "status": "active",
        "connections": 2,
        "ingress": [
          { "hostname": "grafana.example.com", "service": "http://localhost:3000", "path": "" },
          { "hostname": "api.example.com", "service": "http://localhost:8080", "path": "/api/*" }
        ],
        "has_access_policy": true,
        "origincert_expiry": "2027-01-15T00:00:00Z"
      }
    ],
    "access_applications": [
      { "domain": "grafana.example.com", "session_duration": "24h", "policies_count": 2, "mfa_required": true }
    ]
  }
}
```

## Management Actions

### Tunnel Management

| Action | Description | Priority |
|--------|-------------|----------|
| List tunnels | Show all configured tunnels with status | P0 |
| Tunnel detail | Connection count, ingress rules, cert expiry | P0 |
| Create tunnel | `cloudflared tunnel create <name>` | P2 |
| Delete tunnel | `cloudflared tunnel delete <name>` | P2 |
| Start tunnel | Start tunnel service | P1 |
| Stop tunnel | Stop tunnel service | P1 |
| Restart tunnel | Restart cloudflared service | P1 |

### Ingress Management

| Action | Description | Priority |
|--------|-------------|----------|
| List ingress rules | Show hostname → service mappings | P0 |
| Add ingress rule | Add new hostname mapping | P2 |
| Remove ingress rule | Remove a hostname mapping | P2 |
| "Expose privately" | Mark a service as tunnel-only with Access enforcement | P1 |
| "Require MFA" | Toggle MFA requirement on an Access Application | P1 |
| "Restrict to team" | Show identity-based access rules | P1 |

### Access Application Management (via Cloudflare API)

| Action | Description | Priority |
|--------|-------------|----------|
| List Access apps | All Access Applications for the account | P1 |
| Access policy detail | Groups, rules, session duration | P1 |
| Enable MFA | Require MFA for an application | P1 |
| Set session duration | Configure session length | P2 |
| Add policy rule | Add email/domain/IP restriction | P2 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/cloudflare/status` | Full Cloudflare tunnel status |
| GET | `/api/v1/nodes/:id/cloudflare/tunnels` | List tunnels |
| GET | `/api/v1/nodes/:id/cloudflare/tunnels/:name` | Tunnel detail |
| POST | `/api/v1/nodes/:id/cloudflare/tunnels/:name/restart` | Restart tunnel |
| GET | `/api/v1/nodes/:id/cloudflare/ingress` | List ingress rules |
| GET | `/api/v1/nodes/:id/cloudflare/access` | Access Application info |
| PUT | `/api/v1/settings/integrations/cloudflare` | Update Cloudflare API credentials |

### Cloudflare API Credentials

Stored in settings:

```jsonc
{
  "api_token": "****",              // Cloudflare API token (scoped)
  "account_id": "abc123",
  "zone_id": "def456",
  "domain": "example.com"
}
```

The Cloudflare API token requires permissions:
- `zone:read`
- `access:apps:read`, `access:apps:write`
- `access:policies:read`, `access:policies:write`

## Frontend UI

### Settings — Cloudflare Integration

- API token input (masked)
- Account ID
- Zone / Domain selector
- Connection test button
- Status: configured / not configured / error

### Node Detail — Cloudflare Section

**Tunnel card:**
- Tunnel name, ID, status badge
- Connection count
- Ingress rules table:
  - Hostname, local service, path
  - Has Access policy (yes/no badge)
  - Actions: "Expose privately", "Require MFA", "Restrict to team"

**Access Applications:**
- Domain, session duration, policy count
- MFA required indicator

### UX Labels & Actions

- **"Expose privately"** — ensures service is only accessible via Cloudflare Tunnel with Access Application. If no Access app exists, prompts to create one.
- **"Require MFA"** — toggles MFA requirement on the Access Application for that hostname.
- **"Restrict to team"** — shows current identity providers and group restrictions.

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| Tunnel disconnected | Tunnel status went from active to inactive | critical |
| Quick tunnel detected | Random trycloudflare URL active without auth | high |
| Expiring cert | Origin certificate expires in < 30 days | warning |
| Missing Access policy | Service routed through tunnel without Access | high |
| Cert expired | Origin certificate expired | critical |

## Agent Implementation Details

### Package Structure

```
agent/internal/cloudflare/
├── cloudflare.go       # Main integration
├── tunnel.go           # Tunnel detection and status
├── ingress.go          # Ingress rule parsing (config.yml)
├── access.go           # Access Application check (requires API token)
└── command.go          # Tunnel lifecycle commands
```

### Key Interactions

```bash
# Status
cloudflared tunnel list
cloudflared tunnel info <name>
cloudflared version

# Management
cloudflared tunnel create <name>
cloudflared tunnel delete <name>
systemctl restart cloudflared
```

### Config Files

```
~/.cloudflared/config.yml            # Main config (ingress rules)
~/.cloudflared/cert.pem              # Origin certificate
/etc/systemd/system/cloudflared*     # Systemd service
```

### Cloudflare API Endpoints Used

```
GET  /accounts/:id/access/apps
GET  /accounts/:id/access/apps/:id
POST /accounts/:id/access/apps
PUT  /accounts/:id/access/apps/:id
DELETE /accounts/:id/access/apps/:id
GET  /accounts/:id/access/apps/:id/policies
POST /accounts/:id/access/apps/:id/policies
```

## Implementation Order

1. Agent: cloudflared installed/running detection
2. Agent: tunnel list + status
3. Agent: ingress rule parsing (config.yml)
4. Agent: status report integration
5. Backend: API credential storage + validation
6. Backend: read-only endpoints (status, tunnels, ingress)
7. Frontend: Cloudflare integration settings page
8. Frontend: tunnel status on node detail
9. Backend: Cloudflare Access API integration (via API token)
10. Frontend: ingress rules table + "Expose privately" action
11. Backend: tunnel lifecycle commands
12. Agent: tunnel restart command
13. Incident generation
