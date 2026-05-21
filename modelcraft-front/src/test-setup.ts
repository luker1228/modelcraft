// Minimal DOM polyfill for vitest running in node environment.
// Only provides what the unit tests actually use.
// If jsdom or happy-dom is later added as a dev dependency, remove this file
// and set `test.environment: 'jsdom'` in vitest.config.ts instead.

if (typeof globalThis.document === 'undefined') {
  // Minimal classList implementation.
  const makeClassList = () => {
    const classes = new Set<string>()
    return {
      add: (...tokens: string[]) => tokens.forEach((t) => classes.add(t)),
      remove: (...tokens: string[]) => tokens.forEach((t) => classes.delete(t)),
      contains: (token: string) => classes.has(token),
      toggle: (token: string, force?: boolean) => {
        if (force === undefined ? classes.has(token) : !force) {
          classes.delete(token)
          return false
        }
        classes.add(token)
        return true
      },
    }
  }

  // Minimal createElement mock — returns a plain object shaped like an element.
  const createElement = (tag: string): object => ({
    tagName: tag.toUpperCase(),
    nodeType: 1,
    classList: makeClassList(),
    scrollIntoView: () => { /* no-op in test environment */ },
  })

  Object.defineProperty(globalThis, 'document', {
    value: { createElement },
    writable: true,
    configurable: true,
  })
}
