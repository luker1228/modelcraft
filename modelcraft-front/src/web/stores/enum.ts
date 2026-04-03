import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { EnumDefinition, EnumOption } from '@/types'

type Enum = EnumDefinition

interface EnumState {
  // 状态
  enums: Enum[]
  selectedEnum: Enum | null
  loading: boolean
  error: string | null

  // Actions
  setEnums: (enums: Enum[]) => void
  addEnum: (enumItem: Enum) => void
  updateEnum: (id: string, updates: Partial<Enum>) => void
  removeEnum: (id: string) => void
  setSelectedEnum: (enumItem: Enum | null) => void
  findEnumById: (id: string) => Enum | undefined
  findEnumByName: (name: string) => Enum | undefined
  addOptionToEnum: (enumId: string, option: EnumOption) => void
  updateOptionInEnum: (enumId: string, optionId: string, updates: Partial<EnumOption>) => void
  removeOptionFromEnum: (enumId: string, optionId: string) => void
  reorderOptionsInEnum: (enumId: string, options: EnumOption[]) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearEnums: () => void
}

export const useEnumStore = create<EnumState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      enums: [],
      selectedEnum: null,
      loading: false,
      error: null,

      // Actions
      setEnums: (enums) => set({ enums }),

      addEnum: (enumItem) => set((state) => ({
        enums: [...state.enums, enumItem]
      })),

      updateEnum: (id, updates) => set((state) => ({
        enums: state.enums.map(enumItem => 
          enumItem.id === id ? { ...enumItem, ...updates } : enumItem
        ),
        selectedEnum: state.selectedEnum?.id === id 
          ? { ...state.selectedEnum, ...updates }
          : state.selectedEnum
      })),

      removeEnum: (id) => set((state) => ({
        enums: state.enums.filter(enumItem => enumItem.id !== id),
        selectedEnum: state.selectedEnum?.id === id ? null : state.selectedEnum
      })),

      setSelectedEnum: (enumItem) => set({ selectedEnum: enumItem }),

      findEnumById: (id) => {
        const { enums } = get()
        return enums.find(enumItem => enumItem.id === id)
      },

      findEnumByName: (name) => {
        const { enums } = get()
        return enums.find(enumItem => enumItem.name === name)
      },

      addOptionToEnum: (enumId, option) => set((state) => ({
        enums: state.enums.map(enumItem => 
          enumItem.id === enumId 
            ? { ...enumItem, options: [...enumItem.options, option] }
            : enumItem
        ),
        selectedEnum: state.selectedEnum?.id === enumId
          ? { ...state.selectedEnum, options: [...state.selectedEnum.options, option] }
          : state.selectedEnum
      })),

      updateOptionInEnum: (enumId, optionId, updates) => set((state) => ({
        enums: state.enums.map(enumItem => 
          enumItem.id === enumId 
            ? {
                ...enumItem,
                options: enumItem.options.map(option =>
                  option.id === optionId ? { ...option, ...updates } : option
                )
              }
            : enumItem
        ),
        selectedEnum: state.selectedEnum?.id === enumId
          ? {
              ...state.selectedEnum,
              options: state.selectedEnum.options.map(option =>
                option.id === optionId ? { ...option, ...updates } : option
              )
            }
          : state.selectedEnum
      })),

      removeOptionFromEnum: (enumId, optionId) => set((state) => ({
        enums: state.enums.map(enumItem => 
          enumItem.id === enumId 
            ? { ...enumItem, options: enumItem.options.filter(option => option.id !== optionId) }
            : enumItem
        ),
        selectedEnum: state.selectedEnum?.id === enumId
          ? { ...state.selectedEnum, options: state.selectedEnum.options.filter(option => option.id !== optionId) }
          : state.selectedEnum
      })),

      reorderOptionsInEnum: (enumId, options) => set((state) => ({
        enums: state.enums.map(enumItem => 
          enumItem.id === enumId 
            ? { ...enumItem, options }
            : enumItem
        ),
        selectedEnum: state.selectedEnum?.id === enumId
          ? { ...state.selectedEnum, options }
          : state.selectedEnum
      })),

      setLoading: (loading) => set({ loading }),

      setError: (error) => set({ error }),

      clearEnums: () => set({
        enums: [],
        selectedEnum: null,
        loading: false,
        error: null,
      }),
    }),
    {
      name: 'enum-store',
    }
  )
)