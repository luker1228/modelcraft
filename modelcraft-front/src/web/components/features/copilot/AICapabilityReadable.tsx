'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { PROJECT_REQUIRED_ROUTE_POLICY, ROUTE_CATALOG } from '@web/lib/route-catalog'

/**
 * Must be mounted INSIDE a <CopilotKit> provider tree.
 *
 * Injects two pieces of knowledge into the Agent context on every render:
 *   1. aiTargets  — current page's registered AiTarget elements (id, label, description, type)
 *   2. routeCatalog — all navigable pages (routeTemplate, title, description, keywords, requiresProject)
 */
export const AICapabilityReadable = memo(function AICapabilityReadable() {
  const { getAll } = useAICapabilityContext()
  const targets = getAll()

  useCopilotReadable({
    description:
      '当前页面已注册的 AI 高亮目标（AiTarget）。' +
      '调用 ui_present_proposal 时，ui.highlight 的 targetId 必须从这个列表中选取。',
    value: targets.map((c) => ({
      id: c.id,
      label: c.label,
      description: c.description,
      type: c.type,
    })),
  })

  useCopilotReadable({
    description:
      '系统所有可导航页面目录（routeCatalog）。' +
      '调用 ui_present_proposal 时，ui.navigate 的 route 字段必须从 routeTemplate 派生，' +
      '将 :orgName、:projectSlug 等参数替换为当前会话的实际值。' +
      'requiresProject=true 表示该页面需要项目上下文（projectSlug / projectName）。',
    value: ROUTE_CATALOG,
  })

  useCopilotReadable({
    description:
      '项目依赖页面/工具的前置策略。' +
      '对于 requiresProject=true 的 route 或项目级工具，必须先 list_projects；' +
      '有项目则让用户选择项目，无项目则推荐先创建项目。',
    value: PROJECT_REQUIRED_ROUTE_POLICY,
  })

  return null
})
