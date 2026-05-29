import { useQuery } from '@tanstack/react-query'
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
} from 'lucide-react'
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
