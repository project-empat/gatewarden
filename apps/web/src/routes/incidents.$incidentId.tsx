import { createRoute } from '@tanstack/react-router'
import { Route as incidentsRoute } from './incidents'
import { IncidentDetailPage } from '@/components/pages/IncidentDetailPage'

export const Route = createRoute({
  getParentRoute: () => incidentsRoute,
  path: '$incidentId',
  component: () => <IncidentDetailPage />,
})
