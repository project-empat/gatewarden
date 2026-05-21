import React from 'react'
import ReactDOM from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/authStore'
import { routeTree } from './routeTree.gen'
import './index.css'

const queryClient = new QueryClient()

const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  context: { queryClient, auth: undefined! },
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
  interface RouterContext {
    auth: ReturnType<typeof useAuthStore>
  }
}

function InnerApp() {
  const auth = useAuthStore()

  React.useEffect(() => {
    auth.init()
  }, [])

  return <RouterProvider router={router} context={{ queryClient, auth }} />
}

const root = ReactDOM.createRoot(document.getElementById('root')!)

root.render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <InnerApp />
    </QueryClientProvider>
  </React.StrictMode>,
)
