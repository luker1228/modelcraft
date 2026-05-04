'use client'

import type { ApolloClient } from '@apollo/client'
import { createContext, useContext } from 'react'

/**
 * RecordAccessAdapter — workspace 数据访问边界。
 *
 * 由 DevelopRecordWorkspace / RuntimeRecordWorkspace 各自在顶层创建并通过 context 注入。
 * 共享的子组件（OneToManyRelationManagerSection、RecordRelationManagerDialog）
 * 通过 useRecordAccessAdapter() 获取，不再直接依赖 workspaceMode 或特定 client factory。
 *
 * develop workspace:
 *   - managementClient = project-scoped developer client
 *   - managementContext = project-scoped URI context
 *   - createRuntimeClient = createModelRuntimeClient(databaseName, modelName)
 *
 * runtime workspace:
 *   - managementClient = end-user scoped client
 *   - managementContext = end-user URI context
 *   - createRuntimeClient = 始终返回 createEndUserScopedClient（忽略 databaseName/modelName）
 */
export interface RecordAccessAdapter {
  /** 用于模型元数据（schema、逻辑外键等）查询的 Apollo client */
  managementClient: ApolloClient<object>
  /** 用于模型元数据查询的 Apollo link context（包含 uri） */
  managementContext: { uri: string }
  /**
   * 创建指定模型的 runtime client（用于 record CRUD 操作）。
   * develop 场景：基于 databaseName/modelName 构建 model-specific endpoint
   * runtime 场景：始终使用 end-user scoped endpoint
   */
  createRuntimeClient: (databaseName: string, modelName: string) => ApolloClient<object>
}

const RecordAccessAdapterContext = createContext<RecordAccessAdapter | null>(null)

export const RecordAccessAdapterProvider = RecordAccessAdapterContext.Provider

/**
 * 获取当前 workspace 注入的 RecordAccessAdapter。
 * 必须在 RecordAccessAdapterProvider 的子树中调用。
 */
export function useRecordAccessAdapter(): RecordAccessAdapter {
  const adapter = useContext(RecordAccessAdapterContext)
  if (!adapter) {
    throw new Error(
      'useRecordAccessAdapter must be called within a RecordAccessAdapterProvider. ' +
      'Make sure the component is rendered inside DevelopRecordWorkspace or RuntimeRecordWorkspace.'
    )
  }
  return adapter
}
