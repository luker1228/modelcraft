'use client'

import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { useContext, createContext, useMemo } from 'react'
import { useOrganizationStore } from '@shared/stores/organization'
import { generateUUID } from '@/shared/utils/uuid'
import { useAuthStore } from '@shared/stores/auth-store'

/**
 * Build Runtime GraphQL endpoint URL
 * URL format: /graphql/org/:orgName/project/:projectSlug/db/:databaseName/model/:modelName
 */
export function buildRuntimeEndpoint(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): string {
  return `/graphql/org/${orgName}/project/${projectSlug}/db/${databaseName}/model/${modelName}`
}

/**
 * Shared auth link factory — Bearer token + x-request-id header
 */
function createAuthLink() {
  return setContext((_, { headers }: { headers?: Record<string, string> }) => {
    const token = typeof window !== 'undefined' ? useAuthStore.getState().accessToken : null
    return {
      headers: {
        ...(headers ?? {}),
        authorization: token ? `Bearer ${token}` : '',
        'x-request-id': generateUUID(),
      },
    }
  })
}

/**
 * Org-Scoped Apollo Client
 * Endpoint: /graphql/org/{orgName}/
 * Handles: Projects, Clusters, Users, Roles, Organization management
 */
function createOrgScopedClient() {
  const httpLink = createHttpLink({
    uri: (operation) => {
      // Get current organization from store
      const currentOrg = typeof window !== 'undefined' ? useOrganizationStore.getState().currentOrg : null

      // Use org-scoped endpoint if org is available
      if (currentOrg) {
        return `/graphql/org/${currentOrg}/`
      }

      // Fallback when org is not yet loaded
      return '/api/graphql'
    },
  })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache({
      typePolicies: {
        Model: {
          keyFields: ['id'],
        },
        Field: {
          keyFields: ['name'],
        },
      },
    }),
    defaultOptions: {
      watchQuery: {
        errorPolicy: 'all',
      },
      query: {
        errorPolicy: 'all',
      },
      mutate: {
        errorPolicy: 'all',
      },
    },
  })
}

/**
 * Project-Scoped Apollo Client factory
 * Endpoint: /graphql/org/{orgName}/project/{projectSlug}/
 * Handles: Models, Fields, Enums, Logical Foreign Keys
 *
 * Creates a fresh instance per project — NOT a singleton — to avoid cache conflicts
 * between different projects.
 */
export function createProjectScopedClient(
  orgName: string,
  projectSlug: string
): ApolloClient<object> {
  const uri = `/graphql/org/${orgName}/project/${projectSlug}/`

  const httpLink = createHttpLink({ uri })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache({
      typePolicies: {
        Model: {
          keyFields: ['id'],
        },
        Field: {
          keyFields: ['name'],
        },
      },
    }),
    defaultOptions: {
      watchQuery: {
        errorPolicy: 'all',
      },
      query: {
        errorPolicy: 'all',
      },
      mutate: {
        errorPolicy: 'all',
      },
    },
  })
}

/**
 * Create a Runtime Apollo Client for a specific model
 * URL format: /graphql/org/:orgName/project/:projectSlug/db/:databaseName/model/:modelName
 */
export function createModelRuntimeClient(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): ApolloClient<object> {
  const uri = buildRuntimeEndpoint(orgName, projectSlug, databaseName, modelName)

  const httpLink = createHttpLink({ uri })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: {
        errorPolicy: 'all',
      },
      query: {
        errorPolicy: 'all',
      },
      mutate: {
        errorPolicy: 'all',
      },
    },
  })
}

// Org-scoped singleton
let orgScopedClient: ApolloClient<object> | null = null

/**
 * Get or create Org-Scoped Apollo Client (singleton)
 * Use for org-level operations: Projects, Clusters, Users, Roles
 */
export function getOrgScopedClient(): ApolloClient<object> {
  if (!orgScopedClient) {
    orgScopedClient = createOrgScopedClient()
  }
  return orgScopedClient
}

/**
 * Hook to get the appropriate Apollo Client based on context.
 * - If projectSlug is provided: returns a project-scoped client (fresh instance)
 *   pointing to /graphql/org/{orgName}/project/{projectSlug}/
 * - Otherwise: returns the org-scoped singleton
 *   pointing to /graphql/org/{orgName}/
 *
 * @param projectSlug - The project slug for project-scoped operations
 * @param orgNameOverride - Explicit org name. If omitted, falls back to the organization store.
 *   Pass this when orgName is already available from URL params to avoid a stale-store edge case
 *   where currentOrg has not yet been set and the client would fall back to the org-level endpoint.
 */
export function useProjectScopedClient(
  projectSlug?: string,
  orgNameOverride?: string
): ApolloClient<object> {
  const currentOrg = useOrganizationStore((s) => s.currentOrg)
  const resolvedOrg = orgNameOverride ?? currentOrg

  return useMemo(() => {
    if (projectSlug && resolvedOrg) {
      return createProjectScopedClient(resolvedOrg, projectSlug)
    }
    return getOrgScopedClient()
  }, [projectSlug, resolvedOrg])
}

// React Context for org-scoped client
const OrgScopedClientContext = createContext<ApolloClient<object> | null>(null)

/**
 * Hook to access the Org-Scoped Apollo Client.
 * Falls back to singleton if no context provider is present.
 *
 * @deprecated Prefer useProjectScopedClient() for project-level operations.
 * Kept for backward compatibility with existing consumers (ModelRecordWorkspace, InsertFieldSheet, FormRenderer).
 */
export function useDesignTimeClient(): ApolloClient<object> {
  const client = useContext(OrgScopedClientContext)
  if (!client) {
    return getOrgScopedClient()
  }
  return client
}

// Export context for provider usage if needed
export { OrgScopedClientContext as DesignTimeClientContext }
