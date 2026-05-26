# Fail2Ban Integration

## Overview

Fail2Ban is a critical part of host-level brute-force protection. Gatewarden must not only detect and report Fail2Ban status but provide active management: inspecting jails, viewing ban lists, managing whitelists, and configuring jail parameters — all from the dashboard.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Installed | Check `fail2ban-client --version` or dpkg status | P0 |
| Running | Check systemd service status (`fail2ban.service`) | P0 |
| Enabled | Check if service is set to start on boot | P0 |
| Jails list | `fail2ban-client status` — parse jail names | P0 |
| Per-jail status | `fail2ban-client status <jail>` — currently banned, total banned, failed count | P0 |
| Configuration files | Parse `/etc/fail2ban/jail.local`, `/etc/fail2ban/jail.d/*.conf`, `/etc/fail2ban/jail.conf` | P1 |
| Database | Read `/var/lib/fail2ban/fail2ban.sqlite3` for historical ban data | P2 |
| Whitelist | Parse `ignoreip` from jail configs | P1 |
| Ban action | Detect which action is configured per jail (iptables, nftables, ufw) | P1 |

### Agent Checks

- Is `fail2ban` installed? (apt package or binary)
- Is the `fail2ban.service` systemd unit active and enabled?
- What jails are configured? Which are currently active?
- Per jail: currently banned IPs, total ban count, failed login count
- What is the current `bantime`, `findtime`, `maxretry` for each jail?
- Which IPs are whitelisted (`ignoreip`)?
- What action backend is in use (iptables, nftables, ufw)?
- Are there any stale/invalid jails referencing non-existent log files?

### Agent Report Format

Add a `fail2ban` section to the agent's periodic status report:

```jsonc
{
  "fail2ban": {
    "installed": true,
    "running": true,
    "enabled": true,
    "version": "1.0.2",
    "jails": [
      {
        "name": "sshd",
        "active": true,
        "currently_banned": 3,
        "total_banned": 47,
        "failed_count": 234,
        "bantime": 3600,
        "findtime": 600,
        "maxretry": 5,
        "action": "nftables-multiport",
        "log_path": "/var/log/auth.log",
        "ignoreip": ["127.0.0.1", "10.0.0.0/8"],
        "banned_ips": [
          { "ip": "203.0.113.1", "banned_at": "2026-05-25T14:30:00Z", "expires_at": "2026-05-25T15:30:00Z" },
          { "ip": "198.51.100.1", "banned_at": "2026-05-25T14:22:00Z", "expires_at": "2026-05-25T15:22:00Z" }
        ]
      },
      {
        "name": "nginx-http-auth",
        "active": false,
        "currently_banned": 0,
        "total_banned": 12,
        "failed_count": 56,
        "bantime": 86400,
        "findtime": 600,
        "maxretry": 10,
        "action": "ufw",
        "log_path": "/var/log/nginx/error.log",
        "ignoreip": ["127.0.0.1"],
        "banned_ips": []
      }
    ]
  }
}
```

## Management Actions

### Jail Management

| Action | Description | Priority |
|--------|-------------|----------|
| List jails | Show all configured jails with status | P0 |
| View jail detail | Per-jail stats, config, banned IPs | P0 |
| Enable jail | `fail2ban-client start <jail>` | P1 |
| Disable jail | `fail2ban-client stop <jail>` | P1 |
| Set bantime | Override the bantime for a jail | P2 |
| Set findtime | Override the findtime window | P2 |
| Set maxretry | Override the failure threshold | P2 |

### Ban Management

| Action | Description | Priority |
|--------|-------------|----------|
| View banned IPs | List currently banned IPs per jail | P0 |
| Unban IP | `fail2ban-client set <jail> unbanip <ip>` | P0 |
| Unban IP all jails | `fail2ban-client unban <ip>` | P1 |
| Ban IP manually | `fail2ban-client set <jail> banip <ip>` | P2 |
| View ban history | Historical ban data from Fail2Ban database | P1 |

### Whitelist Management

| Action | Description | Priority |
|--------|-------------|----------|
| View whitelist | Show `ignoreip` per jail | P1 |
| Add to whitelist | Add IP/subnet to `ignoreip` | P1 |
| Remove from whitelist | Remove IP/subnet from `ignoreip` | P2 |

### Configuration Management

| Action | Description | Priority |
|--------|-------------|----------|
| View config | Show effective jail configuration | P1 |
| Update log path | Change which log file a jail monitors | P2 |
| Update action | Switch between iptables/nftables/ufw backend | P2 |

## Backend API

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/fail2ban/status` | Full Fail2Ban status for a node |
| GET | `/api/v1/nodes/:id/fail2ban/jails` | List jails with summary |
| GET | `/api/v1/nodes/:id/fail2ban/jails/:name` | Jail detail (config + banned IPs) |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/unban` | Unban an IP from a jail |
| POST | `/api/v1/nodes/:id/fail2ban/unban` | Unban an IP from all jails |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/ban` | Manually ban an IP |
| PUT | `/api/v1/nodes/:id/fail2ban/jails/:name/config` | Update jail config (bantime, findtime, maxretry) |
| GET | `/api/v1/nodes/:id/fail2ban/whitelist` | View whitelist per jail |
| POST | `/api/v1/nodes/:id/fail2ban/whitelist/add` | Add IP to whitelist |
| POST | `/api/v1/nodes/:id/fail2ban/whitelist/remove` | Remove IP from whitelist |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/start` | Enable/start a jail |
| POST | `/api/v1/nodes/:id/fail2ban/jails/:name/stop` | Disable/stop a jail |

