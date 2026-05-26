# SSH Hardening

## Overview

SSH is the primary remote access vector for Linux infrastructure. Gatewarden checks SSH configuration against security best practices, surfaces risks, and can apply hardening fixes automatically or on demand.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| SSH installed | Check `sshd` binary or process | P0 |
| SSH running | Check `sshd` process or systemd service | P0 |
| Port | Parse `Port` directive in sshd_config | P0 |
| Password auth | Check `PasswordAuthentication` setting | P0 |
| Root login | Check `PermitRootLogin` setting | P0 |
| Key-only | Confirm `PubkeyAuthentication` is enabled and password auth is off | P0 |
| Public exposure | Detect if SSH listens on 0.0.0.0 | P0 |
| Protocol version | Check `Protocol` directive | P1 |
| AllowUsers/AllowGroups | Check if access is restricted to specific users | P1 |
| MaxAuthTries | Check max authentication attempts | P1 |
| ClientAliveInterval | Check connection timeout settings | P1 |
| ChallengeResponse | Check if challenge-response auth is enabled | P1 |
| PAM auth | Check `UsePAM` and its impact on auth methods | P1 |
| ListenAddress | Which addresses SSH binds to | P0 |
| HostKey algorithms | Check which host key types are in use | P2 |
| Ciphers/MACs | Check cipher and MAC algorithm strength | P2 |
| SSH config files | `/etc/ssh/sshd_config` and `/etc/ssh/sshd_config.d/*.conf` | P0 |

### Agent Checks

- Is SSH server installed and running?
- What port is SSH on?
- Is password authentication disabled?
- Is root login prohibited?
- Is public key authentication enabled?
- Is SSH listening on 0.0.0.0 (publicly accessible)?
- Are there any non-standard but insecure settings?
- Are there alternative SSH config overrides in `sshd_config.d/`?
- Is there an SSH firewall rule? Is it restrictive enough?
- Are there any SSH sessions currently active?

### Agent Report Format

```jsonc
{
  "ssh": {
    "installed": true,
    "running": true,
    "port": 22,
    "alternate_ports": [],
    "password_auth": false,
    "root_login": false,
    "pubkey_auth": true,
    "publicly_exposed": true,
    "listen_addresses": ["0.0.0.0"],
    "challenge_response": false,
    "pam_auth": false,
    "max_auth_tries": 3,
    "client_alive_interval": 300,
    "client_alive_count_max": 3,
    "allow_users": "deploy admin",
    "allow_groups": "sshers",
    "protocol": 2,
    "config_files": ["/etc/ssh/sshd_config"],
    "config_d_files": ["/etc/ssh/sshd_config.d/hardening.conf"],
    "active_sessions": 2,
    "firewall_coverage": true,
    "firewall_restrictive": false,
    "exposure_tier": "public"    // "public" | "tailscale" | "vpn" | "private"
  }
}
```

### Security Score Calculation

Each SSH setting contributes to an overall SSH hardening score (0-100):

| Setting | Weight | Best value | Partial credit |
|---------|--------|------------|----------------|
| Password auth disabled | 25 | false | — |
| Root login disabled | 20 | false | `prohibit-password` = 10 |
| Pubkey auth enabled | 10 | true | — |
| Non-default port | 5 | != 22 | — |
| MaxAuthTries ≤ 3 | 10 | ≤ 3 | ≤ 5 = 5 |
| ClientAliveInterval < 300 | 5 | < 300 | — |
| ChallengeResponse disabled | 10 | false | — |
| AllowUsers/AllowGroups set | 10 | set | — |
| Not publicly exposed | 5 | false | Tailscale-only = 3 |

## Management Actions

### One-Click Hardening

| Action | Description | Priority |
|--------|-------------|----------|
| Apply all hardening | Apply all recommended SSH hardening settings | P0 |
| Disable password auth | Set `PasswordAuthentication no` | P0 |
| Disable root login | Set `PermitRootLogin no` | P0 |
| Change SSH port | Change to non-standard port | P1 |
| Set MaxAuthTries | Reduce to 3 | P1 |
| Enable key-only mode | Ensure password auth off + pubkey on | P0 |
| Restrict users | Set `AllowUsers` to specific list | P2 |
| Apply config file | Write hardening config to `sshd_config.d/hardening.conf` | P0 |
| Restart SSH | `systemctl restart sshd` (with config test) | P0 |
| Test config | `sshd -t` before restart | P0 |

### Exposure Management

