# Agent Features

## Overview

Go-based Linux agent deployed on target machines. Static binary, minimal dependencies. Installed via one-liner:

```bash
curl -sSL https://install.gatewarden.com | sudo bash
```

## Core Requirements

- [x] Go-based Linux agent
- [x] Static binary (single-file deployment)
- [x] Ubuntu/Debian support (apt-based install flow)
- [x] amd64 + arm64 architectures
- [x] Outbound-only connection to API server
- [x] Systemd service for lifecycle management
- [x] Periodic heartbeat + status reporting

---

## 1. Secure Connectivity

Status reporting and integration checks for connectivity tools.

| Feature | Description | Priority |
|---------|-------------|----------|
| Cloudflare Tunnel | Detect running `cloudflared` tunnels, ingress, Access policy — [full spec](cloudflare-features.md) | P0 |
| Tailscale | Detect Tailscale status, node IP, peers, Serve/Funnel — [full spec](tailscale-features.md) | P0 |
| MFA enforcement | Check if services behind tunnels require authentication | P1 |
| Identity mapping | Report which identities can access which local services | P2 |

### Agent Checks

- Is `cloudflared` installed and running?
- What tunnels are configured?
- Is Tailscale connected?
- What Tailscale node name/IP?
- Are there services accessible without Cloudflare Access auth?

### UX Labels

- "Expose privately" — toggle a service from public to tunnel-only
- "Require MFA" — indicate whether MFA is enforced at tunnel level
- "Restrict to team" — show identity-aware access rules

---

## 2. Host Security

Attack detection, firewall management, and brute-force protection.

| Feature | Description | Priority |
|---------|-------------|----------|
| CrowdSec | Detect CrowdSec installation, alerts, bouncers, decisions — [full spec](crowdsec-features.md) | P0 |
| UFW/nftables | Inspect firewall rules, detect open ports, manage rules — [full spec](firewall-features.md) | P0 |
| SSH hardening | Check SSH config and apply hardening — [full spec](ssh-hardening-features.md) | P0 |
| Docker exposure | Scan containers for exposed ports, sockets, and misconfigs — [full spec](docker-security-features.md) | P0 |
| journald/auth log | Parse auth.log / journald for failed SSH, sudo, su attempts — [full spec](authlog-features.md) | P0 |
| Fail2Ban management | Detect, configure, and manage Fail2Ban jails, bans, and whitelist — [full spec](fail2ban-features.md) | P0 |
| Brute-force blocking | Report IPs currently blocked by CrowdSec/Fail2Ban/UFW | P1 |
| Geo-blocking | Detect geo-IP blocking rules (if any) | P1 |
| Suspicious IP | Flag IPs with known bad reputation (via CrowdSec API) | P2 |

### Agent Checks

- Is CrowdSec installed? Is it running? How many decisions active?
- Is Fail2Ban installed? What jails are active? ([full spec](fail2ban-features.md))
- What are the current UFW/nftables rules?
- Which ports are listening? Which are exposed to 0.0.0.0?
- Is SSH configured securely?
- Are any Docker containers publishing ports to 0.0.0.0?
- Is the Docker socket exposed?
- Recent failed login attempts (count, source IPs, usernames)

### UX Labels

- "SSH exposed publicly" — warning when SSH listens on 0.0.0.0:22
- "Docker socket exposed" — warning when /var/run/docker.sock is accessible
- "Grafana publicly reachable" — warning when Grafana is on 0.0.0.0
- "Multiple failed root login attempts blocked" — alert from auth log

---

## 3. Operational Visibility

Status reporting that feeds the dashboard and security graph.

| Feature | Description | Priority |
|---------|-------------|----------|
| Exposed services | Report all listening services and their exposure | P0 |
| Incidents | Flag security-relevant events for the dashboard | P0 |
| Attack attempts | Report brute-force attempts, port scans, suspicious traffic | P0 |
| Authentication activity | Report logins (success/fail), sudo usage | P0 |
| System health | CPU, memory, disk, uptime | P1 |
| Active tunnels | Report active tunnel connections | P0 |
| Blocked IPs | Report IPs currently blocked | P1 |
| Connected nodes | Report which other nodes this node can reach | P2 |

---

## Agent Report Format

The agent sends periodic reports to the API as JSON via POST. Suggested interval: 60 seconds for core checks, 300 seconds for slow checks.

```jsonc
{
  "node_id": "uuid",
  "timestamp": "2026-05-26T12:00:00Z",
  "hostname": "web-01",
  "os": "ubuntu-22.04",
  "kernel": "5.15.0-xxx",
  "uptime_seconds": 1234567,
  "docker": {
    "running_containers": [
      {
        "id": "abc123",
        "name": "grafana",
        "image": "grafana/grafana:latest",
        "ports": [
          { "internal": 3000, "external": 3000, "exposure": "0.0.0.0" }
        ],
        "socket_exposed": false
      }
    ],
    "total_containers": 5
  },
  "firewall": {
    "active": "ufw",
    "rules_count": 12,
    "status": "active",
    "rules": [
      { "port": 22, "proto": "tcp", "allowed": ["10.0.0.0/8"], "denied": ["0.0.0.0/0"] }
    ]
  },
  "ssh": {
    "port": 22,
    "password_auth": false,
    "root_login": false,
    "key_only": true,
    "publicly_exposed": false
  },
  "crowdsec": {
    "installed": true,
    "running": true,
    "active_decisions": 12,
    "bouncers": ["firewall-bouncer"],
    "alerts_last_hour": 45
  },
  "tailscale": {
    "installed": true,
    "running": true,
    "node_name": "web-01",
    "node_ip": "100.x.x.x",
    "online": true
  },
  "cloudflare_tunnel": {
    "installed": true,
    "running": true,
    "tunnel_id": "uuid",
    "tunnel_name": "web-tunnel",
    "ingress_rules": [
      { "service": "http://localhost:3000", "hostname": "grafana.example.com" }
    ]
  },
  "auth_log": {
    "failed_ssh_last_hour": 23,
    "failed_ssh_top_ips": [
      { "ip": "203.0.113.1", "attempts": 12 },
      { "ip": "198.51.100.1", "attempts": 8 }
    ],
    "failed_root_last_hour": 5,
    "sudo_usage_last_hour": 3
  },
  "ports": {
    "listening": [
      { "port": 22, "proto": "tcp", "process": "sshd", "exposed": false },
      { "port": 3000, "proto": "tcp", "process": "grafana", "exposed": true },
      { "port": 443, "proto": "tcp", "process": "nginx", "exposed": true }
    ]
  },
  "system": {
    "cpu_percent": 23.5,
    "memory_percent": 45.2,
    "disk_percent": 62.1
  }
}
```

## Implementation Order

1. Port scanning + listening services
2. Docker discovery (Docker SDK) — [full spec](docker-security-features.md)
3. Firewall inspection (parse `ufw status`, `nft list ruleset`) — [full spec](firewall-features.md)
4. SSH config parsing (`/etc/ssh/sshd_config`) — [full spec](ssh-hardening-features.md)
5. journald/auth log parsing (tail + regex) — [full spec](authlog-features.md)
6. Fail2Ban detection and status (fail2ban-client) — [full spec](fail2ban-features.md)
7. CrowdSec status (local API or `cscli`) — [full spec](crowdsec-features.md)
8. Tailscale status (`tailscale status`) — [full spec](tailscale-features.md)
9. Cloudflare Tunnel status (`cloudflared tunnel list`) — [full spec](cloudflare-features.md)
10. System health metrics
11. Reporter client (HTTP POST + heartbeat)

