/**
 * Generate a UUID v4 string with fallback for environments where crypto.randomUUID is not available
 * (non-HTTPS contexts, older browsers, or server-side rendering)
 */
export function generateUUID(): string {
  // Try native crypto.randomUUID if available
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    try {
      return crypto.randomUUID()
    } catch {
      // Fall through to fallback implementation
    }
  }

  // Fallback implementation: generate UUID v4 using crypto.getRandomValues
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    const uuid = new Uint8Array(16)
    crypto.getRandomValues(uuid)
    
    // Set version (4) and variant bits according to RFC 4122
    uuid[6] = (uuid[6] & 0x0f) | 0x40
    uuid[8] = (uuid[8] & 0x3f) | 0x80

    return Array.from(uuid)
      .map((byte, index) => {
        const hex = byte.toString(16).padStart(2, '0')
        if (index === 4 || index === 6 || index === 8 || index === 10) {
          return '-' + hex
        }
        return hex
      })
      .join('')
  }

  // Last resort: generate a random string (not a valid UUID, but sufficient for request IDs)
  const chars = '0123456789abcdef'
  let uuid = ''
  for (let i = 0; i < 36; i++) {
    if (i === 8 || i === 13 || i === 18 || i === 23) {
      uuid += '-'
    } else {
      uuid += chars[Math.floor(Math.random() * 16)]
    }
  }
  return uuid
}
