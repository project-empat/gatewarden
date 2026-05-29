import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Settings as SettingsIcon, Save, Cloud, Users } from 'lucide-react'
import { api } from '@/api/client'
import { useEffect, useState } from 'react'

interface AppSettings {
  agent_auto_approve: boolean
  heartbeat_interval: number
  log_retention_days: number
  cloudflare_api_token: string
  tailscale_api_key: string
  tailscale_tailnet: string
}

export function SettingsPage() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<AppSettings | null>(null)

  const { data: settings, isLoading } = useQuery<AppSettings>({
    queryKey: ['settings'],
    queryFn: () => api.get('api/settings').json(),
  })

  useEffect(() => {
    if (settings && !form) {
      setForm(settings)
    }
  }, [settings, form])

  const updateMutation = useMutation({
    mutationFn: (data: AppSettings) =>
      api.put('api/settings', { json: data }).json(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  const update = (patch: Partial<AppSettings>) => {
    if (form) setForm({ ...form, ...patch })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Settings</h1>
          <p className="text-base-content/60 text-sm mt-1">System configuration</p>
        </div>
        {form && (
          <button
            className="btn btn-primary btn-sm"
            onClick={() => updateMutation.mutate(form)}
            disabled={updateMutation.isPending}
          >
            <Save className="w-4 h-4" />
            {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
          </button>
        )}
      </div>

      {updateMutation.isSuccess && (
        <div className="alert alert-success text-sm">Settings saved successfully.</div>
      )}
      {updateMutation.isError && (
        <div className="alert alert-error text-sm">Failed to save settings.</div>
      )}

      {/* General Settings */}
      <div className="bg-base-100 rounded-xl border border-base-content/5 divide-y divide-base-content/10">
        <div className="px-5 py-4 border-b border-base-content/5">
          <h2 className="font-semibold">General</h2>
        </div>

        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Agent Auto-Approve</p>
            <p className="text-sm text-base-content/60">Automatically approve new agent connections</p>
          </div>
          <input
            type="checkbox"
            className="toggle toggle-primary"
            checked={form?.agent_auto_approve ?? true}
            onChange={(e) => update({ agent_auto_approve: e.target.checked })}
          />
        </div>

        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Heartbeat Interval</p>
            <p className="text-sm text-base-content/60">How often agents report status</p>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="number"
              className="input input-bordered input-sm w-20 text-right"
              value={form?.heartbeat_interval ?? 60}
              onChange={(e) => update({ heartbeat_interval: parseInt(e.target.value) || 60 })}
              min={10}
              max={600}
            />
            <span className="text-sm text-base-content/60">seconds</span>
          </div>
        </div>

        <div className="p-5 flex items-center justify-between">
          <div>
            <p className="font-medium">Log Retention</p>
            <p className="text-sm text-base-content/60">How long to keep event history</p>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="number"
              className="input input-bordered input-sm w-20 text-right"
              value={form?.log_retention_days ?? 30}
              onChange={(e) => update({ log_retention_days: parseInt(e.target.value) || 30 })}
              min={1}
              max={365}
            />
            <span className="text-sm text-base-content/60">days</span>
          </div>
        </div>
      </div>

      {/* Cloudflare Integration */}
      <div className="bg-base-100 rounded-xl border border-base-content/5 divide-y divide-base-content/10">
        <div className="px-5 py-4 border-b border-base-content/5 flex items-center gap-2">
          <Cloud className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">Cloudflare Integration</h2>
        </div>

        <div className="p-5">
          <div className="form-control">
            <label className="label">
              <span className="label-text">API Token</span>
            </label>
            <input
              type="password"
              className="input input-bordered w-full font-mono text-sm"
              placeholder="Enter your Cloudflare API token"
              value={form?.cloudflare_api_token ?? ''}
              onChange={(e) => update({ cloudflare_api_token: e.target.value })}
            />
            <label className="label">
              <span className="label-text-alt text-base-content/40">
                Required to list tunnels and monitor tunnel health. Needs Cloudflare Tunnel:Read permissions.
              </span>
            </label>
          </div>
          <p className="text-xs text-base-content/30 mt-2">
            Configure Cloudflare API token to enable tunnel listing, health monitoring, and "Expose privately" actions.
          </p>
        </div>
      </div>

      {/* Tailscale Integration */}
      <div className="bg-base-100 rounded-xl border border-base-content/5 divide-y divide-base-content/10">
        <div className="px-5 py-4 border-b border-base-content/5 flex items-center gap-2">
          <Users className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">Tailscale Integration</h2>
        </div>

        <div className="p-5 space-y-4">
          <div className="form-control">
            <label className="label">
              <span className="label-text">API Key</span>
            </label>
            <input
              type="password"
              className="input input-bordered w-full font-mono text-sm"
              placeholder="Enter your Tailscale API key"
              value={form?.tailscale_api_key ?? ''}
              onChange={(e) => update({ tailscale_api_key: e.target.value })}
            />
            <label className="label">
              <span className="label-text-alt text-base-content/40">
                Generate from the Tailscale admin console. Needs ACL read permissions.
              </span>
            </label>
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Tailnet Name</span>
            </label>
            <input
              type="text"
              className="input input-bordered w-full"
              placeholder="e.g., my-org.github"
              value={form?.tailscale_tailnet ?? ''}
              onChange={(e) => update({ tailscale_tailnet: e.target.value })}
            />
            <label className="label">
              <span className="label-text-alt text-base-content/40">
                Your tailnet name from the admin console (e.g., &lt;org&gt;.github or &lt;org&gt;.ts.net)
              </span>
            </label>
          </div>

          <p className="text-xs text-base-content/30">
            Configure Tailscale API access to list nodes, check ACL versions, and detect MFA enforcement.
          </p>
        </div>
      </div>

      {/* About */}
      <div className="bg-base-100 rounded-xl p-5 border border-base-content/5">
        <div className="flex items-center gap-3 mb-4">
          <SettingsIcon className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">About</h2>
        </div>
        <p className="text-sm text-base-content/60">
          Gatewarden v0.1.0 — Infrastructure Security Control Plane
        </p>
      </div>
    </div>
  )
}
