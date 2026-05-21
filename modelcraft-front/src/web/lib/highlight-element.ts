import type { RefObject } from 'react'

/**
 * Tailwind classes applied to highlighted elements.
 * Matches the existing table-row highlight style in DevelopRecordWorkspace.
 */
export const HIGHLIGHT_CLASSES = [
  'bg-amber-50',
  'ring-4',
  'ring-amber-400',
  'ring-offset-4',
  'animate-pulse',
  'transition-all',
] as const

/**
 * Tracks the active highlight timeout so rapid successive calls don't leak.
 */
let activeTimeoutId: ReturnType<typeof setTimeout> | null = null

/**
 * Apply amber highlight to a DOM element and auto-remove after durationMs.
 * Silently skips if ref.current is null (element unmounted).
 * Cancels any previous pending timeout to prevent memory leaks.
 */
export function highlightElement(
  ref: RefObject<HTMLElement | null>,
  durationMs = 5000,
): void {
  const el = ref.current
  if (!el) return

  // Cancel any previous highlight timer
  if (activeTimeoutId !== null) {
    clearTimeout(activeTimeoutId)
    activeTimeoutId = null
  }

  // Scroll into view if not visible
  el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })

  // Apply highlight
  el.classList.add(...HIGHLIGHT_CLASSES)

  // Auto-remove after duration
  activeTimeoutId = setTimeout(() => {
    el.classList.remove(...HIGHLIGHT_CLASSES)
    activeTimeoutId = null
  }, durationMs)
}
