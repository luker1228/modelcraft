// src/web/components/features/copilot/SharedCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { toast } from 'sonner'
import { AiProposalCard } from './AiProposalCard'
import { useNavigationProposal } from '@web/hooks/ai/use-navigation-proposal'
import type { AgentUiResponse } from './types'

/**
 * Frontend tools shared between admin and end-user agents.
 * Mount inside any CopilotKit context tree.
 *
 * Registers:
 *   show_toast               — agent sends a one-line notification to the user
 *   show_navigation_proposal — agent sends a proposal card with candidate actions
 */
export const SharedCopilotActions = memo(function SharedCopilotActions() {
  const { handleCandidateClick } = useNavigationProposal()

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

  useCopilotAction({
    name: 'show_navigation_proposal',
    description:
      '向用户展示 AI 导航提案卡片，用户可点击候选项执行页面跳转、元素高亮或发送澄清消息',
    parameters: [
      {
        name: 'response',
        type: 'object',
        description: 'AgentUiResponse — 包含 message、candidates 等字段的导航提案',
        required: true,
      },
    ],
    // handler is required by CopilotKit to classify this as a valid "frontend" action.
    // The actual interaction is handled by the render function below.
    handler: async () => {},
    render: (props) => {
      const response = props.args.response as AgentUiResponse | undefined
      if (!response?.candidates) {
        return <></>
      }
      return (
        <AiProposalCard
          message={response.message ?? ''}
          candidates={response.candidates}
          onSelect={handleCandidateClick}
        />
      )
    },
  })

  return null
})
