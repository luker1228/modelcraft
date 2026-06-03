'use client'

import { memo } from 'react'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ENDUSER_SUGGESTIONS = [
  { title: '新手引导：带我了解这个系统', message: '我是新用户，请带我了解这个系统' },
  { title: '用自然语言帮我筛选数据', message: '帮我筛选数据，我来描述条件' },
  { title: '我看不到想要的数据，帮我排查', message: '我看不到想要的数据，请帮我排查' },
  { title: '这些字段分别是什么意思？', message: '请解释当前表格里各个字段的含义' },
  { title: '帮我统计一下数据', message: '帮我统计一下当前模型的数据情况' },
]

/**
 * Registers end-user sidebar quick suggestions.
 * Static knowledge (onboarding/troubleshooting) has been migrated to the backend
 * agent system prompt (agents/enduser_agent.py). Only UI suggestions remain here.
 */
export const EndUserCopilotKnowledge = memo(function EndUserCopilotKnowledge() {
  useCopilotChatSuggestions({
    suggestions: ENDUSER_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
