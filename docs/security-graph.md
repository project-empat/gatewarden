# Infrastructure Security Graph

## Overview

The Infrastructure Security Graph is Gatewarden's killer feature. It maps relationships between users, services, containers, tunnels, firewall rules, identities, and attacks — providing a visual understanding of your infrastructure's security posture.

## Concept

```
Internet
  │
  ▼
Cloudflare Tunnel ─── Access Policy ─── MFA Required
  │
  ▼
Docker Host ─── Port 3000 ─── Container: grafana
  │                              │
  │                              └── Image: grafana/grafana:11.0.0
  │                              └── Resource Limits: none (⚠️)
  │
  ├── Port 22 ─── SSH ─── Password Auth: disabled 👍
  │                           Root Login: prohibited 👍
  │                           Publicly exposed: ❌
  │
  ├── Port 443 ─── nginx
  │
  └── Firewall ─── UFW Active ─── Default: deny incoming
                        ├── Allow 22/tcp from any  (⚠️ SSH public)
                        ├── Allow 443/tcp from any
                        └── Allow 3000/tcp from 127.0.0.1

Fail2Ban ─── Active Jails: sshd (3 banned), nginx-http-auth (0 banned)
  │
  └── Attack: SSH brute force from 203.0.113.1 (12 attempts in 5 min)

CrowdSec ─── Active Decisions: 24
  ├── 203.0.113.1 banned (ssh_bf)
  └── 198.51.100.1 banned (http_scan)
```

## Relationship Model

### Entity Types

| Entity | Description | Attributes |
|--------|-------------|------------|
| Node | Physical or virtual machine | hostname, IP, OS, status |
| Container | Docker container | name, image, ports, status, security issues |
| Service | Listening service | port, protocol, process, exposure |
| Tunnel | Cloudflare Tunnel / Tailscale | name, status, ingress rules |
| Firewall | UFW / nftables rules | backend, rules, default policy |
| Identity | User or team | email, role, access level |
| Attack | Security event | source IP, type, count, timespan |
| Incident | Alert or finding | severity, category, status |
| Policy | Automation rule | conditions, actions |
| Integration | Third-party tool | CrowdSec, Fail2Ban, Cloudflare, Tailscale |

### Relationship Types

| Relationship | Example | Risk |
|-------------|---------|------|
| `exposes` | Service → Internet Tunnel | Exposes service to public |
| `protects` | Firewall → Service | Restricts access to service |
| `blocks` | Fail2Ban → Attack IP | Blocks malicious IP |
| `detects` | CrowdSec → Attack | Alerts on malicious activity |
| `runs_on` | Container → Node | Container deployed on host |
| `configured_by` | Policy → Firewall | Policy manages firewall rule |
| `connected_via` | Node → Tailscale | Node in tailnet mesh |
| `served_by` | Service → Container | Service runs inside container |
| `accessed_by` | Identity → Service | User/team can access service |
| `reported_by` | Incident → Service | Incident references service |

## Graph Data Model

### Node Attributes (Cytoscape.js / D3.js compatible)

