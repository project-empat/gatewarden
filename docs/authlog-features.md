# Authentication Log Monitoring (journald / auth.log)

## Overview

System authentication logs (journald / `/var/log/auth.log`) are the primary source for detecting brute-force attacks, unauthorized access attempts, and privilege escalation. Gatewarden parses these logs to surface actionable security events.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Log source detection | Detect available sources: `journald` vs `auth.log` | P0 |
| journald available | Check `journalctl` binary and access | P0 |
| auth.log available | Check `/var/log/auth.log` exists and readable | P0 |
| Failed SSH attempts | Count failed password/publickey auth attempts | P0 |
| Failed root login | Count root-specific failed attempts | P0 |
| Successful logins | Count successful SSH logins | P0 |
| Top offending IPs | Aggregate failures by source IP | P0 |
| sudo usage | Detect sudo/su authentication attempts | P0 |
| Invalid users | Count attempts for non-existent users | P1 |
| User enumeration | Detect patterns suggesting username enumeration | P1 |
| Connection timing | Time-windowed analysis (last 5m, 1h, 24h) | P0 |
| Unusual times | Detect logins outside normal hours | P2 |
| New user accounts | Detect recently created user accounts | P2 |
| SSH key events | Detect key-based auth events | P1 |
| Port connection tracking | Detect connection establishes outside expected patterns | P1 |
| Service-specific failures | Detect auth failures for specific services (vsftpd, postfix, etc.) | P1 |

### Agent Checks

- What auth log sources are available (journald, auth.log, both)?
- How many failed SSH attempts in the last 5 minutes / 1 hour / 24 hours?
- What are the top 10 source IPs for failed attempts?
- How many failed root login attempts?
- How many successful logins? From which IPs?
- How many sudo authentication failures?
- Are there any invalid/unknown user login attempts?
- What usernames are being targeted most frequently?
- Is the log source reliable (file not rotated out, journald not corrupted)?

### Agent Report Format

```jsonc
{
  "auth_log": {
    "source": "journald",           // "journald" | "auth.log" | "both"
    "journald_available": true,
    "auth_log_available": false,
    "time_windows": {
      "last_5_minutes": {
        "failed_ssh": 12,
        "failed_root": 3,
        "successful_logins": 1,
        "sudo_failures": 0,
        "invalid_users": 5,
        "unique_source_ips": 4
      },
      "last_hour": {
        "failed_ssh": 89,
        "failed_root": 23,
        "successful_logins": 3,
        "sudo_failures": 1,
        "invalid_users": 34,
        "unique_source_ips": 18
      },
      "last_24h": {
        "failed_ssh": 456,
        "failed_root": 89,
        "successful_logins": 15,
        "sudo_failures": 4,
        "invalid_users": 123,
        "unique_source_ips": 67
      }
    },
    "top_source_ips": [
      { "ip": "203.0.113.1", "attempts": 45, "last_seen": "2026-05-26T11:59:00Z" },
      { "ip": "198.51.100.1", "attempts": 23, "last_seen": "2026-05-26T11:55:00Z" },
      { "ip": "192.0.2.1", "attempts": 12, "last_seen": "2026-05-26T11:50:00Z" }
    ],
    "targeted_usernames": [
      { "username": "root", "attempts": 89 },
      { "username": "admin", "attempts": 45 },
      { "username": "ubuntu", "attempts": 12 },
      { "username": "deploy", "attempts": 8 }
    ],
    "recent_events": [
      {
        "timestamp": "2026-05-26T11:59:00Z",
        "type": "failed_password",
        "username": "root",
        "source_ip": "203.0.113.1",
        "service": "sshd",
        "message": "Failed password for root from 203.0.113.1 port 45231 ssh2"
      }
    ],
    "sudo_events_last_hour": [
      {
        "timestamp": "2026-05-26T11:45:00Z",
        "username": "deploy",
        "command": "/usr/bin/apt update",
        "success": true
      }
    ],
    "new_accounts_last_24h": [
      { "username": "ci-user", "created_at": "2026-05-26T08:00:00Z", "created_by": "root" }
    ],
    "log_health": {
      "last_entry_seen": "2026-05-26T11:59:30Z",
      "entries_last_5m": 45,
      "rotation_detected": false,
      "backpressure": false
    }
  }
}
```

