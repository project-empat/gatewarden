import type { SecuritySummary } from './common'

export function ConnectivityStats({ summary }: { summary?: SecuritySummary }) {
  if (!summary) return null

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
        <p className="text-xs text-base-content/40 mb-1">Exposed Services</p>
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span>SSH Public</span>
            <span className={summary.exposed_ssh > 0 ? 'text-error font-bold' : 'text-success'}>{summary.exposed_ssh}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>SSH Password Auth</span>
            <span className={summary.password_auth_ssh > 0 ? 'text-error font-bold' : 'text-success'}>{summary.password_auth_ssh}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Docker Socket Exposed</span>
            <span className={summary.docker_exposed > 0 ? 'text-error font-bold' : 'text-success'}>{summary.docker_exposed}</span>
          </div>
        </div>
      </div>

      <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
        <p className="text-xs text-base-content/40 mb-1">CrowdSec</p>
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span>Nodes</span>
            <span>{summary.crowdsec_nodes} / {summary.online_nodes}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Active Decisions</span>
            <span className={summary.total_decisions > 0 ? 'text-warning' : ''}>{summary.total_decisions}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Alerts (1h)</span>
            <span className={summary.total_alerts > 0 ? 'text-error font-bold' : ''}>{summary.total_alerts}</span>
          </div>
        </div>
      </div>

      <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
        <p className="text-xs text-base-content/40 mb-1">Fail2Ban</p>
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span>Total Jails</span>
            <span>{summary.fail2ban_jails_total}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Currently Banned</span>
            <span className={summary.fail2ban_bans_total > 0 ? 'text-error font-bold' : 'text-success'}>{summary.fail2ban_bans_total}</span>
          </div>
        </div>
      </div>

      <div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
        <p className="text-xs text-base-content/40 mb-1">Incidents</p>
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span>Open</span>
            <span className={summary.open_incidents > 0 ? 'text-error font-bold' : 'text-success'}>{summary.open_incidents}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>High/Critical</span>
            <span className={summary.high_severity > 0 ? 'text-error font-bold' : 'text-success'}>{summary.high_severity}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Total</span>
            <span>{summary.total_incidents}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
