import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { LoginPage } from '@/components/pages/LoginPage'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})
