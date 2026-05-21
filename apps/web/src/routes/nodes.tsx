import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { NodesPage } from '@/components/pages/NodesPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/nodes',
  component: () => (
    <DashboardLayout>
      <NodesPage />
    </DashboardLayout>
  ),
})
