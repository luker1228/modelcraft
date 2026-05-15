'use client'

import { useContext } from 'react'
import { useCopilotAction, CopilotContext } from '@copilotkit/react-core'

interface FilterCopilotActionsProps {
  onSetFilter: (filterJson: string) => void
  onClearFilter: () => void
}

/**
 * Registers CopilotKit frontend actions for the filter panel.
 *
 * Designed to be mounted conditionally — only when CopilotKit context is
 * present. Both tenant-admin and end-user data views can use this component.
 *
 * Usage:
 *   const hasCopilot = useContext(CopilotContext) !== null
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
 */
export { CopilotContext }
