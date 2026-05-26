import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, Plus, X } from 'lucide-react'
import { api } from '@/api/client'

interface Policy {
  id: string
  name: string
  description: string
  enabled: boolean
  severity: string
  triggers: string
  actions: string
  created_at: string
  updated_at: string
}

export function PoliciesPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Policy | null>(null)

  const { data: policies, isLoading } = useQuery<Policy[]>({
    queryKey: ['policies'],
    queryFn: () => api.get('api/policies').json(),
    refetchInterval: 30_000,
  })

  const createMutation = useMutation({
    mutationFn: (data: Partial<Policy>) =>
      api.post('api/policies', { json: data }).json(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] })
      setShowForm(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Policy> }) =>
      api.put(`api/policies/${id}`, { json: data }).json(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] })
      setEditing(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`api/policies/${id}`).json(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['policies'] }),
  })

  const toggleMutation = useMutation({
    mutationFn: (id: string) => api.post(`api/policies/${id}/toggle`).json(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['policies'] }),
  })

  const severityBadge = (sev: string) => {
    const classes: Record<string, string> = {
      critical: 'badge-error',
      high: 'badge-warning',
      medium: 'badge-info',
      low: 'badge-ghost',
    }
    return `badge ${classes[sev] ?? 'badge-ghost'}`
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Policies</h1>
          <p className="text-base-content/60 text-sm mt-1">
            Automated response rules for security events
          </p>
        </div>
        <button
          className="btn btn-primary btn-sm"
          onClick={() => {
            setEditing(null)
            setShowForm(true)
          }}
        >
          <Plus className="w-4 h-4" />
          New Policy
        </button>
      </div>

      {/* Policy Form Modal */}
      {(showForm || editing) && (
        <PolicyForm
          initial={editing}
          onSave={(data) => {
            if (editing) {
              updateMutation.mutate({ id: editing.id, data })
            } else {
              createMutation.mutate(data)
            }
          }}
          onClose={() => {
            setShowForm(false)
            setEditing(null)
          }}
        />
      )}

      {/* Empty State */}
      {(!policies || policies.length === 0) && (
        <div className="bg-base-100 rounded-xl p-12 text-center border border-base-content/5">
          <Shield className="w-12 h-12 mx-auto text-base-content/30 mb-3" />
          <p className="font-medium">No policies defined</p>
          <p className="text-sm text-base-content/60 mt-1">
            Create your first automated response policy to react to security events.
          </p>
          <button
            className="btn btn-primary btn-sm mt-4"
            onClick={() => setShowForm(true)}
          >
            <Plus className="w-4 h-4" />
            Create Policy
          </button>
        </div>
      )}

      {/* Policy List */}
      {policies && policies.length > 0 && (
        <div className="space-y-3">
          {policies.map((policy) => (
            <div
              key={policy.id}
              className="bg-base-100 rounded-xl p-5 border border-base-content/5 shadow-sm"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={severityBadge(policy.severity)}>
                      {policy.severity}
                    </span>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        className="toggle toggle-sm toggle-primary"
                        checked={policy.enabled}
                        onChange={() => toggleMutation.mutate(policy.id)}
                      />
                    </label>
                  </div>
                  <h3 className="font-semibold">{policy.name}</h3>
                  {policy.description && (
                    <p className="text-sm text-base-content/60 mt-1">
                      {policy.description}
                    </p>
                  )}
                  <div className="flex gap-4 mt-2 text-xs text-base-content/40">
                    <div>
                      <span className="font-medium">Triggers:</span>{' '}
                      {policy.triggers}
                    </div>
                    <div>
                      <span className="font-medium">Actions:</span>{' '}
                      {policy.actions}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <button
                    className="btn btn-ghost btn-xs"
                    onClick={() => {
                      setEditing(policy)
                      setShowForm(true)
                    }}
                  >
                    Edit
                  </button>
                  <button
                    className="btn btn-ghost btn-xs text-error"
                    onClick={() => {
                      if (confirm('Delete this policy?')) {
                        deleteMutation.mutate(policy.id)
                      }
                    }}
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function PolicyForm({
  initial,
  onSave,
  onClose,
}: {
  initial: Policy | null
  onSave: (data: Partial<Policy>) => void
  onClose: () => void
}) {
  const [name, setName] = useState(initial?.name ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [severity, setSeverity] = useState(initial?.severity ?? 'high')
  const [triggers, setTriggers] = useState(initial?.triggers ?? '')
  const [actions, setActions] = useState(initial?.actions ?? '')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave({ name, description, severity, triggers, actions, enabled: initial?.enabled ?? true })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-base-100 rounded-2xl shadow-2xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-6 border-b border-base-content/10">
          <h2 className="text-lg font-bold">
            {initial ? 'Edit Policy' : 'New Policy'}
          </h2>
          <button className="btn btn-ghost btn-sm btn-square" onClick={onClose}>
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="form-control">
            <label className="label">
              <span className="label-text">Name</span>
            </label>
            <input
              type="text"
              className="input input-bordered w-full"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder="e.g., Block Brute Force IPs"
            />
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Description</span>
            </label>
            <textarea
              className="textarea textarea-bordered w-full"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              placeholder="When this policy triggers..."
            />
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Severity</span>
            </label>
            <select
              className="select select-bordered w-full"
              value={severity}
              onChange={(e) => setSeverity(e.target.value)}
            >
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Triggers</span>
            </label>
            <textarea
              className="textarea textarea-bordered w-full font-mono text-sm"
              value={triggers}
              onChange={(e) => setTriggers(e.target.value)}
              rows={3}
              placeholder="e.g., Failed SSH logins > 10 in 5 minutes"
            />
            <label className="label">
              <span className="label-text-alt text-base-content/40">
                What event conditions fire this policy
              </span>
            </label>
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Actions</span>
            </label>
            <textarea
              className="textarea textarea-bordered w-full font-mono text-sm"
              value={actions}
              onChange={(e) => setActions(e.target.value)}
              rows={3}
              placeholder="e.g., Block IP via firewall, send notification"
            />
            <label className="label">
              <span className="label-text-alt text-base-content/40">
                What to do when this policy triggers
              </span>
            </label>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" className="btn btn-ghost" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary">
              {initial ? 'Save Changes' : 'Create Policy'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
