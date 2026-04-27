import { useMemo } from 'react'

/**
 * Returns an Apollo operation context that routes requests to the org-scoped
 * BFF endpoint: /api/bff/graphql/org/{orgName}
 */
export function useOrgScopedContext(orgName: string | undefined) {
  return useMemo(() => {
    if (!orgName) return undefined
    return { uri: `/api/bff/graphql/org/${orgName}` }
  }, [orgName])
}

/**
 * Returns an Apollo operation context that routes requests to the project-scoped
 * BFF endpoint: /api/bff/graphql/org/{orgName}/project/{projectSlug}/
 */
export function useProjectScopedContext(orgName: string | undefined, projectSlug: string | undefined) {
  return useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/api/bff/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])
}
