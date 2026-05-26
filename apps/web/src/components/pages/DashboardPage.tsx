import { useQuery } from '@tanstack/react-query'
import {
  Server,
  AlertTriangle,
  CheckCircle,
  RefreshCw,
} from 'lucide-react'
import { api } from '@/api/client'

interface Stats {
  total_nodes: number
  online_nodes: number
  total_incidents: number
  open_incidents: number
}

export function DashboardPage() {
  const { data: stats, isLoading } = useQuery<Stats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('api/dashboard/stats').json(),
    refetchInterval: 30_000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  const cards = [
    {
      title: 'Total Nodes',
      value: stats?.total_nodes ?? 0,
      icon: Server,
      color: 'text-primary',
      bg: 'bg-primary/10',
    },
    {
      title: 'Online Nodes',
      value: stats?.online_nodes ?? 0,
      icon: CheckCircle,
      color: 'text-success',
      bg: 'bg-success/10',
    },
    {
      title: 'Total Incidents',
      value: stats?.total_incidents ?? 0,
      icon: AlertTriangle,
      color: 'text-warning',
      bg: 'bg-warning/10',
    },
    {
      title: 'Open Incidents',
      value: stats?.open_incidents ?? 0,
      icon: RefreshCw,
      color: 'text-error',
      bg: 'bg-error/10',
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-base-content/60 text-sm mt-1">Overview of your infrastructure security</p>
      </div>

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

      <div className="bg-base-100 rounded-xl p-6 border border-base-content/5">
        <h2 className="text-lg font-semibold mb-2">Real-time Events</h2>
        <p className="text-sm text-base-content/60">
          Connect agents to see live security events from your infrastructure.
        </p>
      </div>
    </div>
  )
}
