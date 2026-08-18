import { useQuery } from '@tanstack/react-query'
import { useMutation } from '@tanstack/react-query'
import { useParams, Link } from '@tanstack/react-router'
import { useEffect, useRef } from 'react'
import cytoscape from 'cytoscape'
import {
  ArrowLeft,
  Server,
  Wifi,
  XCircle,
  Terminal,
  Clock,
  Shield,
  Lock,
  AlertTriangle,
  CheckCircle,
  Database,
  Globe,
  Activity,
  Users,
  FileWarning,
  Bug,
  Fingerprint,
  Ban,
  ShieldOff,
  Unlock,
  UserCheck,
  Eye,
} from 'lucide-react'
import { api } from '@/api/client'

interface Node {
  id: string
  name: string
  hostname: string
  ip: string
  os: string
  status: string
  labels: string[]
  last_seen: string
  created_at: string
}

interface Incident {
  id: string
  node_id: string
  severity: string
  title: string
  message: string
  status: string
  created_at: string
}

// Agent report types matching proto
interface AgentReport {
  os: string
  kernel: string
  uptime_seconds: number
  ports?: { listening: ListeningPort[] }
  docker?: DockerStatus
  firewall?: FirewallStatus
  ssh?: SSHStatus
  crowdsec?: CrowdSecStatus
  fail2ban?: Fail2BanStatus
  tailscale?: TailscaleStatus
  cloudflare_tunnel?: CloudflareStatus
  auth_log?: AuthLogStatus
  system?: SystemHealth
  fim?: { mode: string; files?: Array<{ path: string; hash: string }> }
  packages?: { installed?: Array<{ name: string; version: string }>; security_updates_pending: number }
}

interface VulnerablePackage {
  name: string
  version: string
  cve_count: number
  top_cve?: string
  summary?: string
}

interface FIMFile {
  path: string
  changed_at?: string
}

interface ListeningPort {
  port: number
  proto: string
  process: string
  exposed: boolean
}

interface DockerStatus {
  total_containers: number
  running_containers: Array<{ id: string; name: string; image: string }>
  socket_exposed: boolean
}

interface FirewallStatus {
  active_backend: string
  ufw?: { active: boolean; rules: Array<{ action: string; port: number; proto: string }> }
  nftables?: { active: boolean }
}

interface SSHStatus {
  port: number
  password_auth: boolean
  root_login: string
  pubkey_auth: boolean
  publicly_exposed: boolean
  max_auth_tries: number
}

interface CrowdSecStatus {
  installed: boolean
  running: boolean
  active_decisions: number
  alerts_last_hour: number
  bouncers: string[]
}

interface Fail2BanStatus {
  installed: boolean
  running: boolean
  jails: Array<{ name: string; active: boolean; currently_banned: number }>
}

interface TailscaleStatus {
  installed: boolean
  running: boolean
  node_name: string
  node_ip: string
  online: boolean
  peers_count: number
 acl_version?: string
}

interface CloudflareStatus {
  installed: boolean
  running: boolean
  tunnels: Array<{ id: string; name: string; status: string }>
}

interface AuthLogStatus {
  failed_ssh_last_hour: number
  failed_root_last_hour: number
  failed_ssh_top_ips?: Array<{ ip: string; attempts: number }>
  sudo_usage_last_hour: number
}

interface SystemHealth {
  cpu_percent: number
  memory_percent: number
  disk_percent: number
}

