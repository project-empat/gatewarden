import { Route as rootRoute } from './routes/__root'
import { Route as loginRoute } from './routes/login'
import { Route as dashboardRoute } from './routes/dashboard'
import { Route as nodesRoute } from './routes/nodes'
import { Route as incidentsRoute } from './routes/incidents'
import { Route as settingsRoute } from './routes/settings'

const routeTree = rootRoute.addChildren([
  loginRoute,
  dashboardRoute,
  nodesRoute,
  incidentsRoute,
  settingsRoute,
])

export { routeTree }
