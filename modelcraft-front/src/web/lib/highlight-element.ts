import type { RefObject } from 'react'

/**
 * Tailwind classes applied to highlighted elements.
 * Matches the existing table-row highlight style in DevelopRecordWorkspace.
 */
export const HIGHLIGHT_CLASSES = [
  'bg-amber-50',
  'ring-2',
  'ring-amber-300',
  'ring-offset-1',
  'transition-all',
] as const

/**
 * Apply amber highlight to a DOM element and auto-remove after durationMs.
 * Silently skips if ref.current is null (element unmounted).
 */
export function highlightElement(
  ref: RefObject<HTMLElement | null>,
  durationMs = 5000,
): void {
  const el = ref.current
  if (!el) return

  // Scroll into view if not visible
  el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })

  // Apply highlight
  el.classList.add(...HIGHLIGHT_CLASSES)

  // Auto-remove after duration
  setTimeout(() => {
    el.classList.remove(...HIGHLIGHT_CLASSES)
  }, durationMs)
}
