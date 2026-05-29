import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { GraphPage } from '@/components/pages/GraphPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/graph',
  component: () => (
    <DashboardLayout>
      <GraphPage />
    </DashboardLayout>
  ),
})
