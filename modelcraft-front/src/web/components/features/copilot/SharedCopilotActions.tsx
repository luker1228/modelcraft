// src/web/components/features/copilot/SharedCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { AiProposalCard } from './AiProposalCard'
import { useNavigationProposal } from '@web/hooks/ai/use-navigation-proposal'
import type { AgentUiResponse } from './types'

/**
 * Frontend tools shared between admin and end-user agents.
 * Mount inside any CopilotKit context tree.
 *
 * Registers:
 *   ui_present_proposal — agent sends a proposal card with candidate actions
 */
export const SharedCopilotActions = memo(function SharedCopilotActions() {
  const { handleCandidateClick } = useNavigationProposal()

  useCopilotAction({
    name: 'ui_present_proposal',
    description:
      '向用户展示 AI 导航提案卡片，用户可点击候选项执行页面跳转、元素高亮或发送澄清消息',
    parameters: [
      {
        name: 'response',
        type: 'object',
        description: 'AgentUiResponse — 导航提案，包含 message 和 candidates 列表',
        required: true,
        attributes: [
          {
            name: 'kind',
            type: 'string',
            description: '固定值 "proposal"',
            required: true,
          },
          {
            name: 'proposalId',
            type: 'string',
            description: '唯一提案 ID，如 "nav-001"',
            required: true,
          },
          {
            name: 'proposalType',
            type: 'string',
            description: '"navigation" | "highlight" | "clarification" | "mixed"',
            required: true,
          },
          {
            name: 'message',
            type: 'string',
            description: '向用户展示的说明文字',
            required: true,
          },
          {
            name: 'query',
            type: 'string',
            description: '用户的原始查询（原样复制）',
            required: false,
          },
          {
            name: 'candidates',
            type: 'object[]',
            description: '候选项列表，每项为 action_candidate 或 clarification_candidate',
            required: true,
            attributes: [
              {
                name: 'id',
                type: 'string',
                description: '候选项唯一 ID，如 "c1"',
                required: true,
              },
              {
                name: 'type',
                type: 'string',
                description: '"action_candidate"（可执行跳转）或 "clarification_candidate"（需澄清）',
                required: true,
              },
              {
                name: 'title',
                type: 'string',
                description: '展示给用户的标题',
                required: true,
              },
              {
                name: 'description',
                type: 'string',
                description: '候选项说明',
                required: false,
              },
              {
                name: 'category',
                type: 'string',
                description: '"page" | "model" | "table" | "field" | "setting" | "action"',
                required: false,
              },
              {
                name: 'isPrimary',
                type: 'boolean',
                description: '是否为首选项（展示"推荐"标签）',
                required: false,
              },
              {
                name: 'actions',
                type: 'object[]',
                description: 'action_candidate 时必填，clarification_candidate 时省略',
                required: false,
                attributes: [
                  {
                    name: 'type',
                    type: 'string',
                    description: '"ui.navigate"（页面跳转）/ "ui.highlight"（高亮元素）/ "ui.guide"（组合引导）',
                    required: true,
                  },
                  {
                    name: 'args',
                    type: 'object',
                    description:
                      'ui.navigate: { route } — route 从 routeCatalog 派生，替换 :orgName/:projectSlug；' +
                      'ui.highlight/ui.guide: { targetId } — targetId 从 aiTargets 选取（ui.guide 可同时带 route）',
                    required: true,
                    attributes: [
                      {
                        name: 'route',
                        type: 'string',
                        description: 'ui.navigate 时填写完整路径，如 /org/acme/project/main/model-editor',
                        required: false,
                      },
                      {
                        name: 'targetId',
                        type: 'string',
                        description: 'ui.highlight 时填写，必须是 aiTargets 中已注册的 id',
                        required: false,
                      },
                      {
                        name: 'message',
                        type: 'string',
                        description: '高亮时的提示文字（可选）',
                        required: false,
                      },
                      {
                        name: 'durationMs',
                        type: 'number',
                        description: '高亮持续毫秒，默认 5000',
                        required: false,
                      },
                      {
                        name: 'scrollIntoView',
                        type: 'boolean',
                        description: '是否自动滚动到可见区域，默认 true',
                        required: false,
                      },
                    ],
                  },
                ],
              },
              {
                name: 'payload',
                type: 'object',
                description: 'clarification_candidate 时填写，描述推测的用户意图',
                required: false,
                attributes: [
                  {
                    name: 'intent',
                    type: 'string',
                    description: '意图标识，如 "configure_rbac"',
                    required: false,
                  },
                  {
                    name: 'userMeaning',
                    type: 'string',
                    description: '对用户目的的自然语言描述',
                    required: false,
                  },
                ],
              },
            ],
          },
        ],
      },
    ],
    // handler is required by CopilotKit to classify this as a valid "frontend" action.
    // The actual interaction is handled by the render function below.
    handler: async () => {},
    render: (props) => {
      // DeepSeek may return `response` as a JSON string instead of an object
      // when the tool schema is not fully specified. Parse defensively.
      let response = props.args.response as AgentUiResponse | string | undefined
      if (typeof response === 'string') {
        try {
          response = JSON.parse(response) as AgentUiResponse
        } catch {
          return <></>
        }
      }
      if (!response || typeof response !== 'object' || !('candidates' in response)) {
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
