'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

/**
 * Must be mounted INSIDE a <CopilotKit> provider tree.
 * Reads the current page's registered capabilities and injects them
 * into the AI context via useCopilotReadable on every change.
 */
export const AICapabilityReadable = memo(function AICapabilityReadable() {
  const { getAll } = useAICapabilityContext()
  const capabilities = getAll()

  useCopilotReadable({
    description: '当前页面可用的 UI 操作（点击 [ACTION:id] chip 可高亮对应元素）',
    value: capabilities.map((c) => ({
      id: c.id,
      label: c.label,
      description: c.description,
    })),
    // Empty array is fine — CopilotKit injects it but AI will see an empty list
    // and correctly skip using [ACTION:] markers.
  })

  return null
})
