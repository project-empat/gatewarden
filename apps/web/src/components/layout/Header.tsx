import { useAuthStore } from '@/stores/authStore'
import { Sun, Moon, Menu } from 'lucide-react'

export function Header() {
  const user = useAuthStore((s) => s.user)
  const toggleTheme = () => {
    const html = document.documentElement
    const current = html.getAttribute('data-theme')
    html.setAttribute('data-theme', current === 'dark' ? 'light' : 'dark')
  }

  return (
    <header className="navbar bg-base-100 border-b border-base-content/10 px-4 sticky top-0 z-30">
      <div className="flex-1 flex items-center gap-3">
        <label htmlFor="sidebar-drawer" className="btn btn-ghost btn-square lg:hidden">
          <Menu className="w-6 h-6" />
        </label>
        <span className="text-sm text-base-content/60">
          Welcome back, <span className="font-semibold text-base-content">{user?.email ?? 'admin'}</span>
        </span>
      </div>
      <div className="flex-none flex items-center gap-2">
        <button className="btn btn-ghost btn-square" onClick={toggleTheme}>
          <Sun className="w-5 h-5 hidden dark:inline" />
          <Moon className="w-5 h-5 inline dark:hidden" />
        </button>
      </div>
    </header>
  )
}
