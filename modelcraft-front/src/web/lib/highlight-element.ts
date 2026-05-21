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
 * Options for highlighting an element.
 */
export type HighlightOptions = {
  /** Tooltip message shown near the element (caller is responsible for rendering). */
  message?: string
  /** Duration in milliseconds before highlight is removed. Defaults to 5000. */
  durationMs?: number
  /** Whether to scroll the element into view. Defaults to true. */
  scrollIntoView?: boolean
}

/**
 * Apply amber highlight to a DOM element and auto-remove after durationMs.
 * Second argument is either a legacy durationMs number or a HighlightOptions object.
 * Silently skips if ref.current is null (element unmounted).
 * Cancels any previous pending timeout to prevent memory leaks.
 */
export function highlightElement(
  ref: RefObject<HTMLElement | null>,
  optionsOrDuration: HighlightOptions | number = {},
): void {
  const el = ref.current
  if (!el) return

  // Normalize arguments: support both legacy number and new options object
  const opts: HighlightOptions =
    typeof optionsOrDuration === 'number'
      ? { durationMs: optionsOrDuration }
      : optionsOrDuration

  const { durationMs = 5000, scrollIntoView = true } = opts

  // Cancel any previous highlight timer
  if (activeTimeoutId !== null) {
    clearTimeout(activeTimeoutId)
    activeTimeoutId = null
  }

  // Scroll into view if not visible
  if (scrollIntoView) {
    el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }

  // Apply highlight
  el.classList.add(...HIGHLIGHT_CLASSES)

  // Auto-remove after duration
  activeTimeoutId = setTimeout(() => {
    el.classList.remove(...HIGHLIGHT_CLASSES)
    activeTimeoutId = null
  }, durationMs)
}

/**
 * Highlight a raw HTMLElement directly (used by AiTarget / executeActions).
 */
export function highlightHTMLElement(
  el: HTMLElement,
  optionsOrDuration: HighlightOptions | number = {},
): void {
  highlightElement({ current: el }, optionsOrDuration)
}
