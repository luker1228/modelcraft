import { useCallback, useMemo, useState } from 'react'
import { queryModelEnumContext } from '@api-client/model-enum/public'
import { useProjectScopedClient } from '@api-client/apollo/public'
import { REMOVE_FIELD } from '@/api-client/model'
import { toast } from 'sonner'
import { isSystemGeneratedLabelField } from '@/shared/model/system-field'
import type {
  CreateEnumFieldFormValues,
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
  UpdateFieldMetaFormValues,
} from '@/types'
import { useCreateEnumFieldPage } from './use-create-enum-field-page'
import { useEditFieldPage } from './use-edit-field-page'
import type { ModelEditorState } from './use-model-editor-state'
import type { EditorModelField } from './types'

interface UseFieldOperationsParams {
  orgName: string
  projectSlug: string
  state: ModelEditorState
}

type FieldPageMode = 'edit' | 'create-enum'

export function useFieldOperations({ orgName, projectSlug, state }: UseFieldOperationsParams) {
  const projectClient = useProjectScopedClient(projectSlug)
  const modelId = state.editModelData?.id ?? ''

  // 解构稳定的 setter 引用，避免 useCallback 依赖整个 state 对象
  const { setEditModelData } = state

  const [fieldPageMode, setFieldPageMode] = useState<FieldPageMode>('edit')
  const [contextLoading, setContextLoading] = useState(false)
  const [contextError, setContextError] = useState<ModelEnumDomainError | null>(null)
  const [sourceOptions, setSourceOptions] = useState<EnumSourceOption[]>([])
  const [relationOptions, setRelationOptions] = useState<EnumRelationOption[]>([])

  const enumOptions = useMemo(
    () => Array.from(new Set(sourceOptions.map((source) => source.enumName))),
    [sourceOptions],
  )

  const updateFieldInEditorState = useCallback(
    (fieldName: string, updater: (field: EditorModelField) => EditorModelField) => {
      setEditModelData((prev) => {
        if (!prev) {
          return prev
        }

        return {
          ...prev,
          fields: prev.fields.map((field) => (field.name === fieldName ? updater(field) : field)),
        }
      })
    },
    [setEditModelData],
  )

  const appendFieldToEditorState = useCallback(
    (field: EditorModelField) => {
      setEditModelData((prev) => {
        if (!prev) {
          return prev
        }

        const duplicated = prev.fields.some((item) => item.name === field.name)
        if (duplicated) {
          return prev
        }

        return {
          ...prev,
          fields: [...prev.fields, field],
        }
      })
    },
    [setEditModelData],
  )

  const refreshEnumContext = useCallback(async () => {
    if (!modelId) {
      setSourceOptions([])
      setRelationOptions([])
      return null
    }

    setContextLoading(true)
    setContextError(null)

    try {
      const result = await queryModelEnumContext(
        { orgName, projectSlug, modelId },
        projectClient,
      )

      if (result.error) {
        setContextError(result.error)
        setSourceOptions([])
        setRelationOptions([])
        return null
      }

      setSourceOptions(result.enumSources)
      setRelationOptions(result.relations)
      return result
    } catch {
      const error: ModelEnumDomainError = {
        type: 'Unknown',
        code: 'UNKNOWN',
        message: '加载 ENUM 上下文失败，请稍后重试。',
      }
      setContextError(error)
      setSourceOptions([])
      setRelationOptions([])
      return null
    } finally {
      setContextLoading(false)
    }
  }, [modelId, orgName, projectSlug, projectClient])

  const handleToggleDeprecate = async (field: EditorModelField) => {
    if (!state.editModelData) {
      return
    }

    if (isSystemGeneratedLabelField(field, state.editModelData.fields)) {
      toast('系统生成字段不可编辑')
      return
    }

    const deprecated = !field.isDeprecated
    updateFieldInEditorState(field.name, (current) => ({
      ...current,
      isDeprecated: deprecated,
    }))
    toast.success(deprecated ? '字段已废弃' : '已取消废弃')
  }

  const handleRemoveField = async (field: EditorModelField) => {
    if (!field.isDeprecated || !state.editModelData) {
      return
    }

    if (isSystemGeneratedLabelField(field, state.editModelData.fields)) {
      toast('系统生成字段不可编辑')
      return
    }

    try {
      await projectClient.mutate({
        mutation: REMOVE_FIELD,
        variables: {
          modelID: state.editModelData.id,
          fieldName: field.name,
        },
        refetchQueries: ['GetModel', 'GetModelJsonSchema'],
      })

      setEditModelData((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          fields: prev.fields.filter((item) => item.name !== field.name),
        }
      })
      toast.success('字段已删除')
    } catch {
      toast.error('删除失败，请重试')
    }
  }

  const createEnumFieldPage = useCreateEnumFieldPage({
    orgName,
    projectSlug,
    modelId,
    onSuccess: (values: CreateEnumFieldFormValues) => {
      appendFieldToEditorState({
        name: values.name,
        title: values.title,
        description: values.description,
        format: 'ENUM',
        enum: {
          name: values.relateEnumName,
        },
      })
      toast.success('ENUM 字段创建成功')
      state.setEditFieldOpen(false)
      void refreshEnumContext()
    },
  })

  const editFieldPage = useEditFieldPage({
    orgName,
    projectSlug,
    modelId,
    fieldName: state.editingField?.name ?? '',
    onSuccess: (values: UpdateFieldMetaFormValues) => {
      if (!state.editingField) {
        return
      }

      updateFieldInEditorState(state.editingField.name, (current) => ({
        ...current,
        title: values.title ?? current.title,
        description: values.description,
      }))

      toast.success('字段保存成功')
      state.setEditFieldOpen(false)
    },
  })

  const handleOpenEditField = (field: EditorModelField) => {
    const allFields = state.editModelData?.fields ?? []
    if (isSystemGeneratedLabelField(field, allFields)) {
      toast('系统生成字段不可编辑')
      return
    }

    setFieldPageMode('edit')
    state.setEditingField(field)
    state.setEditFieldTitle(field.title || '')
    state.setEditFieldDescription(field.description || '')
    state.setEditFieldOpen(true)
    setContextError(null)
  }

  const handleOpenCreateEnumField = () => {
    setFieldPageMode('create-enum')
    state.setEditingField(null)
    state.setEditFieldTitle('')
    state.setEditFieldDescription('')
    state.setEditFieldOpen(true)
    createEnumFieldPage.clearError()
    setContextError(null)
    void refreshEnumContext()
  }

  const handleCloseFieldPage = () => {
    state.setEditFieldOpen(false)
    setContextError(null)
  }

  const handleSubmitEditField = async (values: UpdateFieldMetaFormValues) => {
    await editFieldPage.submit(values)
  }

  const handleSaveField = () => {
    void handleSubmitEditField({
      title: state.editFieldTitle,
      description: state.editFieldDescription,
    })
  }

  return {
    fieldPageMode,
    enumOptions,
    sourceOptions,
    relationOptions,
    contextLoading,
    contextError,
    createEnumFieldLoading: createEnumFieldPage.loading,
    createEnumFieldError: createEnumFieldPage.error,
    editFieldLoading: editFieldPage.loading,
    editFieldError: editFieldPage.error,
    handleToggleDeprecate,
    handleRemoveField,
    handleOpenEditField,
    handleSaveField,
    handleOpenCreateEnumField,
    handleCloseFieldPage,
    handleSubmitCreateEnumField: createEnumFieldPage.submit,
    handleSubmitEditField,
  }
}

export type FieldOperations = ReturnType<typeof useFieldOperations>
