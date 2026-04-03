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

  // JWT no longer contains organization info
  // Use the last selected org from localStorage
  const lastOrgId = localStorage.getItem('lastSelectedOrgId')
  if (lastOrgId) {
    // Try to get org name from org memberships cache
    const orgMembershipsCache = localStorage.getItem('org_memberships_cache')
    if (orgMembershipsCache) {
      try {
        const memberships = JSON.parse(orgMembershipsCache) as Array<{ orgId: string; orgName: string }>
        const org = memberships.find((m) => m.orgId === lastOrgId)
        if (org) return org.orgName
      } catch {
        // Ignore parse errors for cache
      }
    }
  }

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
  return '/org-selector' // no org context, let user select one
}

/**
 * Get organization-specific welcome path
 * @returns Organization-scoped welcome path
 */
export function getWelcomePath(): string {
  return getOrgPath('/welcome')
}
