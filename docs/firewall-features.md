# Firewall Management (UFW / nftables)

## Overview

Gatewarden provides visibility into firewall rules across all managed nodes and enables rule management directly from the dashboard. Supports both UFW (Ubuntu default) and nftables (Debian default, modern Linux firewall).

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| UFW installed | Check `ufw` binary or dpkg status | P0 |
| UFW active | `ufw status` verbose | P0 |
| nftables installed | Check `nft` binary | P0 |
| nftables active | `nft list ruleset` | P0 |
| Which is active | Detect active firewall backend | P0 |
| Rules count | Total rules across all chains | P0 |
| Default policy | Default inbound/outbound policy | P0 |
| Rules detail | Per-rule: port, proto, action, source/dest, interface | P0 |
| Application profiles | UFW app profiles (`ufw app list`) | P1 |
| Rate limiting | Rules with limit/connlimit modifiers | P1 |
| Logging status | UFW logging level, nftables log rules | P1 |
| IPv6 rules | Separate IPv6 rule set status | P1 |

### Agent Checks

- Is UFW installed? Is it active?
- Is nftables installed? Are there any rules loaded?
- Which firewall is the active/default on this system?
- What are the default policies (allow/deny for in/out)?
- What rules are currently configured?
- Which ports are explicitly allowed? Which are denied?
- Are there any allow rules for 0.0.0.0/0 on sensitive ports (22, 443, etc.)?
- Are there rate-limiting rules? (connection limits)
- Is firewall logging active?
- Are IPv6 rules consistent with IPv4 rules?
- Is there a difference between the firewall rules and actual listening services?

### Agent Report Format

```jsonc
{
  "firewall": {
    "active_backend": "ufw",
    "ufw": {
      "installed": true,
      "active": true,
      "logging": "low",
      "default_incoming": "deny",
      "default_outgoing": "allow",
      "rules": [
        { "action": "allow", "port": 22, "proto": "tcp", "from": "any", "to": "any", "interface": "" },
        { "action": "allow", "port": 443, "proto": "tcp", "from": "any", "to": "any", "interface": "" },
        { "action": "deny", "port": 3306, "proto": "tcp", "from": "any", "to": "any", "interface": "" },
        { "action": "allow", "port": 22, "proto": "tcp", "from": "10.0.0.0/8", "to": "any", "interface": "eth0" }
      ],
      "enabled_app_profiles": ["OpenSSH", "Nginx Full"],
      "status_raw": "Status: active\nLogging: low\nDefault: deny (incoming), allow (outgoing)\n..."
    },
    "nftables": {
      "installed": true,
      "active": false,
      "tables_count": 0,
      "rules_count": 0,
      "tables": []
    }
  }
}
```

## Management Actions

### Rule Management

| Action | Description | Priority |
|--------|-------------|----------|
| List rules | View all rules across chains | P0 |
| Add allow rule | Allow a port/proto from a source | P0 |
| Add deny rule | Deny a port/proto from a source | P0 |
| Delete rule | Remove a rule by ID or parameters | P0 |
| Reorder rules | Move a rule's position in the chain | P1 |
| Add rate limit | Add connection limit rule | P2 |
| Add interface rule | Restrict rule to specific interface | P1 |
| Enable logging | Turn on firewall logging for a rule | P1 |

### Policy Management

| Action | Description | Priority |
|--------|-------------|----------|
| View defaults | Show default incoming/outgoing policies | P0 |
| Set default deny | Set default incoming to deny | P0 |
| Set default allow | Set default incoming to allow (warning) | P1 |
| Enable firewall | Turn on UFW / apply nftables rules | P0 |
| Disable firewall | Turn off UFW / flush nftables | P1 |
| Reset firewall | Reset to defaults | P2 |

### Application Profiles (UFW)

| Action | Description | Priority |
|--------|-------------|----------|
| List app profiles | `ufw app list` | P1 |
| Allow app | `ufw allow <app>` | P1 |
| Deny app | `ufw deny <app>` | P1 |
| Show app detail | `ufw app info <app>` | P2 |

### Insight & Diagnostics

