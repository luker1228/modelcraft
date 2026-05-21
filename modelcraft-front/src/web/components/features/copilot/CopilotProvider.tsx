'use client'

import { memo, useMemo, Suspense } from 'react'
import dynamic from 'next/dynamic'
import type { Project } from '@/types'
import { CopilotAvailableContext } from '@web/components/features/end-user-data/FilterCopilotActions'
import { useAuthStore } from '@shared/stores/auth-store'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { SharedCopilotActions } from './SharedCopilotActions'
import { AdminCopilotKnowledge } from './AdminCopilotKnowledge'
import { EndUserCopilotActions } from './EndUserCopilotActions'
import { EndUserCopilotKnowledge } from './EndUserCopilotKnowledge'
import { AICapabilityReadable } from './AICapabilityReadable'
import { AIChipMessage } from './AIChipMessage'

const CopilotKit = dynamic(
  () => import('@copilotkit/react-core').then(mod => mod.CopilotKit),
  { ssr: false }
)

const CopilotSidebar = dynamic(
  () => import('@copilotkit/react-ui').then(mod => mod.CopilotSidebar),
  { ssr: false }
)

interface CopilotProviderProps {
  children: React.ReactNode
  selectedProject: Project | null
  orgName: string
}

/**
 * Inner provider for the admin (tenant) surface.
 * Mounts SharedCopilotActions + AdminCopilotKnowledge inside CopilotKit context.
 */
const CopilotProvider = memo(({ children, selectedProject, orgName }: CopilotProviderProps) => {
  const accessToken = useAuthStore((s) => s.accessToken)

  const copilotContext = useMemo(() => ({
    projectId: selectedProject?.id || 'default',
    projectSlug: selectedProject?.slug || 'Default Project',
    orgName,
  }), [selectedProject?.id, selectedProject?.slug, orgName])

  const headers = useMemo<Record<string, string> | undefined>(
    () => accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    [accessToken]
  )

  const initialMessage = useMemo(() => {
    const projectSlug = selectedProject?.slug || 'Default Project'
    return `你好！我是 ModelCraft AI 助手，当前项目：${projectSlug}。

我可以帮助你：

• 创建和管理数据库集群
• 设计数据模型和字段
• 配置枚举类型
• 管理项目

请问有什么可以帮助你的？`
  }, [selectedProject?.slug])

  return (
    <CopilotKit
      runtimeUrl="/api/copilotkit"
      agent="modelcraft_admin_agent"
      headers={headers}
      properties={copilotContext}
    >
      <SharedCopilotActions />
      <AdminCopilotKnowledge />
      <AICapabilityReadable />
      {children}
      <CopilotSidebar
        labels={{
          title: 'ModelCraft AI 助手',
          initial: initialMessage,
        }}
        defaultOpen={false}
        clickOutsideToClose={true}
        AssistantMessage={AIChipMessage}
      />
    </CopilotKit>
  )
})

CopilotProvider.displayName = 'CopilotProvider'

/**
 * Wrapper for admin (tenant) routes — org/* and project/* layouts.
 */
export const CopilotWrapper = memo(({
  children,
  selectedProject,
  orgName,
}: CopilotProviderProps) => {
  return (
    <CopilotAvailableContext.Provider value={true}>
      <Suspense fallback={children}>
        <CopilotProvider selectedProject={selectedProject} orgName={orgName}>
          {children}
        </CopilotProvider>
      </Suspense>
    </CopilotAvailableContext.Provider>
  )
})

CopilotWrapper.displayName = 'CopilotWrapper'

interface EndUserCopilotWrapperProps {
  children: React.ReactNode
  orgName: string
  projectSlug: string
}

/**
 * Wrapper for end-user routes — mounts enduser-specific tools, knowledge, and sidebar.
 */
export const EndUserCopilotWrapper = memo(({
  children,
  orgName,
  projectSlug,
}: EndUserCopilotWrapperProps) => {
  const accessToken = useEndUserAuthStore((s) => s.accessToken)

  const copilotContext = useMemo(() => ({
    orgName,
    projectSlug,
  }), [orgName, projectSlug])

  const headers = useMemo<Record<string, string> | undefined>(
    () => accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    [accessToken]
  )

  const initialMessage = useMemo(() => `你好！我是 ModelCraft AI 助手，当前项目：${projectSlug}。

我可以帮助你：

• 查询和筛选数据
• 分析数据记录

请问有什么可以帮助你的？`, [projectSlug])

  return (
    <CopilotAvailableContext.Provider value={true}>
      <Suspense fallback={children}>
        <CopilotKit
          runtimeUrl="/api/copilotkit"
          agent="modelcraft_enduser_agent"
          headers={headers}
          properties={copilotContext}
        >
          <SharedCopilotActions />
          <EndUserCopilotKnowledge />
          <EndUserCopilotActions orgName={orgName} projectSlug={projectSlug} />
          {children}
          <CopilotSidebar
            labels={{
              title: 'ModelCraft AI 助手',
              initial: initialMessage,
            }}
            defaultOpen={false}
            AssistantMessage={AIChipMessage}
          />
        </CopilotKit>
      </Suspense>
    </CopilotAvailableContext.Provider>
  )
})

EndUserCopilotWrapper.displayName = 'EndUserCopilotWrapper'
