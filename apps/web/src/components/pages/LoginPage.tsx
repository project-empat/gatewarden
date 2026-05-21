import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { ShieldCheckIcon } from '@heroicons/react/24/outline'
import { useAuthStore } from '@/stores/authStore'

export function LoginPage() {
  const [email, setEmail] = useState('admin@gatewarden.dev')
  const [password, setPassword] = useState('admin')
  const [error, setError] = useState('')
  const login = useAuthStore((s) => s.login)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await login(email, password)
      navigate({ to: '/dashboard' })
    } catch {
      setError('Invalid credentials')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-base-200">
      <div className="card w-full max-w-md bg-base-100 shadow-xl">
        <div className="card-body p-8">
          <div className="flex flex-col items-center gap-3 mb-6">
            <ShieldCheckIcon className="w-12 h-12 text-primary" />
            <h1 className="text-2xl font-bold">Gatewarden</h1>
            <p className="text-base-content/60 text-sm">Infrastructure Security Control Plane</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="form-control">
              <label className="label">
                <span className="label-text">Email</span>
              </label>
              <input
                type="email"
                className="input input-bordered w-full"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            <div className="form-control">
              <label className="label">
                <span className="label-text">Password</span>
              </label>
              <input
                type="password"
                className="input input-bordered w-full"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>

            {error && (
              <div className="alert alert-error text-sm py-2">{error}</div>
            )}

            <button type="submit" className="btn btn-primary w-full">
              Sign In
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
