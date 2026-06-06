import { useMemo } from 'react'
import { buildProjectScopedEndpoint } from '@api-client/apollo/clients'

/**
 * Returns an Apollo operation context that routes requests to the org-scoped
 * BFF endpoint: /api/bff/graphql/org/{orgName}
 * URI has no trailing slash to match Next.js App Router route handler exactly.
 */
export function useOrgScopedContext(orgName: string | undefined) {
  return useMemo(() => {
    if (!orgName) return undefined
    return { uri: `/api/bff/graphql/org/${orgName}` }
  }, [orgName])
}

/**
 * Returns an Apollo operation context that routes requests to the project-scoped
 * BFF endpoint: /api/bff/graphql/org/{orgName}/project/{projectSlug}
 * Shares the path builder with createProjectScopedClient so both are always consistent.
 * URI has no trailing slash to match Next.js App Router route handler exactly.
 */
export function useProjectScopedContext(orgName: string | undefined, projectSlug: string | undefined) {
  return useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: buildProjectScopedEndpoint(orgName, projectSlug) }
  }, [orgName, projectSlug])
}
