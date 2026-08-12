import { Server, AlertTriangle, CheckCircle, RefreshCw } from 'lucide-react'
import type { Stats } from './common'

export function StatsCards({ stats }: { stats?: Stats }) {
  const cards = [
    { title: 'Total Nodes', value: stats?.total_nodes ?? 0, icon: Server, color: 'text-primary', bg: 'bg-primary/10' },
    { title: 'Online Nodes', value: stats?.online_nodes ?? 0, icon: CheckCircle, color: 'text-success', bg: 'bg-success/10' },
    { title: 'Total Incidents', value: stats?.total_incidents ?? 0, icon: AlertTriangle, color: 'text-warning', bg: 'bg-warning/10' },
    { title: 'Open Incidents', value: stats?.open_incidents ?? 0, icon: RefreshCw, color: 'text-error', bg: 'bg-error/10' },
  ]

  return (
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
  )
}
