'use client'

import { memo } from 'react'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ADMIN_SUGGESTIONS = [
  { title: '帮我创建第一个数据模型', message: '帮我创建第一个数据模型，带我到对应页面' },
  { title: '权限配置在哪里？', message: '我想配置数据权限，带我去对应页面' },
  { title: '我有哪些项目？', message: '列出我所有的项目' },
  { title: '帮我导航到模型管理', message: '帮我去数据模型管理页面' },
  { title: '数据库连不上，帮我排查', message: '数据库连接有问题，带我去检查配置' },
]

/**
 * Registers admin sidebar quick suggestions.
 * Static knowledge (onboarding/troubleshooting) has been migrated to the backend
 * agent system prompt (agents/admin_agent.py). Only UI suggestions remain here.
 */
export const AdminCopilotKnowledge = memo(function AdminCopilotKnowledge() {
  useCopilotChatSuggestions({
    suggestions: ADMIN_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
