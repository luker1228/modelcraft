'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { isValidOrgName, isValidProjectSlug } from '@web/hooks/organization/use-organization'

interface RouteValidatorProps {
  orgName?: string
  projectSlug?: string
  children: React.ReactNode
}

/**
 * Route parameter validator component
 * 
 * Validates organization and project names in URL paths
 * Redirects to error page if validation fails
 * 
 * Valid format: lowercase letters, numbers, hyphens, underscores
 * Invalid characters will trigger a redirect
 * 
 * @example
 * ```tsx
 * <RouteValidator orgName="my-org" projectSlug="my-project">
 *   <YourComponent />
 * </RouteValidator>
 * ```
 */
export function RouteValidator({ 
  orgName, 
  projectSlug, 
  children 
}: RouteValidatorProps) {
  const router = useRouter()

  useEffect(() => {
    // Validate organization name if provided
    if (orgName && !isValidOrgName(orgName)) {
      console.error(`Invalid organization name: ${orgName}`)
      router.push('/404')
      return
    }

    // Validate project slug if provided
    if (projectSlug && !isValidProjectSlug(projectSlug)) {
      console.error(`Invalid project slug: ${projectSlug}`)
      router.push('/404')
      return
    }
  }, [orgName, projectSlug, router])

  return <>{children}</>
}

/**
 * Hook for client-side route validation
 * 
 * Returns validation status for org and project names
 * Useful for conditional rendering based on route validity
 * 
 * @param orgName Organization name to validate
 * @param projectSlug Project name to validate
 * @returns Object with validation results
 * 
 * @example
 * ```tsx
 * const { isValid, errors } = useRouteValidation(orgName, projectSlug?)
 * if (!isValid) {
 *   return <ErrorPage errors={errors} />
 * }
 * ```
 */
export function useRouteValidation(
  orgName?: string,
  projectSlug?: string
): {
  isValid: boolean
  errors: string[]
} {
  const errors: string[] = []

  if (orgName && !isValidOrgName(orgName)) {
    errors.push(`Invalid organization name: ${orgName}. Only lowercase letters, numbers, hyphens, and underscores are allowed.`)
  }

  if (projectSlug && !isValidProjectSlug(projectSlug)) {
    errors.push(`Invalid project slug: ${projectSlug}. Only lowercase letters, numbers, hyphens, and underscores are allowed.`)
  }

  return {
    isValid: errors.length === 0,
    errors,
  }
}
