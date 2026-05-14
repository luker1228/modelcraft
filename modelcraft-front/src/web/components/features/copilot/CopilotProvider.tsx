'use client'

import { memo, useMemo, Suspense } from 'react'
import dynamic from 'next/dynamic'
import type { Project } from '@/types'

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
 * Wraps children with CopilotKit context and provides AI assistant sidebar
 * Lazy-loaded to optimize initial page load performance
 * 
 * @param children - Child components to wrap
 * @param selectedProject - Currently selected project for context
 */
const CopilotProvider = memo(({ children, selectedProject, orgName }: CopilotProviderProps) => {
  // Memoize copilot context to prevent unnecessary re-renders
  const copilotContext = useMemo(() => ({
    projectId: selectedProject?.id || 'default',
    projectSlug: selectedProject?.slug || 'Default Project',
    orgName: orgName,
  }), [selectedProject?.id, selectedProject?.slug, orgName])

  // Memoize initial message to prevent recreation
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
      properties={copilotContext}
    >
      {children}
      <CopilotSidebar
        labels={{
          title: "ModelCraft AI 助手",
          initial: initialMessage,
        }}
        defaultOpen={true}
        clickOutsideToClose={true}
      />
    </CopilotKit>
  )
})

CopilotProvider.displayName = 'CopilotProvider'

/**
 * AI Assistant floating button
 * 
 * Displays a floating action button to open the AI assistant
 */
export const AIAssistantButton = memo(({ onClick }: { onClick: () => void }) => {
  return (
    <button
      onClick={onClick}
      className="fixed bottom-6 right-6 z-50 flex size-12 cursor-pointer items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg shadow-primary/25 transition-all duration-200 hover:bg-primary/90 hover:shadow-xl hover:shadow-primary/30"
      title="打开 AI 助手"
      aria-label="打开 AI 助手"
    >
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        width="20" 
        height="20" 
        viewBox="0 0 24 24" 
        fill="none" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
        strokeLinejoin="round"
      >
        <path d="M12 8V4H8"/>
        <rect width="16" height="12" x="4" y="8" rx="2"/>
        <path d="M2 14h2"/>
        <path d="M20 14h2"/>
        <path d="M15 13v2"/>
        <path d="M9 13v2"/>
      </svg>
    </button>
  )
})

AIAssistantButton.displayName = 'AIAssistantButton'

/**
 * Wrapper component for lazy-loaded CopilotKit
 * 
 * Provides suspense boundary and fallback UI
 */
export const CopilotWrapper = memo(({
  children,
  selectedProject,
  orgName,
}: CopilotProviderProps) => {
  return (
    <Suspense fallback={children}>
      <CopilotProvider selectedProject={selectedProject} orgName={orgName}>
        {children}
      </CopilotProvider>
    </Suspense>
  )
})

CopilotWrapper.displayName = 'CopilotWrapper'
