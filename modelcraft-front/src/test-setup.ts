// Minimal DOM polyfill for vitest running in node environment.
// Only provides what the unit tests actually use.
// If jsdom or happy-dom is later added as a dev dependency, remove this file
// and set `test.environment: 'jsdom'` in vitest.config.ts instead.

if (typeof globalThis.document === 'undefined') {
  // Minimal createElement mock — returns a plain object shaped like an element.
  const createElement = (tag: string): object => ({ tagName: tag.toUpperCase(), nodeType: 1 })
  Object.defineProperty(globalThis, 'document', {
    value: { createElement },
    writable: true,
    configurable: true,
  })
}
