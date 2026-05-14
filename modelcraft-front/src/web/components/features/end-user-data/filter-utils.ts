/**
 * Check if a string is a valid JSON object (not array, not primitive).
 */
export function isValidJson(value: string): boolean {
  if (!value.trim()) return false
  try {
    const parsed = JSON.parse(value)
    return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)
  } catch {
    return false
  }
}

/**
 * Pretty-print JSON. Returns original string on parse failure.
 */
export function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

/**
 * Count top-level filter conditions for the filter button badge.
 *
 * Returns:
 * - number  — length of AND/OR array
 * - '•'     — non-AND/OR structure (single-field condition)
 * - null    — no active filter (empty, invalid JSON, or empty object)
 *
 * AI note: this reads `whereJsonCommitted`. AI agents should write a valid
 * where JSON string and this function will derive the badge automatically.
 */
export function getFilterCount(whereJson: string | null): number | '•' | null {
  if (!whereJson) return null
  try {
    const parsed = JSON.parse(whereJson)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null
    if (Array.isArray(parsed.AND)) return parsed.AND.length || null
    if (Array.isArray(parsed.OR)) return parsed.OR.length || null
    const keys = Object.keys(parsed).filter((k) => k !== 'NOT')
    return keys.length > 0 ? '•' : null
  } catch {
    return null
  }
}
