import { Users, Wifi } from 'lucide-react'
import { InfoBadge, type TSDevice } from './common'

export function TailscaleWidget({ configured, devices, loading }: {
  configured: boolean
  devices?: TSDevice[]
  loading: boolean
}) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-4 border-b border-base-content/5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">Tailscale Devices</h2>
        </div>
        {!configured && <span className="text-xs text-base-content/30">Not configured</span>}
      </div>
      {!configured ? (
        <div className="p-6 text-center">
          <Wifi className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
          <p className="text-sm text-base-content/50">
            Configure your Tailscale API key in Settings to see devices.
          </p>
        </div>
      ) : loading ? (
        <div className="p-6 text-center">
          <span className="loading loading-spinner loading-sm" />
        </div>
      ) : !devices || devices.length === 0 ? (
        <div className="p-6 text-center">
          <p className="text-sm text-base-content/50">No devices found in tailnet.</p>
        </div>
      ) : (
        <div className="divide-y divide-base-content/5">
          <div className="px-5 py-2 text-xs text-base-content/30 flex justify-between">
            <span>
              {devices.length} device{devices.length !== 1 ? 's' : ''}
            </span>
            <span>{devices.filter((d) => d.online).length} online</span>
          </div>
          {devices.slice(0, 10).map((d) => (
            <div key={d.id} className="px-5 py-3 flex items-center justify-between">
              <div className="flex items-center gap-3 min-w-0">
                <div>
                  <p className="text-sm font-medium">{d.name || d.hostname}</p>
                  <p className="text-xs text-base-content/40">
                    {d.os} &middot; {d.addresses?.[0]}
                  </p>
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
          {devices.length > 10 && (
            <div className="px-5 py-2 text-xs text-center text-base-content/30">
              + {devices.length - 10} more devices
            </div>
          )}
        </div>
      )}
    </div>
  )
}
