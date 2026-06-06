'use client'

import { memo, useMemo, Suspense } from 'react'
import dynamic from 'next/dynamic'
import { usePathname } from 'next/navigation'
import { CopilotAvailableContext } from '@web/components/features/end-user-data/FilterCopilotActions'
import { useAuthStore } from '@shared/stores/auth-store'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { useAppStore } from '@web/stores/app'
import { SharedCopilotActions } from './SharedCopilotActions'
import { AdminCopilotKnowledge } from './AdminCopilotKnowledge'
import { EndUserCopilotActions } from './EndUserCopilotActions'
import { EndUserCopilotKnowledge } from './EndUserCopilotKnowledge'
import { AIChipMessage } from './AIChipMessage'
import { RoutePageKnowledge } from './RoutePageKnowledge'

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
  orgName: string
}

/**
 * Inner provider for the admin (tenant) surface.
 * Reads selectedProject from useAppStore so project layout can update it
 * without creating a nested CopilotKit instance.
 */
const CopilotProvider = memo(({ children, orgName }: CopilotProviderProps) => {
  const accessToken = useAuthStore((s) => s.accessToken)
  const pathname = usePathname()
  const selectedProject = useAppStore((s) => s.selectedProject)

  const copilotContext = useMemo(() => ({
    projectId: selectedProject?.id || '',
    projectSlug: selectedProject?.slug || '',
    orgName,
    currentRoute: pathname,
  }), [selectedProject?.id, selectedProject?.slug, orgName, pathname])

  const headers = useMemo<Record<string, string> | undefined>(
    () => accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    [accessToken]
  )

  const initialMessage = useMemo(() => {
    const projectText = selectedProject?.slug ? `当前项目：${selectedProject.slug}` : '当前未选择项目'
    return `你好！我是 ModelCraft AI 助手，${projectText}。

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
      showDevConsole={false}
    >
      <SharedCopilotActions />
      <AdminCopilotKnowledge />
      <RoutePageKnowledge />
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
 * selectedProject is read from useAppStore; no prop needed.
 */
export const CopilotWrapper = memo(({
  children,
  orgName,
}: Omit<CopilotProviderProps, never>) => {
  return (
    <CopilotAvailableContext.Provider value={true}>
      <Suspense fallback={children}>
        <CopilotProvider orgName={orgName}>
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
 *
 * Note: the current end-user data route does not use this wrapper on purpose.
 * The CopilotSidebar floating entry is more suitable for admin workflows and
 * currently blocks the end-user table view. Keep this wrapper as the reserved
 * integration point so end-user Copilot can be turned back on later.
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
          showDevConsole={false}
        >
          <SharedCopilotActions />
          <EndUserCopilotKnowledge />
          <EndUserCopilotActions orgName={orgName} projectSlug={projectSlug} />
          {/* AICapabilityReadable intentionally omitted: end-user surface is read-only (data query/view), no page-action capability registration */}
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
