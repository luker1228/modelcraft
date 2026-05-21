'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ADMIN_ONBOARDING = `
新手引导共 5 步：

Step 1 [创建项目]
  目标：创建第一个项目
  工具：open_create_project(slug, title)
  验证：调用 list_projects，确认项目出现

Step 2 [配置数据库集群]
  目标：连接一个数据库
  工具：navigate_to_cluster，引导用户在页面上手动填写连接信息
  提示：集群配置需要用户提供数据库连接信息，agent 无法代替操作

Step 3 [创建数据模型]
  目标：在项目下创建第一个模型
  引导流程（必须按顺序执行）：
    1. 调用 list_databases 查询可用数据库列表
    2. 调用 guide_select_database("请先选择要在哪个数据库中创建模型") — 高亮数据库选择器
    3. 在文字中告诉用户："数据库选择器已高亮（左侧金色边框），请点击选择一个数据库，然后告诉我数据库名称"
    4. 等待用户回复选择的数据库名 OR 用户已明确指定数据库时直接进入第 5 步
    5. 调用 guide_create_model("请点击新建模型按钮") — 高亮新建模型按钮
    6. 在文字中告诉用户："新建模型按钮已高亮（左侧金色边框），请点击它，填写模型名称后保存"
  验证：调用 list_models 确认模型存在

Step 4 [添加字段]
  目标：给模型添加字段
  工具：navigate_to_model(db, model)
  提示：字段编辑在右侧面板，由用户手动操作

Step 5 [查看数据]
  目标：进入数据视图，确认配置完成
  工具：navigate_to_data(db, model)
`.trim()

const ADMIN_TROUBLESHOOTING = `
常见问题排查：

问题：数据库连接失败
  → show_toast("正在带你去检查集群配置", "info")
  → navigate_to_cluster，引导检查 host/port/credentials

问题：找不到模型或字段
  → 调用 list_models(db) 确认模型名是否正确
  → list_models 返回空则模型未创建，建议执行新手引导 Step 3

问题：权限被拒绝
  → navigate_to_rbac(section="users")，检查用户角色分配

问题：字段显示异常
  → navigate_to_model(db, model)，检查字段类型和配置
`.trim()

const ADMIN_SUGGESTIONS = [
  { title: '新手引导：带我完成初始配置', message: '我是新用户，请帮我按步骤完成初始配置' },
  { title: '帮我创建第一个数据模型', message: '帮我创建第一个数据模型' },
  { title: '数据库连不上，帮我排查', message: '数据库连接有问题，请帮我排查' },
  { title: '我有哪些项目？', message: '列出我所有的项目' },
  { title: '解释当前页面的功能', message: '请解释当前页面有哪些功能' },
]

/**
 * Injects admin knowledge base and sidebar suggestions into CopilotKit context.
 * Must be mounted inside a CopilotKit provider tree.
 */
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
