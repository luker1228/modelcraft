'use client'
/* eslint-disable no-restricted-syntax */

import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { useContext, createContext, useMemo } from 'react'
import { useOrganizationStore } from '@shared/stores/organization'
import { generateUUID } from '@/shared/utils/uuid'
import { useAuthStore } from '@shared/stores/auth-store'
import { refreshAccessToken } from '@api-client/auth/public'

// Gateway base URL — empty string means same-origin (when behind a reverse proxy)
const GATEWAY_URL = ''

function isEndUserPath(): boolean {
  if (typeof window === 'undefined') return false
  return window.location.pathname.startsWith('/end-user/')
}

export function buildRuntimeEndpoint(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): string {
  return `${GATEWAY_URL}/api/bff/graphql/org/${orgName}/project/${projectSlug}/db/${databaseName}/model/${modelName}`
}

export function buildEndUserRuntimeEndpoint(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): string {
  return `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}/db/${databaseName}/model/${modelName}`
}

function createAuthLink() {
  return setContext(async (_, { headers }: { headers?: Record<string, string> }) => {
    let token = typeof window !== 'undefined' ? useAuthStore.getState().accessToken : null
    if (!token && typeof window !== 'undefined' && !isEndUserPath()) {
      token = await refreshAccessToken()
    }

    const nextHeaders: Record<string, string> = {
      ...(headers ?? {}),
      'x-client-request-id': generateUUID(),
    }

    if (token) {
      nextHeaders.authorization = `Bearer ${token}`
    }

    return { headers: nextHeaders }
  })
}

/**
 * Org-Scoped Apollo Client
 * Endpoint: /api/bff/graphql/org/{orgName}/
 * Gateway validates the developer Bearer token, strips the Authorization header,
 * and injects X-User-ID before forwarding to the backend. The backend trusts X-User-ID
 * as the authenticated developer identity; it does not perform its own token validation.
 */
function createOrgScopedClient() {
  const httpLink = createHttpLink({
    uri: () => {
      const currentOrg = typeof window !== 'undefined' ? useOrganizationStore.getState().currentOrg : null
      return currentOrg
        ? `${GATEWAY_URL}/api/bff/graphql/org/${currentOrg}/`
        : `${GATEWAY_URL}/api/bff/graphql/org/`
    },
    credentials: 'include',
  })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache({
      typePolicies: {
        Model: { keyFields: ['id'] },
        Field: { keyFields: ['name'] },
      },
    }),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

/**
 * Project-Scoped Apollo Client factory
 * Endpoint: /api/bff/graphql/org/{orgName}/project/{projectSlug}/
 * Creates a fresh instance per project to avoid cache conflicts.
 * Gateway validates the developer Bearer token, strips Authorization, and injects X-User-ID.
 */
export function createProjectScopedClient(
  orgName: string,
  projectSlug: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/org/${orgName}/project/${projectSlug}/`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache({
      typePolicies: {
        Model: { keyFields: ['id'] },
        Field: { keyFields: ['name'] },
      },
    }),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

/**
 * End-User Scoped Apollo Client factory
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}
 * Uses the End-User Bearer token (not the design-time admin token).
 */
export function createEndUserScopedClient(
  orgName: string,
  projectSlug: string,
  endUserToken: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  const authLink = setContext((_, { headers }: { headers?: Record<string, string> }) => ({
    headers: {
      ...(headers ?? {}),
      authorization: `Bearer ${endUserToken}`,
    },
  }))

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

/**
 * End-User Org-Scoped Apollo Client factory
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}/
 * Uses the End-User Bearer token for org-level queries (e.g. findUsers).
 */
export function createEndUserOrgScopedClient(
  orgName: string,
  endUserToken: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  const authLink = setContext((_, { headers }: { headers?: Record<string, string> }) => ({
    headers: {
      ...(headers ?? {}),
      authorization: `Bearer ${endUserToken}`,
    },
  }))

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

export function createModelRuntimeClient(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): ApolloClient<object> {
  const uri = buildRuntimeEndpoint(orgName, projectSlug, databaseName, modelName)
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createAuthLink().concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

/**
 * End-User Runtime Apollo Client factory
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
 * Routes through Gateway (JWT validation + X-User-Type: end_user injection).
 * Uses the End-User Bearer token.
 */
export function createEndUserModelRuntimeClient(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string,
  endUserToken: string
): ApolloClient<object> {
  const uri = buildEndUserRuntimeEndpoint(orgName, projectSlug, databaseName, modelName)
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  const authLink = setContext((_, { headers }: { headers?: Record<string, string> }) => ({
    headers: {
      ...(headers ?? {}),
      authorization: `Bearer ${endUserToken}`,
    },
  }))

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}

let orgScopedClient: ApolloClient<object> | null = null

export function getOrgScopedClient(): ApolloClient<object> {
  if (!orgScopedClient) {
    orgScopedClient = createOrgScopedClient()
  }
  return orgScopedClient
}

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

const OrgScopedClientContext = createContext<ApolloClient<object> | null>(null)

/**
 * @deprecated Prefer useProjectScopedClient() for project-level operations.
 */
export function useDesignTimeClient(): ApolloClient<object> {
  const client = useContext(OrgScopedClientContext)
  return client ?? getOrgScopedClient()
}

export { OrgScopedClientContext as DesignTimeClientContext }
