'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ADMIN_ONBOARDING = `
ModelCraft 新手引导（AI 导航助手）：

AI 助手通过 ui_present_proposal 帮助你导航到正确页面并高亮目标区域。
你可以直接用自然语言描述想做的事，AI 会给你一个或多个候选方案，点击后自动跳转。

常见操作引导：
- 创建数据模型 → AI 带你到"数据模型编辑器"，高亮新建模型入口
- 配置权限 → AI 带你到"RBAC 角色管理"，高亮相关配置区域
- 添加数据库连接 → AI 带你到"项目设置"，高亮集群配置区域
- 管理终端用户 → AI 带你到"终端用户管理"页面

数据查询（不需要导航）：
- 直接询问数据，AI 会调用后端工具查询并在对话中返回结果
`.trim()

const ADMIN_TROUBLESHOOTING = `
常见问题：

问题：数据库连接失败
  → 告诉 AI "帮我检查集群配置"，AI 会导航到项目设置并高亮集群配置区域

问题：找不到模型或字段
  → 告诉 AI 模型名称，AI 可以调用 list_models / get_model_fields 查询

问题：权限配置在哪里
  → 告诉 AI "帮我配置权限"，AI 会导航到 RBAC 相关页面
`.trim()

const ADMIN_SUGGESTIONS = [
  { title: '帮我创建第一个数据模型', message: '帮我创建第一个数据模型，带我到对应页面' },
  { title: '权限配置在哪里？', message: '我想配置数据权限，带我去对应页面' },
  { title: '我有哪些项目？', message: '列出我所有的项目' },
  { title: '帮我导航到模型管理', message: '帮我去数据模型管理页面' },
  { title: '数据库连不上，帮我排查', message: '数据库连接有问题，带我去检查配置' },
]

export const AdminCopilotKnowledge = memo(function AdminCopilotKnowledge() {
  useCopilotReadable({
    description: 'ModelCraft 新手引导操作手册（管理员）',
    value: ADMIN_ONBOARDING,
  })

  useCopilotReadable({
    description: 'ModelCraft 常见问题排查手册（管理员）',
    value: ADMIN_TROUBLESHOOTING,
  })

  useCopilotChatSuggestions({
    suggestions: ADMIN_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
