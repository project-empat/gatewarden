import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { ReportsPage } from '@/components/pages/ReportsPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/reports',
  component: () => (
    <DashboardLayout>
      <ReportsPage />
    </DashboardLayout>
  ),
})
