'use client'

import { useState } from 'react'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { highlightElement } from '@web/lib/highlight-element'

/**
 * Dev-only floating panel for manually triggering AI capability highlights.
 * Rendered only when NODE_ENV === 'development'.
 *
 * Usage: mount anywhere inside <AICapabilityProvider>.
 * It is already included in AICapabilityProvider — no extra wiring needed.
 */
export function AICapabilityDebugPanel() {
  const [open, setOpen] = useState(false)
  const { getAll } = useAICapabilityContext()

  if (process.env.NODE_ENV !== 'development') return null

  const capabilities = getAll()

  // Hide entirely when this provider has nothing registered —
  // avoids the outer (org-level) panel overlapping the inner (project-level) one.
  if (capabilities.length === 0 && !open) return null

  return (
    <div className="fixed bottom-4 left-4 z-[9999] flex flex-col-reverse items-start gap-1.5">
      {/* Toggle button */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 rounded-full bg-amber-500 px-3 py-1 font-mono text-[11px] font-semibold text-white shadow-lg transition-colors hover:bg-amber-600"
      >
        <span>AI</span>
        <span className="rounded bg-amber-400/60 px-1">{capabilities.length}</span>
        <span>{open ? '▲' : '▼'}</span>
      </button>

      {/* Capability list */}
      {open && (
        <div className="min-w-[200px] rounded-lg border border-amber-200 bg-white p-1.5 shadow-xl">
          <p className="mb-1 px-2 font-mono text-[10px] text-muted-foreground">
            点击高亮元素
          </p>
          {capabilities.length === 0 ? (
            <p className="px-2 py-1 text-xs text-muted-foreground">暂无注册的 capability</p>
          ) : (
            <ul className="flex flex-col gap-0.5">
              {capabilities.map((cap) => (
                <li key={cap.id}>
                  <button
                    type="button"
                    onClick={() => highlightElement(cap.ref)}
                    title={cap.description ?? cap.id}
                    className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-left transition-colors hover:bg-amber-50 active:bg-amber-100"
                  >
                    <span className="text-[8px] text-amber-500">●</span>
                    <span className="text-xs font-medium text-foreground">{cap.label}</span>
                    <span className="ml-auto font-mono text-[10px] text-muted-foreground/60">
                      {cap.id}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
