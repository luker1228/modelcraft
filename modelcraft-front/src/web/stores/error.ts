import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { GraphQLErrorInfo, GraphQLErrorContext } from '@web/components/common/GraphQLErrorDialog'

interface ErrorState {
  // 状态
  isErrorDialogOpen: boolean
  currentErrors: GraphQLErrorInfo[]
  currentContext?: GraphQLErrorContext
  errorHistory: Array<{
    id: string
    errors: GraphQLErrorInfo[]
    context?: GraphQLErrorContext
    timestamp: string
  }>

  // Actions
  showErrorDialog: (errors: GraphQLErrorInfo[], context?: GraphQLErrorContext) => void
  hideErrorDialog: () => void
  clearErrors: () => void
  addToHistory: (errors: GraphQLErrorInfo[], context?: GraphQLErrorContext) => void
  clearHistory: () => void
}

export const useErrorStore = create<ErrorState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      isErrorDialogOpen: false,
      currentErrors: [],
      currentContext: undefined,
      errorHistory: [],

      // Actions
      showErrorDialog: (errors, context) => {
        const { addToHistory } = get()
        
        // 添加到历史记录
        addToHistory(errors, context)
        
        set({
          isErrorDialogOpen: true,
          currentErrors: errors,
          currentContext: context,
        })
      },

      hideErrorDialog: () => set({
        isErrorDialogOpen: false,
      }),

      clearErrors: () => set({
        currentErrors: [],
        currentContext: undefined,
      }),

      addToHistory: (errors, context) => {
        const { errorHistory } = get()
        const newEntry = {
          id: Date.now().toString(),
          errors,
          context,
          timestamp: new Date().toISOString(),
        }

        // 保持最近的50条错误记录
        const updatedHistory = [newEntry, ...errorHistory].slice(0, 50)

        set({ errorHistory: updatedHistory })
      },

      clearHistory: () => set({
        errorHistory: [],
      }),
    }),
    {
      name: 'error-store',
    }
  )
)