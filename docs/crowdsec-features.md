# CrowdSec Integration

## Overview

CrowdSec provides distributed brute-force protection and threat intelligence. Gatewarden integrates with CrowdSec to surface alerts, active decisions, bouncer status, and suspicious IPs — with the ability to manage configurations and trigger remediation from the dashboard.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Installed | Check `cscli` binary or dpkg status | P0 |
| Running | Check systemd service status (`crowdsec.service`) | P0 |
| Version | `cscli version` | P0 |
| LAPI status | `cscli lapi status` | P0 |
| Active decisions | `cscli decisions list` — count and list | P0 |
| Alerts | `cscli alerts list` — paginated | P0 |
| Bouncers | `cscli bouncers list` — name, IP, type, last pull | P0 |
| Metrics | `cscli metrics` — parsed summary | P1 |
| Hub status | `cscli hub list` — installed collections/parsers/scenarios | P1 |
| Remediation components | `cscli remediation-components list` | P2 |

### Agent Checks

- Is CrowdSec installed? (binary check or dpkg)
- Is `crowdsec.service` systemd unit active and enabled?
- Is the Local API (LAPI) responding?
- How many active decisions? What are the top banned IPs?
- How many alerts in the last hour? Last 24 hours?
- Which bouncers are connected? When did they last pull?
- What collections are installed (default linux collection)?
- Are there any expired/stale decisions clogging the list?

### Agent Report Format

```jsonc
{
  "crowdsec": {
    "installed": true,
    "running": true,
    "enabled": true,
    "version": "1.6.3",
    "lapi_healthy": true,
    "active_decisions_count": 24,
    "alerts_last_hour": 12,
    "alerts_last_24h": 89,
    "top_banned_ips": [
      { "ip": "203.0.113.1", "decisions": 8, "reasons": ["ssh_bf", "http_scan"] },
      { "ip": "198.51.100.1", "decisions": 5, "reasons": ["http_scan"] }
    ],
    "bouncers": [
      { "name": "firewall-bouncer", "type": "iptables", "ip": "127.0.0.1", "last_pull": "2026-05-26T11:55:00Z", "status": "ok" },
      { "name": "blocklist-mirror", "type": "blocklist", "ip": "127.0.0.1", "last_pull": "2026-05-26T11:50:00Z", "status": "ok" }
    ],
    "collections": [
      "crowdsecurity/linux",
      "crowdsecurity/nginx",
      "crowdsecurity/sshd"
    ],
    "metrics": {
      "parsed_events_total": 45231,
      "unparsed_events_total": 234,
      "total_alerts": 1234,
      "ip_enrichments": 890
    }
  }
}
```

## Management Actions

### Decision Management

| Action | Description | Priority |
|--------|-------------|----------|
| List decisions | View active ban decisions with IP, reason, expiration | P0 |
| View alert detail | Full alert info: timestamp, source IP, scenario, remediation | P1 |
| Add decision | Manually ban an IP (`cscli decisions add`) | P1 |
| Delete decision | Unban an IP by ID (`cscli decisions delete`) | P0 |
| Flush decisions | Delete all decisions | P2 |

### Bouncer Management

| Action | Description | Priority |
|--------|-------------|----------|
| List bouncers | Show connected bouncers and status | P0 |
| Add bouncer | Generate new bouncer API key | P2 |
| Delete bouncer | Remove a bouncer | P2 |

### Configuration

| Action | Description | Priority |
|--------|-------------|----------|
| View scenarios | List installed scenarios with status | P1 |
| View collections | Installed collections | P1 |
| Install collection | `cscli collections install` | P2 |
| Remove collection | `cscli collections remove` | P2 |
| Reload CrowdSec | `systemctl reload crowdsec` | P1 |
| Restart CrowdSec | `systemctl restart crowdsec` | P1 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/crowdsec/status` | Full CrowdSec status |
| GET | `/api/v1/nodes/:id/crowdsec/alerts` | List alerts (paginated) |
| GET | `/api/v1/nodes/:id/crowdsec/decisions` | List active decisions |
| POST | `/api/v1/nodes/:id/crowdsec/decisions` | Add a ban decision |
| DELETE | `/api/v1/nodes/:id/crowdsec/decisions/:id` | Remove a decision |
| GET | `/api/v1/nodes/:id/crowdsec/bouncers` | List bouncers |
| GET | `/api/v1/nodes/:id/crowdsec/metrics` | Parsed metrics |
| POST | `/api/v1/nodes/:id/crowdsec/reload` | Reload CrowdSec |
| POST | `/api/v1/nodes/:id/crowdsec/restart` | Restart CrowdSec |

## Frontend UI

### Node Detail — CrowdSec Section

**Status card:**
- Installed / Not installed badge
- Running / Stopped status
- LAPI health indicator
- Version
- Active decisions count
- Alerts last hour

**Decisions table:**
- IP, reason/scenario, expiration time, age
- Bulk delete / individual unban

**Alerts table:**
- Timestamp, source IP, scenario, action taken
- Click to expand full alert detail

**Bouncers list:**
- Name, type, IP, last pull, status indicator

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| CrowdSec down | Service stopped unexpectedly | critical |
| Decision spike | > 100 new decisions in 5 minutes | critical |
| Bouncer offline | Bouncer hasn't pulled in > 5 minutes | warning |
| Attack detected | Alert triggered (any scenario) | high |

### Policy Integration

- "If CrowdSec is not installed on nodes with public SSH, alert and suggest install"
- "If bouncer is stale, restart CrowdSec service"
- "On SSH brute force alert from CrowdSec, add temporary firewall rule"

## Agent Implementation Details

### Package Structure

```
agent/internal/crowdsec/
├── crowdsec.go       # Main integration — detection, status
├── client.go         # cscli command wrapper
├── decisions.go      # Decision list/add/delete parsing
├── alerts.go         # Alert list/detail parsing
├── bouncers.go       # Bouncer listing
├── metrics.go        # Metrics parsing
└── command.go        # Command execution (reload, restart, install)
```

### Key Interactions

```bash
# Status
cscli version
cscli lapi status
cscli decisions list -o json
cscli alerts list -o json
cscli bouncers list -o json
cscli metrics -o json
cscli hub list -o json

# Management
cscli decisions add --ip <ip> --duration <duration> --reason <reason>
cscli decisions delete --id <id>
cscli collections install <collection>
cscli collections remove <collection>
```

### Config Files

```
/etc/crowdsec/config.yaml            # Main config
/etc/crowdsec/acquis.yaml            # Log acquisition config
/etc/crowdsec/parsers/               # Parser configurations
/etc/crowdsec/scenarios/             # Scenario configurations
/var/lib/crowdsec/data/              # Database and state
```

### Error Handling

- CrowdSec not installed → return clear status, suggest install via apt
- LAPI unreachable → report degraded status, queue commands
- Permission denied → agent runs as root via systemd
- Invalid IP for decision → validate on API side
- cscli command timeout → set 10s timeout, handle partial output

## Implementation Order

1. Agent: installed/running detection
2. Agent: decisions list parsing
3. Agent: alerts list parsing
4. Agent: bouncer list parsing
5. Agent: metrics parsing
6. Agent: status report integration
7. Backend: read-only endpoints
8. Backend: decision management (add/delete)
9. Agent: command execution (add/delete decision, reload)
10. Frontend: CrowdSec section on node detail
11. Incident generation from CrowdSec events
12. Policy integration
