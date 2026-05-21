import { useEffect, useRef } from 'react'
import { useAuthStore } from '@/stores/authStore'

type EventCallback = (event: { type: string; data: unknown }) => void

export function useSSE(onEvent: EventCallback) {
  const token = useAuthStore((s) => s.token)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    if (!token) return

    const es = new EventSource(`/api/events?token=${token}`)
    esRef.current = es

    es.onmessage = (msg) => {
      try {
        const data = JSON.parse(msg.data)
        onEvent({ type: msg.type || 'message', data })
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      es.close()
    }

    return () => es.close()
  }, [token, onEvent])

  return esRef
}
