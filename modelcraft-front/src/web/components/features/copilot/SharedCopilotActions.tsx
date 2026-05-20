// src/web/components/features/copilot/SharedCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { toast } from 'sonner'

/**
 * Frontend tools shared between admin and end-user agents.
 * Mount inside any CopilotKit context tree.
 *
 * Registers:
 *   show_toast — agent sends a one-line notification to the user
 */
export const SharedCopilotActions = memo(function SharedCopilotActions() {
  useCopilotAction({
    name: 'show_toast',
    description: '向用户显示一条临时通知消息（不需要用户在聊天框内查看）',
    parameters: [
      {
        name: 'message',
        type: 'string',
        description: '通知内容',
        required: true,
      },
      {
        name: 'type',
        type: 'string',
        description: 'success | error | info | warning（默认 info）',
        required: false,
      },
    ],
    handler: async ({ message, type }: { message: string; type?: string }) => {
      const fn = (type === 'success' ? toast.success
        : type === 'error' ? toast.error
        : type === 'warning' ? toast.warning
        : toast.info)
      fn(message)
      return 'toast displayed'
    },
  })

  return null
})