### Agent Commands

Management actions require the API to send commands to the agent. Since agents connect outbound-only, use one of:

- **SSE command channel** — API pushes a command via SSE, agent picks it up on next event
- **Polling** — agent checks for pending commands on each heartbeat response
- **WebSocket** — persistent bidirectional channel (post-MVP)

Recommended for MVP: agent checks for pending commands after each heartbeat response. The API returns a list of queued commands, the agent executes them and reports results.

## Frontend UI

### Node Detail — Fail2Ban Tab

A dedicated Fail2Ban section on the node detail page:

**Overview card:**
- Installation status (installed / not installed)
- Service status (running / stopped / disabled)
- Version
- Total jails count
- Total currently banned IPs

**Jails table:**

| Column | Description |
|--------|-------------|
| Jail name | Link to jail detail |
| Status | Active / Inactive badge |
| Currently banned | Count |
| Total banned | Count |
| Failed count | Count |
| Action | Action badge text |
| Bantime | Human-readable (e.g. "1h", "24h") |
| Controls | Start/Stop buttons |

**Jail detail (expandable row or modal):**
- Configuration parameters: bantime, findtime, maxretry, log path, action
- Whitelist (ignoreip) — view and edit
- Banned IPs list — table with IP, banned at, expires at, Unban button

### Incidents Integration

Fail2Ban events should create incidents in the main feed:

| Incident | Trigger | Severity |
|----------|---------|----------|
| Jail stopped | A previously active jail is now stopped | warning |
| High ban rate | > 50 bans in 5 minutes on any single jail | critical |
| Fail2Ban not running | Service was running but now stopped | critical |
| Fail2Ban not installed | Node without Fail2Ban where SSH is publicly exposed | warning |
| Config error | Jail references missing log file | warning |

### Policy Integration

Policies should be able to trigger Fail2Ban actions:

- "If > 10 failed SSH attempts in 5 minutes, ensure Fail2Ban sshd jail is active"
- "If public SSH exposure detected, install and configure Fail2Ban if missing"
- "Whitelist trusted IP ranges across all jails"
- "On attack detection, temporarily reduce maxretry to 3"

## Implementation Order

1. Agent: installed/running/enabled detection
2. Agent: jail list + per-jail status (parse `fail2ban-client status`)
3. Agent: jail config parsing (bantime, findtime, maxretry, ignoreip)
4. Agent: banned IPs list per jail
5. Agent: report format integration into existing reporter
6. Backend: Fail2Ban schema extensions (jail configs, ban events, whitelist)
7. Backend: read-only endpoints (status, jails, jail detail)
8. Backend: unban command endpoint + agent command queue
9. Agent: command execution (unban, unban all, ban, start/stop jail)
10. Frontend: Fail2Ban tab on node detail page
11. Frontend: banned IPs table + unban flow
12. Frontend: jail detail view with config display
13. Backend: jail config update endpoint
14. Agent: jail config modification (set bantime/findtime/maxretry)
15. Backend: whitelist management endpoints
16. Agent: whitelist modification
17. Frontend: whitelist management UI
18. Incident generation from Fail2Ban events
19. Policy integration for Fail2Ban actions

## Agent Implementation Details

### Package Structure

```
agent/internal/fail2ban/
├── fail2ban.go          # Main integration — detection, status, commands
├── client.go            # fail2ban-client wrapper
├── config.go            # Configuration file parsing (jail.local, jail.d/*.conf)
├── jail.go              # Jail-specific types and parsing
├── ban.go               # Ban list parsing
├── whitelist.go         # Whitelist management
└── command.go           # Command execution (unban, ban, start, stop, config)
```

### Key Interactions

**fail2ban-client commands used:**

```bash
# Status
fail2ban-client status                          # -> jail names list
fail2ban-client status <jail>                   # -> per-jail stats

# Management
fail2ban-client set <jail> unbanip <ip>         # Unban single IP
fail2ban-client unban <ip>                      # Unban from all jails
fail2ban-client set <jail> banip <ip>           # Manually ban
fail2ban-client set <jail> bantime <seconds>    # Set ban time
fail2ban-client set <jail> findtime <seconds>   # Set find time
fail2ban-client set <jail> maxretry <count>     # Set max retries
fail2ban-client start <jail>                    # Start/enable jail
fail2ban-client stop <jail>                     # Stop/disable jail

# Config reload
fail2ban-client reload                          # Reload all config
fail2ban-client reload <jail>                   # Reload specific jail
```

**Config files:**

```
/etc/fail2ban/jail.conf          # Default configuration (read-only)
/etc/fail2ban/jail.local         # Local overrides (writable by agent)
/etc/fail2ban/jail.d/*.conf      # Drop-in config snippets
```

The agent writes config changes to `jail.local` or creates drop-in files in `jail.d/`, then triggers a reload.

### Error Handling

- Commands to a node where Fail2Ban is not installed → return error, surface in UI
- Commands to a stopped Fail2Ban service → return error, suggest starting the service first
- Invalid jail name → return error
- Invalid IP format for ban/unban → validate on API side before forwarding
- Permission denied → agent must run with appropriate capabilities (usually runs as root via systemd)
- Config reload failures → capture stderr, surface error in UI with details
