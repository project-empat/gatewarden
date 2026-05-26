# Policies & Automated Remediation

## Overview

Policies define automated responses to security events. Gatewarden's policy engine watches for conditions across all nodes and triggers actions — notifications, remediation commands, or escalations — without manual intervention.

## Core Concepts

A policy consists of:

```
Policy
├── Name + Description
├── Enabled / Disabled
├── Severity (for incidents generated)
│
├── Triggers (OR logic — any match triggers)
│   ├── Event type filter
│   ├── Node selector
│   ├── Threshold + time window
│   └── Conditions (AND logic within a trigger)
│
└── Actions (run in order)
    ├── Notify (webhook, email)
    ├── Remediate (agent command)
    ├── Create incident
    └── Escalate (if context matches)
```

### Trigger Model

```jsonc
{
  "triggers": [
    {
      "event_types": ["auth_failure", "port_exposed", "docker_issue", "attack_detected", "service_down", "config_change"],
      "node_selector": {
        "all": false,
        "node_ids": ["uuid1", "uuid2"],
        "tags": ["production", "public-facing"],
        "os": ["ubuntu"],
        "exposure": ["public"]          // "public" | "tailscale" | "vpn" | "private"
      },
      "conditions": {
        "operator": "and",              // "and" | "or"
        "rules": [
          { "field": "attempt_count", "comparison": "gt", "value": 10 },
          { "field": "time_window_minutes", "value": 5 },
          { "field": "source_ip_reputation", "comparison": "lt", "value": 50 }
        ]
      }
    }
  ]
}
```

### Action Model

```jsonc
{
  "actions": [
    {
      "type": "notify",
      "config": {
        "channel": "webhook",
        "url": "https://hooks.slack.com/...",
        "template": "incident"
      }
    },
    {
      "type": "remediate",
      "config": {
        "target": "source_node",         // "source_node" | "all_nodes" | "specific_nodes"
        "node_ids": [],
        "command": {
          "action": "fail2ban_ban_ip",
          "params": {
            "ip": "{{.SourceIP}}",
            "duration": "1h",
            "jail": "sshd"
          }
        }
      }
    },
    {
      "type": "create_incident",
      "config": {
        "severity": "high",
        "category": "attack",
        "title_template": "SSH brute force from {{.SourceIP}}"
      }
    }
  ]
}
```

## Available Remediation Actions

### Fail2Ban Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `fail2ban_ban_ip` | ip, duration, jail | Ban an IP in a specific jail |
| `fail2ban_unban_ip` | ip, jail | Unban an IP |
| `fail2ban_start_jail` | jail | Enable a jail |
| `fail2ban_stop_jail` | jail | Disable a jail |
| `fail2ban_set_bantime` | jail, seconds | Change ban duration |
| `fail2ban_set_maxretry` | jail, count | Change max retries |

### Firewall Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `firewall_block_ip` | ip, port, proto | Add deny rule for an IP |
| `firewall_allow_ip` | ip, port, proto | Add allow rule for an IP |
| `firewall_add_rule` | action, port, proto, source | Add generic rule |
| `firewall_delete_rule` | ruleId | Remove a rule |
| `firewall_enable` | — | Enable firewall |
| `firewall_set_default` | policy, direction | Set default policy |

### SSH Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `ssh_harden` | — | Apply all SSH hardening |
| `ssh_disable_password_auth` | — | Disable password authentication |
| `ssh_disable_root_login` | — | Disable root login |
| `ssh_restrict_to_localhost` | — | Bind SSH to 127.0.0.1 |
| `ssh_restart` | — | Test config and restart SSHD |

### Docker Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `docker_restart_container` | container | Restart a container |
| `docker_stop_container` | container | Stop a container |

### Cloudflare Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `cloudflare_enable_access` | hostname | Add Access policy for a hostname |
| `cloudflare_require_mfa` | hostname | Require MFA on an Access app |

