import { useCallback, useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useLazyQuery } from '@apollo/client'
import { ApolloError } from '@apollo/client/errors'
import { useRouter } from 'next/navigation'
import { useProjectScopedClient, getOrgScopedClient } from '@api-client/apollo/public'
import { TEST_CLUSTER_CONNECTION } from '@/api-client/cluster'
import { CREATE_MODEL, UPDATE_MODEL, DELETE_MODEL } from '@/api-client/model'
import { GET_MODEL, GET_MODELS, GET_MODELS_FOR_RELATION } from '@/api-client/model'
import { useDatabases } from '@web/hooks/database/use-databases'
import { toast } from 'sonner'
import type { ModelEditorState } from './use-model-editor-state'
import type {
  EditorModel,
  EditorModelDetail,
  TestConnectionResult,
  CreateModelResult,
  DeleteModelResult,
  UpdateModelMetaResult,
  ModelsQueryData,
  ModelQueryData,
} from './types'

interface UseModelCRUDParams {
  orgName: string
  projectSlug: string
  state: ModelEditorState
}

function getApolloErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApolloError) {
    const graphQLErrorMessage = error.graphQLErrors.find((item) => !!item.message)?.message
    if (graphQLErrorMessage) return graphQLErrorMessage

    if (error.networkError) {
      return error.networkError.message || fallback
    }
  }

  if (error instanceof Error && error.message) {
    return error.message
  }

  return fallback
}

