'use client'

import { ApolloClient, ApolloError, ApolloProvider, InMemoryCache, createHttpLink, from } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'
import { createGlobalErrorHandler } from '@web/hooks/error/use-graphql-error-handler'
import { removeToken } from '@api-client/auth/public'
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'
import { refreshAccessToken } from '@api-client/auth/public'
import { useOrganizationStore } from '@shared/stores/organization'
import { generateUUID } from '@shared/utils/uuid'
import { useAuthStore } from '@shared/stores/auth-store'
import { buildXAction } from '@api-client/apollo/x-action'

// HTTP链接配置 - 使用动态 URI
// Org endpoint: /graphql/org/{orgName}/
// Handles: Projects, Clusters, Users, Roles, Organization management
const httpLink = createHttpLink({
  uri: (operation) => {
    // Get current organization from store
    const currentOrg = typeof window !== 'undefined' ? useOrganizationStore.getState().currentOrg : null

    // Use org-scoped endpoint if org is available
    if (currentOrg) {
      return `/api/bff/graphql/org/${currentOrg}`
    }

    // Fallback when org is not yet loaded
    return '/api/graphql'
  },
})

// 认证链接（如果需要）
const authLink = setContext(async (operation: { operationName?: string; query: import('@apollo/client').DocumentNode }, { headers }: { headers?: Record<string, string> }) => {
  // 获取认证token（如果有的话）
  let token = typeof window !== 'undefined' ? useAuthStore.getState().accessToken : null
  if (!token && typeof window !== 'undefined') {
    token = await refreshAccessToken()
  }

  const nextHeaders: Record<string, string> = {
    ...headers,
    'x-client-request-id': generateUUID(),
  }

  if (token) {
    nextHeaders.authorization = `Bearer ${token}`
  }

  const xAction = buildXAction(operation)
  if (xAction) {
    nextHeaders['X-Action'] = xAction
  }

  return {
    headers: nextHeaders,
  }
})

// 错误处理链接
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  // Handle 401 network errors - redirect to login
  if (networkError && 'statusCode' in networkError) {
    const statusCode = (networkError as Error & { statusCode?: number }).statusCode
    if (statusCode === 401) {
      console.error('[Auth error]: Token expired or invalid, redirecting to login')
      removeToken()
      if (typeof window !== 'undefined') {
        window.location.href = TENANT_LOGIN_PATH
      }
      return
    }
  }

  if (graphQLErrors) {
    for (const error of graphQLErrors) {
      const { message, locations, path, extensions } = error
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`,
        extensions
      )

      // Handle authentication errors in GraphQL responses
      const code = extensions?.code as string | undefined
      if (code === 'UNAUTHENTICATED' || code === 'AUTH_MISSING_TOKEN' || code === 'AUTH_INVALID_TOKEN') {
        console.error('[Auth error]: Unauthenticated, redirecting to login')
        removeToken()
        if (typeof window !== 'undefined') {
          window.location.href = TENANT_LOGIN_PATH
        }
        return
      }
    }
  }

  if (networkError) {
    console.error(`[Network error]: ${networkError}`)
  }

  // 调用全局错误处理器
  if (typeof window !== 'undefined') {
    const errorHandler = createGlobalErrorHandler()
    if (graphQLErrors || networkError) {
      const apolloError = new ApolloError({
        graphQLErrors: graphQLErrors || [],
        networkError,
      })
      errorHandler(apolloError)
    }
  }
})

// 缓存配置
const cache = new InMemoryCache({
  typePolicies: {
    Project: {
      keyFields: ['id'],
    },
    Cluster: {
      keyFields: ['id'],
    },
    Model: {
      keyFields: ['id'],
    },
    Field: {
      // Field 没有独立的 id，作为 Model 的嵌套字段存储
      // 使用 false 禁用 normalization，让 Field 内联存储在父对象中
      keyFields: false,
    },
    Enum: {
      keyFields: ['id'],
    },
    EnumOption: {
      // EnumOption 没有独立的 id，作为 EnumDefinition 的嵌套字段存储
      keyFields: false,
    },
    Organization: {
      keyFields: ['id'],
    },
    Role: {
      keyFields: ['id'],
    },
  },
})

// 创建 Apollo Client 实例
function makeClient() {
  return new ApolloClient({
    link: from([
      errorLink,
      authLink,
      httpLink,
    ]),
    cache,
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

// Apollo Client 单例
let apolloClient: ApolloClient<Record<string, unknown>> | null = null

function getClient() {
  if (!apolloClient) {
    apolloClient = makeClient()
  }
  return apolloClient
}

/**
 * Reset Apollo Client cache.
 * Call this when switching organizations to ensure fresh data.
 */
export function resetApolloCache() {
  if (apolloClient) {
    apolloClient.resetStore()
  }
}

export function ApolloWrapper({ children }: { children: React.ReactNode }) {
  return (
    <ApolloProvider client={getClient()}>
      {children}
    </ApolloProvider>
  )
}

// 导出客户端供其他地方使用
export { getClient as apolloClient }
