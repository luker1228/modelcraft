'use client'

import { useEffect, useState, useMemo, useRef } from 'react'
import { useParams } from 'next/navigation'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { LoadingScreen } from '@web/components/common/LoadingScreen'
import { CopilotWrapper, AIAssistantButton } from '@web/components/features/copilot/CopilotProvider'
import { RouteValidator } from '@web/components/common/RouteValidator'
import { useAppStore } from '@web/stores/app'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { useCopilotReadable } from '@copilotkit/react-core'
import { ProjectCopilotActions } from '@web/components/features/copilot/ProjectCopilotActions'
import { WorkspaceAIRefContext } from '@web/contexts/workspace-ai-ref-context'
import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'
import "@copilotkit/react-ui/styles.css"

interface ProjectLayoutProps {
  children: React.ReactNode
}

function ProjectAIContext({
  orgName,
  projectSlug,
  workspaceAiRef,
}: {
  orgName: string
  projectSlug: string
  workspaceAiRef: React.MutableRefObject<DevelopRecordWorkspaceAIRef | null>
}) {
  useCopilotReadable({
    description: '当前 AI 上下文',
    value: {
      layer: 'project',
      orgName,
      projectSlug,
      availableActions: [
        'navigate_to_org',
        'navigate_to_model',
        'navigate_to_data',
        'open_create_model',
        'open_create_record',
        'open_edit_record',
        'highlight_records',
        'set_filter',
        'clear_filter',
        'list_models',
        'get_model_fields',
        'query_model',
        'nl2filter',
      ],
    },
  })

  return (
    <ProjectCopilotActions
      orgName={orgName}
      projectSlug={projectSlug}
      workspaceAiRef={workspaceAiRef}
    />
  )
}

/**
 * Project-scoped layout component
 * 
 * Provides the main application layout with:
 * - Authentication guard
 * - Unified sidebar navigation
 * - Unified top bar with path-style breadcrumbs
 * - AI assistant (lazy-loaded)
 * - Project context synchronization
 * 
 * Route: /org/[orgName]/project/[projectSlug]/*
 */
export default function ProjectLayout({ children }: ProjectLayoutProps) {
  const { isLoading: authLoading } = useRequireAuth()
  const params = useParams()
  
  // Extract route parameters
  const projectSlug = useMemo(() => {
    return params?.projectSlug as string
  }, [params?.projectSlug])

  const orgName = useMemo(() => {
    return params?.orgName as string
  }, [params?.orgName])
  
  // Global state
  const selectedProject = useAppStore((state) => state.selectedProject)
  const setSelectedProject = useAppStore((state) => state.setSelectedProject)
  
  // Lazy-load CopilotKit only when user requests it
  const [showCopilot, setShowCopilot] = useState(false)

  // Ref for AI actions to access the workspace
  const workspaceAiRef = useRef<DevelopRecordWorkspaceAIRef | null>(null)

  // Sync URL path to store when project changes
  useEffect(() => {
    // Skip if no project slug in URL
    if (!projectSlug) return

    // Skip if project is already selected
    if (selectedProject?.slug === projectSlug) return

    // TODO: Replace with actual API call to fetch project details
    // For now, create a temporary project object
    setSelectedProject({
      id: projectSlug,
      slug: projectSlug,
      title: projectSlug === 'default' ? 'Default Project' : projectSlug,
      description: '',
      status: 'ACTIVE' as const,
      orgName: orgName || 'default',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
  }, [projectSlug, orgName, selectedProject?.slug, setSelectedProject])

  // Show loading state while checking authentication
  if (authLoading) {
    return <LoadingScreen message="Authenticating..." />
  }

  // Main content layout with route validation
  const mainContent = (
    <RouteValidator orgName={orgName} projectSlug={projectSlug}>
      <AppLayout showProjectNav>
        {children}
        {/* AI Assistant button - only show when CopilotKit not loaded */}
        {!showCopilot && (
          <AIAssistantButton onClick={() => setShowCopilot(true)} />
        )}
      </AppLayout>
    </RouteValidator>
  )

  // Conditionally wrap with CopilotKit when user activates it
  if (showCopilot) {
    return (
      <WorkspaceAIRefContext.Provider value={workspaceAiRef}>
        <CopilotWrapper selectedProject={selectedProject} orgName={orgName}>
          <ProjectAIContext
            orgName={orgName}
            projectSlug={projectSlug}
            workspaceAiRef={workspaceAiRef}
          />
          {mainContent}
        </CopilotWrapper>
      </WorkspaceAIRefContext.Provider>
    )
  }

  return (
    <WorkspaceAIRefContext.Provider value={workspaceAiRef}>
      {mainContent}
    </WorkspaceAIRefContext.Provider>
  )
}
