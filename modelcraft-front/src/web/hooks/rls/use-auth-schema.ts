'use client'

import * as React from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { toast } from 'sonner'
import { useProjectScopedClient } from '@api-client/apollo/public'
import { GET_PROJECT_AUTH_SCHEMA } from '@/api-client/rls'
import { SET_PROJECT_AUTH_SCHEMA } from '@/api-client/rls'
import type { AuthVariable, AuthVariableInput } from '@/types/rls'

// ── GraphQL response types ──────────────────────────────────────────

interface ProjectAuthSchemaData {
  variables: AuthVariable[]
}

interface GetProjectAuthSchemaData {
  projectAuthSchema: ProjectAuthSchemaData | null
}

interface SetProjectAuthSchemaPayload {
  authSchema: ProjectAuthSchemaData | null
  error: {
    __typename: string
    message: string
    suggestion?: string
  } | null
}

interface SetProjectAuthSchemaData {
  setProjectAuthSchema: SetProjectAuthSchemaPayload
}

// ── Hook implementation ────────────────────────────────────────────

interface UseAuthSchemaReturn {
  /** 认证变量列表（含内置 uid） */
  authSchema: AuthVariable[]
  /** 加载状态 */
  loading: boolean
  /** 错误信息 */
  error: Error | undefined
  /** 更新认证变量配置 */
  updateAuthSchema: (variables: AuthVariableInput[]) => Promise<boolean>
  /** 更新中状态 */
  updating: boolean
  /** 重新获取 */
  refetch: () => void
}

/**
 * 获取和更新 Project 认证变量配置
 *
 * @param orgName - 组织名称
 * @param projectSlug - 项目 slug
 * @returns UseAuthSchemaReturn
 */
export function useAuthSchema(
  _orgName: string,
  projectSlug: string
): UseAuthSchemaReturn {
  const projectClient = useProjectScopedClient(projectSlug)

  // 查询认证变量配置
  const { data, loading, error, refetch } = useQuery<GetProjectAuthSchemaData>(
    GET_PROJECT_AUTH_SCHEMA,
    {
      client: projectClient,
      skip: !projectSlug,
      onError: (err) => {
        toast.error('获取认证变量配置失败', {
          description: err.message,
        })
      },
    }
  )

  // 更新认证变量配置
  const [setProjectAuthSchema, { loading: updating }] =
    useMutation<SetProjectAuthSchemaData>(SET_PROJECT_AUTH_SCHEMA, {
      client: projectClient,
      onCompleted: (mutationData) => {
        const error = mutationData?.setProjectAuthSchema?.error
        if (error) {
          const description = error.suggestion
            ? `${error.message}。建议: ${error.suggestion}`
            : error.message
          toast.error('更新认证变量配置失败', { description })
        } else {
          toast.success('认证变量配置更新成功')
        }
      },
      onError: (err) => {
        toast.error('更新认证变量配置失败', {
          description: err.message,
        })
      },
    })

  const updateAuthSchema = async (
    variables: AuthVariableInput[]
  ): Promise<boolean> => {
    if (!projectSlug) {
      toast.error('项目信息缺失')
      return false
    }

    try {
      const result = await setProjectAuthSchema({
        variables: {
          input: {
            variables,
          },
        },
      })

      const payload = result.data?.setProjectAuthSchema
      if (payload?.error) {
        return false
      }
      return true
    } catch {
      return false
    }
  }

  const variables = data?.projectAuthSchema?.variables
  const authSchema: AuthVariable[] = React.useMemo(
    () => [
      { name: 'uid', source: 'jwt.user_id', type: 'UUID', isBuiltin: true },
      ...(variables || []),
    ],
    [variables]
  )

  return {
    authSchema,
    loading,
    error,
    updateAuthSchema,
    updating,
    refetch,
  }
}
