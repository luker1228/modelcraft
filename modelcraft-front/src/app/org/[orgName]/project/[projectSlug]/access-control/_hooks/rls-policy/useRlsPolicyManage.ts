/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */

import { useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import {
  UPSERT_RLS_POLICY,
  DELETE_RLS_POLICY,
  GET_RLS_POLICIES,
  VALIDATE_RLS_EXPR,
} from '@/api-client/rls-policy'
import type { RlsAction, RlsExprType } from '@/generated/graphql'

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
  validateRlsExpression: (input: {
    expression: string
    exprType: RlsExprType
    sampleInput?: string
  }) => Promise<{ success: boolean; message?: string }>
  upserting: boolean
  deleting: boolean
  validating: boolean
}

export function useRlsPolicyManage({ projectSlug, modelId }: UseRlsPolicyManageProps): UseRlsPolicyManageReturn {
  const client = useProjectScopedClient(projectSlug)

  const [upsertMutation, { loading: upserting }] = useMutation(UPSERT_RLS_POLICY, { client })
  const [deleteMutation, { loading: deleting }] = useMutation(DELETE_RLS_POLICY, { client })
  const [validateMutation, { loading: validating }] = useMutation(VALIDATE_RLS_EXPR, { client })

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

  const validateRlsExpression = useCallback(
    async (input: { expression: string; exprType: RlsExprType; sampleInput?: string }) => {
      const result = await validateMutation({
        variables: {
          input: {
            modelId,
            exprType: input.exprType,
            expression: input.expression,
            sampleInput: input.sampleInput ?? null,
          },
        },
      })

      const payload = result.data?.validateRLSExpr
      const error = payload?.error
      if (error) {
        return {
          success: false,
          message: error.suggestion ? `${error.message}：${error.suggestion}` : error.message,
        }
      }

      const validation = payload?.result
      if (!validation?.valid) {
        const firstError = validation?.errors?.[0]
        return {
          success: false,
          message: firstError
            ? `${firstError.path ? `${firstError.path}: ` : ''}${firstError.message}`
            : '表达式校验未通过',
        }
      }

      const dryRun = payload?.dryRun
      if (dryRun?.sql) {
        const rawParams: unknown[] = Array.isArray(dryRun.params) ? [...dryRun.params] : []
        const dryRunParams = rawParams.filter((param): param is string => typeof param === 'string')
        return {
          success: true,
          message: `${dryRun.sql}${dryRunParams.length ? ` | params: ${dryRunParams.join(', ')}` : ''}`,
        }
      }
      if (typeof dryRun?.result === 'boolean') {
        return {
          success: dryRun.result,
          message: dryRun.result ? 'Input Check 返回 true' : 'Input Check 返回 false',
        }
      }

      return { success: true, message: 'Dry run 通过' }
    },
    [validateMutation, modelId],
  )

  return { upsertPolicy, deletePolicy, validateRlsExpression, upserting, deleting, validating }
}
