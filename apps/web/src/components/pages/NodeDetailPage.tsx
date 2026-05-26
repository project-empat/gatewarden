import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from '@tanstack/react-router'
import {
  ArrowLeft,
  Server,
  Wifi,
  XCircle,
  Terminal,
  Clock,
} from 'lucide-react'
import { api } from '@/api/client'

interface Node {
  id: string
  name: string
  hostname: string
  ip: string
  os: string
  status: string
  labels: string[]
  last_seen: string
  created_at: string
}

interface Incident {
  id: string
  node_id: string
  severity: string
  title: string
  message: string
  status: string
  created_at: string
}

export function NodeDetailPage() {
  const { nodeId } = useParams({ from: '/nodes/$nodeId' })

  const { data: node, isLoading: nodeLoading } = useQuery<Node>({
    queryKey: ['node', nodeId],
    queryFn: () => api.get(`api/nodes/${nodeId}`).json(),
  })

  const { data: incidents } = useQuery<Incident[]>({
    queryKey: ['node-incidents', nodeId],
    queryFn: () =>
      api
        .get('api/incidents')
        .json()
      .then((all) =>
        (all as Incident[]).filter((i) => i.node_id === nodeId)
      )
  })

  if (nodeLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  if (!node) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <Server className="w-12 h-12 mx-auto text-base-content/30 mb-3" />
          <p className="font-medium">Node not found</p>
          <Link to="/nodes" className="btn btn-ghost btn-sm mt-2">
            Back to nodes
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-base-content/40">
        <Link to="/nodes" className="hover:text-base-content/70 transition-colors">
          Nodes
        </Link>
        <span>/</span>
        <span className="text-base-content/70">{node.name}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Link to="/nodes" className="btn btn-ghost btn-sm btn-square">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{node.name}</h1>
            <p className="text-sm text-base-content/60">{node.hostname}</p>
          </div>
        </div>
        <span
          className={`inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-full ${
            node.status === 'online'
              ? 'bg-success/10 text-success'
              : 'bg-base-200 text-base-content/50'
          }`}
        >
          {node.status === 'online' ? (
            <Wifi className="w-3.5 h-3.5" />
          ) : (
            <XCircle className="w-3.5 h-3.5" />
          )}
          {node.status}
        </span>
      </div>

      {/* Details Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-base-100 rounded-xl p-5 border border-base-content/5">
          <h2 className="text-sm font-semibold text-base-content/40 uppercase tracking-wider mb-3">
            Overview
          </h2>
          <dl className="space-y-3">
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">IP Address</dt>
              <dd className="text-sm font-medium">{node.ip}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">OS</dt>
              <dd className="text-sm font-medium">{node.os || 'Unknown'}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">Hostname</dt>
              <dd className="text-sm font-medium">{node.hostname}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">Labels</dt>
              <dd className="text-sm font-medium">
                {node.labels?.length
                  ? node.labels.map((l) => (
                      <span key={l} className="badge badge-sm badge-ghost mr-1">
                        {l}
                      </span>
                    ))
                  : 'None'}
              </dd>
            </div>
          </dl>
        </div>

        <div className="bg-base-100 rounded-xl p-5 border border-base-content/5">
          <h2 className="text-sm font-semibold text-base-content/40 uppercase tracking-wider mb-3">
            Status
          </h2>
          <dl className="space-y-3">
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">Last Seen</dt>
              <dd className="text-sm font-medium">
                <div className="flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5 text-base-content/40" />
                  {new Date(node.last_seen).toLocaleString()}
                </div>
              </dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">Registered</dt>
              <dd className="text-sm font-medium">
                {new Date(node.created_at).toLocaleDateString()}
              </dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-sm text-base-content/60">Status</dt>
              <dd className="text-sm font-medium capitalize">{node.status}</dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Incidents Section */}
      <div className="bg-base-100 rounded-xl border border-base-content/5">
        <div className="px-5 py-4 border-b border-base-content/5">
          <h2 className="font-semibold">Incidents</h2>
        </div>
        {(!incidents || incidents.length === 0) ? (
          <div className="p-8 text-center">
            <Terminal className="w-8 h-8 mx-auto text-base-content/30 mb-2" />
            <p className="text-sm text-base-content/50">No incidents for this node</p>
          </div>
        ) : (
          <div className="divide-y divide-base-content/5">
            {incidents.map((inc) => {
              const severityClass: Record<string, string> = {
                critical: 'badge-error',
                high: 'badge-warning',
                medium: 'badge-info',
                low: 'badge-ghost',
              }
              return (
                <div key={inc.id} className="px-5 py-3 flex items-center justify-between">
                  <div className="flex items-center gap-3 min-w-0">
                    <span className={`badge ${severityClass[inc.severity] ?? 'badge-ghost'}`}>
                      {inc.severity}
                    </span>
                    <span className="text-sm font-medium truncate">{inc.title}</span>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    <span className={`text-xs ${inc.status === 'open' ? 'text-error' : 'text-success'}`}>
                      {inc.status}
                    </span>
                    <Link
                      to="/incidents"
                      className="btn btn-ghost btn-xs"
                    >
                      View
                    </Link>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
