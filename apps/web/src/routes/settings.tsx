import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { SettingsPage } from '@/components/pages/SettingsPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: () => (
    <DashboardLayout>
      <SettingsPage />
    </DashboardLayout>
  ),
})
