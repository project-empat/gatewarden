import { useQuery } from '@tanstack/react-query'
import { ServerIcon, WifiIcon, XCircleIcon } from '@heroicons/react/24/outline'
import { api } from '@/api/client'

interface Node {
  id: string
  name: string
  hostname: string
  ip: string
  os: string
  status: string
  last_seen: string
}

export function NodesPage() {
  const { data: nodes, isLoading } = useQuery<Node[]>({
    queryKey: ['nodes'],
    queryFn: () => api.get('api/nodes').json(),
    refetchInterval: 15_000,
  })

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
          <h1 className="text-2xl font-bold">Nodes</h1>
          <p className="text-base-content/60 text-sm mt-1">Managed infrastructure nodes</p>
        </div>
        <button className="btn btn-primary btn-sm">Add Node</button>
      </div>

      <div className="overflow-x-auto bg-base-100 rounded-xl border border-base-content/5">
        <table className="table table-zebra">
          <thead>
            <tr>
              <th>Name</th>
              <th>Hostname</th>
              <th>IP</th>
              <th>OS</th>
              <th>Status</th>
              <th>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {nodes?.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center text-base-content/40 py-12">
                  <ServerIcon className="w-8 h-8 mx-auto mb-2" />
                  <p>No nodes registered yet. Install an agent on your server.</p>
                </td>
              </tr>
            )}
            {nodes?.map((node) => (
              <tr key={node.id}>
                <td className="font-medium">{node.name}</td>
                <td>{node.hostname}</td>
                <td>{node.ip}</td>
                <td>{node.os}</td>
                <td>
                  <span
                    className={`inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full ${
                      node.status === 'online'
                        ? 'bg-success/10 text-success'
                        : 'bg-base-200 text-base-content/50'
                    }`}
                  >
                    {node.status === 'online' ? (
                      <WifiIcon className="w-3 h-3" />
                    ) : (
                      <XCircleIcon className="w-3 h-3" />
                    )}
                    {node.status}
                  </span>
                </td>
                <td className="text-sm text-base-content/60">
                  {new Date(node.last_seen).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
