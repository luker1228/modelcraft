import { useQuery, useMutation } from '@apollo/client'
import { GET_MODELS, GET_MODEL_GROUPS } from '@/api-client/model'
import { CREATE_MODEL, UPDATE_MODEL, DELETE_MODEL } from '@/api-client/model'
import { useModelStore, useAppStore } from '@web/stores'
import type { Model, ModelGroup, CreateModelInput, UpdateModelMetaInput } from '@/types'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'

// ── GraphQL response types ──────────────────────────────────────────

interface ModelEdge {
  node: Model
  cursor: string
}

interface ModelConnection {
  edges: ModelEdge[]
  pageInfo: {
    hasNextPage: boolean
    hasPreviousPage: boolean
    startCursor: string | null
    endCursor: string | null
  }
  totalCount: number
}

interface GetModelsData {
  models: ModelConnection
}

interface GetModelGroupsData {
  modelGroups: ModelGroup[]
}

interface CreateModelPayload {
  model: Model | null
  error: { __typename: string; message: string } | null
}

interface UpdateModelMetaPayload {
  success: boolean
  model: Model | null
  error: { __typename: string; message: string } | null
}

interface DeleteModelPayload {
  success: boolean
  error: { __typename: string; message: string } | null
}

interface CreateModelData {
  createModel: CreateModelPayload
}

interface UpdateModelMetaData {
  updateModelMeta: UpdateModelMetaPayload
}

interface DeleteModelData {
  deleteModel: DeleteModelPayload
}

// ── Hook implementations ────────────────────────────────────────────

export function useModels() {
  const selectedProject = useAppStore((state) => state.selectedProject)
  const { setModels, addModel, updateModel, removeModel } = useModelStore()
  const { markStep } = useOnboarding()

  // 查询模型列表
  const { data, loading, error, refetch } = useQuery<GetModelsData>(GET_MODELS, {
    variables: {
      input: {
        databaseName: selectedProject?.databaseName || '',
      },
    },
    skip: !selectedProject?.slug,
    onCompleted: (queryData) => {
      if (queryData?.models?.edges) {
        const models = queryData.models.edges.map((edge) => edge.node)
        setModels(models)
      }
    },
  })

  // 创建模型
  const [createModelMutation, { loading: creating }] = useMutation<CreateModelData>(CREATE_MODEL, {
    onCompleted: (mutationData) => {
      if (mutationData?.createModel?.model) {
        addModel(mutationData.createModel.model)
        markStep('create_model')
      }
    },
    refetchQueries: [{ query: GET_MODELS }],
  })

  // 更新模型
  const [updateModelMutation, { loading: updating }] = useMutation<UpdateModelMetaData>(UPDATE_MODEL, {
    onCompleted: (mutationData) => {
      if (mutationData?.updateModelMeta?.model) {
        updateModel(mutationData.updateModelMeta.model.id, mutationData.updateModelMeta.model)
      }
    },
  })

  // 删除模型
  const [deleteModelMutation, { loading: deleting }] = useMutation<DeleteModelData>(DELETE_MODEL, {
    refetchQueries: [{ query: GET_MODELS }],
  })

  const createModel = async (input: CreateModelInput) => {
    if (!selectedProject?.slug) {
      throw new Error('请先选择项目')
    }

    try {
      const result = await createModelMutation({
        variables: {
          input,
        },
      })
      return result.data?.createModel
    } catch (err) {
      console.error('创建模型失败:', err)
      throw err
    }
  }

  const updateModelById = async (id: string, input: UpdateModelMetaInput) => {
    if (!selectedProject?.slug) {
      throw new Error('请先选择项目')
    }

    try {
      const result = await updateModelMutation({
        variables: {
          id,
          input,
        },
      })
      return result.data?.updateModelMeta
    } catch (err) {
      console.error('更新模型失败:', err)
      throw err
    }
  }

  const deleteModel = async (id: string) => {
    if (!selectedProject?.slug) {
      throw new Error('请先选择项目')
    }

    try {
      const result = await deleteModelMutation({
        variables: {
          id,
        },
      })
      if (result.data?.deleteModel?.success) {
        removeModel(id)
      }
      return result.data?.deleteModel
    } catch (err) {
      console.error('删除模型失败:', err)
      throw err
    }
  }

  const models = data?.models?.edges?.map((edge) => edge.node) || []

  return {
    models,
    loading,
    error,
    creating,
    updating,
    deleting,
    refetch,
    createModel,
    updateModel: updateModelById,
    deleteModel,
  }
}

export function useModelGroups() {
  const selectedProject = useAppStore((state) => state.selectedProject)

  const { data, loading, error, refetch } = useQuery<GetModelGroupsData>(GET_MODEL_GROUPS, {
    skip: !selectedProject?.slug,
  })

  return {
    modelGroups: data?.modelGroups || [],
    loading,
    error,
    refetch,
  }
}
