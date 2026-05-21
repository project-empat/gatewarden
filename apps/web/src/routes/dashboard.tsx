import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { DashboardPage } from '@/components/pages/DashboardPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: () => (
    <DashboardLayout>
      <DashboardPage />
    </DashboardLayout>
  ),
})
