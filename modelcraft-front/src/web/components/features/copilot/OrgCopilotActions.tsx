'use client'

import { useCopilotAction } from '@copilotkit/react-core'
import { useRouter } from 'next/navigation'

interface OrgCopilotActionsProps {
  orgName: string
  /** 打开新建项目 Sheet — 由 workspace page 传入 */
  onOpenCreateProject?: (prefill: { slug?: string; title?: string; description?: string }) => void
  /** 高亮某个项目卡片 — 由 workspace page 传入 */
  onHighlightProject?: (slug: string, reason: string) => void
}

export function OrgCopilotActions({ orgName, onOpenCreateProject, onHighlightProject }: OrgCopilotActionsProps) {
  const router = useRouter()

  useCopilotAction({
    name: 'navigate_to_project',
    description: '跳转到指定项目的工作区',
    parameters: [
      { name: 'slug', type: 'string', description: '项目 slug', required: true },
    ],
    handler: async ({ slug }: { slug: string }) => {
      router.push(`/org/${orgName}/project/${slug}`)
      return `已跳转到项目 ${slug}`
    },
  })

  useCopilotAction({
    name: 'navigate_to_settings',
    description: '跳转到组织设置页面',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/settings`)
      return '已跳转到设置页面'
    },
  })

  useCopilotAction({
    name: 'open_create_project',
    description: '打开新建项目表单，可预填 slug、title、description。用户需手动点击 Create 按钮完成创建。',
    parameters: [
      { name: 'slug', type: 'string', description: '项目 slug（英文小写+连字符）', required: false },
      { name: 'title', type: 'string', description: '项目显示名称', required: false },
      { name: 'description', type: 'string', description: '项目描述', required: false },
    ],
    handler: async ({ slug, title, description }: { slug?: string; title?: string; description?: string }) => {
      if (onOpenCreateProject) {
        onOpenCreateProject({ slug, title, description })
        return '已打开新建项目表单，字段已预填，请确认后点击 Create。'
      }
      return '当前页面不支持新建项目操作，请先导航到 workspace 页。'
    },
  })

  useCopilotAction({
    name: 'highlight_project',
    description: '在项目列表中高亮指定项目，并显示说明原因',
    parameters: [
      { name: 'slug', type: 'string', description: '要高亮的项目 slug', required: true },
      { name: 'reason', type: 'string', description: '高亮原因，显示为 tooltip', required: true },
    ],
    handler: async ({ slug, reason }: { slug: string; reason: string }) => {
      if (onHighlightProject) {
        onHighlightProject(slug, reason)
        return `已高亮项目 ${slug}：${reason}`
      }
      return '当前页面不支持高亮操作，请先导航到 workspace 页。'
    },
  })

  return null
}
