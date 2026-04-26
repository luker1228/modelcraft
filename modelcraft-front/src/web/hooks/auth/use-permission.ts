import { useMemo } from 'react'
import { getToken, getUserInfoFromToken } from '@api-client/auth/public'

/**
 * Check if a permission list grants the required permission.
 * Supports three matching modes:
 *  1. Global wildcard: "*" matches any permission
 *  2. Resource wildcard: "resource:*" matches any action on that resource
 *  3. Exact match: "resource:action" matches exactly
 */
function checkPermission(userPermissions: string[], required: string): boolean {
  for (const perm of userPermissions) {
    // Global wildcard
    if (perm === '*') return true
    // Exact match
    if (perm === required) return true
    // Resource wildcard: "project:*" matches "project:create"
    if (perm.endsWith(':*')) {
      const permResource = perm.slice(0, -2)
      const parts = required.split(':')
      if (parts.length === 2 && parts[0] === permResource) return true
    }
  }
  return false
}

/**
 * Get current user's permissions from JWT token.
 * NOTE: Permissions are no longer stored in JWT.
 * This returns an empty array for now.
 * TODO: Fetch permissions from /api/user/permissions endpoint
 */
function getUserPermissions(): string[] {
  if (typeof window === 'undefined') return []
  const token = getToken()
  if (!token) return []
  // Permissions removed from JWT - need to fetch from API
  return []
}

/**
 * Hook to check if the current user has a specific permission.
 *
 * @example
 * const canCreate = usePermission('project:create')
 */
export function usePermission(permission: string): boolean {
  return useMemo(() => {
    const perms = getUserPermissions()
    return checkPermission(perms, permission)
  }, [permission])
}

/**
 * Hook to check if the current user has ANY of the specified permissions.
 *
 * @example
 * const canManage = useHasAnyPermission(['user:invite', 'user:remove'])
 */
export function useHasAnyPermission(permissions: string[]): boolean {
  return useMemo(() => {
    const perms = getUserPermissions()
    return permissions.some((p) => checkPermission(perms, p))
  }, [permissions])
}

/**
 * Hook to check if the current user has ALL of the specified permissions.
 *
 * @example
 * const isAdmin = useHasAllPermissions(['user:invite', 'user:remove', 'user:list'])
 */
export function useHasAllPermissions(permissions: string[]): boolean {
  return useMemo(() => {
    const perms = getUserPermissions()
    return permissions.every((p) => checkPermission(perms, p))
  }, [permissions])
}