| Action | Description | Priority |
|--------|-------------|----------|
| Bind to localhost | Change `ListenAddress` to 127.0.0.1 | P0 |
| Bind to Tailscale IP | Change `ListenAddress` to Tailscale IP | P1 |
| Bind to private IP | Restrict to RFC1918 address | P1 |
| Add firewall rule | Add restrictive firewall rule for SSH port | P0 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/ssh/status` | Full SSH status and score |
| GET | `/api/v1/nodes/:id/ssh/config` | Current parsed sshd_config |
| POST | `/api/v1/nodes/:id/ssh/harden` | Apply all recommended hardening |
| POST | `/api/v1/nodes/:id/ssh/disable-password-auth` | Disable password authentication |
| POST | `/api/v1/nodes/:id/ssh/disable-root-login` | Disable root login |
| POST | `/api/v1/nodes/:id/ssh/set-port` | Change SSH port |
| POST | `/api/v1/nodes/:id/ssh/restrict-users` | Set AllowUsers |
| POST | `/api/v1/nodes/:id/ssh/bind-localhost` | Restrict to localhost |
| POST | `/api/v1/nodes/:id/ssh/bind-tailscale` | Restrict to Tailscale IP |
| POST | `/api/v1/nodes/:id/ssh/restart` | Test config and restart SSHD |
| GET | `/api/v1/nodes/:id/ssh/active-sessions` | List active SSH sessions |

## Frontend UI

### Node Detail — SSH Section

**Score card:**
- Hardening score (0-100) with color (red/yellow/green)
- Exposure tier badge (public / tailscale / vpn / private)
- Running status

**Security checks table:**

| Check | Status | Detail |
|-------|--------|--------|
| Password auth | ✅ Disabled / ❌ Enabled | Current value |
| Root login | ✅ Prohibited / ❌ Allowed | Current value |
| Public key auth | ✅ Enabled / ❌ Disabled | Current value |
| Port | Value | Recommendation if 22 |
| Public exposure | ✅ No / ❌ Yes | Listen addresses |
| MaxAuthTries | Value | Recommended: ≤ 3 |
| Users restricted | ✅ Yes / ❌ No | AllowUsers list |

**Exposure section:**
- Current listen addresses
- If public, show: "SSH is publicly accessible" warning
- Action buttons: "Restrict to localhost", "Bind to Tailscale", "Add firewall rule"

**Hardening panel:**
- "Apply all hardening" button (with confirmation dialog)
- Individual toggle buttons for each setting
- Preview of what will change before applying
- Config test result after changes
- "Restart SSH" button (disabled until config is clean)

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| SSH publicly exposed | SSH listening on 0.0.0.0 without restrictive firewall | critical |
| Password auth enabled | PasswordAuthentication = yes | critical |
| Root login enabled | PermitRootLogin = yes | critical |
| SSH config error | `sshd -t` fails after a change | high |
| No firewall for SSH | SSH reachable without firewall rule | high |
| Weak SSH config | Overall hardening score < 50 | warning |
| Non-standard but weak | Known weak ciphers/MACs in use | warning |

### Policy Integration

- "If SSH is publicly exposed, restrict to private network or Tailscale"
- "If password auth is enabled, disable it automatically"
- "Ensure root SSH login is disabled on all nodes"
- "Apply SSH hardening baseline to new nodes automatically"
- "If > 10 failed SSH attempts in 5 minutes, restrict SSH source IPs"

## Agent Implementation Details

### Package Structure

```
agent/internal/ssh/
├── ssh.go             # Main integration
├── config.go          # sshd_config parsing
├── session.go         # Active session detection
├── hardening.go       # Hardening presets and config generation
└── command.go         # Config write, port change, restart
```

### Key Interactions

```bash
# Status
sshd -T                    # Dump effective config (respects includes)
ss -tlnp | grep sshd       # Listening addresses
ps aux | grep sshd         # Running check
who -a                      # Active sessions

# Config test
sshd -t                    # Returns 0 on valid config

# Restart
systemctl restart sshd     # or ssh.service, depending on distro
```

### Config Files

```
/etc/ssh/sshd_config                   # Main server config
/etc/ssh/sshd_config.d/*.conf          # Drop-in config fragments
/etc/ssh/ssh_host_*_key                # Host keys
/etc/ssh/ssh_host_*_key.pub           # Host key public parts
~/.ssh/authorized_keys                # User authorized keys
```

The agent writes hardening changes as a new drop-in file:

```
/etc/ssh/sshd_config.d/99-gatewarden-hardening.conf
```

This approach:
- Does not modify the main sshd_config (preserves distro defaults)
- Easy to identify and revert (delete the drop-in file)
- Respects SSH's include ordering (last file wins for same directives)
- Safe to apply without touching unrelated settings

### Safe Config Application

1. Agent receives hardening request
2. Agent generates config snippet
3. Agent writes to `/etc/ssh/sshd_config.d/99-gatewarden-hardening.conf`
4. Agent runs `sshd -t` to validate
5. If valid → `systemctl reload sshd` (graceful reload, no session drop)
6. If invalid → revert file, return error with details
7. On failure: restore previous known-good config

### Error Handling

- SSH not installed → show "not installed", link to install instructions
- `sshd -t` failure → capture stderr, return specific config errors
- Config file permission denied → agent must run as root
- Reload failure → report, do not block (SSH continues with old config)
- Distro differences → detect Ubuntu vs Debian (ssh vs sshd service name)
- Concurrent edits → use atomic write (write to temp, rename)

## Implementation Order

1. Agent: installed/running detection
2. Agent: sshd_config primary file parsing
3. Agent: effective config parsing (`sshd -T`)
4. Agent: public exposure detection
5. Agent: active session detection
6. Agent: hardening score calculation
7. Agent: status report integration
8. Backend: read-only endpoints (status, config)
9. Frontend: SSH section on node detail
10. Frontend: security checks table with score
11. Agent: hardening config generation
12. Agent: safe write + config test + reload
13. Backend: hardening endpoints
14. Frontend: one-click hardening panel
15. Incident generation
16. Policy integration
