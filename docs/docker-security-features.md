# Docker Security

## Overview

Docker containers often introduce security risks through exposed ports, mounted sockets, and insecure configurations. Gatewarden scans containers to surface these risks and provides actionable remediation.

## Agent Integration

### Detection & Status Reporting

| Check | Method | Priority |
|-------|--------|----------|
| Docker installed | Check `docker` binary or dpkg status | P0 |
| Docker running | Check `docker info` or systemd `docker.service` | P0 |
| Container list | `docker ps` — all running containers | P0 |
| Published ports | `docker ps --format json` — port mappings | P0 |
| Port exposure level | Detect if mapped to 0.0.0.0 vs 127.0.0.1 | P0 |
| Docker socket exposure | Check `/var/run/docker.sock` permissions and mount | P0 |
| Container images | Image name, tag, version per container | P0 |
| Container status | Running / restarting / paused / exited | P0 |
| Restart policy | `docker inspect` restart policy per container | P1 |
| Privileged mode | Detect containers running `--privileged` | P0 |
| Capabilities | List added/dropped Linux capabilities | P1 |
| Read-only rootfs | Check if root filesystem is read-only | P1 |
| User namespace | Check if `--userns-remap` is configured | P2 |
| Health check | Does container have a health check configured? | P2 |
| Resource limits | Memory/CPU limits set? | P1 |
| Volume mounts | List host path → container mount mappings | P1 |
| Network mode | Bridge / host / overlay / custom | P0 |
| Exposed networks | Which Docker networks is the container on? | P1 |
| Docker version | `docker version` | P0 |
| Docker Compose | Detect docker-compose.yml files | P2 |
| Container run flags | Detect risky flags: `--pid=host`, `--network=host`, `--ipc=host` | P1 |

### Agent Checks

- Is Docker installed? Is the daemon running?
- What containers are currently running?
- Which containers publish ports to 0.0.0.0 (publicly accessible)?
- Which containers publish ports only to 127.0.0.1?
- Is the Docker socket mounted into any container? (classic Docker-in-Docker risk)
- Which containers run in privileged mode?
- Which containers use host networking?
- Which containers have added dangerous capabilities (SYS_ADMIN, NET_ADMIN, SYS_MODULE)?
- Which containers have no memory/CPU limits?
- Which containers run as root inside the container?
- Are there images from untrusted registries?
- Are any containers using the `:latest` tag in production?
- Are there unused/exited containers consuming resources?

### Agent Report Format

```jsonc
{
  "docker": {
    "installed": true,
    "running": true,
    "version": "27.0.1",
    "containers": [
      {
        "id": "abc123def456",
        "name": "grafana",
        "image": "grafana/grafana:11.0.0",
        "status": "running",
        "restart_policy": "unless-stopped",
        "ports": [
          { "container_port": 3000, "host_port": 3000, "host_ip": "0.0.0.0", "exposure": "public" }
        ],
        "privileged": false,
        "network_mode": "bridge",
        "readonly_rootfs": false,
        "has_healthcheck": true,
        "memory_limit": "512m",
        "cpu_limit": 0.5,
        "user": "472:472",
        "capabilities_added": [],
        "capabilities_dropped": ["ALL"],
        "volumes": [
          { "host_path": "/data/grafana", "container_path": "/var/lib/grafana", "mode": "rw" }
        ],
        "socket_mounted": false,
        "pid_mode": "private",
        "ipc_mode": "private",
        "security_issues": []
      },
      {
        "id": "def789abc012",
        "name": "portainer",
        "image": "portainer/portainer-ce:latest",
        "status": "running",
        "restart_policy": "always",
        "ports": [
          { "container_port": 9000, "host_port": 9000, "host_ip": "127.0.0.1", "exposure": "private" },
          { "container_port": 9443, "host_port": 9443, "host_ip": "127.0.0.1", "exposure": "private" }
        ],
        "privileged": false,
        "network_mode": "bridge",
        "readonly_rootfs": false,
        "has_healthcheck": false,
        "memory_limit": "",
        "cpu_limit": 0,
        "user": "root",
        "capabilities_added": [],
        "capabilities_dropped": [],
        "volumes": [
          { "host_path": "/var/run/docker.sock", "container_path": "/var/run/docker.sock", "mode": "rw" }
        ],
        "socket_mounted": true,
        "pid_mode": "private",
        "ipc_mode": "private",
        "security_issues": ["docker_socket_mounted", "running_as_root", "no_resource_limits", "latest_tag"]
      }
    ],
    "total_containers": 5,
    "running_containers": 4,
    "exited_containers": 1,
    "networks": [
      { "name": "bridge", "driver": "bridge", "containers": 2 },
      { "name": "gatewarden_default", "driver": "bridge", "containers": 3 }
    ],
    "socket_exposed": false,
    "userns_remap": false
  }
}
```

### Security Issue Detection

| Issue | Detection | Severity |
|-------|-----------|----------|
| Docker socket mounted | Container has `/var/run/docker.sock` mounted | critical |
| Privileged mode | `--privileged` flag | critical |
| Host network | `--network=host` | high |
| Host PID | `--pid=host` | high |
| Host IPC | `--ipc=host` | high |
| Root user | Container runs as UID 0 | high |
| Public exposure | Port published on 0.0.0.0 | high |
| `:latest` tag | Image uses `:latest` in production | warning |
| Missing healthcheck | No `HEALTHCHECK` instruction | info |
| No memory limit | No `--memory` set | warning |
| No CPU limit | No `--cpus` set | warning |
| Dangerous capabilities | SYS_ADMIN, SYS_MODULE, SYS_RAWIO, etc. | high |
| All capabilities not dropped | `--cap-drop=ALL` not used | info |
| Exited container | Container in exited state | info |
| Unrestricted restart | `restart_policy: always` without stop handling | info |