### System Actions

| Action ID | Parameters | Description |
|-----------|------------|-------------|
| `restart_service` | service | Restart a systemd service |
| `reload_service` | service | Reload a systemd service |
| `execute_command` | command, args | Run arbitrary command (with restricted allowlist) |

## Notification Actions

### Webhook

POST JSON payload to a URL:

```jsonc
{
  "event": "policy_triggered",
  "policy": {
    "id": "uuid",
    "name": "SSH Brute Force Response"
  },
  "incident": {
    "id": "uuid",
    "severity": "critical",
    "title": "SSH brute force from 203.0.113.1",
    "node": "web-01",
    "timestamp": "2026-05-26T12:00:00Z"
  },
  "actions_taken": [
    { "action": "fail2ban_ban_ip", "status": "success", "details": "IP banned in sshd jail for 1h" }
  ]
}
```

### Email (MVP via SMTP)

- Simple text or HTML template
- Configurable recipients
- Per-policy or global configuration

### Webhook Integrations

| Platform | Format | Priority |
|----------|--------|----------|
| Slack | Webhook URL with JSON payload | P0 |
| Discord | Webhook URL | P0 |
| Telegram | Bot token + chat ID | P1 |
| PagerDuty | Events API v2 | P2 |
| OpsGenie | API key | P2 |
| Generic webhook | Custom JSON | P0 |

## Built-in Policy Templates

### 1. SSH Brute Force Response

```yaml
name: SSH Brute Force Response
description: Automatically block IPs with excessive failed SSH attempts
enabled: true
triggers:
  - event_types: [auth_failure]
    conditions:
      rules:
        - { field: attempt_count, comparison: gte, value: 10 }
        - { field: time_window_minutes, value: 5 }
actions:
  - type: remediate
    config:
      command: { action: fail2ban_ban_ip, params: { ip: "{{.SourceIP}}", duration: "1h", jail: sshd } }
  - type: create_incident
    config: { severity: critical, category: attack, title_template: "SSH brute force blocked from {{.SourceIP}}" }
```

### 2. Public SSH Exposure

```yaml
name: Public SSH Exposure
description: Alert when SSH is publicly exposed without password auth disabled
enabled: true
triggers:
  - event_types: [port_exposed]
    conditions:
      rules:
        - { field: port, comparison: eq, value: 22 }
        - { field: exposure, comparison: eq, value: public }
actions:
  - type: create_incident
    config: { severity: high, category: exposure, title_template: "SSH publicly exposed on {{.NodeName}}" }
  - type: notify
    config: { channel: webhook, template: incident }
```

### 3. Docker Socket Mounted

```yaml
name: Docker Socket Exposure
description: Alert when any container mounts the Docker socket
enabled: true
triggers:
  - event_types: [docker_issue]
    conditions:
      rules:
        - { field: issue_type, comparison: eq, value: docker_socket_mounted }
actions:
  - type: create_incident
    config: { severity: critical, category: misconfig, title_template: "Docker socket mounted in container {{.ContainerName}}" }
  - type: notify
    config: { channel: webhook, template: incident }
```

### 4. Firewall Disabled

```yaml
name: Firewall Disabled Alert
description: Alert when host firewall is turned off
enabled: true
triggers:
  - event_types: [config_change]
    conditions:
      rules:
        - { field: change_type, comparison: eq, value: firewall_disabled }
actions:
  - type: create_incident
    config: { severity: critical, category: misconfig, title_template: "Firewall disabled on {{.NodeName}}" }
  - type: remediate
    config:
      command: { action: firewall_enable }
```

### 5. Root Login Target

