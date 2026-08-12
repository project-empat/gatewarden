import { ShieldCheck, ShieldAlert, ShieldX } from 'lucide-react'
import type { SecuritySummary } from './common'

// Compute a security score from 0-100 based on the summary.
export function computeSecurityScore(s: SecuritySummary): {
  score: number
  label: string
  Icon: typeof ShieldCheck
  color: string
} {
  let deductions = 0
  deductions += s.exposed_ssh * 15
  deductions += s.password_auth_ssh * 20
  deductions += s.docker_exposed * 25
  deductions += s.open_incidents * 5
  deductions += s.high_severity * 10
  const score = Math.max(0, Math.min(100, 100 - deductions))

  if (score >= 80) return { score, label: 'Good', Icon: ShieldCheck, color: 'text-success' }
  if (score >= 50) return { score, label: 'Warning', Icon: ShieldAlert, color: 'text-warning' }
  return { score, label: 'Critical', Icon: ShieldX, color: 'text-error' }
}

export function SecurityScoreCard({ summary }: { summary?: SecuritySummary }) {
  if (!summary) return null
  const sc = computeSecurityScore(summary)
  const Icon = sc.Icon

  return (
    <div className="bg-base-100 rounded-xl p-5 border border-base-content/5 flex items-center gap-4">
      <div className={`p-3 rounded-full ${sc.color.replace('text', 'bg')}/10`}>
        <Icon className={`w-8 h-8 ${sc.color}`} />
      </div>
      <div className="flex-1">
        <p className="text-xs text-base-content/40 mb-1">Security Score</p>
        <div className="flex items-baseline gap-2">
          <span className={`text-3xl font-bold ${sc.color}`}>{sc.score}</span>
          <span className={`text-sm font-medium ${sc.color}`}>/ 100</span>
        </div>
        <div className="w-full bg-base-200 rounded-full h-2 mt-2">
          <div
            className={`h-2 rounded-full transition-all duration-500 ${
              sc.score >= 80 ? 'bg-success' : sc.score >= 50 ? 'bg-warning' : 'bg-error'
            }`}
            style={{ width: `${sc.score}%` }}
          />
        </div>
        <p className="text-xs text-base-content/50 mt-1">
          {sc.label} &middot; {summary.online_nodes}/{summary.total_nodes} nodes online
        </p>
      </div>
    </div>
  )
}
