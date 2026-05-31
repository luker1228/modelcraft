'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { KeyRound } from 'lucide-react'
import { useQuery } from '@apollo/client'
import { EndUserAppLayout } from '@web/components/features/layout'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: { orgName: string }
}

// ─── Token 管理占位内容 ─────────────────────────────────────────────────────

function TokenManagementPlaceholder() {
  return (
    <div className="flex h-full flex-col p-6">
      <div className="mb-6">
        <h2 className="text-[18px] font-semibold text-foreground">Token 管理</h2>
        <p className="mt-1 text-[13px] text-muted-foreground">
          个人访问 Token，用于 CLI 和 API 调用
        </p>
      </div>

      <div className="flex flex-1 flex-col items-center justify-center rounded-lg border border-dashed border-border bg-background py-24 text-center">
        <div className="mb-4 flex size-12 items-center justify-center rounded-xl border border-border bg-card shadow-sm">
          <KeyRound className="size-5 text-muted-foreground/50" strokeWidth={1.5} />
        </div>
        <p className="text-[15px] font-semibold text-foreground">即将推出</p>
        <p className="mt-1.5 max-w-xs text-[13px] leading-relaxed text-muted-foreground">
          通过个人访问 Token，你可以在 CLI 工具和外部 API 中以当前身份进行安全认证。
        </p>
        <span className="mt-5 inline-flex items-center rounded-full border border-border bg-muted px-3 py-1 text-xs text-muted-foreground">
          开发中
        </span>
      </div>
    </div>
  )
}

// ─── 项目列表内容（主页） ────────────────────────────────────────────────────

function ProjectsContent({ orgName }: { orgName: string }) {
  const { data, loading } = useQuery<{ endUserProjects: EndUserAccessibleProject[] }>(
    END_USER_PROJECTS
  )
  const projects = data?.endUserProjects ?? []

  return (
    <div className="flex flex-col p-6">
      <div className="mb-6">
        <h2 className="text-[18px] font-semibold text-foreground">项目</h2>
        <p className="mt-1 text-[13px] text-muted-foreground">选择项目进入数据管理</p>
      </div>
      <WorkspaceProjectsTab orgName={orgName} projects={projects} loading={loading} />
    </div>
  )
}

// ─── Token 恢复 hook ─────────────────────────────────────────────────────────

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

// ─── Page entry ──────────────────────────────────────────────────────────────

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const { orgName } = params
  const ready = useEndUserTokenReady(orgName)
  const [activeSection, setActiveSection] = useState<'projects' | 'tokens' | null>(null)

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="size-5 animate-spin rounded-full border-2 border-border border-t-foreground" />
      </div>
    )
  }

  return (
    <div className="h-screen overflow-hidden">
      <EndUserAppLayout
        orgName={orgName}
        activeSection={activeSection === 'tokens' ? 'tokens' : undefined}
        onSectionChange={setActiveSection}
      >
        {activeSection === 'tokens' ? (
          <TokenManagementPlaceholder />
        ) : (
          <ProjectsContent orgName={orgName} />
        )}
      </EndUserAppLayout>
    </div>
  )
}
