'use client'

import { useQuery, useMutation } from '@apollo/client'
import { toast } from 'sonner'
import { useProjectScopedClient } from '@api-client/apollo/public'
import { GET_MODEL_RLS_POLICY } from '@web/graphql/queries/rls'
import { SET_MODEL_RLS_POLICY } from '@web/graphql/mutations/rls'
import type {
  ModelRLSPolicy,
  SetModelRLSPolicyInput,
} from '@/types/rls'

// ── GraphQL response types ──────────────────────────────────────────

interface GetModelRLSPolicyData {
  modelRLSPolicy: ModelRLSPolicy | null
}

interface SetModelRLSPolicyPayload {
  policy: ModelRLSPolicy | null
  error: {
    __typename: string
    message: string
    suggestion?: string
    path?: string
    variable?: string
  } | null
}

interface SetModelRLSPolicyData {
  setModelRLSPolicy: SetModelRLSPolicyPayload
}

// ── Hook implementation ────────────────────────────────────────────

interface UseRLSPolicyReturn {
  /** 当前策略 */
  policy: ModelRLSPolicy | null | undefined
  /** 加载状态 */
  loading: boolean
  /** 错误信息 */
  error: Error | undefined
  /** 更新策略函数 */
  updatePolicy: (input: SetModelRLSPolicyInput) => Promise<ModelRLSPolicy | null | undefined>
  /** 更新中状态 */
  updating: boolean
  /** 重新获取策略 */
  refetch: () => void
}

/**
 * 获取和更新 Model RLS 策略
 *
 * @param modelId - 模型 ID
 * @returns UseRLSPolicyReturn
 */
export function useRLSPolicy(modelId: string, projectSlug?: string): UseRLSPolicyReturn {
  const projectClient = useProjectScopedClient(projectSlug)

  // 查询策略
  const { data, loading, error, refetch } = useQuery<GetModelRLSPolicyData>(
    GET_MODEL_RLS_POLICY,
    {
      client: projectClient,
      variables: { modelId },
      skip: !modelId,
      onError: (err) => {
        toast.error('获取 RLS 策略失败', {
          description: err.message,
        })
      },
    }
  )

  // 更新策略
  const [setModelRLSPolicy, { loading: updating }] = useMutation<SetModelRLSPolicyData>(
    SET_MODEL_RLS_POLICY,
    {
      client: projectClient,
      onCompleted: (mutationData) => {
        const error = mutationData?.setModelRLSPolicy?.error
        if (error) {
          const description = error.suggestion
            ? `${error.message}。建议: ${error.suggestion}`
            : error.message
          toast.error('更新 RLS 策略失败', { description })
        } else {
          toast.success('RLS 策略更新成功')
        }
      },
      onError: (err) => {
        toast.error('更新 RLS 策略失败', {
          description: err.message,
        })
      },
    }
  )

  const updatePolicy = async (
    input: SetModelRLSPolicyInput
  ): Promise<ModelRLSPolicy | null | undefined> => {
    try {
      const result = await setModelRLSPolicy({
        variables: { input },
      })

      const payload = result.data?.setModelRLSPolicy
      if (payload?.error) {
        return null
      }
      return payload?.policy
    } catch {
      return null
    }
  }

  return {
    policy: data?.modelRLSPolicy,
    loading,
    error,
    updatePolicy,
    updating,
    refetch,
  }
}
