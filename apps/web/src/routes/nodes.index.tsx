import { createRoute } from '@tanstack/react-router'
import { Route as nodesRoute } from './nodes'
import { NodesPage } from '@/components/pages/NodesPage'

export const Route = createRoute({
  getParentRoute: () => nodesRoute,
  path: '/',
  component: () => <NodesPage />,
})
