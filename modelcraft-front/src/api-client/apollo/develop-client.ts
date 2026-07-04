'use client'
/* eslint-disable no-restricted-syntax */

import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'

import { useContext, createContext, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { useOrganizationStore } from '@shared/stores/organization'
import { useAuthStore } from '@shared/stores/auth-store'
import { refreshAccessToken } from '@api-client/auth/public'
import {
  GATEWAY_URL,
  buildRuntimeEndpoint,
  buildProjectScopedEndpoint,
  generateUUID,
  buildXAction,
} from './clients'

/**
 * Creates an Apollo Link that injects the X-Action header required by the backend middleware.
 * Format: "{type}:{operationName}", e.g. "query:EndUserProjects", "mutation:CreateEndUser".
 * Backend validates this header to prevent cross-operation replay attacks.
 */
function createAuthLink() {
  return setContext(async (operation, { headers }: { headers?: Record<string, string> }) => {
    try {
      const store = typeof window !== 'undefined' ? useAuthStore.getState() : null
      let token = store?.accessToken ?? null

      // Proactively refresh if token is missing or expired (before sending)
      if (typeof window !== 'undefined') {
        if (!token || store!.isTokenExpired()) {
          token = await refreshAccessToken()
        }
      }

      const nextHeaders: Record<string, string> = {
        ...(headers ?? {}),
        'x-client-request-id': generateUUID(),
      }

      if (token) {
        nextHeaders.authorization = `Bearer ${token}`
      }

      const xAction = buildXAction(operation)
      if (xAction) {
        nextHeaders['X-Action'] = xAction
      }

      return { headers: nextHeaders }
    } catch (err) {
      console.error('[AuthLink] ERROR:', err)
      return { headers }
    }
  })
}

/**
 * Error link for developer clients.
 * Catches 401 / UNAUTHENTICATED responses (e.g. "Invalid or expired token") and
 * redirects to the developer login page after clearing the local session.
 */
function createDevErrorLink() {
  return onError(({ graphQLErrors, networkError }) => {
    if (typeof window === 'undefined') return

    const is401 =
      networkError && 'statusCode' in networkError && (networkError as { statusCode?: number }).statusCode === 401

    const isAuthError = graphQLErrors?.some((e) => {
      const code = e.extensions?.code as string | undefined
      return code === 'UNAUTHENTICATED' || code === 'AUTH_INVALID_TOKEN' || code === 'AUTH_MISSING_TOKEN'
    })

    if (is401 || isAuthError) {
      useAuthStore.getState().clearAccessToken()
      window.location.href = '/login'
    }
  })
}

/**
 * Org-Scoped Apollo Client
 * Endpoint: /api/bff/graphql/org/{orgName}
 * URI has no trailing slash to match Next.js App Router route handler exactly,
 * avoiding a 308 redirect on every request.
 * Gateway validates the developer Bearer token, strips the Authorization header,
 * and injects X-User-ID before forwarding to the backend. The backend trusts X-User-ID
 * as the authenticated developer identity; it does not perform its own token validation.
 */
function createOrgScopedClient() {
  const httpLink = createHttpLink({
    uri: () => {
      const currentOrg = typeof window !== 'undefined' ? useOrganizationStore.getState().currentOrg : null
      return currentOrg
        ? `${GATEWAY_URL}/api/bff/graphql/org/${currentOrg}`
        : `${GATEWAY_URL}/api/bff/graphql/org`
    },
    credentials: 'include',
  })

  return new ApolloClient({
    link: createDevErrorLink().concat(createAuthLink().concat(httpLink)),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

/**
 * Project-Scoped Apollo Client factory
 * Endpoint: /api/bff/graphql/org/{orgName}/project/{projectSlug}
 * URI has no trailing slash to match Next.js App Router route handler exactly,
 * avoiding a 308 redirect on every request.
 * Shares the path builder with useProjectScopedContext so both are always consistent.
 * Gateway validates the developer Bearer token, strips Authorization, and injects X-User-ID.
 */
export function createProjectScopedClient(
  orgName: string,
  projectSlug: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}${buildProjectScopedEndpoint(orgName, projectSlug)}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createDevErrorLink().concat(createAuthLink().concat(httpLink)),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

export function createDevelopModelRuntimeClient(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): ApolloClient<object> {
  const uri = buildRuntimeEndpoint(orgName, projectSlug, databaseName, modelName)
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createDevErrorLink().concat(createAuthLink().concat(httpLink)),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

const orgScopedClients = new Map<string, ApolloClient<object>>()

/**
 * Get or create an org-scoped ApolloClient.
 * Clients are cached per org name so each org has its own InMemoryCache,
 * preventing cross-org data contamination when navigating between orgs.
 */
export function getOrgScopedClient(orgName?: string): ApolloClient<object> {
  const key = orgName ?? '__default__'
  if (!orgScopedClients.has(key)) {
    orgScopedClients.set(key, createOrgScopedClient())
  }
  return orgScopedClients.get(key)!
}

const projectScopedClients = new Map<string, ApolloClient<object>>()

/**
 * Get or create a project-scoped ApolloClient.
 * Clients are cached per org+project key so all hooks within the same project
 * share a single InMemoryCache.  This ensures that refetchQueries from a
 * mutation in one hook actually invalidates the query cache that another hook
 * is watching — otherwise each hook gets its own client/cache and list views
 * never see mutation results.
 */
function getProjectScopedClient(
  orgName: string,
  projectSlug: string,
): ApolloClient<object> {
  const key = `${orgName}/${projectSlug}`
  if (!projectScopedClients.has(key)) {
    projectScopedClients.set(key, createProjectScopedClient(orgName, projectSlug))
  }
  return projectScopedClients.get(key)!
}

/**
 * Synchronously wipe ALL project-scoped Apollo caches (and the org-scoped cache).
 * Call this when the user clicks the global refresh button so that the
 * subsequent React re-mount (via contentRefreshKey) forces every useQuery
 * to fetch fresh data from the network.
 */
export function clearAllScopedCaches() {
  orgScopedClients.forEach((client) => client.cache.reset())
  projectScopedClients.forEach((client) => client.cache.reset())
}

export function useProjectScopedClient(
  projectSlug: string | null | undefined
): ApolloClient<object> {
  // URL params are the source of truth — they update immediately on navigation.
  // Fall back to the store only when the hook is used outside an org-scoped route.
  const params = useParams()
  const orgNameFromUrl = params?.orgName as string | undefined
  const orgNameFromStore = useOrganizationStore((s) => s.currentOrg)
  const resolvedOrg = orgNameFromUrl ?? orgNameFromStore

  if (!resolvedOrg) {
    throw new Error(
      'useProjectScopedClient: currentOrg is not set in the organization store'
    )
  }

  return useMemo(
    // When projectSlug is null the caller is skipping the query (skip: !projectSlug).
    // Return an org-scoped client as a harmless placeholder so hooks rules are satisfied.
    () => projectSlug
      ? getProjectScopedClient(resolvedOrg, projectSlug)
      : getOrgScopedClient(resolvedOrg),
    [projectSlug, resolvedOrg],
  )
}

const OrgScopedClientContext = createContext<ApolloClient<object> | null>(null)

/**
 * @deprecated Prefer useProjectScopedClient() for project-level operations.
 */
export function useDesignTimeClient(): ApolloClient<object> {
  const client = useContext(OrgScopedClientContext)
  return client ?? getOrgScopedClient()
}

export { OrgScopedClientContext as DesignTimeClientContext }