```jsonc
{
  "nodes": [
    { "id": "internet", "label": "Internet", "type": "concept", "icon": "globe" },
    { "id": "tunnel-grafana", "label": "Cloudflare Tunnel", "type": "tunnel", "icon": "cloud", "status": "active" },
    { "id": "node-web01", "label": "web-01", "type": "node", "icon": "server", "hostname": "web-01.example.com", "status": "online" },
    { "id": "container-grafana", "label": "grafana", "type": "container", "icon": "container", "image": "grafana/grafana:11.0.0" },
    { "id": "svc-3000", "label": "grafana:3000", "type": "service", "icon": "activity", "port": 3000, "exposure": "public" },
    { "id": "svc-22", "label": "SSH:22", "type": "service", "icon": "terminal", "port": 22, "exposure": "public" },
    { "id": "firewall", "label": "UFW", "type": "firewall", "icon": "shield", "rules": 12, "active": true },
    { "id": "fail2ban", "label": "Fail2Ban", "type": "integration", "icon": "ban", "jails": 3, "banned": 7 },
    { "id": "crowdsec", "label": "CrowdSec", "type": "integration", "icon": "users", "decisions": 24 },
    { "id": "attack-1", "label": "203.0.113.1", "type": "attack", "icon": "alert-triangle", "attempts": 45 }
  ],
  "edges": [
    { "id": "e1", "source": "internet", "target": "tunnel-grafana", "label": "routes to", "color": "#f59e0b" },
    { "id": "e2", "source": "tunnel-grafana", "target": "node-web01", "label": "terminates on", "color": "#3b82f6" },
    { "id": "e3", "source": "node-web01", "target": "container-grafana", "label": "runs", "color": "#10b981" },
    { "id": "e4", "source": "container-grafana", "target": "svc-3000", "label": "exposes", "color": "#f59e0b" },
    { "id": "e5", "source": "node-web01", "target": "svc-22", "label": "listens on", "color": "#ef4444" },
    { "id": "e6", "source": "node-web01", "target": "firewall", "label": "protected by", "color": "#10b981" },
    { "id": "e7", "source": "firewall", "target": "svc-22", "label": "allows", "color": "#f59e0b" },
    { "id": "e8", "source": "node-web01", "target": "fail2ban", "label": "runs", "color": "#3b82f6" },
    { "id": "e9", "source": "fail2ban", "target": "attack-1", "label": "blocked", "color": "#ef4444" },
    { "id": "e10", "source": "crowdsec", "target": "attack-1", "label": "detected", "color": "#ef4444" }
  ]
}
```

## Graph Queries

### Risk Path Analysis

Find the shortest path from "Internet" to sensitive services:

```
Internet → Cloudflare Tunnel → Docker Host → Grafana
```
Risk: `public_web_app` (medium) — requires valid Cloudflare Access

```
Internet → SSH:22 → Docker Host
```
Risk: `public_ssh` (high) — SSH publicly exposed with password auth disabled but still accessible

### Blast Radius

When an incident occurs, the graph shows all affected entities:

```
Incident: SSH Brute Force (203.0.113.1)
  │
  ├── Affects: web-01 (node)
  ├── Via: SSH Port 22 (service)
  ├── Blocked by: Fail2Ban sshd jail
  ├── Also detected: CrowdSec (ssh_bf scenario)
  └── Also at risk: All containers on web-01
```

### Exposure Analysis

