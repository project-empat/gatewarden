# Frontend Pages

## Stack

- **Build:** Vite
- **Framework:** React, TypeScript
- **Routing:** TanStack Router
- **Data fetching:** TanStack Query
- **Styling:** TailwindCSS + daisyUI
- **Icons:** lucide
- **Brand icons:** thesvg
- **State:** Zustand
- **Theme:** Dark mode (default, with toggle)

## Pages

### Login

| Element | Description |
|---------|-------------|
| Route | `/login` |
| Auth | None (public) |
| Content | Email + password form, submit, link to register |
| States | Loading, error (invalid credentials), success (redirect to dashboard) |

### Dashboard

| Route | `/` (redirect from `/dashboard`) |
|-------|----------------------------------|

The main operational view. NOT a metrics dashboard.

**Widgets (row-based layout):**

- **Security Score** — overall health (good / warning / critical) across all nodes
- **Exposed Services** — count + quick list of publicly exposed services
- **Active Incidents** — list of open incidents sorted by severity
- **Recent Attacks** — live feed of blocked attempts (last 24h)
- **Node Status** — online / offline / degraded node count
- **Quick Actions** — common remediation buttons

**Empty state:** No nodes registered — shows install instructions.

### Nodes

| Route | `/nodes`, `/nodes/:id` |
|-------|------------------------|

**List view (`/nodes`):**

- Table of all registered nodes
- Columns: hostname, IP, status, OS, last seen, incidents count, security score
- Sortable, filterable
- Row click navigates to detail

**Detail view (`/nodes/:id`):**

- **Overview:** hostname, OS, kernel, uptime, IP
- **Security checks** (per-agent capability):
  - SSH status: secure / exposed / misconfigured — [detail](ssh-hardening-features.md)
  - Firewall status: active / inactive, rule count — [detail](firewall-features.md)
  - Docker exposure warnings — [detail](docker-security-features.md)
  - CrowdSec status + active decisions — [detail](crowdsec-features.md)
  - Fail2Ban status: jails, bans, whitelist — [detail](fail2ban-features.md)
  - Tailscale connection status — [detail](tailscale-features.md)
  - Cloudflare Tunnel status — [detail](cloudflare-features.md)
  - Auth log metrics: failed attempts, top IPs — [detail](authlog-features.md)
- **Exposed Ports:** table of listening services with exposure info
- **Incidents:** list of incidents for this node
- **Recent Events:** event feed for this node

**Empty state:** "No nodes registered. Run the install command on your server."

### Incidents

| Route | `/incidents`, `/incidents/:id` |
|-------|-------------------------------|

**List view (`/incidents`):**

- Table of all incidents
- Columns: severity, title, node, category, status, created time
- Filterable by severity, status, category
- Sortable by time, severity
- Color coding: red (critical), yellow (warning), blue (info)

**Detail view (`/incidents/:id`):**

- Title, severity, timestamp
- Description with contextual details
- Affected node link
- Action buttons: Acknowledge, Resolve
- Related events timeline
- Suggested remediation (if automated action exists)

**Empty state:** "No incidents. Your infrastructure looks healthy."

### Policies

| Route | `/policies` |
|-------|-------------|

**List view:**

- Table of defined policies
- Columns: name, enabled, conditions summary, actions summary, last updated
- Toggle to enable/disable

**Create/Edit (modal or inline):**

- Policy name and description
- Condition builder (what triggers this policy):
  - Node selector (all / specific tags)
  - Event type (exposure / attack / misconfig)
  - Threshold (e.g., > 10 failed logins in 5 minutes)
- Action selector (what to do):
  - Notify (webhook, email)
  - Remediate (block IP, apply firewall rule, restart service)

**Empty state:** "No policies defined. Create your first automated response."

### Settings

| Route | `/settings` |
|-------|-------------|

**Sections:**

- **General:** Site name, timezone
- **Notifications:** Webhook URLs, email config
- **Integrations:**
  - Cloudflare API token configuration
  - Tailscale API key
  - CrowdSec API endpoint
- **Security:** JWT secret rotation, session timeout
- **Users:** List of users, invite flow (MVP: single user)

**Empty state (integrations):** "Connect your services to enable security checks."

## Design Guidelines

- Dark mode default, light mode toggle
- Consistent spacing: 4px grid, 16px/24px/32px gaps
- Cards for dashboard widgets
- Tables for list views
- Modals or slide-overs for create/edit forms
- Toasts for confirmations and errors
- Loading skeletons for data-fetching states
- Error boundaries per page

## Major UX Labels / Actions

- "Expose privately" — mark a service as tunnel-only
- "Require MFA" — enforce MFA on a service
- "Restrict to team" — limit access by identity
- "SSH exposed publicly" — warning
- "Docker socket exposed" — warning
- "Grafana publicly reachable" — warning

## State Management

- **Server state:** TanStack Query (cache, refetch, optimistic updates)
- **Client state:** Zustand (theme, sidebar collapse, user session)
- **Auth token:** Stored in memory + httpOnly cookie or localStorage

## Route Tree

```
/
├── login
├── dashboard          (index redirect /)
├── nodes
│   └── $nodeId
├── incidents
│   └── $incidentId
├── policies
└── settings
```

## Implementation Order

1. Vite + React + TypeScript scaffold
2. TailwindCSS + daisyUI setup
3. TanStack Router integration
4. TanStack Query setup
5. Zustand store (auth, theme)
6. Dark mode toggle
7. API client layer
8. Login page (form + auth flow)
9. Dashboard layout + placeholder widgets
10. Nodes list page
11. Nodes detail page
12. Incidents list page
13. Incidents detail page
14. Policies list + create/edit
15. Settings page
16. Real-time event subscription (SSE)
