import { useCallback, useState } from 'react'
import { createEnumField } from '@bff/model-enum/public'
import type { CreateEnumFieldFormValues, ModelEnumDomainError } from '@/types'

interface UseCreateEnumFieldPageParams {
  orgName: string
  projectSlug: string
  modelId: string
  onSuccess?: (values: CreateEnumFieldFormValues) => void
}

export interface UseCreateEnumFieldPageReturn {
  submit: (values: CreateEnumFieldFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
  clearError: () => void
}

const UNKNOWN_ERROR: ModelEnumDomainError = {
  type: 'Unknown',
  code: 'UNKNOWN',
  message: '创建 ENUM 字段失败，请稍后重试。',
}

export function useCreateEnumFieldPage({
  orgName,
  projectSlug,
  modelId,
  onSuccess,
}: UseCreateEnumFieldPageParams): UseCreateEnumFieldPageReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ModelEnumDomainError | null>(null)

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const submit = useCallback(
    async (values: CreateEnumFieldFormValues) => {
      setLoading(true)
      setError(null)

      try {
        const result = await createEnumField({
          orgName,
          projectSlug,
          modelId,
          name: values.name,
          title: values.title,
          description: values.description,
          relateEnumName: values.relateEnumName,
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
