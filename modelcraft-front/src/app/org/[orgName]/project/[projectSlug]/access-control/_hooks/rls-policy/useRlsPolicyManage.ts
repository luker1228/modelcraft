/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */

import { useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import {
  UPSERT_RLS_POLICY,
  DELETE_RLS_POLICY,
  GET_RLS_POLICIES,
} from '@/api-client/rls-policy'
import type { RlsAction } from '@/generated/graphql'

interface UseRlsPolicyManageProps {
  projectSlug: string
  modelId: string
}

interface UpsertInput {
  policyName: string
  action: RlsAction
  role: string
  usingExpr?: string
  withCheckExpr?: string
}

interface UseRlsPolicyManageReturn {
  upsertPolicy: (input: UpsertInput) => Promise<{ success: boolean; errorMessage?: string }>
  deletePolicy: (id: string) => Promise<{ success: boolean; errorMessage?: string }>
  upserting: boolean
  deleting: boolean
}

export function useRlsPolicyManage({ projectSlug, modelId }: UseRlsPolicyManageProps): UseRlsPolicyManageReturn {
  const client = useProjectScopedClient(projectSlug)

  const [upsertMutation, { loading: upserting }] = useMutation(UPSERT_RLS_POLICY, { client })
  const [deleteMutation, { loading: deleting }] = useMutation(DELETE_RLS_POLICY, { client })

  const upsertPolicy = useCallback(
    async (input: UpsertInput) => {
      const result = await upsertMutation({
        variables: {
          modelId,
          input: {
            policyName: input.policyName,
            action: input.action,
            role: input.role,
            usingExpr: input.usingExpr ?? null,
            withCheckExpr: input.withCheckExpr ?? null,
          },
        },
        refetchQueries: [{ query: GET_RLS_POLICIES, variables: { modelId } }],
      })

      const payload = result.data?.upsertRlsPolicy
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message }
      }
      return { success: true }
    },
    [upsertMutation, modelId],
  )

  const deletePolicy = useCallback(
    async (id: string) => {
      const result = await deleteMutation({
        variables: { id },
        refetchQueries: [{ query: GET_RLS_POLICIES, variables: { modelId } }],
      })

      const payload = result.data?.deleteRlsPolicy
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message }
      }
      return { success: true }
    },
    [deleteMutation, modelId],
  )

  return { upsertPolicy, deletePolicy, upserting, deleting }
}
