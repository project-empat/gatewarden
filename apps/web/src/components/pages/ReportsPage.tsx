import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Shield,
  AlertTriangle,
  Server,
  Activity,
  CheckCircle,
  XCircle,
  FileText,
  BarChart3,
  HardDrive,
} from 'lucide-react'
import { api } from '@/api/client'

// ---- API response types ----

interface PostureReport {
  generated_at: string
  node_count: number
  online_count: number
  offline_count: number
  incident_count: number
  open_incident_count: number
  resolved_count: number
  exposures: {
    ssh_public: number
    ssh_password_auth: number
    docker_socket: number
  }
  integration_coverage: {
    total_nodes: number
    crowdsec_nodes: number
    fail2ban_nodes: number
    tailscale_nodes: number
    cloudflare_nodes: number
  }
  firewall_summary: {
    total_with_firewall: number
    active_firewalls: number
    inactive_firewalls: number
  }
}

interface IncidentSummary {
  by_severity: Record<string, number>
  by_status: Record<string, number>
  total: number
  open_count: number
  resolved_count: number
  top_affected_nodes: Array<{ node_id: string; hostname: string; count: number }>
}

interface NodeHealth {
  id: string
  hostname: string
  ip: string
  status: string
  last_seen: string
  os: string
  cpu_percent: number
  memory_percent: number
  disk_percent: number
  uptime_seconds: number
}

interface NodeHealthReport {
  nodes: NodeHealth[]
}

// ---- Helpers ----

