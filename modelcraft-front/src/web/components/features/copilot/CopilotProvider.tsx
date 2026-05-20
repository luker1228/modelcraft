'use client'

import { memo, useMemo, Suspense } from 'react'
import dynamic from 'next/dynamic'
import type { Project } from '@/types'
import { CopilotAvailableContext } from '@web/components/features/end-user-data/FilterCopilotActions'
import { useAuthStore } from '@shared/stores/auth-store'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'

// Lazy load CopilotKit components to reduce initial bundle size
const CopilotKit = dynamic(
  () => import("@copilotkit/react-core").then(mod => mod.CopilotKit),
  { ssr: false }
)

const CopilotSidebar = dynamic(
  () => import("@copilotkit/react-ui").then(mod => mod.CopilotSidebar),
  { ssr: false }
)

interface CopilotProviderProps {
  children: React.ReactNode
  selectedProject: Project | null
  orgName: string
}

/**
 * CopilotKit Provider component
 *
 * Wraps children with CopilotKit context and provides AI assistant sidebar.
 * Forwards the tenant access token so the agent can authenticate GraphQL calls.
 */
const CopilotProvider = memo(({ children, selectedProject, orgName }: CopilotProviderProps) => {
  const accessToken = useAuthStore((s) => s.accessToken)

  const copilotContext = useMemo(() => ({
    projectId: selectedProject?.id || 'default',
    projectSlug: selectedProject?.slug || 'Default Project',
    orgName: orgName,
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
      agent="modelcraft_agent"
      headers={headers}
      properties={copilotContext}
    >
      {children}
      <CopilotSidebar
        labels={{
          title: "ModelCraft AI 助手",
          initial: initialMessage,
        }}
        defaultOpen={false}
        clickOutsideToClose={true}
      />
    </CopilotKit>
  )
})

CopilotProvider.displayName = 'CopilotProvider'

/**
 * Wrapper component for lazy-loaded CopilotKit (tenant-admin routes)
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

/**
 * CopilotKit wrapper for End-User routes.
 *
 * Forwards the end-user access token so the agent can authenticate GraphQL calls.
 */
interface EndUserCopilotWrapperProps {
  children: React.ReactNode
  orgName: string
  projectSlug: string
}

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
          agent="modelcraft_agent"
          headers={headers}
          properties={copilotContext}
        >
          {children}
          <CopilotSidebar
            labels={{
              title: "ModelCraft AI 助手",
              initial: initialMessage,
            }}
            defaultOpen={false}
          />
        </CopilotKit>
      </Suspense>
    </CopilotAvailableContext.Provider>
  )
})

EndUserCopilotWrapper.displayName = 'EndUserCopilotWrapper'
