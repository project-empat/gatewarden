import { createRoute, Outlet } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/nodes',
  component: () => (
    <DashboardLayout>
      <Outlet />
    </DashboardLayout>
  ),
})
