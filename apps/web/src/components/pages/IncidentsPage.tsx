import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CheckCircleIcon } from '@heroicons/react/24/outline'
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

export function IncidentsPage() {
  const queryClient = useQueryClient()

  const { data: incidents, isLoading } = useQuery<Incident[]>({
    queryKey: ['incidents'],
    queryFn: () => api.get('api/incidents').json(),
    refetchInterval: 15_000,
  })

  const resolveMutation = useMutation({
    mutationFn: (id: string) => api.put(`api/incidents/${id}/resolve`).json(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['incidents'] }),
  })

  const severityBadge = (severity: string) => {
    const classes: Record<string, string> = {
      critical: 'badge-error',
      high: 'badge-warning',
      medium: 'badge-info',
      low: 'badge-ghost',
    }
    return `badge ${classes[severity] ?? 'badge-ghost'}`
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Incidents</h1>
        <p className="text-base-content/60 text-sm mt-1">Security incidents across your infrastructure</p>
      </div>

      <div className="space-y-3">
        {incidents?.length === 0 && (
          <div className="bg-base-100 rounded-xl p-12 text-center border border-base-content/5">
            <CheckCircleIcon className="w-12 h-12 mx-auto text-success mb-3" />
            <p className="font-medium">No incidents</p>
            <p className="text-sm text-base-content/60">Your infrastructure is secure.</p>
          </div>
        )}
        {incidents?.map((inc) => (
          <div
            key={inc.id}
            className={`bg-base-100 rounded-xl p-5 border shadow-sm ${
              inc.status === 'resolved' ? 'border-base-content/5 opacity-60' : 'border-base-content/10'
            }`}
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className={severityBadge(inc.severity)}>{inc.severity}</span>
                  <span className="text-xs text-base-content/40">
                    {new Date(inc.created_at).toLocaleString()}
                  </span>
                  {inc.status === 'resolved' && (
                    <span className="badge badge-outline text-xs">Resolved</span>
                  )}
                </div>
                <h3 className="font-semibold">{inc.title}</h3>
                {inc.message && (
                  <p className="text-sm text-base-content/60 mt-1">{inc.message}</p>
                )}
              </div>
              {inc.status === 'open' && (
                <button
                  className="btn btn-ghost btn-sm text-success"
                  onClick={() => resolveMutation.mutate(inc.id)}
                >
                  <CheckCircleIcon className="w-4 h-4" />
                  Resolve
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
