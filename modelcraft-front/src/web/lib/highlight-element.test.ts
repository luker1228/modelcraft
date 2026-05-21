import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { highlightElement, HIGHLIGHT_CLASSES } from './highlight-element'

describe('highlightElement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('adds highlight classes to the element', () => {
    const el = document.createElement('button')
    highlightElement({ current: el })
    for (const cls of HIGHLIGHT_CLASSES) {
      expect(el.classList.contains(cls)).toBe(true)
    }
  })

  it('removes highlight classes after durationMs', () => {
    const el = document.createElement('button')
    highlightElement({ current: el }, 1000)
    vi.advanceTimersByTime(1000)
    for (const cls of HIGHLIGHT_CLASSES) {
      expect(el.classList.contains(cls)).toBe(false)
    }
  })

  it('does not throw when ref.current is null', () => {
    expect(() => highlightElement({ current: null })).not.toThrow()
  })

  it('default duration is 5000ms', () => {
    const el = document.createElement('button')
    highlightElement({ current: el })
    vi.advanceTimersByTime(4999)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(true)
    vi.advanceTimersByTime(1)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(false)
  })

  it('cancels previous timeout when called again before duration elapses', () => {
    const el = document.createElement('button')
    const ref = { current: el }
    highlightElement(ref, 1000)
    // Call again before first timeout fires
    highlightElement(ref, 1000)
    // Advance past original timeout — classes should still be present (reset)
    vi.advanceTimersByTime(999)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(true)
    // Advance past the second timeout
    vi.advanceTimersByTime(1)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(false)
  })
})
