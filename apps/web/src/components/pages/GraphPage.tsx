import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Shield,
  Server,
  AlertTriangle,
  ZoomIn,
  ZoomOut,
  Maximize2,
  SlidersHorizontal,
  X,
} from 'lucide-react'
import cytoscape, { type Core, type EventObject } from 'cytoscape'
import { api } from '@/api/client'

// Colors from the security-graph spec
const TYPE_COLORS: Record<string, { color: string; bg: string; shape: string }> = {
  internet:     { color: '#6b7280', bg: '#6b7280', shape: 'ellipse' },
  node:         { color: '#3b82f6', bg: '#1e40af', shape: 'round-rectangle' },
  container:    { color: '#10b981', bg: '#065f46', shape: 'rectangle' },
  service:      { color: '#f59e0b', bg: '#92400e', shape: 'ellipse' },
  tunnel:       { color: '#8b5cf6', bg: '#5b21b6', shape: 'diamond' },
  firewall:     { color: '#06b6d4', bg: '#0e7490', shape: 'hexagon' },
  integration:  { color: '#ec4899', bg: '#9d174d', shape: 'hexagon' },
  identity:     { color: '#14b8a6', bg: '#0f766e', shape: 'round-rectangle' },
  attack:       { color: '#ef4444', bg: '#991b1b', shape: 'triangle' },
  incident:     { color: '#f97316', bg: '#9a3412', shape: 'diamond' },
}

const EDGE_COLORS: Record<string, string> = {
  'routes to':     '#6b7280',
  'terminates on': '#6b7280',
  'listens on':    '#3b82f6',
  'exposes':       '#ef4444',
  'protected by':  '#06b6d4',
  'allows':        '#10b981',
  'runs':          '#10b981',
  'exposes via':   '#f59e0b',
  'blocked':       '#ef4444',
  'detected':      '#f59e0b',
  'reported by':   '#f97316',
  'has issue':     '#ef4444',
}

interface GraphElem {
  data: {
    id: string
    label: string
    type: string
    status?: string
    severity?: string
    port?: number
    hostname?: string
    image?: string
    detail?: string
    node_id?: string
    source?: string
    target?: string
    color?: string
  }
}

interface GraphResponse {
  elements: GraphElem[]
}

