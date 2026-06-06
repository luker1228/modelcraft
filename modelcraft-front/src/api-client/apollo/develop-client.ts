'use client'
/* eslint-disable no-restricted-syntax */

import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'

import { useContext, createContext, useMemo } from 'react'
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

/** Returns true when the current page is an end-user route (vs. developer route). */
function isEndUserPath(): boolean {
  if (typeof window === 'undefined') return false
  return window.location.pathname.startsWith('/end-user/')
}

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
      if (typeof window !== 'undefined' && !isEndUserPath()) {
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

export function useProjectScopedClient(
  projectSlug: string
): ApolloClient<object> {
  const resolvedOrg = useOrganizationStore((s) => s.currentOrg)

  if (!projectSlug) {
    throw new Error(
      `useProjectScopedClient: projectSlug is required and must be non-empty, got "${projectSlug}"`
    )
  }
  if (!resolvedOrg) {
    throw new Error(
      'useProjectScopedClient: currentOrg is not set in the organization store'
    )
  }

  return useMemo(
    () => createProjectScopedClient(resolvedOrg, projectSlug),
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
