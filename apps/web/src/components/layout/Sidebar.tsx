import { Link, useLocation } from '@tanstack/react-router'
import {
  ShieldCheck,
  Server,
  AlertTriangle,
  Settings,
  LogOut,
  BarChart3,
  Shield,
  Share2,
  FileText,
  Bug,
} from 'lucide-react'
import { useAuthStore } from '@/stores/authStore'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: BarChart3 },
  { path: '/graph', label: 'Graph', icon: Share2 },
  { path: '/reports', label: 'Reports', icon: FileText },
  { path: '/nodes', label: 'Nodes', icon: Server },
  { path: '/incidents', label: 'Incidents', icon: AlertTriangle },
  { path: '/vulnerabilities', label: 'Vulnerabilities', icon: Bug },
  { path: '/policies', label: 'Policies', icon: Shield },
  { path: '/settings', label: 'Settings', icon: Settings },
]

export function Sidebar() {
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)

  return (
    <div className="drawer-side z-40">
      <label htmlFor="sidebar-drawer" className="drawer-overlay" />
      <aside className="bg-base-300 text-base-content min-h-screen w-64 flex flex-col">
        <div className="flex items-center gap-3 px-6 py-5 border-b border-base-content/10">
          <ShieldCheck className="w-8 h-8 text-primary" />
          <span className="text-xl font-bold tracking-tight">Gatewarden</span>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path ||
              (item.path !== '/' && location.pathname.startsWith(item.path))
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-content'
                    : 'text-base-content/70 hover:bg-base-200 hover:text-base-content'
                }`}
              >
                <Icon className="w-5 h-5" />
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="px-3 py-4 border-t border-base-content/10">
          <button
            onClick={logout}
            className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-base-content/70 hover:bg-base-200 hover:text-base-content w-full transition-colors"
          >
            <LogOut className="w-5 h-5" />
            Logout
          </button>
        </div>
      </aside>
    </div>
  )
}