export function GraphPage() {
  const containerRef = useRef<HTMLDivElement>(null)
  const cyRef = useRef<Core | null>(null)
  const [selectedNode, setSelectedNode] = useState<GraphElem | null>(null)
  const [filterTypes, setFilterTypes] = useState<Set<string>>(
    new Set(['internet', 'node', 'service', 'container', 'firewall', 'integration', 'incident', 'attack']),
  )
  const [showFilters, setShowFilters] = useState(false)

  const { data: graphData, isLoading, error } = useQuery<GraphResponse>({
    queryKey: ['graph'],
    queryFn: () => api.get('api/graph').json(),
    refetchInterval: 30_000,
  })

  // Build cytoscape stylesheet
  const elements = graphData?.elements ?? []

  // Initialize cytoscape when data changes
  useEffect(() => {
    if (!containerRef.current || elements.length === 0) return

    // Destroy previous instance
    if (cyRef.current) {
      cyRef.current.destroy()
      cyRef.current = null
    }

    const nodes = elements.filter((el) => !el.data.source && !el.data.target)
    const edges = elements.filter((el) => el.data.source && el.data.target)

    const cy = cytoscape({
      container: containerRef.current,
      elements: [
        ...nodes.map((n) => ({
          group: 'nodes' as const,
          data: { ...n.data },
        })),
        ...edges.map((e) => ({
          group: 'edges' as const,
          data: {
            ...e.data,
            source: e.data.source!,
            target: e.data.target!,
          },
        })),
      ],
      style: [
        // Default node style
        {
          selector: 'node',
          style: {
            label: 'data(label)',
            'text-valign': 'bottom',
            'text-halign': 'center',
            'text-margin-y': 8,
            color: '#e5e7eb',
            'font-size': '11px',
            'font-weight': 'bold',
            'text-outline-width': 2,
            'text-outline-color': '#1f2937',
            width: 'label',
            height: 'label',
            padding: '12px',
            'border-width': 2,
            'border-color': '#374151',
          },
        },
        // Node by type
        ...Object.entries(TYPE_COLORS).map(([type, spec]) => ({
          selector: `node[type = "${type}"]`,
          style: {
            'background-color': spec.bg,
            'border-color': spec.color,
            shape: spec.shape as cytoscape.Css.NodeShape,
          },
        })),
        // Node size for different types
        {
          selector: 'node[type = "internet"]',
          style: { width: 60, height: 60, padding: '0px', 'font-size': '13px' },
        },
        {
          selector: 'node[type = "node"]',
          style: { 'border-width': 3 },
        },
        {
          selector: 'node[type = "incident"]',
          style: { 'background-opacity': 0.3 },
        },
        {
          selector: 'node[type = "service"][status = "public"]',
          style: { 'border-color': '#ef4444', 'border-width': 3 },
        },
        {
          selector: 'node[type = "service"][status = "private"]',
          style: { 'border-color': '#10b981', 'border-width': 3 },
        },
        // Edge styles
        {
          selector: 'edge',
          style: {
            width: 2,
            'curve-style': 'bezier',
            'target-arrow-shape': 'triangle',
            'arrow-scale': 1.2,
            'line-color': '#4b5563',
            'target-arrow-color': '#4b5563',
            'font-size': '9px',
            color: '#9ca3af',
            'text-background-color': '#1f2937',
            'text-background-opacity': 0.8,
            'text-background-padding': '2px',
          },
        },
        // Edge by label color
        ...Object.entries(EDGE_COLORS).map(([label, color]) => ({
          selector: `edge[label = "${label}"]`,
          style: {
            'line-color': color,
            'target-arrow-color': color,
          },
        })),
        // Hover effects
        {
          selector: 'node:selected',
          style: {
            'border-color': '#fbbf24',
            'border-width': 4,
          },
        },
        {
          selector: 'edge:selected',
          style: {
            width: 3,
          },
        },
      ],
      layout: {
        name: 'breadthfirst',
        directed: true,
        spacingFactor: 1.5,
        animate: true,
        animationDuration: 500,
      },
      minZoom: 0.3,
      maxZoom: 4,
      wheelSensitivity: 0.3,
    })

    // Event handlers
    cy.on('tap', 'node', (evt: EventObject) => {
      const node = evt.target
      const data = node.data()
      const allEls = elements.find(
        (el) => el.data.id === data.id || (el.data.source !== undefined && false),
      )
      if (allEls) {
        setSelectedNode(allEls)
      } else {
        // Find the matching node element
        const match = elements.find((el) => el.data.id === data.id && !el.data.source)
        setSelectedNode(match || null)
      }
    })

    cy.on('tap', (evt: EventObject) => {
      if (evt.target === cy) {
        setSelectedNode(null)
      }
    })

    cyRef.current = cy

    return () => {
      cy.destroy()
      cyRef.current = null
    }
  }, [elements])

  // Filter nodes by type
  const filterGraph = (type: string) => {
    setFilterTypes((prev) => {
      const next = new Set(prev)
      if (next.has(type)) {
        next.delete(type)
      } else {
        next.add(type)
      }
      return next
    })
  }

  // Apply filter to cytoscape
  useEffect(() => {
    if (!cyRef.current) return
    cyRef.current.nodes().forEach((node) => {
      const nodeType = (node.data() as any).type as string
      if (filterTypes.has(nodeType)) {
        ;(node as any).show()
      } else {
        ;(node as any).hide()
      }
    })
    // Also hide edges whose source or target is hidden
    cyRef.current.edges().forEach((edge) => {
      const source = edge.source() as any
      const target = edge.target() as any
      if (source.visible() && target.visible()) {
        ;(edge as any).show()
      } else {
        ;(edge as any).hide()
      }
    })
  }, [filterTypes])

  // Layout controls
  const applyLayout = (name: string) => {
    if (!cyRef.current) return
    cyRef.current.layout({
      name,
      directed: true,
      spacingFactor: name === 'breadthfirst' ? 1.5 : 1,
      animate: true,
      animationDuration: 400,
    } as any).run()
  }

  const fitToScreen = () => {
    cyRef.current?.fit(undefined, 50)
  }

  const zoomIn = () => {
    cyRef.current?.zoom(cyRef.current.zoom() * 1.3)
  }

  const zoomOut = () => {
    cyRef.current?.zoom(cyRef.current.zoom() * 0.7)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <span className="loading loading-spinner loading-lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center space-y-2">
          <AlertTriangle className="w-12 h-12 text-error mx-auto" />
          <p className="text-error">Failed to load security graph</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Shield className="w-6 h-6 text-primary" />
            Security Graph
          </h1>
          <p className="text-sm text-base-content/60 mt-1">
            Infrastructure relationship map — internet to services to containers
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            className="btn btn-ghost btn-sm"
            onClick={() => setShowFilters(!showFilters)}
            title="Toggle filters"
          >
            <SlidersHorizontal className="w-4 h-4" />
            Filters
          </button>
        </div>
      </div>

      <div className="flex gap-4">
        {/* Legend and filters */}
        <div className={`space-y-3 w-56 shrink-0 transition-opacity ${showFilters ? 'opacity-100' : 'opacity-60'}`}>
          <div className="bg-base-200 rounded-xl p-3 border border-base-content/5 space-y-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-base-content/50">Legend</h3>
            {Object.entries(TYPE_COLORS).map(([type, spec]) => (
              <label
                key={type}
                className="flex items-center gap-2 cursor-pointer group"
              >
                <input
                  type="checkbox"
                  checked={filterTypes.has(type)}
                  onChange={() => filterGraph(type)}
                  className="checkbox checkbox-xs"
                  style={{ '--chkbg': spec.color } as React.CSSProperties}
                />
                <span
                  className="w-2.5 h-2.5 rounded-sm shrink-0"
                  style={{ backgroundColor: spec.color }}
                />
                <span className="text-xs capitalize text-base-content/70 group-hover:text-base-content transition-colors">
                  {type}
                </span>
              </label>
            ))}
          </div>

          <div className="bg-base-200 rounded-xl p-3 border border-base-content/5 space-y-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-base-content/50">Controls</h3>
            <div className="flex flex-wrap gap-1">
              <button className="btn btn-xs btn-ghost" onClick={() => applyLayout('breadthfirst')}>
                Tree
              </button>
              <button className="btn btn-xs btn-ghost" onClick={() => applyLayout('circle')}>
                Circle
              </button>
              <button className="btn btn-xs btn-ghost" onClick={() => applyLayout('grid')}>
                Grid
              </button>
            </div>
            <div className="flex gap-1">
              <button className="btn btn-xs btn-ghost" onClick={zoomIn}>
                <ZoomIn className="w-3.5 h-3.5" />
              </button>
              <button className="btn btn-xs btn-ghost" onClick={zoomOut}>
                <ZoomOut className="w-3.5 h-3.5" />
              </button>
              <button className="btn btn-xs btn-ghost" onClick={fitToScreen}>
                <Maximize2 className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>

          <div className="bg-base-200 rounded-xl p-3 border border-base-content/5">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-base-content/50">Stats</h3>
            <div className="mt-2 space-y-1 text-xs text-base-content/60">
              <p>Nodes: {elements.filter((e) => !e.data.source).length}</p>
              <p>Edges: {elements.filter((e) => e.data.source).length}</p>
            </div>
          </div>
        </div>

        {/* Graph canvas */}
        <div className="flex-1 bg-base-200 rounded-xl border border-base-content/5 relative min-h-[500px]">
          <div
            ref={containerRef}
            className="w-full h-[calc(100vh-14rem)] min-h-[500px]"
          />
          {elements.length === 0 && !isLoading && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <div className="text-center space-y-2">
                <Server className="w-12 h-12 text-base-content/20 mx-auto" />
                <p className="text-base-content/40 text-sm">
                  No nodes registered yet. The graph will appear here.
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Node detail panel */}
        {selectedNode && (
          <div className="w-72 shrink-0 bg-base-200 rounded-xl border border-base-content/5 p-4 space-y-3 overflow-y-auto max-h-[calc(100vh-14rem)]">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span
                  className="w-3 h-3 rounded-sm"
                  style={{
                    backgroundColor: TYPE_COLORS[selectedNode.data.type]?.color || '#6b7280',
                  }}
                />
                <span className="text-xs font-medium uppercase tracking-wider text-base-content/50">
                  {selectedNode.data.type}
                </span>
              </div>
              <button
                className="btn btn-ghost btn-xs"
                onClick={() => setSelectedNode(null)}
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
            <h3 className="font-semibold text-base">{selectedNode.data.label}</h3>
            <div className="space-y-1.5 text-xs">
              {selectedNode.data.hostname && (
                <DetailRow label="Hostname" value={selectedNode.data.hostname} />
              )}
              {selectedNode.data.status && (
                <DetailRow label="Status" value={selectedNode.data.status} />
              )}
              {selectedNode.data.severity && (
                <DetailRow label="Severity" value={selectedNode.data.severity} />
              )}
              {selectedNode.data.port && selectedNode.data.port > 0 && (
                <DetailRow label="Port" value={String(selectedNode.data.port)} />
              )}
              {selectedNode.data.image && (
                <DetailRow label="Image" value={selectedNode.data.image} />
              )}
              {selectedNode.data.detail && (
                <DetailRow label="Details" value={selectedNode.data.detail} />
              )}
              {selectedNode.data.node_id && (
                <DetailRow label="Node ID" value={selectedNode.data.node_id} />
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-2">
      <span className="text-base-content/50">{label}</span>
      <span className="text-right font-mono max-w-[160px] truncate" title={value}>
        {value}
      </span>
    </div>
  )
}