```yaml
name: Root Login Attacks
description: Escalate when root account is repeatedly targeted
enabled: false
triggers:
  - event_types: [auth_failure]
    conditions:
      rules:
        - { field: username, comparison: eq, value: root }
        - { field: attempt_count, comparison: gte, value: 5 }
        - { field: time_window_minutes, value: 5 }
actions:
  - type: remediate
    config:
      command: { action: fail2ban_ban_ip, params: { ip: "{{.SourceIP}}", duration: "24h", jail: sshd } }
  - type: create_incident
    config: { severity: critical, category: attack, title_template: "Root account targeted from {{.SourceIP}}" }
  - type: notify
    config: { channel: webhook, template: incident }
```

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/policies` | List all policies |
| POST | `/api/v1/policies` | Create a new policy |
| GET | `/api/v1/policies/:id` | Get policy detail |
| PUT | `/api/v1/policies/:id` | Update a policy |
| DELETE | `/api/v1/policies/:id` | Delete a policy |
| POST | `/api/v1/policies/:id/toggle` | Enable/disable a policy |
| GET | `/api/v1/policies/templates` | List built-in policy templates |
| POST | `/api/v1/policies/from-template` | Create policy from a template |
| GET | `/api/v1/policies/:id/history` | Policy execution history |
| GET | `/api/v1/policies/:id/test` | Dry-run a policy against current state |

## Frontend UI

### Policies List Page (`/policies`)

**Route:** `/policies`

| Column | Description |
|--------|-------------|
| Name | Policy name (click to edit) |
| Enabled | Toggle switch |
| Triggers summary | Event types + conditions (collapsed) |
| Actions summary | Action types (notify, remediate, incident) |
| Last triggered | Timestamp |
| Status | OK / Warning / Error (from last execution) |

**Empty state:** "No policies defined. Create your first automated response." with "Create from template" CTA.

### Policy Editor

**Create/Edit page:**

**Section 1: Basic Info**
- Name (text)
- Description (textarea)
- Enabled (toggle)
- Severity (dropdown: info, warning, high, critical)

**Section 2: Triggers**

Trigger builder with dynamic form rows:

- Event type (multi-select: auth_failure, port_exposed, docker_issue, attack_detected, service_down, config_change)
- Node selector:
  - All nodes (toggle)
  - Specific nodes (multi-select)
  - Tags (tag input)
  - OS filter (checkboxes)
  - Exposure filter (checkboxes)
- Conditions (AND/OR group):
  - Field (select): attempt_count, time_window_minutes, port, exposure, username, issue_type
  - Comparison: gt, gte, lt, lte, eq, neq, contains, matches
  - Value: dynamic input (number, text, select depending on field)

**Section 3: Actions**

Action builder:

- Add action (dropdown: notify, remediate, create_incident)
- Remove action (X button)
- Reorder actions (drag handle)

**Notify action form:**
- Channel: webhook (URL input), email (email input)
- Template: incident, custom (message template with {{.Variable}} support)

**Remediate action form:**
- Target: source_node, all_nodes, specific_nodes
- Action: dropdown of all available remediation actions
- Parameters: dynamic form based on selected action

**Create incident action form:**
- Severity, category, title template

**Section 4: Test & Save**
- "Test with current state" button (dry run)
- Save / Cancel

### Policy Execution History

**Route:** `/policies/:id/history`

Table of policy executions:

| Column | Description |
|--------|-------------|
| Timestamp | When triggered |
| Triggering event | Summary of what matched |
| Actions taken | List of actions with status |
| Status | Success / Partial / Failed |

### Policy Templates

**Route:** `/policies/templates`

Grid of built-in templates with:
- Name, description, severity
- Triggers summary
- Actions summary
- "Create from template" button

## Policy Engine Architecture

### Execution Flow

```
Agent Report / Event
       │
       ▼
  Event Bus (internal channel)
       │
       ▼
  Policy Matcher
       │
       ├── Load all enabled policies
       ├── For each policy:
       │   ├── Match event type
       │   ├── Match node selector
       │   ├── Evaluate conditions (threshold, time window)
       │   └── If matched: queue for execution
       │
       ▼
  Policy Executor
       │
       ├── Execute actions in order
       ├── On failure:
       │   ├── Retry (up to 3 times)
       │   └── Mark policy as error
       ├── Record execution history
       └── Create incident if configured
