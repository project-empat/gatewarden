import { createRoute } from '@tanstack/react-router'
import { Route as nodesRoute } from './nodes'
import { NodeDetailPage } from '@/components/pages/NodeDetailPage'

export const Route = createRoute({
  getParentRoute: () => nodesRoute,
  path: '$nodeId',
  component: () => <NodeDetailPage />,
})
