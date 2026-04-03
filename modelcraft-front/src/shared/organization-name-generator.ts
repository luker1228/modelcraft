/**
 * Organization Name Generator
 * 
 * Generates human-readable organization names for Casdoor registration.
 * 
 * - Default: 12 random alphanumeric characters (e.g., "a3k9m2x7p5q1")
 * - Custom: User input + 6 random characters (e.g., "mycompany_x9k2j7")
 * 
 * Uses crypto.getRandomValues() for secure randomness.
 */

const CHARS = 'abcdefghijklmnopqrstuvwxyz0123456789'
const LETTERS = 'abcdefghijklmnopqrstuvwxyz'

/**
 * Generate a random 12-character organization name
 * Format: [a-z][a-z0-9]{11}
 * 
 * @returns Random organization name (e.g., "a3k9m2x7p5q1")
 * 
 * @example
 * generateRandomOrgName() // "a3k9m2x7p5q1"
 * generateRandomOrgName() // "x7m3p9k2n5q8"
 */
export function generateRandomOrgName(): string {
  const array = new Uint8Array(12)
  crypto.getRandomValues(array)
  
  // First character must be a letter
  const result = [LETTERS[array[0] % LETTERS.length]]
  
  // Remaining 11 characters can be any alphanumeric
  for (let i = 1; i < 12; i++) {
    result.push(CHARS[array[i] % CHARS.length])
  }
  
  return result.join('')
}

/**
 * Append a 6-character random suffix to a base name
 * Format: {base}_{random6}
 * 
 * @param base - User-provided base name (1-12 characters)
 * @returns Name with random suffix (e.g., "mycompany_x9k2j7")
 * 
 * @example
 * appendRandomSuffix("acme") // "acme_x9k2j7"
 * appendRandomSuffix("my-company") // "my-company_p3q8r2"
 */
export function appendRandomSuffix(base: string): string {
  const array = new Uint8Array(6)
  crypto.getRandomValues(array)
  
  const suffix = Array.from(array, byte => CHARS[byte % CHARS.length]).join('')
  return `${base}_${suffix}`
}

/**
 * Generate a 6-character random suffix
 * Format: [a-z0-9]{6}
 * 
 * @returns Random 6-character suffix (e.g., "x9k2j7")
 * 
 * @example
 * generateRandomSuffix() // "x9k2j7"
 * generateRandomSuffix() // "p3q8r2"
 */
export function generateRandomSuffix(): string {
  const array = new Uint8Array(6)
  crypto.getRandomValues(array)
  
  return Array.from(array, byte => CHARS[byte % CHARS.length]).join('')
}
