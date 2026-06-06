import { useEffect } from 'react'
import { useLazyQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_LOGICAL_FOREIGN_KEYS, GET_MODEL } from '@/api-client/model'
import {
  CREATE_LOGICAL_FOREIGN_KEY,
  DELETE_LOGICAL_FOREIGN_KEY,
} from '@/api-client/model'
import { toast } from 'sonner'
import type { ModelEditorState } from './use-model-editor-state'
import type {
  EditorModelDetail,
  ForeignKeysQueryData,
  ModelQueryData,
  CreateLogicalForeignKeyResult,
  DeleteLogicalForeignKeyResult,
} from './types'

interface UseForeignKeysParams {
  projectSlug: string
  state: ModelEditorState
}

export function useForeignKeys({ projectSlug, state }: UseForeignKeysParams) {
  const projectClient = useProjectScopedClient(projectSlug)

  const [fetchForeignKeys] = useLazyQuery<ForeignKeysQueryData>(GET_LOGICAL_FOREIGN_KEYS, {
    fetchPolicy: 'network-only',
    client: projectClient,
  })

  const [fetchModelDetail] = useLazyQuery<ModelQueryData>(GET_MODEL, {
    fetchPolicy: 'network-only',
    client: projectClient,
  })

  // Load foreign keys when edit model opens
  useEffect(() => {
    if (state.editModelId && state.editModelOpen && state.editModelData) {
      loadForeignKeys(state.editModelId)
    }
  }, [state.editModelId, state.editModelOpen, state.editModelData]) // eslint-disable-line react-hooks/exhaustive-deps

  // Lazy load ref model fields when ref model is selected
  useEffect(() => {
    if (!state.fkRefModelId || !projectSlug) {
      state.setFkRefModelDetail(null)
      return
    }
    let cancelled = false
    state.setFkRefModelLoading(true)
    fetchModelDetail({ variables: { id: state.fkRefModelId } })
      .then(result => {
        if (cancelled) return
        const m = result.data?.model?.model
        if (m) state.setFkRefModelDetail(m as EditorModelDetail)
      })
      .finally(() => {
        if (!cancelled) state.setFkRefModelLoading(false)
      })
    return () => { cancelled = true }
  }, [state.fkRefModelId, projectSlug]) // eslint-disable-line react-hooks/exhaustive-deps

  const loadForeignKeys = async (modelId: string) => {
    state.setFkLoading(true)
    try {
      const result = await fetchForeignKeys({
        variables: { modelId },
      })
      state.setFkList(result.data?.logicalForeignKeys ?? [])
    } catch {
      state.setFkList([])
    } finally {
      state.setFkLoading(false)
    }
  }

  const handleCreateFK = async () => {
    if (!state.editModelId || !state.fkRefModelId) return
    const validMappings = state.fkMappings.filter(m => m.sourceField && m.targetField)
    if (validMappings.length === 0) return

    state.setFkSubmitting(true)
    try {
      const result = await projectClient.mutate<{ createLogicalForeignKey?: { result?: CreateLogicalForeignKeyResult } }>({
        mutation: CREATE_LOGICAL_FOREIGN_KEY,
        variables: {
          input: {
            modelId: state.editModelId,
            refModelId: state.fkRefModelId,
            sourceFields: validMappings.map(m => m.sourceField),
            targetFields: validMappings.map(m => m.targetField),
          },
        },
      })
      const r = result.data?.createLogicalForeignKey?.result
      if (r?.__typename === 'LogicalForeignKey') {
        toast.success('外键创建成功')
        state.setFkFormOpen(false)
        state.setFkRefModelId('')
        state.setFkMappings([{ sourceField: '', targetField: '' }])
        await loadForeignKeys(state.editModelId)
      } else if (r?.__typename === 'FKColumnsNotFoundError') {
        toast.error(`字段不存在：${(r as { message: string }).message}`)
      } else if (r?.__typename === 'FKFieldCountMismatchError') {
        toast.error('源字段与目标字段数量不匹配')
      }
    } catch {
      toast.error('创建外键失败，请重试')
    } finally {
      state.setFkSubmitting(false)
    }
  }

  const handleDeleteFK = async (pairId: string) => {
    try {
      const result = await projectClient.mutate<{ deleteLogicalForeignKey?: { result?: DeleteLogicalForeignKeyResult } }>({
        mutation: DELETE_LOGICAL_FOREIGN_KEY,
        variables: { pairId },
      })
      const r = result.data?.deleteLogicalForeignKey?.result
      if (r?.__typename === 'DeleteLogicalForeignKeySuccess') {
        toast.success('外键已删除')
        if (state.editModelId) await loadForeignKeys(state.editModelId)
      } else if (r?.__typename === 'FKPairHasRelateFieldsError') {
        toast.error('该外键关联了关系字段，请先删除相关字段')
      } else if (r?.__typename === 'FKNotFoundError') {
        toast.error('外键不存在')
      }
    } catch {
      toast.error('删除外键失败，请重试')
    } finally {
      state.setFkDeleteConfirm(null)
    }
  }

  return {
    loadForeignKeys,
    handleCreateFK,
    handleDeleteFK,
  }
}

export type ForeignKeyOps = ReturnType<typeof useForeignKeys>
