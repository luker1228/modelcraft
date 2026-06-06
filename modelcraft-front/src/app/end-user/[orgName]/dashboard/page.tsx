'use client'

import { useMemo } from 'react'
import { useQuery } from '@apollo/client'
import { ApolloProvider } from '@apollo/client'
import { EndUserAppLayout } from '@web/components/features/layout/EndUserAppLayout'
import { createEndUserOrgScopedClient } from '@api-client/apollo/end-user-client'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import { useEndUserTokenReady } from '@web/hooks/end-user/useEndUserTokenReady'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: { orgName: string }
}

function DashboardContent({ orgName }: { orgName: string }) {
  const { data, loading } = useQuery<{ endUserProjects: EndUserAccessibleProject[] }>(
    END_USER_PROJECTS
  )
  const projects = data?.endUserProjects ?? []

  return (
    <EndUserAppLayout orgName={orgName} activePage="projects">
      <div className="h-full overflow-y-auto">
        <WorkspaceProjectsTab orgName={orgName} projects={projects} loading={loading} />
      </div>
    </EndUserAppLayout>
  )
}

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const { orgName } = params
  const ready = useEndUserTokenReady(orgName)

  const client = useMemo(() => createEndUserOrgScopedClient(orgName), [orgName])

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="size-5 animate-spin rounded-full border-2 border-border border-t-foreground" />
      </div>
    )
  }

  return (
    <ApolloProvider client={client}>
      <DashboardContent orgName={orgName} />
    </ApolloProvider>
  )
}
