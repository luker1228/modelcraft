'use client'

import { useState, useCallback } from 'react'
import { useMutation } from '@apollo/client'
import { toast } from 'sonner'
import { VALIDATE_RLS_EXPR } from '@web/graphql/mutations/rls'
import type {
  RLSExprType,
  ValidateRLSExprInput,
  ValidationResult,
} from '@/types/rls'

// ── GraphQL response types ──────────────────────────────────────────

interface ValidateRLSExprPayload {
  result: ValidationResult | null
  error: {
    __typename: string
    message: string
    suggestion?: string
    path?: string
  } | null
}

interface ValidateRLSExprData {
  validateRLSExpr: ValidateRLSExprPayload
}

// ── Hook implementation ────────────────────────────────────────────

interface UseRLSExprValidationReturn {
  /** 校验函数 */
  validate: (
    modelId: string,
    exprType: RLSExprType,
    expression: string
  ) => Promise<ValidationResult | null>
  /** 校验中状态 */
  validating: boolean
  /** 上次校验结果 */
  result: ValidationResult | null
  /** 重置结果 */
  resetResult: () => void
}

/**
 * RLS 表达式校验 Hook
 *
 * @returns UseRLSExprValidationReturn
 */
export function useRLSExprValidation(): UseRLSExprValidationReturn {
  const [result, setResult] = useState<ValidationResult | null>(null)

  // 校验 mutation
  const [validateRLSExpr, { loading: validating }] = useMutation<ValidateRLSExprData>(
    VALIDATE_RLS_EXPR,
    {
      onCompleted: (mutationData) => {
        const payload = mutationData?.validateRLSExpr
        if (payload?.error) {
          const description = payload.error.suggestion
            ? `${payload.error.message}。建议: ${payload.error.suggestion}`
            : payload.error.message
          toast.error('表达式校验失败', { description })
        }
      },
      onError: (err) => {
        toast.error('表达式校验失败', {
          description: err.message,
        })
      },
    }
  )

  const validate = useCallback(
    async (
      modelId: string,
      exprType: RLSExprType,
      expression: string
    ): Promise<ValidationResult | null> => {
      const input: ValidateRLSExprInput = {
        modelId,
        exprType,
        expression,
      }

      try {
        const response = await validateRLSExpr({
          variables: { input },
        })

        const payload = response.data?.validateRLSExpr
        if (payload?.error) {
          const errorResult: ValidationResult = {
            valid: false,
            errors: [
              {
                path: payload.error.path || exprType,
                message: payload.error.message,
                code: 'VALIDATION_ERROR',
              },
            ],
          }
          setResult(errorResult)
          return errorResult
        }

        const validationResult = payload?.result || null
        setResult(validationResult)
        return validationResult
      } catch {
        const errorResult: ValidationResult = {
          valid: false,
          errors: [
            {
              path: exprType,
              message: '校验请求失败',
              code: 'REQUEST_ERROR',
            },
          ],
        }
        setResult(errorResult)
        return errorResult
      }
    },
    [validateRLSExpr]
  )

  const resetResult = useCallback(() => {
    setResult(null)
  }, [])

  return {
    validate,
    validating,
    result,
    resetResult,
  }
}
