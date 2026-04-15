import { useCallback, useState } from 'react'
import { createEnumLabelField } from '@bff/model-enum/public'
import type { CreateEnumLabelFieldFormValues, ModelEnumDomainError } from '@/types'

interface UseCreateEnumLabelFieldPageParams {
  orgName: string
  projectSlug: string
  modelId: string
  onSuccess?: (values: CreateEnumLabelFieldFormValues) => void
}

export interface UseCreateEnumLabelFieldPageReturn {
  submit: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
  clearError: () => void
}

const UNKNOWN_ERROR: ModelEnumDomainError = {
  type: 'Unknown',
  code: 'UNKNOWN',
  message: '创建 ENUM_LABEL 字段失败，请稍后重试。',
}

export function useCreateEnumLabelFieldPage({
  orgName,
  projectSlug,
  modelId,
  onSuccess,
}: UseCreateEnumLabelFieldPageParams): UseCreateEnumLabelFieldPageReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ModelEnumDomainError | null>(null)

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const submit = useCallback(
    async (values: CreateEnumLabelFieldFormValues) => {
      setLoading(true)
      setError(null)

      try {
        const enumRelationID = values.enumRelationId?.trim()
        if (!enumRelationID) {
          setError({
            type: 'InvalidInput',
            message: '创建 ENUM_LABEL 前必须先创建 relation。',
          })
          return
        }

        const result = await createEnumLabelField({
          orgName,
          projectSlug,
          modelId,
          name: values.name,
          title: values.title,
          description: values.description,
          sourceFieldName: values.sourceFieldName,
          enumRelationId: enumRelationID,
        })

        if (!result.success) {
          setError(result.error ?? UNKNOWN_ERROR)
          return
        }

        onSuccess?.(values)
      } catch {
        setError(UNKNOWN_ERROR)
      } finally {
        setLoading(false)
      }
    },
    [modelId, onSuccess, orgName, projectSlug],
  )

  return {
    submit,
    loading,
    error,
    clearError,
  }
}
