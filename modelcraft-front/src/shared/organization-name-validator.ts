/**
 * Organization Name Validator
 * 
 * Validates organization names for user registration.
 * 
 * Validation Rules:
 * - Default: 12 alphanumeric chars, starts with letter
 * - Custom: 1-12 char base + underscore + 6 random chars
 * - Only lowercase letters, digits, hyphens
 * - No reserved names (admin, api, etc.)
 */

/**
 * Reserved organization names that cannot be used
 * These are reserved for system routes and functionality
 */
export const RESERVED_NAMES = [
  'admin',
  'api',
  'auth',
  'login',
  'logout',
  'register',
  'dashboard',
  'settings',
  'built-in',
  'default',
  'system',
  'help',
  'docs',
  'support',
  'about',
  'terms',
  'privacy',
  'security',
  'legal',
  'contact',
  'public',
  'assets',
  'static',
  'www',
  'app',
  'internal',
  'root'
]

/**
 * Result of organization name validation
 */
export interface ValidationResult {
  valid: boolean
  error?: string
}

/**
 * Validate default organization name format
 * Format: ^[a-z][a-z0-9]{11}$ (12 chars total)
 * 
 * @param name - Organization name to validate
 * @returns True if valid default name format
 * 
 * @example
 * isValidDefaultName("a3k9m2x7p5q1") // true
 * isValidDefaultName("x7m3p9k2n5q8") // true
 * isValidDefaultName("123invalid") // false (starts with digit)
 * isValidDefaultName("short") // false (too short)
 */
export function isValidDefaultName(name: string): boolean {
  // 12 chars, starts with letter, only lowercase alphanumeric
  return /^[a-z][a-z0-9]{11}$/.test(name)
}

/**
 * Validate custom organization name base (before suffix)
 * Format: ^[a-z][a-z0-9-]{0,11}$ (1-12 chars)
 * 
 * @param base - Base name to validate (before underscore)
 * @returns True if valid custom base format
 * 
 * @example
 * isValidCustomBase("mycompany") // true
 * isValidCustomBase("my-company") // true
 * isValidCustomBase("a") // true (single letter)
 * isValidCustomBase("123company") // false (starts with digit)
 * isValidCustomBase("My-Company") // false (uppercase)
 * isValidCustomBase("my_company") // false (underscore not allowed in base)
 */
export function isValidCustomBase(base: string): boolean {
  // 1-12 chars, starts with letter, lowercase alphanumeric or hyphens
  return /^[a-z][a-z0-9-]{0,11}$/.test(base)
}

/**
 * Validate final organization name (default or custom with suffix)
 * Accepts either:
 * - Default: 12 alphanumeric chars starting with letter
 * - Custom: 1-12 char base + underscore + 6 char suffix
 * 
 * @param name - Full organization name to validate
 * @returns True if valid final name format
 * 
 * @example
 * isValidFinalName("a3k9m2x7p5q1") // true (default)
 * isValidFinalName("mycompany_x9k2j7") // true (custom)
 * isValidFinalName("my-company_p3q8r2") // true (custom with hyphen)
 * isValidFinalName("mycompany") // false (no suffix)
 * isValidFinalName("MyCompany_x9k2j7") // false (uppercase)
 */
export function isValidFinalName(name: string): boolean {
  // Default: 12 chars OR custom: {1-12}_{6}
  return isValidDefaultName(name) || /^[a-z][a-z0-9-]{0,11}_[a-z0-9]{6}$/.test(name)
}

/**
 * Check if organization name base is reserved
 * 
 * @param name - Organization name to check (extracts base if has suffix)
 * @returns True if name uses reserved keyword
 * 
 * @example
 * isReservedName("admin") // true
 * isReservedName("admin_x9k2j7") // true (base is reserved)
 * isReservedName("mycompany_x9k2j7") // false
 */
export function isReservedName(name: string): boolean {
  const base = name.split('_')[0]
  return RESERVED_NAMES.includes(base)
}

/**
 * Comprehensive organization name validation
 * Checks format, length, and reserved names
 * 
 * @param name - Organization name to validate
 * @returns Validation result with error message if invalid
 * 
 * @example
 * validateOrgName("a3k9m2x7p5q1") // { valid: true }
 * validateOrgName("mycompany_x9k2j7") // { valid: true }
 * validateOrgName("") // { valid: false, error: "Organization name is required" }
 * validateOrgName("123invalid") // { valid: false, error: "Invalid name format" }
 * validateOrgName("admin_x9k2j7") // { valid: false, error: "..." }
 */
export function validateOrgName(name: string): ValidationResult {
  if (!name) {
    return { 
      valid: false, 
      error: 'Organization name is required' 
    }
  }
  
  if (!isValidFinalName(name)) {
    return { 
      valid: false, 
      error: 'Invalid name format. Name must start with a letter and contain only lowercase letters, digits, and hyphens.' 
    }
  }
  
  if (isReservedName(name)) {
    const base = name.split('_')[0]
    return { 
      valid: false, 
      error: `"${base}" is a reserved name. Try "${base}-org" or "${base}team" instead.` 
    }
  }
  
  return { valid: true }
}

/**
 * Validate custom base name for user input
 * Provides detailed error messages for user feedback
 * 
 * @param base - Custom base name to validate (before suffix)
 * @returns Validation result with specific error message
 * 
 * @example
 * validateCustomBase("mycompany") // { valid: true }
 * validateCustomBase("") // { valid: false, error: "..." }
 * validateCustomBase("123company") // { valid: false, error: "..." }
 * validateCustomBase("admin") // { valid: false, error: "..." }
 */
export function validateCustomBase(base: string): ValidationResult {
  if (!base) {
    return { 
      valid: false, 
      error: 'Organization name base is required' 
    }
  }
  
  if (base.length > 12) {
    return { 
      valid: false, 
      error: 'Organization name base must be 12 characters or less' 
    }
  }
  
  if (!/^[a-z]/.test(base)) {
    return { 
      valid: false, 
      error: 'Organization name must start with a lowercase letter' 
    }
  }
  
  if (!/^[a-z0-9-]+$/.test(base)) {
    return { 
      valid: false, 
      error: 'Organization name can only contain lowercase letters, digits, and hyphens' 
    }
  }
  
  if (!isValidCustomBase(base)) {
    return { 
      valid: false, 
      error: 'Invalid organization name format' 
    }
  }
  
  if (RESERVED_NAMES.includes(base)) {
    return { 
      valid: false, 
      error: `"${base}" is a reserved name. Try "${base}-org" or "${base}team" instead.` 
    }
  }
  
  return { valid: true }
}
