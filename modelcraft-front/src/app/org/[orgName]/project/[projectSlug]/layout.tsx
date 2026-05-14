'use client'

import { useEffect, useState, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { LoadingScreen } from '@web/components/common/LoadingScreen'
import { CopilotWrapper, AIAssistantButton } from '@web/components/features/copilot/CopilotProvider'
import { RouteValidator } from '@web/components/common/RouteValidator'
import { useAppStore } from '@web/stores/app'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import "@copilotkit/react-ui/styles.css"

interface ProjectLayoutProps {
  children: React.ReactNode
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
      <CopilotWrapper selectedProject={selectedProject} orgName={orgName}>
        {mainContent}
      </CopilotWrapper>
    )
  }

  return mainContent
}