## Auth Event Types

### SSH Events

| Event | Log Pattern | Severity |
|-------|-------------|----------|
| Failed password | `Failed password for` | warning |
| Failed publickey | `Failed publickey for` | info |
| Invalid user | `Invalid user` from | high |
| Root login attempt | `Failed password for root` | high |
| Successful login | `Accepted password` / `Accepted publickey` | info |
| Connection closed | `Connection closed by authenticating user` | warning |
| Break-in attempt | `PAM service(sshd) ignoring max retries` | critical |
| Reverse mapping | `reverse mapping checking getaddrinfo` | info |

### Sudo Events

| Event | Log Pattern | Severity |
|-------|-------------|----------|
| Successful sudo | `sudo: .* COMMAND=` | info |
| Failed sudo | `sudo: .* authentication failure` | warning |
| Root sudo | `sudo: .* COMMAND=/usr/bin/su -` | info |
| Sudo with NOPASSWD | `sudo: .* NOPASSWD: COMMAND=` | info |

### User Management Events

| Event | Log Pattern | Severity |
|-------|-------------|----------|
| User created | `useradd` or `new user` | info |
| User deleted | `userdel` or `delete user` | info |
| Password change | `passwd: password changed` | info |
| Group modified | `groupadd` / `groupmod` | info |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/authlog/status` | Full auth log status and summary |
| GET | `/api/v1/nodes/:id/authlog/summary` | Aggregated metrics (5m, 1h, 24h windows) |
| GET | `/api/v1/nodes/:id/authlog/failed-attempts` | Failed attempts with top IPs |
| GET | `/api/v1/nodes/:id/authlog/recent-events` | Recent auth events (paginated) |
| GET | `/api/v1/nodes/:id/authlog/top-ips` | Top offending source IPs |
| GET | `/api/v1/nodes/:id/authlog/targeted-users` | Most targeted usernames |
| GET | `/api/v1/nodes/:id/authlog/health` | Log source health metrics |

## Frontend UI

### Node Detail — Authentication Log Section

**Summary card:**
- Log source (journald / auth.log / both)
- Data freshness indicator
- Events per time window

**Time-windowed metrics tabs:**
- 5 minutes | 1 hour | 24 hours
- Each tab shows failed SSH, successful logins, sudo events, invalid users

**Threat indicators:**
- Color-coded rate: green (< 10/hr), yellow (10-50/hr), red (> 50/hr)
- Trend arrow (increasing/decreasing compared to previous window)

**Top source IPs table:**

| Column | Description |
|--------|-------------|
| IP address | Source IP |
| Attempts | Count |
| Last seen | Timestamp |
| Status | Currently blocked? (cross-referenced with Fail2Ban/CrowdSec) |
| Action | Block this IP button (adds to firewall or Fail2Ban) |

**Targeted usernames table:**
- Username, attempt count
- Highlight if "root" or other sensitive accounts are targeted

**Recent events feed:**
- Scrollable log of recent auth events
- Timestamp, type badge (color-coded), username, source IP, message
- Auto-refresh with SSE

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| Brute force in progress | > 50 failed SSH attempts in 5 minutes | critical |
| Root targeted | > 10 failed root login attempts in 5 minutes | critical |
| User enumeration | > 5 invalid user attempts in 5 minutes | high |
| Sudo failure spike | > 5 sudo auth failures in 5 minutes | high |
| Suspicious login hour | Successful login from new IP at unusual time | warning |
| New account created | New user account detected | info |
| Log source stale | No auth log entries in > 5 minutes | warning |

### Policy Integration

- "If failed SSH > threshold in time window, add source IP to Fail2Ban"
- "If root login attempts spike, block all root SSH access"
- "If invalid user attempts spike, enable additional rate limiting"
- "Alert on sudo failures for non-admin users"

## Agent Implementation Details

### Package Structure

```
agent/internal/journald/
├── journald.go        # Main integration
├── reader.go          # Log source detection and reading
├── parser.go          # Log line parsing and classification
├── metrics.go         # Time-windowed metric aggregation
├── sudo.go            # Sudo event detection
├── users.go           # User management event detection
└── health.go          # Log health monitoring
```

### Key Interactions

**journald:**
```bash
# Recent SSH auth failures
journalctl -u sshd --since "5 minutes ago" --no-pager -o short-iso

