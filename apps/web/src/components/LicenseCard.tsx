import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound, ShieldCheck, Lock, CheckCircle2, XCircle } from 'lucide-react'
import { api } from '@/api/client'

interface License {
  id: string
  plan: string
  status: string
  seats: number
  features: string[]
  expires_at: string
}

interface FeatureStatus {
  name: string
  enabled: boolean
}

interface LicenseInfo {
  build_mode: string
  license: License
}

interface FeaturesInfo {
  build_mode: string
  features: FeatureStatus[]
}

const featureLabels: Record<string, string> = {
  core_gateway: 'Core Gateway',
  basic_rbac: 'Basic RBAC',
  basic_audit: 'Basic Audit',
  advanced_rbac: 'Advanced RBAC',
  sso_oidc: 'SSO (OIDC)',
  sso_saml: 'SSO (SAML)',
  audit_export: 'Audit Export',
  audit_stream: 'Audit Streaming',
  policy_engine: 'Advanced Policy Engine',
  automation: 'Advanced Automation',
  msp_multi_tenant: 'MSP Multi-Tenancy',
  msp_isolation: 'MSP Isolation',
}

export function LicenseCard() {
  const queryClient = useQueryClient()
  const [licenseKey, setLicenseKey] = useState('')
  const [showKey, setShowKey] = useState(false)

  const { data: info, isLoading } = useQuery<LicenseInfo>({
    queryKey: ['license'],
    queryFn: () => api.get('api/license').json(),
  })

  const { data: features } = useQuery<FeaturesInfo>({
    queryKey: ['license-features'],
    queryFn: () => api.get('api/license/features').json(),
  })

  const activateMutation = useMutation({
    mutationFn: (key: string) =>
      api.post('api/license/activate', { json: { license_key: key } }).json(),
    onSuccess: () => {
      setLicenseKey('')
      queryClient.invalidateQueries({ queryKey: ['license'] })
      queryClient.invalidateQueries({ queryKey: ['license-features'] })
    },
  })

  if (isLoading) {
    return (
      <div className="bg-base-100 rounded-xl border border-base-content/5 p-5">
        <div className="skeleton h-4 w-40" />
      </div>
    )
  }

  const license = info?.license
  const isOSS = info?.build_mode !== 'enterprise'
  const isActive = license?.status === 'active'

  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5">
      <div className="px-5 py-4 border-b border-base-content/5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <KeyRound className="w-5 h-5 text-base-content/40" />
          <h2 className="font-semibold">License</h2>
        </div>
        <span
          className={`badge badge-sm ${
            isActive ? 'badge-success' : isOSS ? 'badge-ghost' : 'badge-warning'
          }`}
        >
          {isOSS ? 'OSS Build' : isActive ? `${license?.plan} plan` : license?.status}
        </span>
      </div>

      <div className="p-5 space-y-4">
        <div className="flex items-center gap-3">
          <ShieldCheck className="w-8 h-8 text-primary" />
          <div>
            <p className="font-semibold capitalize">{license?.plan ?? 'free'} Edition</p>
            <p className="text-sm text-base-content/60">
              {isActive
                ? `Expires ${license?.expires_at ? new Date(license.expires_at).toLocaleDateString() : '—'}`
                : 'Core free features are always available'}
            </p>
          </div>
        </div>

        {/* Feature flags */}
        {features && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {features.features.map((f) => (
              <div
                key={f.name}
                className="flex items-center gap-2 text-sm text-base-content/70"
                title={f.name}
              >
                {f.enabled ? (
                  <CheckCircle2 className="w-4 h-4 text-success shrink-0" />
                ) : (
                  <XCircle className="w-4 h-4 text-base-content/25 shrink-0" />
                )}
                <span>{featureLabels[f.name] ?? f.name}</span>
                {!f.enabled && <Lock className="w-3 h-3 text-base-content/30" />}
              </div>
            ))}
          </div>
        )}

        {/* Activation */}
        {isOSS ? (
          <div className="alert alert-warning text-sm">
            This OSS build does not support license activation. Build with the
            enterprise workspace (<code className="font-mono">make build-enterprise</code>) to
            unlock premium features.
          </div>
        ) : (
          <div className="flex gap-2">
            <input
              type={showKey ? 'text' : 'password'}
              className="input input-bordered input-sm flex-1 font-mono"
              placeholder="Enter license key"
              value={licenseKey}
              onChange={(e) => setLicenseKey(e.target.value)}
            />
            <button
              className="btn btn-primary btn-sm"
              onClick={() => licenseKey && activateMutation.mutate(licenseKey)}
              disabled={!licenseKey || activateMutation.isPending}
            >
              {activateMutation.isPending ? 'Activating...' : 'Activate'}
            </button>
            <button
              className="btn btn-ghost btn-sm"
              onClick={() => setShowKey(!showKey)}
              aria-label="Toggle key visibility"
            >
              {showKey ? 'Hide' : 'Show'}
            </button>
          </div>
        )}

        {activateMutation.isError && (
          <div className="alert alert-error text-sm">
            {String((activateMutation.error as Error)?.message ?? 'Activation failed')}
          </div>
        )}
        {activateMutation.isSuccess && (
          <div className="alert alert-success text-sm">License activated.</div>
        )}
      </div>
    </div>
  )
}
