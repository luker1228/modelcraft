'use client'

import { createContext, useContext } from 'react'
import { useCopilotAction, CopilotContext } from '@copilotkit/react-core'

interface FilterCopilotActionsProps {
  onSetFilter: (filterJson: string) => void
  onClearFilter: () => void
}

/**
 * Explicit opt-in context for components that need to check whether
 * a CopilotKitProvider is present in the tree.
 *
 * Problem background:
 * - @copilotkitnext/react's CopilotKitContext has a DEFAULT value of
 *   { copilotkit: null }. This means useContext() always returns a
 *   non-null object, making `!== null` guards useless.
 * - @copilotkit/react-core's CopilotContext similarly defaults to a
 *   non-null emptyCopilotContext object.
 * - Calling ANY hook from @copilotkitnext/react (e.g. useCopilotKit,
 *   useFrontendTool) outside a provider triggers null.subscribe() crash
 *   in the library's internal useEffect.
 *
 * Solution:
 * - CopilotWrapper (the component that actually renders CopilotKitProvider)
 *   wraps its children in <CopilotAvailableContext.Provider value={true}>.
 * - Components check useContext(CopilotAvailableContext) instead of
 *   relying on the library's own context default values.
 */
export const CopilotAvailableContext = createContext<boolean>(false)

/**
 * Returns true only when a real CopilotKitProvider is present in the tree.
 * Must be used together with CopilotAvailableContext.Provider (see CopilotWrapper).
 */
export function useCopilotKitAvailable(): boolean {
  return useContext(CopilotAvailableContext)
}

/**
 * Registers CopilotKit frontend actions for the filter panel.
 *
 * Designed to be mounted conditionally — only when CopilotKit context is
 * present. Both tenant-admin and end-user data views can use this component.
 *
 * Usage:
 *   const hasCopilot = useCopilotKitAvailable()
 *   {hasCopilot && <FilterCopilotActions onSetFilter={...} onClearFilter={...} />}
 */
export function FilterCopilotActions({ onSetFilter, onClearFilter }: FilterCopilotActionsProps) {
  useCopilotAction({
    name: 'set_filter',
    description:
      '设置 FilterPanel 的 where 筛选条件。接受 ModelCraft filter JSON 字符串，例如: {"AND":[{"name":{"contains":"张"}}]}',
    parameters: [
      {
        name: 'filter_json',
        type: 'string',
        description: 'ModelCraft where JSON 字符串',
        required: true,
      },
    ],
    handler: async ({ filter_json }: { filter_json: string }) => {
      onSetFilter(filter_json)
    },
  })

  useCopilotAction({
    name: 'clear_filter',
    description: '清空 FilterPanel 的所有筛选条件，恢复全量数据展示',
    parameters: [],
    handler: async () => {
      onClearFilter()
    },
  })

  return null
}

/**
 * Re-export CopilotContext for convenience so callers don't need a direct
 * dependency on @copilotkit/react-core just to check context availability.
 *
 * @deprecated Use useCopilotKitAvailable() instead — CopilotContext default
 * value is non-null so `useContext(CopilotContext) !== null` is always true.
 */
export { CopilotContext }