## Management Actions

### Container Management

| Action | Description | Priority |
|--------|-------------|----------|
| List containers | All containers with security status | P0 |
| Container detail | Full security analysis per container | P0 |
| Restart container | `docker restart <container>` | P1 |
| Stop container | `docker stop <container>` | P1 |
| Pause container | `docker pause <container>` | P2 |

### Remediation Actions

| Action | Description | Priority |
|--------|-------------|----------|
| Restrict port binding | Change from 0.0.0.0 to 127.0.0.1 | P1 |
| Remove Docker socket | Recommend removing socket mount from compose | P0 |
| Add resource limits | Generate `docker run --memory/--cpus` flags | P2 |
| Add healthcheck | Generate HEALTHCHECK Dockerfile instruction | P2 |
| Drop capabilities | Suggest `--cap-drop=ALL --cap-add=...` | P1 |
| Run as non-root | Highlight containers running as root | P1 |
| Pin image tag | Suggest replacing `:latest` with specific version | P1 |

## Backend API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes/:id/docker/status` | Docker daemon status |
| GET | `/api/v1/nodes/:id/docker/containers` | List containers with security analysis |
| GET | `/api/v1/nodes/:id/docker/containers/:name` | Container security detail |
| GET | `/api/v1/nodes/:id/docker/issues` | Aggregated security issues across all containers |
| GET | `/api/v1/nodes/:id/docker/networks` | List Docker networks |
| POST | `/api/v1/nodes/:id/docker/containers/:name/restart` | Restart container |
| POST | `/api/v1/nodes/:id/docker/containers/:name/stop` | Stop container |

## Frontend UI

### Node Detail — Docker Section

**Status card:**
- Docker installed / not installed
- Daemon running / stopped
- Total containers / running / exited
- Docker version

**Containers table:**

| Column | Description |
|--------|-------------|
| Name | Container name |
| Image | Image:tag |
| Status | Running / Exited / Paused badge |
| Ports | Port mappings with exposure color (red=public, green=private) |
| Privileged | Warning badge if true |
| Socket mount | Warning badge if true |
| Issues count | Number of security issues |
| Actions | Restart, Stop buttons |

**Container detail (expandable or modal):**
- Full container information
- Port mappings with exposure level
- Volume mounts (highlight socket mounts)
- Resource limits
- User/privilege info
- Security issues list with severity per issue
- Remediation suggestions for each issue

**Cluster security summary:**
- Total security issues across all containers
- Count by severity (critical, high, warning, info)
- Quick action: "View all security issues"

### Incident Integration

| Incident | Trigger | Severity |
|----------|---------|----------|
| Docker socket mounted | Any container mounts the Docker socket | critical |
| Privileged container | Any container runs in privileged mode | critical |
| Container publicly exposed | Port published on 0.0.0.0 | high |
| Docker daemon stopped | Docker was running but now stopped | high |
| Container running as root | Container runs as UID 0 | warning |
| No resource limits | Container without memory or CPU limits | warning |
| Container using :latest | Production container using latest tag | warning |
| Exited container | Container in exited state > 1 hour | info |

### Policy Integration

- "If Docker socket is mounted in any container, alert and recommend removal"
- "If a container is publicly exposed, suggest moving behind tunnel"
- "If Docker is not installed but ports > 1024 are listening, suggest container audit"
- "Ensure all containers have resource limits in production"

## Agent Implementation Details

### Package Structure

```
agent/internal/docker/
├── docker.go          # Main integration
├── containers.go      # Container listing and inspection
├── images.go          # Image details
├── networks.go        # Network listing
├── security.go        # Security issue detection
└── command.go         # Container lifecycle commands (restart, stop)
```

### Key Interactions

```bash
# Status
docker info --format '{{.ServerVersion}}'
docker version

# Container listing
docker ps -a --format '{{json .}}'
docker ps --format 'table {{.ID}}\t{{.Image}}\t{{.Ports}}'

# Container inspect (detailed)
docker inspect <container>

# Network listing
docker network ls
docker network inspect <network>

# Container management
docker restart <container>
docker stop <container>
docker pause <container>
```

The agent uses the Docker Engine API (via `docker` CLI or client SDK):

```
GET /v1.47/containers/json
GET /v1.47/containers/{id}/json
GET /v1.47/networks
GET /v1.47/networks/{id}
POST /v1.47/containers/{id}/restart
POST /v1.47/containers/{id}/stop
```

### Go SDK

For more robust integration, use the official Docker Go SDK:

```go
import "github.com/docker/docker/client"

cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
```

### Error Handling

- Docker not installed → show "not installed", suggest package
- Docker daemon not running → report status, commands unavailable
- Permission denied → agent must be in `docker` group or run as root
- Container not found → return clear error
- Docker API version mismatch → use version negotiation
- Socket timeout → set 5s timeout on API calls

## Implementation Order

1. Agent: Docker installed/daemon detection
2. Agent: container listing with basic info
3. Agent: port mapping and exposure detection
4. Agent: Docker socket mount detection
5. Agent: privileged mode and capability detection
6. Agent: resource limit detection
7. Agent: security issue aggregation
8. Agent: status report integration
9. Backend: read-only endpoints
10. Frontend: Docker section on node detail
11. Frontend: containers table and security badges
12. Frontend: container detail with issue list
13. Backend: container lifecycle endpoints
14. Agent: container restart/stop commands
15. Incident generation
16. Policy integration
