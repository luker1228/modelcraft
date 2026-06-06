'use client'

import { useEffect, useState, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery } from '@apollo/client'
import { ApolloProvider } from '@apollo/client'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { EndUserAppLayout } from '@web/components/features/layout/EndUserAppLayout'
import { createEndUserOrgScopedClient } from '@api-client/apollo/end-user-client'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: { orgName: string }
}

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
  error?: { code?: string }
}

function useEndUserTokenReady(orgName: string): boolean {
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)
  const router = useRouter()

  const [ready, setReady] = useState(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) return true

    if (typeof window !== 'undefined') {
      const savedToken = sessionStorage.getItem(`eu_token_${orgName}`)
      const savedExpiresAt = Number(sessionStorage.getItem(`eu_token_expires_at_${orgName}`) ?? '0')
      if (savedToken && Date.now() < savedExpiresAt - 5 * 60 * 1000) {
        const expiresIn = Math.floor((savedExpiresAt - Date.now()) / 1000)
        useEndUserAuthStore.getState().setAccessToken(savedToken, expiresIn)
        return true
      }
    }
    return false
  })

  useEffect(() => {
    const storeState = useEndUserAuthStore.getState()
    if (storeState.accessToken && !storeState.isTokenExpired()) {
      setReady(true)
      return
    }

    void (async () => {
      try {
        const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/refresh`, {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ orgName }),
        })
        if (!res.ok) {
          router.replace(`/end-user/${orgName}/login`)
          return
        }
        const data = (await res.json()) as RefreshResponse
        if (data.accessToken) {
          let expiresIn = 3600
          if (data.expiresAt) {
            const ms = new Date(data.expiresAt).getTime() - Date.now()
            if (ms > 0) expiresIn = Math.floor(ms / 1000)
          }
          setAccessToken(data.accessToken, expiresIn)
          setReady(true)
        } else {
          router.replace(`/end-user/${orgName}/login`)
        }
      } catch {
        router.replace(`/end-user/${orgName}/login`)
      }
    })()
  }, [orgName, setAccessToken, router])

  return ready
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
