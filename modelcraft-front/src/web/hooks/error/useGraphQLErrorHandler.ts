import { useCallback } from 'react'
import { ApolloError } from '@apollo/client'
import { useErrorStore } from '@web/stores/error'
import type { GraphQLErrorInfo, GraphQLErrorContext } from '@web/components/common/GraphQLErrorDialog'

export function useGraphQLErrorHandler() {
  const { showErrorDialog } = useErrorStore()

  const handleError = useCallback((
    error: ApolloError,
    operationName?: string,
    operationType?: 'query' | 'mutation' | 'subscription',
    variables?: Record<string, unknown>
  ) => {
    // 转换GraphQL错误格式
    const graphqlErrors: GraphQLErrorInfo[] = error.graphQLErrors.map(err => ({
      message: err.message,
      extensions: err.extensions,
      locations: err.locations?.map(loc => ({
        line: loc.line,
        column: loc.column
      })),
      path: err.path ? [...err.path] : undefined,
    }))

    // 如果有网络错误，也添加到错误列表
    if (error.networkError) {
      graphqlErrors.push({
        message: `网络错误: ${error.networkError.message}`,
        extensions: {
          code: 'NETWORK_ERROR',
        },
      })
    }

    // 构建上下文信息
    const context: GraphQLErrorContext = {
      operationName,
      operationType,
      variables,
      networkError: error.networkError,
      timestamp: new Date().toISOString(),
      userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : undefined,
      url: typeof window !== 'undefined' ? window.location.href : undefined,
    }

    // 显示错误弹窗
    showErrorDialog(graphqlErrors, context)
  }, [showErrorDialog])

  const handleCustomError = useCallback((
    message: string,
    code?: string,
    additionalInfo?: Record<string, unknown>
  ) => {
    const error: GraphQLErrorInfo = {
      message,
      extensions: {
        code: code || 'CUSTOM_ERROR',
        ...additionalInfo,
      },
    }

    const context: GraphQLErrorContext = {
      timestamp: new Date().toISOString(),
      userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : undefined,
      url: typeof window !== 'undefined' ? window.location.href : undefined,
    }

    showErrorDialog([error], context)
  }, [showErrorDialog])

  return {
    handleError,
    handleCustomError,
  }
}

// 全局错误处理函数，可以在Apollo Client中使用
export function createGlobalErrorHandler() {
  return (error: ApolloError) => {
    console.error('GraphQL Error:', error)
    
    // 在开发环境下，自动显示错误弹窗
    if (process.env.NODE_ENV === 'development') {
      const { showErrorDialog } = useErrorStore.getState()
      
      const graphqlErrors: GraphQLErrorInfo[] = error.graphQLErrors.map(err => ({
        message: err.message,
        extensions: err.extensions,
        locations: err.locations?.map(loc => ({
          line: loc.line,
          column: loc.column
        })),
        path: err.path ? [...err.path] : undefined,
      }))

      if (error.networkError) {
        graphqlErrors.push({
          message: `网络错误: ${error.networkError.message}`,
          extensions: {
            code: 'NETWORK_ERROR',
          },
        })
      }

      const context: GraphQLErrorContext = {
        operationName: undefined,
        operationType: undefined,
        variables: undefined,
        networkError: error.networkError,
        timestamp: new Date().toISOString(),
        userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : undefined,
        url: typeof window !== 'undefined' ? window.location.href : undefined,
      }

      showErrorDialog(graphqlErrors, context)
    }
  }
}