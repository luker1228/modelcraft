/**
 * JWT token utilities for organization context
 */

/**
 * Get organization name from localStorage token
 * Note: JWT no longer contains org info. Get it from localStorage instead.
 * @returns Organization name or null if not found
 */
export function getCurrentOrgName(): string | null {
  if (typeof window === 'undefined') return null

  // Use recorded default org from localStorage
  const defaultOrgName = localStorage.getItem('defaultOrgName')
  if (defaultOrgName) return defaultOrgName

  return null
}

/**
 * Generate organization-scoped path
 * @param path The path to append to the org route
 * @returns Organization-scoped path or fallback path
 */
export function getOrgPath(path: string): string {
  const orgName = getCurrentOrgName()
  if (orgName) {
    return `/org/${orgName}${path}`
  }
  return '/' // no org context, re-evaluate at root
}

/**
 * Get organization-specific welcome path
 * @returns Organization-scoped welcome path
 */
export function getWelcomePath(): string {
  return getOrgPath('/welcome')
}
