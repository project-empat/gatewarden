import { CheckCircle, XCircle } from 'lucide-react'

// Shared dashboard types and helpers.

export interface Stats {
  total_nodes: number
  online_nodes: number
  total_incidents: number
  open_incidents: number
}

export interface SecuritySummary {
  exposed_ssh: number
  docker_exposed: number
  password_auth_ssh: number
  total_incidents: number
  open_incidents: number
  total_nodes: number
  online_nodes: number
  high_severity: number
  crowdsec_nodes: number
  total_decisions: number
  total_alerts: number
  fail2ban_jails_total: number
  fail2ban_bans_total: number
}

export interface AppSettings {
  cloudflare_api_token: string
  tailscale_api_key: string
  tailscale_tailnet: string
}

export interface CFTunnel {
  id: string
  name: string
  status: string
  connectors?: Array<{ state: string; hostname: string }>
}

export interface TSDevice {
  id: string
  name: string
  hostname: string
  os: string
  online: boolean
  addresses: string[]
  clientVersion: string
}

export interface TimelineEvent {
  id: string
  node_id: string
  type: string
  payload: string
  created_at: string
}

// Security-relevant event types shown on the dashboard timeline.
export const SECURITY_EVENT_TYPES = [
  'auth_failure',
  'ssh_brute_force',
  'ssh_publicly_exposed',
  'ssh_password_auth_enabled',
  'docker_socket_exposed',
  'agent_report',
  'attack_detected',
]

export function InfoBadge({ good, label }: { good: boolean; label: string }) {
  return (
    <span
      className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full ${
        good ? 'bg-success/10 text-success' : 'bg-error/10 text-error'
      }`}
    >
      {good ? <CheckCircle className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
      {label}
    </span>
  )
}