function InfoBadge({ good, label }: { good: boolean; label: string }) {
  return (
    <span className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full ${
      good ? 'bg-success/10 text-success' : 'bg-error/10 text-error'
    }`}>
      {good ? <CheckCircle className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
      {label}
    </span>
  )
}

function StatCard({ title, value, icon: Icon, color }: { title: string; value: string | number; icon: any; color: string }) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 p-4 flex items-center gap-3">
      <div className={`p-2.5 rounded-lg ${color}/10`}>
        <Icon className={`w-5 h-5 ${color}`} />
      </div>
      <div>
        <p className="text-xs text-base-content/40">{title}</p>
        <p className={`text-xl font-bold ${color}`}>{value}</p>
      </div>
    </div>
  )
}

function SectionCard({ title, icon: Icon, children }: { title: string; icon: any; children: React.ReactNode }) {
  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center gap-2">
        <Icon className="w-4 h-4 text-base-content/40" />
        <h2 className="font-semibold text-sm">{title}</h2>
      </div>
      <div className="p-5">{children}</div>
    </div>
  )
}

// ---- Tab Views ----

function PostureTab() {
  const { data, isLoading } = useQuery<PostureReport>({
    queryKey: ['report-posture'],
    queryFn: () => api.get('api/reports/posture').json(),
    refetchInterval: 60_000,
  })

  if (isLoading) return <LoadingSpinner />
  if (!data) return <EmptyState icon={FileText} message="No report data available yet." />

  return (
    <div className="space-y-6">
      <p className="text-xs text-base-content/40">
        Generated {new Date(data.generated_at).toLocaleString()}
      </p>

      {/* Node & Incident Overview */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatCard title="Total Nodes" value={data.node_count} icon={Server} color="text-primary" />
        <StatCard title="Online" value={data.online_count} icon={CheckCircle} color="text-success" />
        <StatCard title="Offline" value={data.offline_count} icon={XCircle} color="text-error" />
        <StatCard title="Open Incidents" value={data.open_incident_count} icon={AlertTriangle} color="text-warning" />
      </div>

      {/* Exposure Summary */}
      <SectionCard title="Exposure Summary" icon={Shield}>
        <div className="space-y-3">
          <ExposureRow label="SSH Publicly Exposed" count={data.exposures.ssh_public} />
          <ExposureRow label="SSH Password Auth Enabled" count={data.exposures.ssh_password_auth} />
          <ExposureRow label="Docker Socket Exposed" count={data.exposures.docker_socket} />
        </div>
      </SectionCard>

      {/* Integration Coverage */}
      <SectionCard title="Integration Coverage" icon={Activity}>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <CovStat label="CrowdSec" count={data.integration_coverage.crowdsec_nodes} total={data.integration_coverage.total_nodes} />
          <CovStat label="Fail2Ban" count={data.integration_coverage.fail2ban_nodes} total={data.integration_coverage.total_nodes} />
          <CovStat label="Tailscale" count={data.integration_coverage.tailscale_nodes} total={data.integration_coverage.total_nodes} />
          <CovStat label="Cloudflare" count={data.integration_coverage.cloudflare_nodes} total={data.integration_coverage.total_nodes} />
        </div>
      </SectionCard>

      {/* Firewall Summary */}
      <SectionCard title="Firewall Summary" icon={Shield}>
        <div className="space-y-3">
          <ExposureRow label="Nodes with Firewall" count={data.firewall_summary.total_with_firewall} />
          <ExposureRow label="Active" count={data.firewall_summary.active_firewalls} />
          <ExposureRow label="Inactive" count={data.firewall_summary.inactive_firewalls} />
        </div>
      </SectionCard>
    </div>
  )
}

function ExposureRow({ label, count }: { label: string; count: number }) {
  return (
    <div className="flex justify-between items-center text-sm">
      <span className="text-base-content/60">{label}</span>
      <span className={count > 0 ? 'text-error font-bold' : 'text-success'}>{count}</span>
    </div>
  )
}

function CovStat({ label, count, total }: { label: string; count: number; total: number }) {
  const pct = total > 0 ? Math.round((count / total) * 100) : 0
  return (
    <div className="bg-base-200/50 rounded-lg p-3 text-center">
      <p className="text-lg font-bold">{count}<span className="text-xs text-base-content/40 ml-1">/ {total}</span></p>
      <p className="text-xs text-base-content/40">{label}</p>
      <div className="w-full bg-base-200 rounded-full h-1.5 mt-1">
        <div className="bg-primary h-1.5 rounded-full" style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}

// ---- Incident Summary Tab ----

function IncidentTab() {
  const { data, isLoading } = useQuery<IncidentSummary>({
    queryKey: ['report-incidents'],
    queryFn: () => api.get('api/reports/incidents').json(),
    refetchInterval: 30_000,
  })

  if (isLoading) return <LoadingSpinner />
  if (!data) return <EmptyState icon={AlertTriangle} message="No incident data available." />

  const severityColors: Record<string, string> = {
    critical: 'text-error',
    high: 'text-warning',
    medium: 'text-info',
    low: 'text-base-content/50',
    info: 'text-base-content/40',
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <StatCard title="Total Incidents" value={data.total} icon={AlertTriangle} color="text-warning" />
        <StatCard title="Open" value={data.open_count} icon={AlertTriangle} color="text-error" />
        <StatCard title="Resolved" value={data.resolved_count} icon={CheckCircle} color="text-success" />
        <StatCard title="Resolution Rate" value={data.total > 0 ? `${Math.round((data.resolved_count / data.total) * 100)}%` : '—'} icon={BarChart3} color="text-info" />
      </div>

      {/* By Severity */}
      <SectionCard title="Incidents by Severity" icon={BarChart3}>
        <div className="space-y-2">
          {['critical', 'high', 'medium', 'low', 'info'].map((sev) => {
            const count = data.by_severity[sev] ?? 0
            if (count === 0 && sev !== 'critical') return null
            const color = severityColors[sev] || 'text-base-content/60'
            const barPct = data.total > 0 ? (count / data.total) * 100 : 0
            return (
              <div key={sev} className="space-y-1">
                <div className="flex justify-between text-sm">
                  <span className={`capitalize ${color}`}>{sev}</span>
                  <span className="font-bold">{count}</span>
                </div>
                <div className="w-full bg-base-200 rounded-full h-2">
                  <div className={`h-2 rounded-full ${color.replace('text', 'bg')}`} style={{ width: `${barPct}%` }} />
                </div>
              </div>
            )
          })}
        </div>
      </SectionCard>

      {/* Top Affected Nodes */}
      {data.top_affected_nodes.length > 0 && (
        <SectionCard title="Most Affected Nodes" icon={Server}>
          <div className="divide-y divide-base-content/5 -mx-5 -mb-5">
            {data.top_affected_nodes.map((n) => (
              <div key={n.node_id} className="px-5 py-3 flex justify-between items-center text-sm">
                <span className="font-medium">{n.hostname}</span>
                <span className="text-base-content/50">{n.count} incident{n.count !== 1 ? 's' : ''}</span>
              </div>
            ))}
          </div>
        </SectionCard>
      )}
    </div>
  )
}

// ---- Node Health Tab ----

function HealthTab() {
  const { data, isLoading } = useQuery<NodeHealthReport>({
    queryKey: ['report-health'],
    queryFn: () => api.get('api/reports/health').json(),
    refetchInterval: 30_000,
  })

  if (isLoading) return <LoadingSpinner />
  if (!data || data.nodes.length === 0) return <EmptyState icon={HardDrive} message="No nodes registered yet." />

  const pctColor = (v: number) => {
    if (v >= 80) return 'text-error'
    if (v >= 60) return 'text-warning'
    return 'text-success'
  }

  const pctBg = (v: number) => {
    if (v >= 80) return 'bg-error'
    if (v >= 60) return 'bg-warning'
    return 'bg-success'
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
        <StatCard title="Total Nodes" value={data.nodes.length} icon={Server} color="text-primary" />
        <StatCard title="Online" value={data.nodes.filter((n) => n.status === 'online').length} icon={CheckCircle} color="text-success" />
        <StatCard title="Offline" value={data.nodes.filter((n) => n.status !== 'online').length} icon={XCircle} color="text-error" />
      </div>

      <div className="overflow-x-auto bg-base-100 rounded-xl border border-base-content/5">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Hostname</th>
              <th>Status</th>
              <th>CPU</th>
              <th>Memory</th>
              <th>Disk</th>
              <th>Uptime</th>
              <th>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {data.nodes.map((n) => (
              <tr key={n.id}>
                <td className="font-medium">
                  {n.hostname}
                  <span className="text-xs text-base-content/40 ml-1">{n.os}</span>
                </td>
                <td>
                  <InfoBadge good={n.status === 'online'} label={n.status} />
                </td>
                <td className={pctColor(n.cpu_percent)}>
                  {n.cpu_percent > 0 ? `${n.cpu_percent.toFixed(1)}%` : '—'}
                </td>
                <td className={pctColor(n.memory_percent)}>
                  {n.memory_percent > 0 ? `${n.memory_percent.toFixed(1)}%` : '—'}
                </td>
                <td className={pctColor(n.disk_percent)}>
                  {n.disk_percent > 0 ? `${n.disk_percent.toFixed(1)}%` : '—'}
                </td>
                <td className="text-base-content/60">
                  {n.uptime_seconds > 0
                    ? `${Math.floor(n.uptime_seconds / 86400)}d`
                    : '—'}
                </td>
                <td className="text-xs text-base-content/40">
                  {new Date(n.last_seen).toLocaleDateString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Health bar chart summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {['cpu_percent', 'memory_percent', 'disk_percent'].map((metric) => {
          const values = data.nodes.map((n) => (n as any)[metric] as number).filter((v) => v > 0)
          const avg = values.length > 0 ? values.reduce((a, b) => a + b, 0) / values.length : 0
          const label = metric.replace('_percent', '').replace('_', ' ')
          return (
            <div key={metric} className="bg-base-100 rounded-xl border border-base-content/5 p-4">
              <p className="text-xs text-base-content/40 capitalize mb-1">Avg {label}</p>
              <p className={`text-2xl font-bold ${pctColor(avg)}`}>{avg.toFixed(1)}%</p>
              <div className="w-full bg-base-200 rounded-full h-2 mt-2">
                <div className={`h-2 rounded-full ${pctBg(avg)}`} style={{ width: `${Math.min(avg, 100)}%` }} />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ---- Shared helpers ----

function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center h-64">
      <span className="loading loading-spinner loading-lg" />
    </div>
  )
}

function EmptyState({ icon: Icon, message }: { icon: any; message: string }) {
  return (
    <div className="flex items-center justify-center h-64">
      <div className="text-center space-y-2">
        <Icon className="w-10 h-10 mx-auto text-base-content/20" />
        <p className="text-sm text-base-content/50">{message}</p>
      </div>
    </div>
  )
}

// ---- Main component ----

const TABS = [
  { id: 'posture', label: 'Security Posture', icon: Shield },
  { id: 'incidents', label: 'Incident Summary', icon: AlertTriangle },
  { id: 'health', label: 'Node Health', icon: Activity },
] as const

type TabId = (typeof TABS)[number]['id']

export function ReportsPage() {
  const urlParams = new URLSearchParams(typeof window !== 'undefined' ? window.location.hash.replace('#', '?') : '')
  const initialTab = (TABS.find((t) => t.id === urlParams.get('tab'))?.id ?? 'posture') as TabId
  const [activeTab, setActiveTab] = useState<TabId>(initialTab)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <FileText className="w-6 h-6 text-primary" />
            Reports
          </h1>
          <p className="text-sm text-base-content/60 mt-1">
            Security posture, incident summary, and node health overview
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="tabs tabs-bordered">
        {TABS.map((tab) => {
          const Icon = tab.icon
          return (
            <button
              key={tab.id}
              className={`tab tab-lg gap-2 ${activeTab === tab.id ? 'tab-active' : ''}`}
              onClick={() => {
                setActiveTab(tab.id)
                window.location.hash = `tab=${tab.id}`
              }}
            >
              <Icon className="w-4 h-4" />
              {tab.label}
            </button>
          )
        })}
      </div>

      {/* Tab Content */}
      {activeTab === 'posture' && <PostureTab />}
      {activeTab === 'incidents' && <IncidentTab />}
      {activeTab === 'health' && <HealthTab />}
    </div>
  )
}

// React.useState workaround - the component above uses it
