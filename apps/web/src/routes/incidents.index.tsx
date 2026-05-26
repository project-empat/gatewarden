import { createRoute } from '@tanstack/react-router'
import { Route as incidentsRoute } from './incidents'
import { IncidentsPage } from '@/components/pages/IncidentsPage'

export const Route = createRoute({
  getParentRoute: () => incidentsRoute,
  path: '/',
  component: () => <IncidentsPage />,
})
