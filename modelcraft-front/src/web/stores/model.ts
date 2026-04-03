import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { Model, Field } from '@/types'

interface ModelState {
  // 状态
  models: Model[]
  selectedModel: Model | null
  loading: boolean
  error: string | null

  // Actions
  setModels: (models: Model[]) => void
  addModel: (model: Model) => void
  updateModel: (id: string, updates: Partial<Model>) => void
  removeModel: (id: string) => void
  setSelectedModel: (model: Model | null) => void
  findModelByName: (name: string) => Model | undefined
  findModelById: (id: string) => Model | undefined
  addFieldToModel: (modelId: string, field: Field) => void
  updateFieldInModel: (modelId: string, fieldId: string, updates: Partial<Field>) => void
  removeFieldFromModel: (modelId: string, fieldId: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearModels: () => void
}

export const useModelStore = create<ModelState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      models: [],
      selectedModel: null,
      loading: false,
      error: null,

      // Actions
      setModels: (models) => set({ models }),

      addModel: (model) => set((state) => ({
        models: [...state.models, model]
      })),

      updateModel: (id, updates) => set((state) => ({
        models: state.models.map(model => 
          model.id === id ? { ...model, ...updates } : model
        ),
        selectedModel: state.selectedModel?.id === id 
          ? { ...state.selectedModel, ...updates }
          : state.selectedModel
      })),

      removeModel: (id) => set((state) => ({
        models: state.models.filter(model => model.id !== id),
        selectedModel: state.selectedModel?.id === id ? null : state.selectedModel
      })),

      setSelectedModel: (model) => set({ selectedModel: model }),

      findModelByName: (name) => {
        const { models } = get()
        return models.find(model => model.name === name)
      },

      findModelById: (id) => {
        const { models } = get()
        return models.find(model => model.id === id)
      },

      addFieldToModel: (modelId, field) => set((state) => ({
        models: state.models.map(model => 
          model.id === modelId 
            ? { ...model, fields: [...model.fields, field] }
            : model
        ),
        selectedModel: state.selectedModel?.id === modelId
          ? { ...state.selectedModel, fields: [...state.selectedModel.fields, field] }
          : state.selectedModel
      })),

      updateFieldInModel: (modelId, fieldId, updates) => set((state) => ({
        models: state.models.map(model => 
          model.id === modelId 
            ? {
                ...model,
                fields: model.fields.map(field =>
                  field.id === fieldId ? { ...field, ...updates } : field
                )
              }
            : model
        ),
        selectedModel: state.selectedModel?.id === modelId
          ? {
              ...state.selectedModel,
              fields: state.selectedModel.fields.map(field =>
                field.id === fieldId ? { ...field, ...updates } : field
              )
            }
          : state.selectedModel
      })),

      removeFieldFromModel: (modelId, fieldId) => set((state) => ({
        models: state.models.map(model => 
          model.id === modelId 
            ? { ...model, fields: model.fields.filter(field => field.id !== fieldId) }
            : model
        ),
        selectedModel: state.selectedModel?.id === modelId
          ? { ...state.selectedModel, fields: state.selectedModel.fields.filter(field => field.id !== fieldId) }
          : state.selectedModel
      })),

      setLoading: (loading) => set({ loading }),

      setError: (error) => set({ error }),

      clearModels: () => set({
        models: [],
        selectedModel: null,
        loading: false,
        error: null,
      }),
    }),
    {
      name: 'model-store',
    }
  )
)