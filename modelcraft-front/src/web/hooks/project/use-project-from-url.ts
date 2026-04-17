'use client'

import { useEffect, useMemo } from 'react'
import { useParams, useRouter, usePathname } from 'next/navigation'
import { useAppStore } from '@web/stores/app'

/**
 * Hook to extract and sync project information from URL path to Zustand store
 * 
 * Route structure: /org/[orgName]/project/[projectSlug]/*
 * 
 * Features:
 * - URL path is the single source of truth
 * - Syncs URL project to store when path changes
 * - Supports page refresh and link sharing
 * - Type-safe project parameter extraction
 * 
 * @returns Object containing projectId, projectSlug, and orgName
 * 
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { projectSlug, orgName } = useProjectFromUrl()
 *   return <div>Current project: {projectSlug}</div>
 * }
 * ```
 */
export function useProjectFromUrl() {
  const params = useParams()
  const router = useRouter()
  const pathname = usePathname()
  
  const selectedProject = useAppStore((state) => state.selectedProject)
  const setSelectedProject = useAppStore((state) => state.setSelectedProject)

  // Extract route parameters with type safety
  const urlProjectSlug = useMemo(() => {
    return params?.projectSlug as string | undefined
  }, [params?.projectSlug])

  const urlOrgName = useMemo(() => {
    return params?.orgName as string | undefined
  }, [params?.orgName])

  // Sync URL path to store when project changes
  useEffect(() => {
    if (!urlProjectSlug) return

    // Only update if project actually changed
    if (selectedProject?.slug === urlProjectSlug) return

    // TODO: Replace with actual API call to fetch project details
    // For now, create a temporary project object
    setSelectedProject({
      id: urlProjectSlug,
      slug: urlProjectSlug,
      title: urlProjectSlug === 'default' ? 'Default Project' : urlProjectSlug,
      description: '',
      status: 'ACTIVE' as const,
      orgName: urlOrgName || 'default',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
  }, [urlProjectSlug, urlOrgName, selectedProject?.slug, setSelectedProject])

  return useMemo(() => ({
    projectId: selectedProject?.id,
    projectSlug: selectedProject?.slug,
    orgName: urlOrgName,
  }), [selectedProject?.id, selectedProject?.slug, urlOrgName])
}

/**
 * Hook to generate project-scoped URLs
 * 
 * In the new routing structure, project name is part of the path, not a query param
 * 
 * @returns Object with helper functions for generating project URLs
 * 
 * @example
 * ```tsx
 * function NavigationButton() {
 *   const { getLink } = useProjectLink()
 *   return <Link href={getLink('dashboard')}>Dashboard</Link>
 * }
 * ```
 */
export function useProjectLink() {
  const params = useParams()
  const selectedProject = useAppStore((state) => state.selectedProject)

  const projectSlug = useMemo(() => {
    return selectedProject?.slug || (params?.projectSlug as string) || 'default'
  }, [selectedProject?.slug, params?.projectSlug])

  const orgName = useMemo(() => {
    return (params?.orgName as string) || 'default'
  }, [params?.orgName])

  return useMemo(() => ({
    /**
     * Generate full project-scoped path
     * @param path Relative path like 'dashboard', 'clusters', 'models'
     * @returns Full project path like '/org/myorg/project/myproject/dashboard'
     */
    getLink: (path: string): string => {
      const cleanPath = path.startsWith('/') ? path.slice(1) : path
      return `/org/${orgName}/project/${projectSlug}/${cleanPath}`
    },
    
    /**
     * Get project root path
     * @returns Project root path like '/org/myorg/project/myproject'
     */
    getProjectRoot: (): string => {
      return `/org/${orgName}/project/${projectSlug}`
    },
    
    projectSlug,
    orgName,
  }), [projectSlug, orgName])
}

/**
 * Hook to extract current project and organization from URL params
 * 
 * Lightweight alternative to useProjectFromUrl when you only need the params
 * without store synchronization
 * 
 * @returns Object containing projectSlug? and orgName from URL
 * 
 * @example
 * ```tsx
 * function Breadcrumb() {
 *   const { projectSlug, orgName } = useProjectParams()
 *   return <nav>{orgName} / {projectSlug}</nav>
 * }
 * ```
 */
export function useProjectParams() {
  const params = useParams()
  
  return useMemo(() => ({
    projectSlug: params?.projectSlug as string | undefined,
    orgName: params?.orgName as string | undefined,
  }), [params?.projectSlug, params?.orgName])
}
