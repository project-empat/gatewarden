import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { PoliciesPage } from '@/components/pages/PoliciesPage'
import { DashboardLayout } from '@/components/layout/DashboardLayout'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/policies',
  component: () => (
    <DashboardLayout>
      <PoliciesPage />
    </DashboardLayout>
  ),
})