```
Node: web-01
  │
  ├── Publicly exposed:
  │   ├── SSH:22 → Internet (any)
  │   └── Grafana:3000 → Cloudflare Tunnel (with Access policy)
  │
  ├── Tailscale-only:
  │   └── Admin panel:9090 (Tailscale IP only)
  │
  └── Private (127.0.0.1):
      └── Redis:6379
```

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/graph` | Full security graph for all nodes |
| GET | `/api/v1/graph/nodes/:id` | Sub-graph for a specific node |
| GET | `/api/v1/graph/risk-paths` | All risk paths from internet to services |
| GET | `/api/v1/graph/relationships` | Relationship query by type |
| GET | `/api/v1/graph/impact/:incidentId` | Blast radius for an incident |
| GET | `/api/v1/graph/summary` | Aggregate graph summary (counts by type) |

## Frontend UI

### Graph Page

**Route:** `/graph`

A full-page interactive graph visualization:

- Force-directed layout (dagre or cose-bilkent)
- Zoom and pan
- Click node to show details
- Click edge to show relationship info
- Filter by entity type (checkboxes)
- Search/find node by name
- Legend for entity colors and shapes

**Node interactions:**
- Hover: tooltip with key attributes
- Click: slide-over panel with entity details and related incidents
- Double-click: center view on node and expand 1-hop neighborhood

**Edge interactions:**
- Hover: relationship type label
- Click: relationship detail

**Context toolbar:**
- Fit to screen
- Zoom to selection
- Toggle labels
- Layout selector (force-directed, hierarchical, radial)
- Export as PNG/SVG
- Auto-layout toggle (animate on changes)

### Node Detail — Graph Mini-View

Inline mini-graph showing the selected node and its immediate neighbors (1-2 hops). Embedded in the node detail page.

### Dashboard Graph Widget

A summary graph showing:
- Topology of all nodes
- Exposure indicators as colored halos
- Active attacks flashing red
- Animated edges for active connections

### Color Scheme

| Entity Type | Color | Shape |
|-------------|-------|-------|
| Internet | #6b7280 (gray) | Cloud icon |
| Node | #3b82f6 (blue) | Server rectangle |
| Container | #10b981 (green) | Box |
| Service | #f59e0b (amber) | Circle |
| Tunnel | #8b5cf6 (purple) | Diamond |
| Firewall | #06b6d4 (cyan) | Shield |
| Integration | #ec4899 (pink) | Hexagon |
| Identity | #14b8a6 (teal) | Person icon |
| Attack | #ef4444 (red) | Triangle (animated) |
| Incident | #f97316 (orange) | Warning diamond |

**Edge colors:**
| Relationship | Color |
|-------------|-------|
| Routing / Network | #6b7280 |
| Blocking / Protection | #10b981 |
| Exposure / Risk | #ef4444 |
| Detection | #f59e0b |
| Deployment | #3b82f6 |

## Implementation Phases

### Phase 1: Backend Graph Construction

- [ ] Define entity types and relationship model
- [ ] Build graph from agent reports (nodes, services, containers)
- [ ] Add integration data (tunnels, firewall, CrowdSec, Fail2Ban)
- [ ] Compute edges automatically from entity attributes
- [ ] Graph serialization endpoint (Cytoscape.js JSON format)

### Phase 2: Static Graph Rendering

- [ ] Integrate graph visualization library (Cytoscape.js or D3.js)
- [ ] Render entity nodes with icons and colors
- [ ] Render relationship edges with labels
- [ ] Basic zoom and pan
- [ ] Legend component

### Phase 3: Interactive Graph

- [ ] Force-directed layout
- [ ] Node click → detail panel
- [ ] Search/filter by entity type
- [ ] Fit-to-screen
- [ ] Export as image

### Phase 4: Intelligence Layer

- [ ] Risk path computation (Internet → sensitive service paths)
- [ ] Blast radius for incidents
- [ ] Animated edges for active attacks
- [ ] Time-travel (show graph state at a point in time)
- [ ] What-if analysis ("what if this port is closed?")

## Technology Considerations

### Graph Libraries

| Library | Pros | Cons |
|---------|------|------|
| Cytoscape.js | Mature, performant, lots of layouts, active dev | Large bundle |
| D3.js force | Flexible, customizable | More code, fewer built-in features |
| vis-network | Simple API, good defaults | Less flexible |
| React Flow | React-native, excellent DX | Focused on flow charts, not graph viz |

**Recommendation:** Cytoscape.js for the main graph (performance at scale), React Flow for the mini-graph on node detail pages (React integration).

### Graph Construction

The graph is constructed server-side from the database:

1. Query all nodes, services, containers, incidents, etc.
2. Infer relationships based on:
   - Service → Node (service runs on node)
   - Container → Node (container runs on node)
   - Service → Container (service port maps to container)
   - Tunnel → Node (tunnel terminates on node)
   - Firewall → Node (firewall protects node)
   - Integration → Node (CrowdSec/Fail2Ban runs on node)
   - Incident → Node/Service (incident references entity)
3. Serialize to Cytoscape.js format
4. Cache for N seconds (graph changes are slow)

### Performance

For larger deployments (50+ nodes, 200+ services):
- Graph rendering with WebGL (Cytoscape.js supports this)
- Server-side graph pruning (return only active entities)
- Lazy loading of sub-graphs (expand on click)
- Debounce auto-layout on data changes

## Implementation Order

1. Entity type and relationship model definitions
2. Server-side graph builder from agent/integration data
3. Graph serialization endpoint
4. Simple static graph renderer (Cytoscape.js)
5. Basic zoom/pan/layout
6. Node detail mini-graph
7. Interactive features (click, filter, search)
8. Risk path computation
9. Incident blast radius
10. Dashboard graph widget
11. Time-travel and what-if analysis
