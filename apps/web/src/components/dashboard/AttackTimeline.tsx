import { Activity, Clock } from 'lucide-react'
import type { TimelineEvent } from './common'

export function AttackTimeline({ events }: { events: TimelineEvent[] }) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="w-4 h-4 text-base-content/40" />
          <h2 className="font-semibold text-sm">Recent Activity</h2>
        </div>
        <span className="text-xs text-base-content/30">{events.length} events</span>
      </div>
      {events.length === 0 ? (
        <div className="p-5 text-center">
          <Clock className="w-6 h-6 mx-auto text-base-content/20 mb-1" />
          <p className="text-xs text-base-content/50">No recent security events</p>
        </div>
      ) : (
        <div className="divide-y divide-base-content/5 max-h-48 overflow-y-auto">
          {events.map((ev) => (
            <div key={ev.id} className="px-5 py-2.5 flex items-center gap-3 text-sm">
              <span
                className={`w-2 h-2 rounded-full shrink-0 ${
                  ev.type.includes('brute') || ev.type.includes('attack') ? 'bg-error' : 'bg-warning'
                }`}
              />
              <span className="text-xs text-base-content/40 w-16 shrink-0 font-mono">
                {new Date(ev.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
              <span className="text-xs flex-1 truncate">{ev.type.replace(/_/g, ' ')}</span>
              <span className="text-xs text-base-content/30 font-mono truncate max-w-[80px]">
                {ev.node_id?.slice(0, 8)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