| Action | Description | Priority |
|--------|-------------|----------|
| Exposure audit | Compare firewall rules vs actual listening ports | P0 |
| Missing rules | Detect services listening without firewall rules | P0 |
| Overly permissive | Flag rules allowing any→any on sensitive ports | P0 |
| Rule conflicts | Detect contradictory or redundant rules | P2 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/firewall/status` | Full firewall summary |
| GET | `/api/v1/nodes/:id/firewall/rules` | List all rules |
| POST | `/api/v1/nodes/:id/firewall/rules` | Add a rule |
| DELETE | `/api/v1/nodes/:id/firewall/rules/:ruleId` | Delete a rule |
| PUT | `/api/v1/nodes/:id/firewall/rules/:ruleId` | Update a rule |
| PUT | `/api/v1/nodes/:id/firewall/defaults` | Set default policies |
| POST | `/api/v1/nodes/:id/firewall/enable` | Enable firewall |
| POST | `/api/v1/nodes/:id/firewall/disable` | Disable firewall |
| POST | `/api/v1/nodes/:id/firewall/reset` | Reset to defaults |
| GET | `/api/v1/nodes/:id/firewall/exposure-audit` | Compare ports vs rules |

## Frontend UI

### Node Detail — Firewall Section

**Status card:**
- Active backend (UFW / nftables / none)
- Status (active / inactive)
- Default policies (incoming: deny, outgoing: allow)
- Total rules count
- Quick actions: Enable, Disable

**Rules table:**

| Column | Description |
|--------|-------------|
| Action | Allow / Deny / Limit |
| Protocol | TCP / UDP / Both |
| Port | Port number or range |
| Source | IP/CIDR |
| Destination | IP/CIDR |
| Interface | Network interface (if specified) |
| Application | UFW app profile name (if applicable) |
| Logging | Log enabled indicator |
| Controls | Delete button |

**Exposure Audit:**
- Tab showing comparison between listening ports and firewall rules
- Highlighted mismatches:
  - Listening on a port not covered by any allow rule
  - Allow rule for a port not currently listening (stale rule)
  - Allow rule from 0.0.0.0/0 on sensitive ports (22, 3306, 6379, etc.)

**Rule Builder (modal):**
- Action: Allow / Deny / Limit
- Direction: In / Out
- Protocol: TCP / UDP / Both
- Port: number or range
- Source CIDR
- Interface (optional)
- Rate limit (optional)
- Description / comment

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| Firewall disabled | Firewall was active but now disabled | critical |
| Port exposed | Listening port without firewall allow rule | high |
| Default allow incoming | Default policy changed to allow all incoming | critical |
| SSH publicly allowed | Port 22 allowed from 0.0.0.0/0 | high |
| Stale rules found | Allow rules for services not running | info |

### Policy Integration

- "If a new listening port is detected without a firewall rule, alert"
- "If firewall is disabled, suggest re-enabling"
- "Automatically add firewall rule when a service is marked as 'expose privately'"
- "On incident escalation, add temporary block rule for source IP"
- "Ensure SSH rate limiting on port 22"

## Agent Implementation Details

### Package Structure

```
agent/internal/firewall/
├── firewall.go        # Main integration — detect active backend
├── ufw.go             # UFW-specific parsing and management
├── nftables.go        # nftables-specific parsing and management
├── rule.go            # Common rule types and formatting
└── command.go         # Rule add/delete, enable/disable, policy commands
```

### Key Interactions

**UFW commands:**
```bash
# Status
ufw status verbose
ufw app list
ufw app info <app>

# Management
ufw enable
ufw disable
ufw default deny incoming
ufw default allow outgoing
ufw allow <port>[/<proto>]
ufw deny <port>[/<proto>]
ufw allow from <source> to any port <port> proto <proto>
ufw delete <rule_num>
ufw reload
```

**nftables commands:**
```bash
# Status
nft list ruleset

# Management — ruleset applied via files or direct commands
nft add rule inet filter input tcp dport <port> accept
nft add rule inet filter input tcp dport <port> drop
nft delete rule inet filter input handle <handle>
nft flush ruleset
```

### Config Files

```
/etc/ufw/ufw.conf                   # UFW main config
/etc/ufw/before.rules               # Pre-defined rules
/etc/ufw/after.rules                # Post-defined rules
/etc/ufw/user.rules                 # User-defined rules
/etc/ufw/applications.d/            # App profile definitions
/etc/nftables.conf                   # nftables ruleset
/etc/nftables/                       # nftables config directory
/etc/sysconfig/nftables.conf         # nftables sysconfig (RHEL)
```

### Rule ID Strategy

UFW rules are identified by their number in `ufw status numbered`. For nftables, rules are identified by their handle.

For the API, assign a stable rule ID based on:
- UFW: rule number (may shift, warn on reorder)
- nftables: handle + chain

A safer approach: generate a content-based hash of the rule parameters and store it client-side when the rule list is fetched. When the agent applies changes, it maps hash → actual rule number/handle at that moment.

### Error Handling

- UFW not installed → show "not available", suggest installation
- `ufw enable` requires non-interactive mode → use `ufw --force enable`
- Permission denied → agent runs as root
- nftables syntax error on rule add → capture stderr, surface error
- Conflicting rules → detect and warn before adding
- Rule number shift → refresh rule list after add/delete

## Implementation Order

1. Agent: UFW installed/active detection + rules parsing
2. Agent: nftables detection + ruleset parsing
3. Agent: default policy detection
4. Agent: status report integration
5. Agent: exposure audit (compare listening ports vs firewall rules)
6. Backend: read-only endpoints (status, rules, exposure-audit)
7. Frontend: firewall section on node detail
8. Frontend: rules table with filtering
9. Backend: rule management endpoints (add/delete)
10. Agent: rule add/delete command execution
11. Frontend: rule builder modal
12. Backend: enable/disable/defaults endpoints
13. Agent: enable/disable command execution
14. Frontend: exposure audit view
15. Incident generation
16. Policy integration