```

### Event Types

| Event Type | Source | Payload Fields |
|------------|--------|----------------|
| `auth_failure` | Auth log parser | source_ip, username, attempt_count, time_window_minutes, node_id |
| `port_exposed` | Port scanner | port, proto, process, exposure, node_id |
| `docker_issue` | Docker scanner | issue_type, container_name, image, node_id |
| `attack_detected` | CrowdSec/Fail2Ban | source_ip, scenario, decision_count, node_id |
| `service_down` | System health | service_name, node_id |
| `config_change` | Agent diff | change_type, old_value, new_value, node_id |
| `incident_created` | Internal | incident_id, severity, category, node_id |

### Template Variables

Available variables for action templates:

| Variable | Source | Example |
|----------|--------|---------|
| `{{.SourceIP}}` | Event | `203.0.113.1` |
| `{{.NodeName}}` | Node | `web-01` |
| `{{.NodeID}}` | Node | `uuid` |
| `{{.Port}}` | Event | `22` |
| `{{.Username}}` | Event | `root` |
| `{{.ContainerName}}` | Event | `grafana` |
| `{{.Severity}}` | Policy | `critical` |
| `{{.IncidentID}}` | Runtime | `uuid` |
| `{{.Timestamp}}` | Runtime | `2026-05-26T12:00:00Z` |

## Implementation Order

### Phase 1: Policy CRUD

- [ ] Policy data model (database schema)
- [ ] Policy CRUD API endpoints
- [ ] Basic policy list UI
- [ ] Policy create/edit form
- [ ] Policy enable/disable toggle

### Phase 2: Trigger Engine

- [ ] Event type definitions and normalization
- [ ] Node selector matching
- [ ] Condition evaluation (threshold, comparison)
- [ ] Time-windowed condition evaluation
- [ ] Policy matcher service

### Phase 3: Action Execution

- [ ] Action registry (supported actions)
- [ ] Agent command queue (API → agent)
- [ ] Sequential action executor
- [ ] Action retry logic
- [ ] Execution history recording

### Phase 4: Notifications

- [ ] Generic webhook sender
- [ ] Template rendering ({{.Variable}} substitution)
- [ ] Incident creation action
- [ ] Notification history

### Phase 5: Templates & UX

- [ ] Built-in policy templates
- [ ] Template import flow
- [ ] Policy test/dry-run
- [ ] Execution history UI
- [ ] Action builder with dynamic forms
- [ ] Node selector UI

### Phase 6: Advanced

- [ ] Multi-condition groups (AND/OR nesting)
- [ ] Escalation chains
- [ ] Approval workflows
- [ ] Policy versioning
- [ ] Scheduled policies (cron-based triggers)

## Database Schema

### `policies`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| name | TEXT | |
| description | TEXT | |
| enabled | BOOLEAN | |
| severity | TEXT | info / warning / high / critical |
| triggers | JSONB | Trigger definitions |
| actions | JSONB | Action definitions |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### `policy_executions`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| policy_id | UUID | FK |
| event_type | TEXT | |
| event_payload | JSONB | |
| node_id | UUID | FK (nullable) |
| results | JSONB | Per-action results |
| status | TEXT | success / partial / failed |
| triggered_at | TIMESTAMPTZ | |
| completed_at | TIMESTAMPTZ | |

### `agent_commands`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| agent_id | UUID | FK |
| policy_id | UUID | FK (nullable) |
| action | TEXT | Action ID |
| params | JSONB | |
| status | TEXT | pending / delivered / running / success / failed |
| result | JSONB | |
| created_at | TIMESTAMPTZ | |
| delivered_at | TIMESTAMPTZ | |
| completed_at | TIMESTAMPTZ | |
