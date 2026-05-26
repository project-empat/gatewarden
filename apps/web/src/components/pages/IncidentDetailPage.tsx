import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, Link } from '@tanstack/react-router'
import { ArrowLeft, CheckCircle, AlertTriangle } from 'lucide-react'
import { api } from '@/api/client'

interface Incident {
  id: string
  node_id: string
  severity: string
  title: string
  message: string
  status: string
  created_at: string
  resolved_at: string | null
}

interface Node {
  id: string
  name: string
  hostname: string
}

export function IncidentDetailPage() {
  const { incidentId } = useParams({ from: '/incidents/$incidentId' })
  const queryClient = useQueryClient()

  const { data: allIncidents, isLoading } = useQuery<Incident[]>({
    queryKey: ['incidents'],
    queryFn: () => api.get('api/incidents').json(),
  })

  const { data: nodes } = useQuery<Node[]>({
    queryKey: ['nodes'],
    queryFn: () => api.get('api/nodes').json(),
  })

  const resolveMutation = useMutation({
    mutationFn: () => api.put(`api/incidents/${incidentId}/resolve`).json(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['incidents'] }),
  })

  const incident = allIncidents?.find((i) => i.id === incidentId)
  const node = nodes?.find((n) => n.id === incident?.node_id)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  if (!incident) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 mx-auto text-base-content/30 mb-3" />
          <p className="font-medium">Incident not found</p>
          <Link to="/incidents" className="btn btn-ghost btn-sm mt-2">
            Back to incidents
          </Link>
        </div>
      </div>
    )
  }

  const severityClass: Record<string, string> = {
    critical: 'badge-error',
    high: 'badge-warning',
    medium: 'badge-info',
    low: 'badge-ghost',
  }

  const severityColor: Record<string, string> = {
    critical: 'text-error',
    high: 'text-warning',
    medium: 'text-info',
    low: 'text-base-content/40',
  }

  return (
    <div className="max-w-3xl space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-base-content/40">
        <Link to="/incidents" className="hover:text-base-content/70 transition-colors">
          Incidents
        </Link>
        <span>/</span>
        <span className="text-base-content/70">{incident.title}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <Link to="/incidents" className="btn btn-ghost btn-sm btn-square">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className={`badge ${severityClass[incident.severity]}`}>
                {incident.severity}
              </span>
              {incident.status === 'resolved' && (
                <span className="badge badge-outline">Resolved</span>
              )}
            </div>
            <h1 className="text-2xl font-bold">{incident.title}</h1>
          </div>
        </div>
        {incident.status === 'open' && (
          <button
            className="btn btn-success btn-sm"
            onClick={() => resolveMutation.mutate()}
          >
            <CheckCircle className="w-4 h-4" />
            Resolve
          </button>
        )}
      </div>

      {/* Details */}
      <div className="bg-base-100 rounded-xl p-6 border border-base-content/5">
        <dl className="space-y-4">
          <div className="flex justify-between items-start">
            <dt className="text-sm text-base-content/60">Severity</dt>
            <dd className={`text-sm font-semibold capitalize ${severityColor[incident.severity]}`}>
              {incident.severity}
            </dd>
          </div>
          <div className="flex justify-between items-start">
            <dt className="text-sm text-base-content/60">Status</dt>
            <dd className="text-sm font-medium capitalize">{incident.status}</dd>
          </div>
          <div className="flex justify-between items-start">
            <dt className="text-sm text-base-content/60">Node</dt>
            <dd className="text-sm font-medium">
              {node ? (
                <Link
                  to="/nodes/$nodeId"
                  params={{ nodeId: node.id }}
                  className="link link-primary"
                >
                  {node.name}
                </Link>
              ) : (
                <span className="text-base-content/40">Unknown</span>
              )}
            </dd>
          </div>
          <div className="flex justify-between items-start">
            <dt className="text-sm text-base-content/60">Created</dt>
            <dd className="text-sm font-medium">
              {new Date(incident.created_at).toLocaleString()}
            </dd>
          </div>
          {incident.resolved_at && (
            <div className="flex justify-between items-start">
              <dt className="text-sm text-base-content/60">Resolved</dt>
              <dd className="text-sm font-medium">
                {new Date(incident.resolved_at).toLocaleString()}
              </dd>
            </div>
          )}
        </dl>
      </div>

      {/* Message */}
      {incident.message && (
        <div className="bg-base-100 rounded-xl p-6 border border-base-content/5">
          <h2 className="text-sm font-semibold text-base-content/40 uppercase tracking-wider mb-2">
            Details
          </h2>
          <p className="whitespace-pre-wrap text-sm">{incident.message}</p>
        </div>
      )}
    </div>
  )
}
