import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import {
  Server,
  AlertTriangle,
  CheckCircle,
  RefreshCw,
  Cloud,
  Globe,
  Users,
  Wifi,
  XCircle,
  Activity,
  Clock,
  Share2,
  ShieldCheck,
  ShieldAlert,
  ShieldX,
} from 'lucide-react'
import cytoscape from 'cytoscape'
import { api } from '@/api/client'

interface Stats {
  total_nodes: number
  online_nodes: number
  total_incidents: number
  open_incidents: number
}

interface AppSettings {
  cloudflare_api_token: string
  tailscale_api_key: string
  tailscale_tailnet: string
}

interface CFTunnel {
  id: string
  name: string
  status: string
  connectors?: Array<{ state: string; hostname: string }>
}

interface TSDevice {
  id: string
  name: string
  hostname: string
  os: string
  online: boolean
  addresses: string[]
  clientVersion: string
}

interface SecuritySummary {
  exposed_ssh: number
  docker_exposed: number
  password_auth_ssh: number
  total_incidents: number
  open_incidents: number
  total_nodes: number
  online_nodes: number
  high_severity: number
  crowdsec_nodes: number
  total_decisions: number
  total_alerts: number
  fail2ban_jails_total: number
  fail2ban_bans_total: number
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

// Event type for the timeline
interface TimelineEvent {
  id: string
  node_id: string
  type: string
  payload: string
  created_at: string
}

// Compute a security score from 0-100 based on the summary
function computeSecurityScore(s: SecuritySummary): { score: number; label: string; Icon: any; color: string } {
  let deductions = 0
  deductions += s.exposed_ssh * 15
  deductions += s.password_auth_ssh * 20
  deductions += s.docker_exposed * 25
  deductions += s.open_incidents * 5
  deductions += s.high_severity * 10
  const score = Math.max(0, Math.min(100, 100 - deductions))

  if (score >= 80) return { score, label: 'Good', Icon: ShieldCheck, color: 'text-success' }
  if (score >= 50) return { score, label: 'Warning', Icon: ShieldAlert, color: 'text-warning' }
  return { score, label: 'Critical', Icon: ShieldX, color: 'text-error' }
}

// Mini connectivity graph widget
function DashboardMiniGraph() {
  const containerRef = useRef<HTMLDivElement>(null)

  const { data: graphData } = useQuery({
    queryKey: ['graph'],
    queryFn: () => api.get('api/graph').json(),
    refetchInterval: 60_000,
  })

  useEffect(() => {
    if (!containerRef.current || !graphData) return
    const resp = graphData as { elements: Array<{ data: Record<string, unknown> }> }
    if (!resp.elements || resp.elements.length === 0) return

    const cy = cytoscape({
      container: containerRef.current,
      elements: resp.elements.map((el) => ({
        group: el.data.source && el.data.target ? ('edges' as const) : ('nodes' as const),
        data: el.data,
      })),
      style: [
        { selector: 'node', style: { label: 'data(label)', color: '#9ca3af', 'font-size': '9px', 'text-valign': 'bottom', 'text-outline-width': 2, 'text-outline-color': '#1f2937', 'background-color': '#1e293b', 'border-color': '#475569', 'border-width': 1.5, width: 20, height: 20 }},
        { selector: 'node[type = "internet"]', style: { 'background-color': '#4b5563', 'border-color': '#6b7280', width: 30, height: 30 }},
        { selector: 'node[type = "node"]', style: { 'background-color': '#1e40af', 'border-color': '#3b82f6', width: 25, height: 25 }},
        { selector: 'node[type = "service"][status = "public"]', style: { 'border-color': '#ef4444', 'border-width': 3 }},
        { selector: 'edge', style: { width: 1, 'line-color': '#374151', 'target-arrow-shape': 'triangle', 'arrow-scale': 0.8, 'curve-style': 'bezier' }},
      ],
      layout: { name: 'breadthfirst', directed: true, spacingFactor: 1.0, animate: true, animationDuration: 300 },
      userZoomingEnabled: false, userPanningEnabled: false, autolock: true,
    } as any)

    return () => { cy.destroy() }
  }, [graphData])

  if (!graphData) return null

  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center gap-2">
        <Share2 className="w-4 h-4 text-base-content/40" />
        <h2 className="font-semibold text-sm">Connectivity Map</h2>
      </div>
      <div ref={containerRef} className="w-full h-48" />
    </div>
  )
}

