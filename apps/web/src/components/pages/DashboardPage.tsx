import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import {
  StatsCards,
  SecurityScoreCard,
  AttackTimeline,
  ConnectivityStats,
  ConnectivityMap,
  CloudflareWidget,
  TailscaleWidget,
} from '@/components/dashboard'
import {
  SECURITY_EVENT_TYPES,
  type AppSettings,
  type CFTunnel,
  type SecuritySummary,
  type Stats,
  type TimelineEvent,
  type TSDevice,
} from '@/components/dashboard/common'

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

  const { data: cfTunnels, isFetching: cfLoading } = useQuery<CFTunnel[]>({
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

  const { data: tsDevices, isFetching: tsLoading } = useQuery<TSDevice[]>({
    queryKey: ['tailscale-devices'],
    queryFn: () => api.get('api/tailscale/devices').json(),
    enabled: tsConfigured,
    refetchInterval: 60_000,
    retry: false,
  })

  const { data: timelineEvents } = useQuery<TimelineEvent[]>({
    queryKey: ['events-history'],
    queryFn: () => api.get('api/events/history').json(),
    refetchInterval: 30_000,
  })

  const securityEvents = (timelineEvents || [])
    .filter((e) => SECURITY_EVENT_TYPES.includes(e.type))
    .slice(0, 10)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-base-content/60 text-sm mt-1">Overview of your infrastructure security</p>
      </div>

      <StatsCards stats={stats} />

      {secSummary && (
        <>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <SecurityScoreCard summary={secSummary} />
            <AttackTimeline events={securityEvents} />
          </div>

          <ConnectivityMap />

          <ConnectivityStats summary={secSummary} />
        </>
      )}

      <CloudflareWidget
        configured={cfConfigured}
        accountName={cfAccounts?.[0]?.name}
        accountID={cfAccountID}
        tunnels={cfTunnels}
        loading={cfLoading}
      />

      <TailscaleWidget configured={tsConfigured} devices={tsDevices} loading={tsLoading} />

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
