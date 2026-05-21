import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { IncidentsPage } from '@/components/pages/IncidentsPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/incidents',
  component: () => (
    <DashboardLayout>
      <IncidentsPage />
    </DashboardLayout>
  ),
})
