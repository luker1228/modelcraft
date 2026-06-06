'use client'
/* eslint-disable no-restricted-syntax */

import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'

import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { refreshEndUserAccessToken } from '@api-client/end-user/end-user-auth-client'
import {
  GATEWAY_URL,
  buildEndUserRuntimeEndpoint,
  buildEndUserProjectScopedEndpoint,
  buildEndUserOrgScopedEndpoint,
  generateUUID,
  buildXAction,
} from './clients'

function createEndUserAuthLink(orgName: string) {
  return setContext(async (operation, { headers }: { headers?: Record<string, string> }) => {
    try {
      let token = typeof window !== 'undefined' ? useEndUserAuthStore.getState().accessToken : null
      if (typeof window !== 'undefined' && useEndUserAuthStore.getState().isTokenExpired()) {
        token = await refreshEndUserAccessToken({ orgName })
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
      console.error('[EndUserAuthLink] ERROR:', err)
      return { headers }
    }
  })
}

/**
 * Error link for end-user clients.
 * Catches 401 / UNAUTHENTICATED responses (e.g. "Invalid or expired token") and
 * redirects to the end-user login page after clearing the local session.
 */
function createEndUserErrorLink(orgName: string) {
  return onError(({ graphQLErrors, networkError }) => {
    if (typeof window === 'undefined') return

    const is401 =
      networkError && 'statusCode' in networkError && (networkError as { statusCode?: number }).statusCode === 401

    const isAuthError = graphQLErrors?.some((e) => {
      const code = e.extensions?.code as string | undefined
      return code === 'UNAUTHENTICATED' || code === 'AUTH_INVALID_TOKEN' || code === 'AUTH_MISSING_TOKEN'
    })

    if (is401 || isAuthError) {
      useEndUserAuthStore.getState().clearSession()
      const loginPath = orgName ? `/end-user/${orgName}/login` : '/end-user/login'
      window.location.href = loginPath
    }
  })
}

/**
 * End-User Scoped Apollo Client factory
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}
 * URI has no trailing slash — matches the pattern of developer clients.
 * Uses the End-User Bearer token (not the design-time admin token).
 */
export function createEndUserScopedClient(
  orgName: string,
  projectSlug: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}${buildEndUserProjectScopedEndpoint(orgName, projectSlug)}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createEndUserErrorLink(orgName).concat(createEndUserAuthLink(orgName).concat(httpLink)),
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
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}
 * URI has no trailing slash — matches the pattern of developer clients.
 * Uses the End-User Bearer token for org-level queries (e.g. findUsers).
 */
export function createEndUserOrgScopedClient(
  orgName: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}${buildEndUserOrgScopedEndpoint(orgName)}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createEndUserErrorLink(orgName).concat(createEndUserAuthLink(orgName).concat(httpLink)),
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
  modelName: string
): ApolloClient<object> {
  const uri = buildEndUserRuntimeEndpoint(orgName, projectSlug, databaseName, modelName)
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  return new ApolloClient({
    link: createEndUserErrorLink(orgName).concat(createEndUserAuthLink(orgName).concat(httpLink)),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}
