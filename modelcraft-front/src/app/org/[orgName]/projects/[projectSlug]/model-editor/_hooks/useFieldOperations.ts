import { useProjectScopedClient } from '@bff/apollo/public'
import { UPDATE_FIELD, REMOVE_FIELD } from '@web/graphql/mutations/model'
import { toast } from 'sonner'
import type { ModelEditorState } from './useModelEditorState'
import type { EditorModelField } from './types'

interface UseFieldOperationsParams {
  projectSlug: string
  state: ModelEditorState
}

export function useFieldOperations({ projectSlug, state }: UseFieldOperationsParams) {
  const projectClient = useProjectScopedClient(projectSlug)

  const handleToggleDeprecate = async (field: EditorModelField) => {
    if (!state.editModelData) return
    // TODO: Backend does not yet have deprecateField/undeprecateField mutations.
    // Toggle locally for UI preview only.
    const newDeprecated = !field.isDeprecated
    state.setEditModelData({
      ...state.editModelData,
      fields: state.editModelData.fields.map(f =>
        f.name === field.name
          ? { ...f, isDeprecated: newDeprecated }
          : f
      )
    })
    toast.success(newDeprecated ? '字段已废弃' : '已取消废弃')
  }

  const handleRemoveField = async (field: EditorModelField) => {
    if (!field.isDeprecated || !state.editModelData) return
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
        fields: state.editModelData.fields.filter(f => f.name !== field.name)
      })
      toast.success('字段已删除')
    } catch {
      toast.error('删除失败，请重试')
    }
  }

  const handleOpenEditField = (field: EditorModelField) => {
    state.setEditingField(field)
    state.setEditFieldTitle(field.title || '')
    state.setEditFieldDescription(field.description || '')
    state.setEditFieldOpen(true)
  }

  const handleSaveField = () => {
    // TODO: save field modification
    console.log('保存字段:', {
      name: state.editingField?.name,
      title: state.editFieldTitle,
      description: state.editFieldDescription,
    })
    state.setEditFieldOpen(false)
  }

  return {
    handleToggleDeprecate,
    handleRemoveField,
    handleOpenEditField,
    handleSaveField,
  }
}

export type FieldOperations = ReturnType<typeof useFieldOperations>
