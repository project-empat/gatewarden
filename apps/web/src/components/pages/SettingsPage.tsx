import { useQuery } from '@tanstack/react-query'
import { Cog6ToothIcon } from '@heroicons/react/24/outline'
import { api } from '@/api/client'

interface Settings {
  agent_auto_approve: boolean
  heartbeat_interval: number
  log_retention_days: number
}

export function SettingsPage() {
  const { data: settings } = useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: () => api.get('api/settings').json(),
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-base-content/60 text-sm mt-1">System configuration</p>
      </div>

      <div className="bg-base-100 rounded-xl border border-base-content/5 divide-y divide-base-content/10">
        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Agent Auto-Approve</p>
            <p className="text-sm text-base-content/60">
              Automatically approve new agent connections
            </p>
          </div>
          <input
            type="checkbox"
            className="toggle toggle-primary"
            defaultChecked={settings?.agent_auto_approve}
          />
        </div>

        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Heartbeat Interval</p>
            <p className="text-sm text-base-content/60">
              How often agents report status (seconds)
            </p>
          </div>
          <span className="text-lg font-semibold">{settings?.heartbeat_interval ?? 60}s</span>
        </div>

        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Log Retention</p>
            <p className="text-sm text-base-content/60">
              How long to keep event history
            </p>
          </div>
          <span className="text-lg font-semibold">{settings?.log_retention_days ?? 30} days</span>
        </div>
      </div>

      <div className="bg-base-100 rounded-xl p-5 border border-base-content/5">
        <div className="flex items-center gap-3 mb-4">
          <Cog6ToothIcon className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">About</h2>
        </div>
        <p className="text-sm text-base-content/60">
          Gatewarden v0.1.0 — Infrastructure Security Control Plane
        </p>
      </div>
    </div>
  )
}
