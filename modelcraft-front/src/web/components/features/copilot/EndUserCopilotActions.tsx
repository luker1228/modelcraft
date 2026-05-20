// src/web/components/features/copilot/EndUserCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { useRouter } from 'next/navigation'

interface EndUserCopilotActionsProps {
  orgName: string
  projectSlug: string
}

/**
 * Frontend navigation tools for end-user routes.
 * Registers: navigate_to_project, navigate_to_workspace
 */
export const EndUserCopilotActions = memo(function EndUserCopilotActions({
  orgName,
  projectSlug: _projectSlug,
}: EndUserCopilotActionsProps) {
  const router = useRouter()

  useCopilotAction({
    name: 'navigate_to_project',
    description: '切换到指定项目（end-user 路由）',
    parameters: [
      { name: 'slug', type: 'string', description: '项目 slug', required: true },
    ],
    handler: async ({ slug }: { slug: string }) => {
      router.push(`/end-user/${orgName}/projects/${slug}/data`)
      return `已切换到项目 ${slug}`
    },
  })

  useCopilotAction({
    name: 'navigate_to_workspace',
    description: '返回项目选择页',
    parameters: [],
    handler: async () => {
      router.push(`/end-user/${orgName}/select-project`)
      return '已返回项目选择页'
    },
  })

  return null
})
