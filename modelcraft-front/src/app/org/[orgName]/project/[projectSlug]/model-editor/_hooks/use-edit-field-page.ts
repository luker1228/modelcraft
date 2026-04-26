import { useCallback, useState } from 'react'
import { updateFieldMeta } from '@api-client/model-enum/public'
import type { ModelEnumDomainError, UpdateFieldMetaFormValues } from '@/types'

interface UseEditFieldPageParams {
  orgName: string
  projectSlug: string
  modelId: string
  fieldName: string
  onSuccess?: (values: UpdateFieldMetaFormValues) => void
}

export interface UseEditFieldPageReturn {
  submit: (values: UpdateFieldMetaFormValues) => Promise<void>
  loading: boolean
  error: ModelEnumDomainError | null
  clearError: () => void
}

const UNKNOWN_ERROR: ModelEnumDomainError = {
  type: 'Unknown',
  code: 'UNKNOWN',
  message: '更新字段失败，请稍后重试。',
}

export function useEditFieldPage({
  orgName,
  projectSlug,
  modelId,
  fieldName,
  onSuccess,
}: UseEditFieldPageParams): UseEditFieldPageReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ModelEnumDomainError | null>(null)

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const submit = useCallback(
    async (values: UpdateFieldMetaFormValues) => {
      if (!fieldName.trim()) {
        setError({
          type: 'InvalidInput',
          message: '缺少字段名称，无法保存。',
          suggestion: '请重新打开字段编辑页后重试。',
        })
        return
      }

      setLoading(true)
      setError(null)

      try {
        const result = await updateFieldMeta({
          orgName,
          projectSlug,
          modelId,
          fieldName,
          title: values.title,
          description: values.description,
          validationConfig: values.validationConfig,
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
    [fieldName, modelId, onSuccess, orgName, projectSlug],
  )

  return {
    submit,
    loading,
    error,
    clearError,
  }
}