function InfoBadge({ good, label }: { good: boolean; label: string }) {
  return (
    <span className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full ${
      good ? 'bg-success/10 text-success' : 'bg-error/10 text-error'
    }`}>
      {good ? <CheckCircle className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
      {label}
    </span>
  )
}

function Section({ title, icon: Icon, children }: { title: string; icon: any; children: React.ReactNode }) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center gap-2">
        <Icon className="w-4 h-4 text-base-content/40" />
        <h2 className="font-semibold text-sm">{title}</h2>
      </div>
      <div className="p-5">{children}</div>
    </div>
  )
}

function ActionButton({ label, icon: Icon, onClick }: { label: string; icon: any; onClick: () => void }) {
  return (
    <button className="btn btn-ghost btn-sm gap-2" onClick={onClick}>
      <Icon className="w-4 h-4" />
      {label}
    </button>
  )
}

function StatusRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex justify-between items-center py-1.5 text-sm">
      <span className="text-base-content/60">{label}</span>
      <span className="font-medium">{children}</span>
    </div>
  )
}

function NodeMiniGraph({ nodeId }: { nodeId: string }) {
  const containerRef = useRef<HTMLDivElement>(null)

  const { data: graphData } = useQuery({
    queryKey: ['node-graph', nodeId],
    queryFn: () => api.get(`api/graph/nodes/${nodeId}`).json(),
    retry: false,
  })

  useEffect(() => {
    if (!containerRef.current || !graphData) return
    const resp = graphData as { elements: Array<{ data: Record<string, unknown> }> }
    if (!resp.elements || resp.elements.length === 0) return

    const cy = cytoscape({
      container: containerRef.current,
      elements: resp.elements.map((el) => ({
        group: el.data.source && el.data.target ? ('edges' as const) : ('nodes' as const),
        data: { ...el.data, id: el.data.id as string },
      })),
      style: [
        { selector: 'node', style: {
            label: 'data(label)', color: '#9ca3af', 'font-size': '10px',
            'text-valign': 'bottom', 'text-outline-width': 2,
            'text-outline-color': '#1f2937', 'background-color': '#1e293b',
            'border-color': '#475569', 'border-width': 2, width: 30, height: 30,
        }},
        { selector: 'node[type = "internet"]', style: { 'background-color': '#4b5563', 'border-color': '#6b7280', width: 40, height: 40 }},
        { selector: 'node[type = "node"]', style: { 'background-color': '#1e40af', 'border-color': '#3b82f6' }},
        { selector: 'node[type = "service"]', style: { 'background-color': '#92400e', 'border-color': '#f59e0b', shape: 'ellipse' }},
        { selector: 'node[type = "container"]', style: { 'background-color': '#065f46', 'border-color': '#10b981', shape: 'rectangle' }},
        { selector: 'node[type = "firewall"]', style: { 'background-color': '#0e7490', 'border-color': '#06b6d4' }},
        { selector: 'node[type = "incident"]', style: { 'background-color': '#991b1b', 'border-color': '#ef4444', 'background-opacity': 0.4 }},
        { selector: 'edge', style: {
            width: 1.5, 'line-color': '#4b5563', 'target-arrow-shape': 'triangle',
            'arrow-scale': 1, 'curve-style': 'bezier',
        }},
      ],
      layout: { name: 'breadthfirst', directed: true, spacingFactor: 1.2, animate: true, animationDuration: 300 },
      userZoomingEnabled: false, userPanningEnabled: false, autolock: true,
    } as any)

    return () => { cy.destroy() }
  }, [graphData])

  if (!graphData) return null

  // Determine mini-graph height based on data
  const elemCount = (graphData as any).elements?.length ?? 0
  const height = elemCount > 8 ? 'h-64' : 'h-48'

  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center gap-2">
        <Activity className="w-4 h-4 text-base-content/40" />
        <h2 className="font-semibold text-sm">Network Graph</h2>
    </div>
      <div ref={containerRef} className={`w-full ${height}`} />
    </div>
  )
}

export function NodeDetailPage() {
  const { nodeId } = useParams({ from: '/nodes/$nodeId' })

  const { data: node, isLoading: nodeLoading } = useQuery<Node>({
    queryKey: ['node', nodeId],
    queryFn: () => api.get(`api/nodes/${nodeId}`).json(),
  })

  const { data: report, isLoading: reportLoading } = useQuery<AgentReport>({
    queryKey: ['node-report', nodeId],
    queryFn: () => api.get(`api/nodes/${nodeId}/report`).json(),
    retry: false,
  })

  const { data: incidents } = useQuery<Incident[]>({
    queryKey: ['node-incidents', nodeId],
    queryFn: () =>
      api.get('api/incidents').json().then((all) =>
        (all as Incident[]).filter((i) => i.node_id === nodeId)
      ),
  })

  const { data: vulnerabilities } = useQuery<VulnerablePackage[]>({
    queryKey: ['node-vulns', nodeId],
    queryFn: () => api.get(`api/nodes/${nodeId}/vulnerabilities`).json(),
  })

  const { data: fimChanges } = useQuery<FIMFile[]>({
    queryKey: ['node-fim', nodeId],
    queryFn: () => api.get(`api/nodes/${nodeId}/fim`).json(),
  })

  if (nodeLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  if (!node) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <Server className="w-12 h-12 mx-auto text-base-content/30 mb-3" />
          <p className="font-medium">Node not found</p>
          <Link to="/nodes" className="btn btn-ghost btn-sm mt-2">Back to nodes</Link>
        </div>
      </div>
    )
  }

  const severityBadge = (sev: string) => {
    const m: Record<string, string> = { critical: 'badge-error', high: 'badge-warning', medium: 'badge-info', low: 'badge-ghost' }
    return `badge ${m[sev] ?? 'badge-ghost'}`
  }

  const createAction = useMutation({
    mutationFn: (data: { nodeId: string; actionType: string; params: Record<string, unknown> }) =>
      api.post('api/actions', { json: { node_id: data.nodeId, action_type: data.actionType, params: data.params } }).json(),
  })

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-base-content/40">
        <Link to="/nodes" className="hover:text-base-content/70 transition-colors">Nodes</Link>
        <span>/</span>
        <span className="text-base-content/70">{node.name}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Link to="/nodes" className="btn btn-ghost btn-sm btn-square">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold">{node.name}</h1>
              {report && report.system && (
                <span className="text-xs text-base-content/40">
                  CPU {report.system.cpu_percent.toFixed(0)}% &middot; Mem {report.system.memory_percent.toFixed(0)}% &middot; Disk {report.system.disk_percent.toFixed(0)}%
                </span>
              )}
            </div>
            <p className="text-sm text-base-content/60">{node.hostname}</p>
          </div>
        </div>
        <span className={`inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-full ${
          node.status === 'online' ? 'bg-success/10 text-success' : 'bg-base-200 text-base-content/50'
        }`}>
          {node.status === 'online' ? <Wifi className="w-3.5 h-3.5" /> : <XCircle className="w-3.5 h-3.5" />}
          {node.status}
        </span>
      </div>

      {/* Overview + Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Section title="Overview" icon={Server}>
          <StatusRow label="IP Address">{node.ip}</StatusRow>
          <StatusRow label="OS">{node.os || 'Unknown'}</StatusRow>
          {report && <StatusRow label="Kernel">{report.kernel || 'Unknown'}</StatusRow>}
          {report && report.uptime_seconds > 0 && (
            <StatusRow label="Uptime">
              {Math.floor(report.uptime_seconds / 86400)}d {Math.floor((report.uptime_seconds % 86400) / 3600)}h
            </StatusRow>
          )}
          <StatusRow label="Hostname">{node.hostname}</StatusRow>
          <StatusRow label="Labels">
            {node.labels?.length ? node.labels.map((l) => (
              <span key={l} className="badge badge-sm badge-ghost mr-1">{l}</span>
            )) : 'None'}
          </StatusRow>
        </Section>

        <Section title="Status" icon={Activity}>
          <StatusRow label="Last Seen">
            <span className="flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-base-content/40" />
              {new Date(node.last_seen).toLocaleString()}
            </span>
          </StatusRow>
          <StatusRow label="Registered">{new Date(node.created_at).toLocaleDateString()}</StatusRow>
          <StatusRow label="Status">{node.status}</StatusRow>
          {report && <StatusRow label="OS Release">{report.os}</StatusRow>}
        </Section>
      </div>

      {/* Mini Network Graph */}
      <NodeMiniGraph nodeId={nodeId} />

      {reportLoading && (
        <div className="flex justify-center py-8">
          <span className="loading loading-spinner loading-sm" />
        </div>
      )}

      {/* SSH Status */}
      {report?.ssh && (
        <Section title="SSH Hardening" icon={Lock}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={!report.ssh.publicly_exposed} label={report.ssh.publicly_exposed ? 'Publicly Exposed' : 'Not Exposed'} />
              <InfoBadge good={!report.ssh.password_auth} label={report.ssh.password_auth ? 'Password Auth On' : 'Password Auth Off'} />
              <InfoBadge good={report.ssh.root_login === 'no' || report.ssh.root_login === 'prohibit-password'} label={`Root Login: ${report.ssh.root_login}`} />
              <InfoBadge good={report.ssh.pubkey_auth} label={report.ssh.pubkey_auth ? 'Pubkey Auth' : 'No Pubkey Auth'} />
            </div>
            <StatusRow label="Port">{report.ssh.port}</StatusRow>
            <StatusRow label="Max Auth Tries">{report.ssh.max_auth_tries}</StatusRow>
          </div>
        </Section>
      )}

      {/* Firewall Status */}
      {report?.firewall && (
        <Section title="Firewall" icon={Shield}>
          <div className="space-y-3">
            <StatusRow label="Active Backend">
              <span className="capitalize">{report.firewall.active_backend}</span>
            </StatusRow>
            {report.firewall.ufw && (
              <>
                <StatusRow label="UFW Active">
                  <InfoBadge good={report.firewall.ufw.active} label={report.firewall.ufw.active ? 'Active' : 'Inactive'} />
                </StatusRow>
                {report.firewall.ufw.rules && report.firewall.ufw.rules.length > 0 && (
                  <div>
                    <p className="text-xs text-base-content/40 mb-1">Rules ({report.firewall.ufw.rules.length})</p>
                    <div className="flex flex-wrap gap-1">
                      {report.firewall.ufw.rules.map((r, i) => (
                        <span key={i} className="badge badge-sm badge-ghost">
                          {r.action} {r.port}/{r.proto}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
            {report.firewall.nftables && (
              <StatusRow label="NFTables">
                <InfoBadge good={report.firewall.nftables.active} label={report.firewall.nftables.active ? 'Active' : 'Inactive'} />
              </StatusRow>
            )}
          </div>
        </Section>
      )}

      {/* Docker Status */}
      {report?.docker && (
        <Section title="Docker" icon={Database}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={!report.docker.socket_exposed} label={report.docker.socket_exposed ? 'Socket Exposed!' : 'Socket Secure'} />
            </div>
            <StatusRow label="Total Containers">{report.docker.total_containers}</StatusRow>
            <StatusRow label="Running">{report.docker.running_containers?.length ?? 0}</StatusRow>
            {report.docker.running_containers && report.docker.running_containers.length > 0 && (
              <div>
                <p className="text-xs text-base-content/40 mb-1">Running Containers</p>
                <div className="flex flex-wrap gap-1">
                  {report.docker.running_containers.map((c) => (
                    <span key={c.id} className="badge badge-sm badge-ghost">{c.name}</span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </Section>
      )}

      {/* CrowdSec Status */}
      {report?.crowdsec && (
        <Section title="CrowdSec" icon={AlertTriangle}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={report.crowdsec.installed} label={report.crowdsec.installed ? 'Installed' : 'Not Installed'} />
              <InfoBadge good={report.crowdsec.running} label={report.crowdsec.running ? 'Running' : 'Stopped'} />
            </div>
            <StatusRow label="Active Decisions">{report.crowdsec.active_decisions}</StatusRow>
            <StatusRow label="Alerts (Last Hour)">{report.crowdsec.alerts_last_hour}</StatusRow>
            {report.crowdsec.bouncers && report.crowdsec.bouncers.length > 0 && (
              <StatusRow label="Bouncers">{report.crowdsec.bouncers.join(', ')}</StatusRow>
            )}
          </div>
        </Section>
      )}

      {/* Fail2Ban Status */}
      {report?.fail2ban && (
        <Section title="Fail2Ban" icon={Fingerprint}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={report.fail2ban.installed} label={report.fail2ban.installed ? 'Installed' : 'Not Installed'} />
              <InfoBadge good={report.fail2ban.running} label={report.fail2ban.running ? 'Running' : 'Stopped'} />
            </div>
            {report.fail2ban.jails && report.fail2ban.jails.length > 0 && (
              <div>
                <p className="text-xs text-base-content/40 mb-2">Jails ({report.fail2ban.jails.length})</p>
                {report.fail2ban.jails.map((j) => (
                  <div key={j.name} className="flex justify-between items-center py-1.5 text-sm border-b border-base-content/5 last:border-0">
                    <div>
                      <span className="font-medium">{j.name}</span>
                      <span className={`ml-2 text-xs ${(j as any).currently_banned > 0 ? 'text-error' : 'text-base-content/40'}`}>
                        {(j as any).currently_banned} banned &middot; {(j as any).total_banned || 0} total
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      {(j as any).currently_banned > 0 && (
                        <button
                          className="btn btn-ghost btn-xs text-success"
                          title="Unban all IPs from this jail"
                          onClick={() => {
                            const ip = prompt('IP to unban from ' + j.name + ':')
                            if (ip) createAction.mutate({ nodeId, actionType: 'fail2ban_unban_ip', params: { ip, jail: j.name } })
                          }}
                        >
                          <Unlock className="w-3 h-3" />
                          Unban
                        </button>
                      )}
                      <button
                        className="btn btn-ghost btn-xs text-error"
                        title="Ban an IP in this jail"
                        onClick={() => {
                          const ip = prompt('IP to ban in ' + j.name + ':')
                          if (ip) createAction.mutate({ nodeId, actionType: 'fail2ban_ban_ip', params: { ip, jail: j.name } })
                        }}
                      >
                        <Ban className="w-3 h-3" />
                        Ban
                      </button>
                    </div>
                  </div>
                ))}
                <div className="mt-3 pt-3 border-t border-base-content/5">
                  <p className="text-xs text-base-content/40 mb-2">Whitelist Management</p>
                  <div className="flex gap-2">
                    <button
                      className="btn btn-ghost btn-xs"
                      onClick={() => {
                        const ip = prompt('Add IP to whitelist:')
                        if (ip) createAction.mutate({ nodeId, actionType: 'fail2ban_unban_ip', params: { ip, jail: 'all' } })
                      }}
                    >
                      Add to Whitelist
                    </button>
                  </div>
              </div>
          </div>
            )}
          </div>
        </Section>
      )}

      {/* Tailscale Status */}
      {report?.tailscale && (
        <Section title="Tailscale" icon={Users}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={report.tailscale.installed} label={report.tailscale.installed ? 'Installed' : 'Not Installed'} />
              <InfoBadge good={report.tailscale.running} label={report.tailscale.running ? 'Running' : 'Stopped'} />
              <InfoBadge good={report.tailscale.online} label={report.tailscale.online ? 'Online' : 'Offline'} />
            </div>
            <StatusRow label="Node Name">{report.tailscale.node_name || '—'}</StatusRow>
            <StatusRow label="Node IP">{report.tailscale.node_ip || '—'}</StatusRow>
            <StatusRow label="Peers">{report.tailscale.peers_count}</StatusRow>
            {report.tailscale.acl_version && (
              <StatusRow label="ACL Hash">
                <span className="font-mono text-xs">{report.tailscale.acl_version}</span>
                <span className="tooltip tooltip-right" data-tip="ACL hash changes may indicate configuration drift">
                  <AlertTriangle className="w-3 h-3 text-warning ml-1 inline" />
                </span>
              </StatusRow>
            )}
            {report.tailscale.acl_version && (
              <div className="bg-warning/5 border border-warning/20 rounded-lg p-2 text-xs text-warning">
                ACL configuration is being tracked. Use the Tailscale admin console to review access policies.
          </div>
            )}
          </div>
        </Section>
      )}

      {/* Cloudflare Tunnel Status */}
      {report?.cloudflare_tunnel && (
        <Section title="Cloudflare Tunnel" icon={Globe}>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <InfoBadge good={report.cloudflare_tunnel.installed} label={report.cloudflare_tunnel.installed ? 'Installed' : 'Not Installed'} />
              <InfoBadge good={report.cloudflare_tunnel.running} label={report.cloudflare_tunnel.running ? 'Running' : 'Stopped'} />
            </div>
            {report.cloudflare_tunnel.tunnels && report.cloudflare_tunnel.tunnels.length > 0 && (
              <div>
                {report.cloudflare_tunnel.tunnels.map((t) => (
                  <div key={t.id} className="flex justify-between items-center py-1 text-sm">
                    <span className="font-medium">{t.name || t.id}</span>
                    <span className={t.status === 'running' ? 'text-success' : 'text-warning'}>{t.status}</span>
                  </div>
                ))}
              </div>
            )}
            {report.cloudflare_tunnel.tunnels && report.cloudflare_tunnel.tunnels.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                <button
                  className="btn btn-ghost btn-xs gap-1 text-info"
                  onClick={() => {
                    const hostname = prompt('Hostname to expose privately:')
                    if (hostname) createAction.mutate({ nodeId, actionType: 'cloudflare_expose_privately', params: { hostname } })
                  }}
                >
                  <Globe className="w-3 h-3" />
                  Expose Privately
                </button>
              </div>
            )}
              </div>
        </Section>
      )}

      {/* Auth Log */}
      {report?.auth_log && (
        <Section title="Authentication Log" icon={Activity}>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-base-200/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold">{report.auth_log.failed_ssh_last_hour}</p>
                <p className="text-xs text-base-content/40">Failed SSH (1h)</p>
              </div>
              <div className="bg-base-200/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold">{report.auth_log.failed_root_last_hour}</p>
                <p className="text-xs text-base-content/40">Root Attempts (1h)</p>
              </div>
            </div>
            <StatusRow label="sudo Usage (1h)">{report.auth_log.sudo_usage_last_hour}</StatusRow>
            {report.auth_log.failed_ssh_top_ips && report.auth_log.failed_ssh_top_ips.length > 0 && (
              <div>
                <p className="text-xs text-base-content/40 mb-1">Top Source IPs</p>
                {report.auth_log.failed_ssh_top_ips.slice(0, 5).map((ip) => (
                  <div key={ip.ip} className="flex justify-between text-xs py-0.5 items-center">
                    <div className="flex items-center gap-2 min-w-0">
                    <span className="font-mono">{ip.ip}</span>
                      {ip.attempts >= 10 && (
                        <span className="badge badge-error badge-xs whitespace-nowrap">Suspicious</span>
                      )}
                      {ip.attempts >= 5 && ip.attempts < 10 && (
                        <span className="badge badge-warning badge-xs whitespace-nowrap">Warning</span>
                      )}
                  </div>
                    <span className={`text-base-content/40 ${ip.attempts >= 10 ? 'text-error font-bold' : ''}`}>{ip.attempts} attempts</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Section>
      )}

      {/* Listening Ports */}
      {report?.ports?.listening && report.ports.listening.length > 0 && (
        <Section title="Listening Ports" icon={Globe}>
          <div className="overflow-x-auto">
            <table className="table table-xs">
              <thead>
                <tr>
                  <th>Port</th>
                  <th>Protocol</th>
                  <th>Process</th>
                  <th>Exposure</th>
                </tr>
              </thead>
              <tbody>
                {report.ports.listening.map((p, i) => (
                  <tr key={i}>
                    <td className="font-mono">{p.port}</td>
                    <td>{p.proto}</td>
                    <td className="text-base-content/60">{p.process}</td>
                    <td>
                      <InfoBadge good={!p.exposed} label={p.exposed ? 'Public' : 'Local'} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      {/* Quick Remediation Actions */}
      {report && node.status === 'online' && (
        <Section title="Quick Actions" icon={Shield}>
          <div className="flex flex-wrap gap-2">
            <ActionButton
              label="Unban IP from SSH"
              icon={Ban}
              onClick={() => {
                const ip = prompt('IP address to unban:')
                if (ip) createAction.mutate({ nodeId, actionType: 'fail2ban_unban_ip', params: { ip, jail: 'sshd' } })
              }}
            />
            <ActionButton
              label="Block IP in Fail2Ban"
              icon={ShieldOff}
              onClick={() => {
                const ip = prompt('IP address to block:')
                if (ip) createAction.mutate({ nodeId, actionType: 'fail2ban_ban_ip', params: { ip, jail: 'sshd' } })
              }}
            />
            <ActionButton
              label="Deny Port (UFW)"
              icon={ShieldOff}
              onClick={() => {
                const port = prompt('Port number to deny:')
                if (port) createAction.mutate({ nodeId, actionType: 'ufw_deny_port', params: { port: parseInt(port), protocol: 'tcp' } })
              }}
            />
            {report?.tailscale?.running && (
              <>
            <ActionButton
                  label="Restrict to Team"
                  icon={Users}
              onClick={() => {
                    const tag = prompt('Tailscale tag to restrict (e.g., prod, internal):')
                    if (tag) createAction.mutate({ nodeId, actionType: 'tailscale_restrict', params: { tag } })
              }}
            />
            <ActionButton
                  label="Require MFA"
                  icon={UserCheck}
              onClick={() => {
                    createAction.mutate({ nodeId, actionType: 'tailscale_require_mfa', params: {} })
              }}
            />
              </>
            )}
            {report?.cloudflare_tunnel?.running && (
            <ActionButton
                label="Expose Privately"
                icon={Eye}
              onClick={() => {
                  const hostname = prompt('Hostname to expose privately via Cloudflare:')
                  if (hostname) createAction.mutate({ nodeId, actionType: 'cloudflare_expose_privately', params: { hostname } })
              }}
            />
            )}
          </div>
        </Section>
      )}

      {/* File Integrity */}
      <Section title="File Integrity" icon={FileWarning}>
        {!report?.fim && (!fimChanges || fimChanges.length === 0) ? (
          <div className="text-center py-4 text-sm text-base-content/50">
            No monitored-file changes detected.
          </div>
        ) : (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-base-200/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold">{report?.fim?.files?.length ?? 0}</p>
                <p className="text-xs text-base-content/40">Monitored Files</p>
              </div>
              <div className={`bg-base-200/50 rounded-lg p-3 text-center ${(fimChanges?.length ?? 0) > 0 ? 'ring-1 ring-warning' : ''}`}>
                <p className={`text-2xl font-bold ${(fimChanges?.length ?? 0) > 0 ? 'text-warning' : ''}`}>
                  {fimChanges?.length ?? 0}
                </p>
                <p className="text-xs text-base-content/40">Changed</p>
              </div>
            </div>
            {(fimChanges ?? []).length > 0 && (
              <div>
                <p className="text-xs text-base-content/40 mb-1">Changed files</p>
                <div className="space-y-1">
                  {fimChanges!.map((f) => (
                    <div key={f.path} className="flex items-center justify-between text-xs py-1">
                      <span className="font-mono truncate">{f.path}</span>
                      <span className="text-warning shrink-0 ml-2">
                        {f.changed_at ? new Date(f.changed_at).toLocaleString() : 'changed'}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </Section>

      {/* Patch Status / Vulnerabilities */}
      <Section title="Patch & Vulnerabilities" icon={Bug}>
        {(!report?.packages && (!vulnerabilities || vulnerabilities.length === 0)) ? (
          <div className="text-center py-4 text-sm text-base-content/50">
            No package or vulnerability data available yet.
          </div>
        ) : (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-base-200/50 rounded-lg p-3 text-center">
                <p className="text-2xl font-bold">{report?.packages?.installed?.length ?? 0}</p>
                <p className="text-xs text-base-content/40">Packages</p>
              </div>
              <div className={`bg-base-200/50 rounded-lg p-3 text-center ${(report?.packages?.security_updates_pending ?? 0) > 0 ? 'ring-1 ring-warning' : ''}`}>
                <p className={`text-2xl font-bold ${(report?.packages?.security_updates_pending ?? 0) > 0 ? 'text-warning' : ''}`}>
                  {report?.packages?.security_updates_pending ?? 0}
                </p>
                <p className="text-xs text-base-content/40">Security Updates Pending</p>
              </div>
            </div>
            {(vulnerabilities ?? []).length > 0 && (
              <div>
                <p className="text-xs text-base-content/40 mb-1">Known CVEs</p>
                <div className="space-y-1">
                  {vulnerabilities!.map((v) => (
                    <div key={`${v.name}@${v.version}`} className="flex items-center justify-between text-xs py-1">
                      <span className="font-mono truncate">{v.name} {v.version}</span>
                      <span className="badge badge-error badge-xs shrink-0 ml-2">{v.cve_count} CVE</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </Section>

      {/* Incidents */}
      <Section title="Incidents" icon={AlertTriangle}>
        {(!incidents || incidents.length === 0) ? (
          <div className="text-center py-4">
            <Terminal className="w-8 h-8 mx-auto text-base-content/30 mb-2" />
            <p className="text-sm text-base-content/50">No incidents for this node</p>
          </div>
        ) : (
          <div className="divide-y divide-base-content/5 -mx-5 -mb-5">
            {incidents.map((inc) => (
              <div key={inc.id} className="px-5 py-3 flex items-center justify-between hover:bg-base-200/30 transition-colors">
                <div className="flex items-center gap-3 min-w-0">
                  <span className={severityBadge(inc.severity)}>{inc.severity}</span>
                  <span className="text-sm font-medium truncate">{inc.title}</span>
                </div>
                <div className="flex items-center gap-3 shrink-0">
                  <span className={`text-xs ${inc.status === 'open' ? 'text-error' : 'text-success'}`}>{inc.status}</span>
                  <Link to="/incidents/$incidentId" params={{ incidentId: inc.id }} className="btn btn-ghost btn-xs">View</Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* No report state */}
      {!reportLoading && !report && (
        <div className="bg-base-100 rounded-xl p-8 text-center border border-base-content/5">
          <Server className="w-10 h-10 mx-auto text-base-content/30 mb-2" />
          <p className="text-sm text-base-content/50">No agent report available yet. Reports appear after the agent checks in.</p>
        </div>
      )}
    </div>
  )
}