export function DashboardPage() {
  const { data: stats, isLoading } = useQuery<Stats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('api/dashboard/stats').json(),
    refetchInterval: 30_000,
  })

  const { data: settings } = useQuery<AppSettings>({
    queryKey: ['settings'],
    queryFn: () => api.get('api/settings').json(),
  })

  const cfConfigured = !!settings?.cloudflare_api_token

  const { data: cfAccounts } = useQuery<any[]>({
    queryKey: ['cloudflare-accounts'],
    queryFn: () => api.get('api/cloudflare/accounts').json(),
    enabled: cfConfigured,
    retry: false,
  })

  const cfAccountID = cfAccounts?.[0]?.id

  const { data: cfTunnels } = useQuery<CFTunnel[]>({
    queryKey: ['cloudflare-tunnels', cfAccountID],
    queryFn: () => api.get(`api/cloudflare/tunnels?account_id=${cfAccountID}`).json(),
    enabled: !!cfAccountID,
    refetchInterval: 60_000,
    retry: false,
  })

  const tsConfigured = !!settings?.tailscale_api_key && !!settings?.tailscale_tailnet
  const { data: secSummary } = useQuery<SecuritySummary>({
    queryKey: ['security-summary'],
    queryFn: () => api.get('api/dashboard/security-summary').json(),
    refetchInterval: 30_000,
  })

  const { data: tsDevices } = useQuery<TSDevice[]>({
    queryKey: ['tailscale-devices'],
    queryFn: () => api.get('api/tailscale/devices').json(),
    enabled: tsConfigured,
    refetchInterval: 60_000,
    retry: false,
  })

  // Timeline events
  const { data: timelineEvents } = useQuery<TimelineEvent[]>({
    queryKey: ['events-history'],
    queryFn: () => api.get('api/events/history').json(),
    refetchInterval: 30_000,
  })

  // Security events for the timeline (filter security-relevant types)
  const securityEvents = (timelineEvents || []).filter((e) =>
    ['auth_failure', 'ssh_brute_force', 'ssh_publicly_exposed', 'ssh_password_auth_enabled',
     'docker_socket_exposed', 'agent_report', 'attack_detected'].includes(e.type)
  ).slice(0, 10)
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  const cards = [
    { title: 'Total Nodes', value: stats?.total_nodes ?? 0, icon: Server, color: 'text-primary', bg: 'bg-primary/10' },
    { title: 'Online Nodes', value: stats?.online_nodes ?? 0, icon: CheckCircle, color: 'text-success', bg: 'bg-success/10' },
    { title: 'Total Incidents', value: stats?.total_incidents ?? 0, icon: AlertTriangle, color: 'text-warning', bg: 'bg-warning/10' },
    { title: 'Open Incidents', value: stats?.open_incidents ?? 0, icon: RefreshCw, color: 'text-error', bg: 'bg-error/10' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-base-content/60 text-sm mt-1">Overview of your infrastructure security</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {cards.map((card) => {
          const Icon = card.icon
          return (
            <div key={card.title} className="stat bg-base-100 rounded-xl shadow-sm border border-base-content/5">
              <div className="flex items-center gap-3">
                <div className={`p-2.5 rounded-lg ${card.bg}`}>
                  <Icon className={`w-6 h-6 ${card.color}`} />
                </div>
                <div>
                  <div className="stat-title text-xs">{card.title}</div>
                  <div className="stat-value text-2xl">{card.value}</div>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Security Summary Row */}
      {secSummary && (
      <>
      {/* Security Score */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-base-100 rounded-xl p-5 border border-base-content/5 flex items-center gap-4">
          {(() => {
            const sc = computeSecurityScore(secSummary)
            const Icon = sc.Icon
            return (
              <>
                <div className={`p-3 rounded-full ${sc.color.replace('text', 'bg')}/10`}>
                  <Icon className={`w-8 h-8 ${sc.color}`} />
                </div>
                <div className="flex-1">
                  <p className="text-xs text-base-content/40 mb-1">Security Score</p>
                  <div className="flex items-baseline gap-2">
                    <span className={`text-3xl font-bold ${sc.color}`}>{sc.score}</span>
                    <span className={`text-sm font-medium ${sc.color}`}>/ 100</span>
                  </div>
                  <div className="w-full bg-base-200 rounded-full h-2 mt-2">
                    <div
                      className={`h-2 rounded-full transition-all duration-500 ${
                        sc.score >= 80 ? 'bg-success' : sc.score >= 50 ? 'bg-warning' : 'bg-error'
                      }`}
                      style={{ width: `${sc.score}%` }}
                    />
                  </div>
                  <p className="text-xs text-base-content/50 mt-1">
                    {sc.label} &middot; {secSummary.online_nodes}/{secSummary.total_nodes} nodes online
                  </p>
                </div>
              </>
            )
          })()}
        </div>

        {/* Attack Timeline */}
        <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
          <div className="px-5 py-3 border-b border-base-content/5 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-base-content/40" />
              <h2 className="font-semibold text-sm">Recent Activity</h2>
            </div>
            <span className="text-xs text-base-content/30">{securityEvents.length} events</span>
          </div>
          {securityEvents.length === 0 ? (
            <div className="p-5 text-center">
              <Clock className="w-6 h-6 mx-auto text-base-content/20 mb-1" />
              <p className="text-xs text-base-content/50">No recent security events</p>
            </div>
          ) : (
            <div className="divide-y divide-base-content/5 max-h-48 overflow-y-auto">
              {securityEvents.map((ev) => (
                <div key={ev.id} className="px-5 py-2.5 flex items-center gap-3 text-sm">
                  <span className={`w-2 h-2 rounded-full shrink-0 ${
                    ev.type.includes('brute') || ev.type.includes('attack') ? 'bg-error' : 'bg-warning'
                  }`} />
                  <span className="text-xs text-base-content/40 w-16 shrink-0 font-mono">
                    {new Date(ev.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                  <span className="text-xs flex-1 truncate">
                    {ev.type.replace(/_/g, ' ')}
                  </span>
                  <span className="text-xs text-base-content/30 font-mono truncate max-w-[80px]">
                    {ev.node_id?.slice(0, 8)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Dashboard Mini Graph */}
      <DashboardMiniGraph />

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
            <p className="text-xs text-base-content/40 mb-1">Exposed Services</p>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>SSH Public</span>
                <span className={secSummary.exposed_ssh > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.exposed_ssh}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>SSH Password Auth</span>
                <span className={secSummary.password_auth_ssh > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.password_auth_ssh}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Docker Socket Exposed</span>
                <span className={secSummary.docker_exposed > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.docker_exposed}</span>
              </div>
            </div>
          </div>

          <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
            <p className="text-xs text-base-content/40 mb-1">CrowdSec</p>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>Nodes</span>
                <span>{secSummary.crowdsec_nodes} / {secSummary.online_nodes}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Active Decisions</span>
                <span className={secSummary.total_decisions > 0 ? 'text-warning' : ''}>{secSummary.total_decisions}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Alerts (1h)</span>
                <span className={secSummary.total_alerts > 0 ? 'text-error font-bold' : ''}>{secSummary.total_alerts}</span>
              </div>
            </div>
          </div>

          <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
            <p className="text-xs text-base-content/40 mb-1">Fail2Ban</p>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>Total Jails</span>
                <span>{secSummary.fail2ban_jails_total}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Currently Banned</span>
                <span className={secSummary.fail2ban_bans_total > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.fail2ban_bans_total}</span>
              </div>
            </div>
          </div>

          <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
            <p className="text-xs text-base-content/40 mb-1">Incidents</p>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>Open</span>
                <span className={secSummary.open_incidents > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.open_incidents}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>High/Critical</span>
                <span className={secSummary.high_severity > 0 ? 'text-error font-bold' : 'text-success'}>{secSummary.high_severity}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Total</span>
                <span>{secSummary.total_incidents}</span>
              </div>
            </div>
          </div>
        </div>
      </>
      )}

      {/* Cloudflare Tunnel Widget */}
      <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
        <div className="px-5 py-4 border-b border-base-content/5 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Cloud className="w-5 h-5 text-base-content/40" />
            <h2 className="font-semibold">Cloudflare Tunnels</h2>
          </div>
          {!cfConfigured && (
            <span className="text-xs text-base-content/30">Not configured</span>
          )}
        </div>
        {!cfConfigured ? (
          <div className="p-6 text-center">
            <Globe className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
            <p className="text-sm text-base-content/50">Configure your Cloudflare API token in Settings to see tunnels.</p>
          </div>
        ) : !cfTunnels ? (
          <div className="p-6 text-center">
            <span className="loading loading-spinner loading-sm" />
          </div>
        ) : cfTunnels.length === 0 ? (
          <div className="p-6 text-center">
            <p className="text-sm text-base-content/50">No tunnels found for account {cfAccountID?.slice(0,8)}...</p>
          </div>
        ) : (
          <div className="divide-y divide-base-content/5">
            {cfTunnels.map((t) => (
              <div key={t.id} className="px-5 py-3 flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">{t.name}</p>
                  <p className="text-xs text-base-content/40">ID: {t.id.slice(0, 8)}...</p>
                </div>
                <div className="flex items-center gap-2">
                  {t.connectors && t.connectors.length > 0 ? (
                    <InfoBadge
                      good={t.connectors.some((c) => c.state === 'healthy')}
                      label={t.connectors.some((c) => c.state === 'healthy') ? 'Healthy' : 'Degraded'}
                    />
                  ) : (
                    <span className="badge badge-ghost badge-sm">{t.status}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
        {cfConfigured && cfAccounts && cfAccounts.length > 0 && (
          <div className="px-5 py-2 border-t border-base-content/5 text-xs text-base-content/30">
            Account: {cfAccounts[0].name} ({cfAccounts[0].id.slice(0,12)}...)
          </div>
        )}
      </div>

      {/* Tailscale Widget */}
      <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
        <div className="px-5 py-4 border-b border-base-content/5 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-base-content/40" />
            <h2 className="font-semibold">Tailscale Devices</h2>
          </div>
          {!tsConfigured && (
            <span className="text-xs text-base-content/30">Not configured</span>
          )}
        </div>
        {!tsConfigured ? (
          <div className="p-6 text-center">
            <Wifi className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
            <p className="text-sm text-base-content/50">Configure your Tailscale API key in Settings to see devices.</p>
          </div>
        ) : !tsDevices ? (
          <div className="p-6 text-center">
            <span className="loading loading-spinner loading-sm" />
          </div>
        ) : tsDevices.length === 0 ? (
          <div className="p-6 text-center">
            <p className="text-sm text-base-content/50">No devices found in tailnet.</p>
          </div>
        ) : (
          <div className="divide-y divide-base-content/5">
            <div className="px-5 py-2 text-xs text-base-content/30 flex justify-between">
              <span>{tsDevices.length} device{tsDevices.length !== 1 ? 's' : ''}</span>
              <span>{tsDevices.filter((d) => d.online).length} online</span>
            </div>
            {tsDevices.slice(0, 10).map((d) => (
              <div key={d.id} className="px-5 py-3 flex items-center justify-between">
                <div className="flex items-center gap-3 min-w-0">
                  <div>
                    <p className="text-sm font-medium">{d.name || d.hostname}</p>
                    <p className="text-xs text-base-content/40">{d.os} &middot; {d.addresses?.[0]}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <InfoBadge good={d.online} label={d.online ? 'Online' : 'Offline'} />
                  {d.clientVersion && (
                    <span className="text-xs text-base-content/30">v{d.clientVersion}</span>
                  )}
                </div>
              </div>
            ))}
            {tsDevices.length > 10 && (
              <div className="px-5 py-2 text-xs text-center text-base-content/30">
                + {tsDevices.length - 10} more devices
              </div>
            )}
          </div>
        )}
      </div>

      {/* Real-time Events placeholder */}
      <div className="bg-base-100 rounded-xl p-6 border border-base-content/5">
        <h2 className="text-lg font-semibold mb-2">Real-time Events</h2>
        <p className="text-sm text-base-content/60">
          Connect agents to see live security events from your infrastructure.
        </p>
      </div>
    </div>
  )
}
