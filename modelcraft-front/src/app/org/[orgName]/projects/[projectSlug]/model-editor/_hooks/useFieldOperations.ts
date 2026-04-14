import { useCallback, useMemo, useState } from 'react'
import { createFieldEnumRelation, queryModelEnumContext } from '@bff/model-enum/public'
import { useProjectScopedClient } from '@bff/apollo/public'
import { REMOVE_FIELD } from '@web/graphql/mutations/model'
import { toast } from 'sonner'
import type {
  CreateEnumFieldFormValues,
  CreateEnumLabelFieldFormValues,
  EnumRelationOption,
  EnumSourceOption,
  ModelEnumDomainError,
  UpdateFieldMetaFormValues,
} from '@/types'
import { useCreateEnumFieldPage } from './use-create-enum-field-page'
import { useCreateEnumLabelFieldPage } from './use-create-enum-label-field-page'
import { useEditFieldPage } from './use-edit-field-page'
import type { ModelEditorState } from './useModelEditorState'
import type { EditorModelField } from './types'

interface UseFieldOperationsParams {
  orgName: string
  projectSlug: string
  state: ModelEditorState
}

type FieldPageMode = 'edit' | 'create-enum' | 'create-enum-label'

export function useFieldOperations({ orgName, projectSlug, state }: UseFieldOperationsParams) {
  const projectClient = useProjectScopedClient(projectSlug)
  const modelId = state.editModelData?.id ?? ''

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
      state.setEditModelData((prev) => {
        if (!prev) {
          return prev
        }

        return {
          ...prev,
          fields: prev.fields.map((field) => (field.name === fieldName ? updater(field) : field)),
        }
      })
    },
    [state],
  )

  const appendFieldToEditorState = useCallback(
    (field: EditorModelField) => {
      state.setEditModelData((prev) => {
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
    [state],
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
      const result = await queryModelEnumContext({
        orgName,
        projectSlug,
        modelId,
      })

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
  }, [modelId, orgName, projectSlug])

  const handleToggleDeprecate = async (field: EditorModelField) => {
    if (!state.editModelData) {
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

    try {
      await projectClient.mutate({
        mutation: REMOVE_FIELD,
        variables: {
          modelID: state.editModelData.id,
          fieldName: field.name,
        },
      })

      state.setEditModelData({
        ...state.editModelData,
        fields: state.editModelData.fields.filter((item) => item.name !== field.name),
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

  const createEnumLabelFieldPage = useCreateEnumLabelFieldPage({
    orgName,
    projectSlug,
    modelId,
    onSuccess: (values: CreateEnumLabelFieldFormValues) => {
      appendFieldToEditorState({
        name: values.name,
        title: values.title,
        description: values.description,
        format: 'ENUM_LABEL',
        enumRelationId: values.enumRelationId,
      })
      toast.success('ENUM_LABEL 字段创建成功')
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

  const handleOpenCreateEnumLabelField = () => {
    setFieldPageMode('create-enum-label')
    state.setEditingField(null)
    state.setEditFieldTitle('')
    state.setEditFieldDescription('')
    state.setEditFieldOpen(true)
    createEnumLabelFieldPage.clearError()
    setContextError(null)
    void refreshEnumContext()
  }

  const handleCloseFieldPage = () => {
    state.setEditFieldOpen(false)
    setContextError(null)
  }

  const handleCreateEnumRelation = async (sourceFieldName: string): Promise<string | null> => {
    if (!modelId) {
      return null
    }

    const source = sourceOptions.find((option) => option.fieldName === sourceFieldName)
    if (!source) {
      setContextError({
        type: 'InvalidInput',
        message: `未找到 source 字段 ${sourceFieldName}。`,
        suggestion: '请重新选择 source 字段。',
      })
      return null
    }

    const relationResult = await createFieldEnumRelation({
      orgName,
      projectSlug,
      modelId,
      sourceFieldName,
      enumName: source.enumName,
      labelFieldName: `${sourceFieldName}_label`,
    })

    if (!relationResult.success) {
      setContextError(relationResult.error)
      return null
    }

    const latestContext = await refreshEnumContext()
    if (!latestContext) {
      return null
    }

    const matchedRelation = latestContext.relations.find(
      (relation) => relation.sourceFieldName === sourceFieldName,
    )

    if (!matchedRelation) {
      return null
    }

    toast.success('relation 创建成功')
    return matchedRelation.id
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
    createEnumLabelFieldLoading: createEnumLabelFieldPage.loading,
    createEnumLabelFieldError: createEnumLabelFieldPage.error,
    editFieldLoading: editFieldPage.loading,
    editFieldError: editFieldPage.error,
    handleToggleDeprecate,
    handleRemoveField,
    handleOpenEditField,
    handleSaveField,
    handleOpenCreateEnumField,
    handleOpenCreateEnumLabelField,
    handleCloseFieldPage,
    handleCreateEnumRelation,
    handleSubmitCreateEnumField: createEnumFieldPage.submit,
    handleSubmitCreateEnumLabelField: createEnumLabelFieldPage.submit,
    handleSubmitEditField,
  }
}

export type FieldOperations = ReturnType<typeof useFieldOperations>
