import { Sidebar } from './Sidebar'
import { Header } from './Header'

export function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="drawer lg:drawer-open">
      <input id="sidebar-drawer" type="checkbox" className="drawer-toggle" />
      <div className="drawer-content flex flex-col">
        <Header />
        <main className="flex-1 p-6 bg-base-200 min-h-screen">
          {children}
        </main>
      </div>
      <Sidebar />
    </div>
  )
}