# All auth events
journalctl _TRANSPORT=audit --since "1 hour ago" --no-pager

# Failed passwords
journalctl -u sshd --since "24 hours ago" | grep "Failed password"

# JSON output (preferred for parsing)
journalctl -u sshd --since "1 hour ago" -o json
```

**auth.log:**
```bash
# Read from file
tail -n 1000 /var/log/auth.log

# Failed SSH
grep "Failed password" /var/log/auth.log

# Seek and tail (for persistent monitoring)
tail -F /var/log/auth.log
```

### Log Parsing Strategy

The agent should:

1. **Detect source**: Try journald first (faster, structured), fall back to auth.log
2. **Read incrementally**: Track last-read position/cursor
3. **Parse with regex**: Classify each line by event type
4. **Aggregate in windows**: Maintain rolling counters (5m, 1h, 24h)
5. **Report diff**: Only send new events since last report

### Time-Windowed Aggregation

```go
type TimeWindowCounters struct {
    FailedSSH       int
    FailedRoot      int
    SuccessfulLogin int
    SudoFailure     int
    InvalidUser     int
    UniqueIPs       map[string]int
    TargetedUsers   map[string]int
    TopIPs          []IPCount
}
```

The agent maintains three rolling windows (5m, 1h, 24h) using a ring buffer of timestamped events. Old events are evicted as time passes.

### Regex Patterns

```go
var patterns = map[string]*regexp.Regexp{
    "failed_password":  regexp.MustCompile(`Failed password for (invalid user )?(?P<user>\S+) from (?P<ip>\S+)`),
    "accepted_login":   regexp.MustCompile(`Accepted (password|publickey) for (?P<user>\S+) from (?P<ip>\S+)`),
    "invalid_user":     regexp.MustCompile(`Invalid user (?P<user>\S+) from (?P<ip>\S+)`),
    "sudo_failure":     regexp.MustCompile(`sudo:.*authentication failure`),
    "sudo_success":     regexp.MustCompile(`sudo:.*COMMAND=.*`),
    "connection_closed": regexp.MustCompile(`Connection closed by authenticating user`),
    "new_user":         regexp.MustCompile(`useradd.*new user.*name=(?P<user>\S+)`),
}
```

### Errored Patterns for Common Cases

- `sshd[1234]: Failed password for root from 203.0.113.1 port 45231 ssh2`
- `sshd[1234]: Accepted publickey for deploy from 10.0.0.1 port 52341 ssh2: RSA SHA256:...`
- `sshd[1234]: Invalid user admin from 198.51.100.1 port 34567`
- `sudo: deploy : TTY=pts/0 ; PWD=/home/deploy ; USER=root ; COMMAND=/usr/bin/apt update`
- `sudo: pam_unix(sudo:auth): authentication failure; logname= uid=1000 euid=0 tty=/dev/pts/0`

### Performance Considerations

- journald querying with `--since` and `-o json` is preferred (lower overhead)
- auth.log tailing uses `tail -F` with seek tracking
- Reports should be incremental (only new events since last report)
- Full re-scan on agent restart or log rotation
- Ring buffer size: 100K events per window (configurable)

### Error Handling

- No log source available → report degraded, surface in UI
- journald permissions → agent must be in `systemd-journal` group or root
- Log rotation → detect via file inode change, re-seek
- journald corruption → fall back to auth.log if available
- High event rate → sample or aggregate during parsing (don't store raw)
- Time-skewed logs → use monotonic clock for windows, log timestamps for events

## Implementation Order

1. Agent: log source detection (journald vs auth.log)
2. Agent: basic failed SSH count (last 5m/1h/24h)
3. Agent: top source IP aggregation
4. Agent: targeted username tracking
5. Agent: sudo event detection
6. Agent: successful login tracking
7. Agent: log health monitoring (stale detection)
8. Agent: status report integration
9. Backend: read-only endpoints
10. Frontend: auth log section on node detail
11. Frontend: time-windowed metric display
12. Frontend: top IPs feed
13. Incident generation from auth events
14. Policy integration (auto-block IPs)
