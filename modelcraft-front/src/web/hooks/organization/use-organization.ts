import { useMemo } from 'react'
import { useParams } from 'next/navigation'
import { getCurrentOrgName } from '@bff/auth/public'

/**
 * Hook to get current organization name from URL params
 * 
 * Use this in org-scoped pages: /org/[orgName]/*
 * 
 * @returns Object containing orgName and isOrgScoped flag
 * 
 * @example
 * ```tsx
 * function OrgHeader() {
 *   const { orgName, isOrgScoped } = useOrganization()
 *   if (!isOrgScoped) return null
 *   return <h1>Organization: {orgName}</h1>
 * }
 * ```
 */
export function useOrganization() {
  const params = useParams()
  const orgName = useMemo(() => {
    return params?.orgName as string | undefined
  }, [params?.orgName])

  return useMemo(() => ({
    orgName: orgName || null,
    isOrgScoped: !!orgName,
  }), [orgName])
}

/**
 * Hook to get user's organization from JWT token
 * 
 * Use this to get the authenticated user's organization without relying on URL
 * 
 * @returns Object containing orgName and hasOrganization flag
 * 
 * @example
 * ```tsx
 * function ProfileWidget() {
 *   const { orgName, hasOrganization } = useUserOrganization()
 *   return <div>Your organization: {orgName || 'None'}</div>
 * }
 * ```
 */
export function useUserOrganization() {
  const orgName = useMemo(() => {
    return getCurrentOrgName()
  }, [])

  return useMemo(() => ({
    orgName,
    hasOrganization: !!orgName,
  }), [orgName])
}

/**
 * Generate organization-scoped URL path
 * 
 * @param orgName Organization name
 * @param path Path after org (e.g., /welcome, /projects)
 * @returns Organization-scoped path like /org/myorg/welcome
 * 
 * @example
 * ```tsx
 * const welcomePath = getOrgPath('myorg', '/welcome')
 * // Returns: /org/myorg/welcome
 * ```
 */
export function getOrgPath(orgName: string, path: string = ''): string {
  const cleanPath = path.startsWith('/') ? path : `/${path}`
  return `/org/${orgName}${cleanPath}`
}

/**
 * Generate project-scoped URL path
 * 
 * @param orgName Organization name
 * @param projectSlug Project name
 * @param path Path after project (e.g., dashboard, clusters)
 * @returns Project-scoped path like /org/myorg/projects/myproject/dashboard
 * 
 * @example
 * ```tsx
 * const dashboardPath = getProjectPath('myorg', 'myproject', 'dashboard')
 * // Returns: /org/myorg/projects/myproject/dashboard
 * ```
 */
export function getProjectPath(
  orgName: string,
  projectSlug?: string,
  path: string = ''
): string {
  const cleanPath = path.startsWith('/') ? path.slice(1) : path
  const pathSuffix = cleanPath ? `/${cleanPath}` : ''
  return `/org/${orgName}/projects/${projectSlug || ''}${pathSuffix}`
}

/**
 * Hook to get both organization and project from URL params
 * 
 * Convenient hook when you need both org and project context
 * 
 * @returns Object containing orgName, projectSlug, and isProjectScoped flag
 * 
 * @example
 * ```tsx
 * function ProjectBreadcrumb() {
 *   const { orgName, projectSlug, isProjectScoped } = useOrgProject()
 *   if (!isProjectScoped) return null
 *   return <nav>{orgName} / {projectSlug}</nav>
 * }
 * ```
 */
export function useOrgProject() {
  const params = useParams()
  
  const orgName = useMemo(() => {
    return params?.orgName as string | undefined
  }, [params?.orgName])

  const projectSlug = useMemo(() => {
    return params?.projectSlug as string | undefined
  }, [params?.projectSlug])

  return useMemo(() => ({
    orgName: orgName || null,
    projectSlug: projectSlug || null,
    isProjectScoped: !!orgName && !!projectSlug,
  }), [orgName, projectSlug])
}

/**
 * Validate organization name format
 * 
 * Valid format: lowercase letters, numbers, hyphens, underscores
 * 
 * @param orgName Organization name to validate
 * @returns true if valid, false otherwise
 * 
 * @example
 * ```tsx
 * if (!isValidOrgName('my-org')) {
 *   throw new Error('Invalid organization name')
 * }
 * ```
 */
export function isValidOrgName(orgName: string): boolean {
  if (!orgName) return false
  // Only allow lowercase letters, numbers, hyphens, and underscores
  return /^[a-z0-9_-]+$/.test(orgName)
}

/**
 * Validate project slug format
 * 
 * Valid format: lowercase letters, numbers, hyphens, underscores
 * 
 * @param projectSlug Project slug to validate
 * @returns true if valid, false otherwise
 * 
 * @example
 * ```tsx
 * if (!isValidProjectSlug('my-project')) {
 *   throw new Error('Invalid project slug')
 * }
 * ```
 */
export function isValidProjectSlug(projectSlug: string): boolean {
  if (!projectSlug) return false
  // Only allow lowercase letters, numbers, hyphens, and underscores
  return /^[a-z0-9_-]+$/.test(projectSlug)
}
