import { useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Share2 } from 'lucide-react'
import cytoscape from 'cytoscape'
import { api } from '@/api/client'

// Mini connectivity graph widget (read-only cytoscape rendering).
export function ConnectivityMap() {
  const containerRef = useRef<HTMLDivElement>(null)

  const { data: graphData } = useQuery({
    queryKey: ['graph'],
    queryFn: () => api.get('api/graph').json(),
    refetchInterval: 60_000,
  })

  useEffect(() => {
    if (!containerRef.current || !graphData) return
    const resp = graphData as { elements: Array<{ data: Record<string, unknown> }> }
    if (!resp.elements || resp.elements.length === 0) return

    const cy = cytoscape({
      container: containerRef.current,
      elements: resp.elements.map((el) => ({
        group: el.data.source && el.data.target ? ('edges' as const) : ('nodes' as const),
        data: el.data,
      })),
      style: [
        { selector: 'node', style: { label: 'data(label)', color: '#9ca3af', 'font-size': '9px', 'text-valign': 'bottom', 'text-outline-width': 2, 'text-outline-color': '#1f2937', 'background-color': '#1e293b', 'border-color': '#475569', 'border-width': 1.5, width: 20, height: 20 } },
        { selector: 'node[type = "internet"]', style: { 'background-color': '#4b5563', 'border-color': '#6b7280', width: 30, height: 30 } },
        { selector: 'node[type = "node"]', style: { 'background-color': '#1e40af', 'border-color': '#3b82f6', width: 25, height: 25 } },
        { selector: 'node[type = "service"][status = "public"]', style: { 'border-color': '#ef4444', 'border-width': 3 } },
        { selector: 'edge', style: { width: 1, 'line-color': '#374151', 'target-arrow-shape': 'triangle', 'arrow-scale': 0.8, 'curve-style': 'bezier' } },
      ],
      layout: { name: 'breadthfirst', directed: true, spacingFactor: 1.0, animate: true, animationDuration: 300 },
      userZoomingEnabled: false,
      userPanningEnabled: false,
      autolock: true,
    } as never)

    return () => {
      cy.destroy()
    }
  }, [graphData])

  if (!graphData) return null

  return (
    <div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
      <div className="px-5 py-3 border-b border-base-content/5 flex items-center gap-2">
        <Share2 className="w-4 h-4 text-base-content/40" />
        <h2 className="font-semibold text-sm">Connectivity Map</h2>
      </div>
      <div ref={containerRef} className="w-full h-48" />
    </div>
  )
}