export function useModelCRUD({ orgName, projectSlug, state }: UseModelCRUDParams) {
  const router = useRouter()
  const projectClient = useProjectScopedClient(projectSlug, orgName)
  const orgClient = getOrgScopedClient()
  const [relationModelsCache, setRelationModelsCache] = useState<Record<string, EditorModel[]>>({})
  const [relationModelsLoading, setRelationModelsLoading] = useState<Record<string, boolean>>({})
  const [relationModelsLoaded, setRelationModelsLoaded] = useState<Record<string, boolean>>({})

  // Connection check
  useEffect(() => {
    if (!projectSlug || !orgName) return

    let cancelled = false

    const check = async () => {
      state.setConnectionChecking(true)
      try {
        const result = await orgClient.mutate<TestConnectionResult>({
          mutation: TEST_CLUSTER_CONNECTION,
          variables: { input: { projectSlug } },
        })
        if (cancelled) return
        const payload = result.data?.testDatabaseConnection
        if (!payload?.success) {
          const msg = payload?.error?.message ?? '数据库连接失败'
          state.setConnectionError(msg)
          state.setConnectionFailed(true)
        }
      } catch {
        if (cancelled) return
        state.setConnectionError('无法连接到数据库集群')
        state.setConnectionFailed(true)
      } finally {
        if (!cancelled) state.setConnectionChecking(false)
      }
    }

    check()
    return () => { cancelled = true }
  }, [projectSlug, orgName]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleGoToCluster = () => {
    router.push(`/org/${orgName}/project/${projectSlug}/cluster`)
  }

  // Fetch databases (skip when connection checking or failed)
  const { databases, loading: databasesLoading } = useDatabases(
    !state.connectionChecking && !state.connectionFailed ? projectSlug : null,
    { initialLimit: 50 }
  )
  const relationDatabaseNames = useMemo(
    () => databases.map((db) => db.name).sort(),
    [databases]
  )

  // Set default selected database when data loads
  useEffect(() => {
    const firstDatabaseName = databases[0]?.name
    if (firstDatabaseName && !state.selectedDatabase) {
      state.setSelectedDatabase(firstDatabaseName)
    }
  }, [databases, state.selectedDatabase]) // eslint-disable-line react-hooks/exhaustive-deps

  // Fetch models
  const { data: modelsData, loading: modelsLoading, refetch: refetchModels } = useQuery<ModelsQueryData>(GET_MODELS, {
    variables: {
      input: {
        databaseName: state.selectedDatabase,
        limit: 100,
      },
    },
    skip: !projectSlug || !state.selectedDatabase || state.connectionChecking || state.connectionFailed,
    client: projectClient,
  })

  const models: EditorModel[] = useMemo(() => {
    if (!modelsData?.models?.edges) return []
    return modelsData.models.edges.map((edge) => edge.node)
  }, [modelsData])

  useEffect(() => {
    // project 切换后清空 relation cache，避免串项目数据
    setRelationModelsCache({})
    setRelationModelsLoading({})
    setRelationModelsLoaded({})
  }, [projectSlug])

  const getRelationModelsForDatabase = useCallback(
    (databaseName: string): EditorModel[] => relationModelsCache[databaseName] ?? [],
    [relationModelsCache]
  )

  const isRelationModelsLoading = useCallback(
    (databaseName: string): boolean => !!relationModelsLoading[databaseName],
    [relationModelsLoading]
  )

  const loadRelationModelsForDatabase = useCallback(
    async (databaseName: string) => {
      if (!projectSlug || !databaseName || state.connectionChecking || state.connectionFailed || databasesLoading) {
        return
      }
      if (relationModelsLoaded[databaseName] || relationModelsLoading[databaseName]) {
        return
      }

      setRelationModelsLoading((prev) => ({ ...prev, [databaseName]: true }))
      try {
        const result = await projectClient.query<ModelsQueryData>({
          query: GET_MODELS_FOR_RELATION,
          variables: {
            input: {
              databaseName,
              limit: 100,
            },
          },
          fetchPolicy: 'network-only',
        })

        const modelsInDatabase = result.data?.models?.edges?.map((edge) => edge.node) ?? []
        setRelationModelsCache((prev) => ({ ...prev, [databaseName]: modelsInDatabase }))
        setRelationModelsLoaded((prev) => ({ ...prev, [databaseName]: true }))
      } catch {
        // 保持 loaded=false，允许用户重试
      } finally {
        setRelationModelsLoading((prev) => ({ ...prev, [databaseName]: false }))
      }
    },
    [
      projectSlug,
      state.connectionChecking,
      state.connectionFailed,
      databasesLoading,
      relationModelsLoaded,
      relationModelsLoading,
      projectClient,
    ]
  )

  const filteredModels = useMemo(() => {
    if (!state.searchQuery) return models
    const query = state.searchQuery.toLowerCase()
    return models.filter(m =>
      m.name.toLowerCase().includes(query) ||
      m.title?.toLowerCase().includes(query)
    )
  }, [models, state.searchQuery])

  // Keep selected model in sync with current database's model list.
  // After switching database, auto select the first available model.
  useEffect(() => {
    if (!state.selectedDatabase) {
      if (state.selectedModelId !== null) {
        state.setSelectedModelId(null)
      }
      return
    }

    if (models.length === 0) {
      if (state.selectedModelId !== null) {
        state.setSelectedModelId(null)
      }
      return
    }

    const hasSelectedModel = !!state.selectedModelId && models.some((model) => model.id === state.selectedModelId)
    if (!hasSelectedModel && state.selectedModelId !== models[0].id) {
      state.setSelectedModelId(models[0].id)
    }
  }, [models, state.selectedDatabase, state.selectedModelId]) // eslint-disable-line react-hooks/exhaustive-deps

  // Lazy load model detail
  const [fetchModelDetail] = useLazyQuery<ModelQueryData>(GET_MODEL, {
    fetchPolicy: 'network-only',
    client: projectClient,
  })

  // Create model
  const handleConfirmCreateModel = async () => {
    if (!state.newModelName.trim() || !state.newModelTitle.trim()) {
      toast.error('请填写模型标识和展示名称')
      return
    }
    if (!state.selectedDatabase || !projectSlug) {
      toast.error('缺少必要参数')
      return
    }

    state.setCreating(true)
    try {
      const result = await projectClient.mutate<CreateModelResult>({
        mutation: CREATE_MODEL,
        variables: {
          input: {
            name: state.newModelName.trim(),
            title: state.newModelTitle.trim(),
            databaseName: state.selectedDatabase,
          },
        },
      })

      const graphQLErrorMessage = result.errors?.find((item) => !!item.message)?.message
      if (graphQLErrorMessage) {
        toast.error(graphQLErrorMessage)
        return
      }

      if (result.data?.createModel?.model) {
        const modelId = result.data.createModel.model.id
        state.setCreateModelOpen(false)
        state.setNewModelName('')
        state.setNewModelTitle('')
        refetchModels()
        state.setSelectedModelId(modelId)
      } else if (result.data?.createModel?.error) {
        toast.error(result.data.createModel.error.message || '创建失败')
      } else {
        toast.error('创建失败')
      }
    } catch (error) {
      console.error('创建模型失败:', error)
      toast.error(getApolloErrorMessage(error, '创建模型失败'))
    } finally {
      state.setCreating(false)
    }
  }

  // Edit model - open drawer
  const handleEditModel = async (modelId: string) => {
    state.setEditModelId(modelId)
    state.setEditModelOpen(true)
    state.setEditModelLoading(true)
    state.setEditModelData(null)

    try {
      const { data } = await fetchModelDetail({
        variables: { id: modelId, withActualSchema: true },
      })

      if (data?.model?.model) {
        state.setEditModelData(data.model.model as EditorModelDetail)
        state.setMetaTitle(data.model.model.title || '')
        state.setMetaDescription(data.model.model.description || '')
        state.setMetaDisplayField(data.model.model.displayField || '')
        state.setFkList([])
        state.setFkFormOpen(false)
        state.setFkMappings([{ sourceField: '', targetField: '' }])
        state.setFkRefModelId('')
        // loadForeignKeys is handled in useForeignKeys hook via effect
      } else if (data?.model?.error) {
        toast.error(data.model.error.message || '获取模型详情失败')
        state.setEditModelOpen(false)
      }
    } catch (error) {
      console.error('获取模型详情失败:', error)
      toast.error(getApolloErrorMessage(error, '获取模型详情失败'))
      state.setEditModelOpen(false)
    } finally {
      state.setEditModelLoading(false)
    }
  }

  // Close edit model drawer
  const handleCloseEditModel = () => {
    state.setEditModelOpen(false)
    state.setEditModelId(null)
    state.setEditModelData(null)
    state.setMetaDisplayField('')
    state.setMetaEditMode(false)
    state.setFkList([])
    state.setFkFormOpen(false)
    state.setFkRefModelId('')
    state.setFkMappings([{ sourceField: '', targetField: '' }])
    state.setFkDeleteConfirm(null)
    state.setFkRefModelDetail(null)
  }

  // Delete model
  const handleDeleteModel = async () => {
    if (!state.modelToDelete) return
    state.setDeletingModel(true)
    try {
      const result = await projectClient.mutate<DeleteModelResult>({
        mutation: DELETE_MODEL,
        variables: { id: state.modelToDelete.id },
      })

      if (result.data?.deleteModel?.success) {
        toast.success('模型删除成功')
        state.setDeleteModelDialogOpen(false)
        state.setModelToDelete(null)
        if (state.selectedModelId === state.modelToDelete.id) {
          state.setSelectedModelId(null)
          handleCloseEditModel()
        }
        refetchModels()
      } else {
        const err = result.data?.deleteModel?.error
        toast.error(err?.message || '删除模型失败')
      }
    } catch (error) {
      console.error('删除模型失败:', error)
      toast.error('删除模型失败')
    } finally {
      state.setDeletingModel(false)
    }
  }

  // Save meta
  const handleSaveMeta = async () => {
    if (!state.editModelId || !projectSlug) return
    state.setMetaSaving(true)
    try {
      const result = await projectClient.mutate<UpdateModelMetaResult>({
        mutation: UPDATE_MODEL,
        variables: {
          id: state.editModelId,
          input: {
            title: state.metaTitle,
            description: state.metaDescription,
            displayField: state.metaDisplayField || '',
          },
        },
      })
      if (result.data?.updateModelMeta?.model) {
        state.setEditModelData(prev => prev
          ? {
            ...prev,
            title: state.metaTitle,
            description: state.metaDescription,
            displayField: state.metaDisplayField,
          }
          : prev,
        )
        toast.success('保存成功')
      } else {
        toast.error(result.data?.updateModelMeta?.error?.message || '保存失败')
      }
    } catch {
      toast.error('保存失败，请重试')
    } finally {
      state.setMetaSaving(false)
    }
  }

  // Refresh model detail (used by InsertFieldSheet onSuccess)
  const refreshModelDetail = async () => {
    if (state.editModelId) {
      const { data } = await fetchModelDetail({ variables: { id: state.editModelId, withActualSchema: true } })
      if (data?.model?.model) {
        state.setEditModelData(data.model.model as EditorModelDetail)
        state.setMetaTitle(data.model.model.title || '')
        state.setMetaDescription(data.model.model.description || '')
        state.setMetaDisplayField(data.model.model.displayField || '')
      }
    }
  }

  return {
    // Data
    databases,
    databasesLoading,
    models,
    relationDatabaseNames,
    getRelationModelsForDatabase,
    loadRelationModelsForDatabase,
    isRelationModelsLoading,
    filteredModels,
    modelsLoading,
    refetchModels,
    fetchModelDetail,

    // Handlers
    handleGoToCluster,
    handleConfirmCreateModel,
    handleEditModel,
    handleCloseEditModel,
    handleDeleteModel,
    handleSaveMeta,
    refreshModelDetail,
  }
}

export type ModelCRUD = ReturnType<typeof useModelCRUD>
