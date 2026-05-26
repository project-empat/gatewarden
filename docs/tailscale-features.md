# Tailscale Integration

## Overview

Tailscale provides secure mesh networking based on WireGuard. Gatewarden integrates to report node connectivity, check ACL configurations, verify MFA enforcement, and surface identity-aware access issues.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Installed | Check `tailscale` binary | P0 |
| Running | Check `tailscaled` systemd service | P0 |
| Node info | `tailscale status` — self node name, IP | P0 |
| Online status | `tailscale status --json` — peer status | P0 |
| Node IP | `tailscale ip` | P0 |
| Version | `tailscale version` | P0 |
| DNS config | `tailscale dns status` | P1 |
| Exit node | Check if node is an exit node | P1 |
| Subnets | Check if node is advertising routes | P1 |
| ACL check | `tailscale debug acls` — last applied ACL hash | P2 |

### Agent Checks

- Is Tailscale installed?
- Is `tailscaled` running?
- What is the Tailscale node name and IP (100.x.x.x)?
- Is the node currently connected to the tailnet?
- How many peers are visible?
- Is MFA enforcement enabled for the tailnet (via ACL)?
- Is this node an exit node?
- Is this node advertising any subnets?
- What version of Tailscale is running?
- Are there any other nodes sharing this device (Tailscale Serve/Funnel)?

### Agent Report Format

```jsonc
{
  "tailscale": {
    "installed": true,
    "running": true,
    "version": "1.70.0",
    "node_name": "web-01",
    "node_ip": "100.82.13.45",
    "online": true,
    "peers_count": 12,
    "is_exit_node": false,
    "advertises_routes": false,
    "served_services": [
      { "port": 80, "proto": "tcp", "type": "serve", "allow_funnel": false },
      { "port": 443, "proto": "tcp", "type": "funnel", "allow_funnel": true }
    ],
    "dns_configured": true,
    "tailnet_name": "example.ts.net",
    "mfa_enabled": true,
    "acl_hash": "abc123def456"
  }
}
```

## Management Actions

### Node Management

| Action | Description | Priority |
|--------|-------------|----------|
| View status | Show tailnet status and peer list | P0 |
| View peers | List all visible peer nodes | P1 |
| Start Tailscale | `systemctl start tailscaled` | P2 |
| Stop Tailscale | `systemctl stop tailscaled` | P2 |
| Restart Tailscale | `systemctl restart tailscaled` | P1 |

### ACL & Access

| Action | Description | Priority |
|--------|-------------|----------|
| View ACL hash | Check last applied ACL version | P1 |
| Check MFA | Detect if MFA is enforced via ACL `check` | P1 |
| Check tags | Show node tags | P2 |

### Serve/Funnel Management

| Action | Description | Priority |
|--------|-------------|----------|
| List serve services | `tailscale serve status` | P1 |
| List funnel services | `tailscale funnel status` | P1 |
| Disable funnel | Remove public internet exposure | P1 |
| "Restrict to tailnet" | Ensure no funnel services are publicly exposed | P1 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/tailscale/status` | Full Tailscale status |
| GET | `/api/v1/nodes/:id/tailscale/peers` | List peer nodes |
| GET | `/api/v1/nodes/:id/tailscale/serve` | List Tailscale Serve/Funnel services |
| POST | `/api/v1/nodes/:id/tailscale/restart` | Restart tailscaled |
| GET | `/api/v1/settings/integrations/tailscale` | Tailscale API credentials |
| PUT | `/api/v1/settings/integrations/tailscale` | Update Tailscale API credentials |

### Tailscale API Credentials (optional, for ACL inspection)

```jsonc
{
  "api_key": "****",
  "tailnet": "example.ts.net"
}
```

The Tailscale API (v2) can be used for:
- Listing tailnet nodes
- Reading ACL (access control list)
- Checking device posture

## Frontend UI

### Settings — Tailscale Integration

- API key input (optional, for ACL features)
- Tailnet name
- Connection test button

### Node Detail — Tailscale Section

**Status card:**
- Installed / Not installed badge
- Running / Stopped indicator
- Node IP (100.x.x.x)
- Node name in tailnet
- Online status
- Tailnet name
- MFA enforcement indicator

**Serve/Funnel warning:**
- If Tailscale Funnel is active (public internet access), show warning
- "Restrict to tailnet" action button

**Peers list:**
- Peer name, IP, online status
- Route information

### UX Labels & Actions

- **"Restrict to tailnet"** — disable Tailscale Funnel (public exposure)
- **"Require MFA"** — show whether tailnet ACL enforces MFA
- **"Private network"** — tag for nodes accessible only via Tailscale

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| Tailscale down | `tailscaled` stopped unexpectedly | warning |
| Funnel active | Tailscale Funnel is exposing a service publicly | high |
| Tailscale not installed | Node without Tailscale exposing SSH | warning |
| Node offline | Peer node went offline | info |
| MFA not enforced | Tailscale ACL doesn't require MFA | warning |

### Policy Integration

- "If Tailscale Funnel is active without explicit approval, alert and disable"
- "If Tailscale is not installed on nodes with SSH exposed, suggest Tailscale"
- "Ensure tailscale MFA enforcement on all nodes"

## Agent Implementation Details

### Package Structure

```
agent/internal/tailscale/
├── tailscale.go       # Main integration
├── status.go          # tailscale status parsing
├── serve.go           # Tailscale Serve/Funnel detection
└── command.go         # Service lifecycle commands
```

### Key Interactions

```bash
# Status
tailscale status --json
tailscale ip
tailscale version
tailscale serve status
tailscale funnel status

# Management
tailscale serve off
tailscale funnel off
systemctl restart tailscaled
```

### Config Files

```
/var/lib/tailscale/tailscaled.state     # State database
/etc/default/tailscaled                 # Service config (optional)
```

### Tailscale API (v2) — for ACL checks

```
GET  /api/v2/tailnet/:tailnet/acl
POST /api/v2/tailnet/:tailnet/acl
GET  /api/v2/tailnet/:tailnet/devices
GET  /api/v2/tailnet/:tailnet/keys
```

## Implementation Order

1. Agent: tailscale installed/running detection
2. Agent: status parsing (node IP, online, peers)
3. Agent: Serve/Funnel detection
4. Agent: status report integration
5. Frontend: Tailscale section on node detail
6. Backend: read-only endpoints
7. Backend: Tailscale API integration (optional)
8. Frontend: Tailscale integration settings
9. Incident generation for funnel/tailscale-down
10. Policy integration
