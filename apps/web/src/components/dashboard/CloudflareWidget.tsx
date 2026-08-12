import { Cloud, Globe } from 'lucide-react'
import { InfoBadge, type CFTunnel } from './common'

interface CloudflareWidgetProps {
  configured: boolean
  accountName?: string
  accountID?: string
  tunnels?: CFTunnel[]
  loading: boolean
}

export function CloudflareWidget({ configured, accountName, accountID, tunnels, loading }: CloudflareWidgetProps) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-4 border-b border-base-content/5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Cloud className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">Cloudflare Tunnels</h2>
        </div>
        {!configured && <span className="text-xs text-base-content/30">Not configured</span>}
      </div>
      {!configured ? (
        <div className="p-6 text-center">
          <Globe className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
          <p className="text-sm text-base-content/50">
            Configure your Cloudflare API token in Settings to see tunnels.
          </p>
        </div>
      ) : loading ? (
        <div className="p-6 text-center">
          <span className="loading loading-spinner loading-sm" />
        </div>
      ) : !tunnels || tunnels.length === 0 ? (
        <div className="p-6 text-center">
          <p className="text-sm text-base-content/50">
            No tunnels found{accountID ? ` for account ${accountID.slice(0, 8)}...` : ''}.
          </p>
        </div>
      ) : (
        <div className="divide-y divide-base-content/5">
          {tunnels.map((t) => (
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
      {configured && accountName && (
        <div className="px-5 py-2 border-t border-base-content/5 text-xs text-base-content/30">
          Account: {accountName} ({accountID?.slice(0, 12)}...)
        </div>
      )}
    </div>
  )
}
