import { create } from 'zustand'
import { api } from '@/api/client'

interface User {
  id: string
  email: string
  role?: string
}

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  init: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  isAuthenticated: false,

  init: () => {
    const token = localStorage.getItem('gatewarden_token')
    const userStr = localStorage.getItem('gatewarden_user')
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr)
        set({ token, user, isAuthenticated: true })
      } catch {
        localStorage.removeItem('gatewarden_token')
        localStorage.removeItem('gatewarden_user')
      }
    }
  },

  login: async (email: string, password: string) => {
    const res: { token: string; user: User } = await api
      .post('api/auth/login', { json: { email, password } })
      .json()

    localStorage.setItem('gatewarden_token', res.token)
    localStorage.setItem('gatewarden_user', JSON.stringify(res.user))
    set({ token: res.token, user: res.user, isAuthenticated: true })
  },

  logout: () => {
    localStorage.removeItem('gatewarden_token')
    localStorage.removeItem('gatewarden_user')
    set({ token: null, user: null, isAuthenticated: false })
  },
}))
