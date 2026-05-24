'use client'

import { useMemo, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useQuery } from '@apollo/client'
import { ApolloProvider } from '@apollo/client'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: { orgName: string }
}

function WorkspaceContent({ orgName }: { orgName: string }) {
  const router = useRouter()
  const activeTab = 'projects'

  const { data, loading } = useQuery<{ endUserProjects: EndUserAccessibleProject[] }>(
    END_USER_PROJECTS
  )

  const projects = data?.endUserProjects ?? []

  const handleLogout = async () => {
    if (!orgName) return
    await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    })
    useEndUserAuthStore.getState().clearSession()
    router.push(`/end-user/${orgName}/login`)
  }

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      {/* 顶部栏 */}
      <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-background px-6">
        <span className="text-base font-semibold text-foreground">{orgName}</span>
        <button
          onClick={() => void handleLogout()}
          className="text-sm text-destructive hover:underline"
        >
          登出
        </button>
      </header>

      {/* Tab 导航 */}
      <nav className="flex border-b bg-background px-6">
        <button
          onClick={() => router.push(`/end-user/${orgName}/workspace`)}
          className={`border-b-2 px-4 py-3 text-sm font-medium transition-colors ${
            activeTab === 'projects'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          Projects
        </button>
        <button
          onClick={() => router.push(`/end-user/${orgName}/workspace/cli`)}
          className="border-b-2 border-transparent px-4 py-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          CLI 使用
        </button>
      </nav>

      {/* 主内容 */}
      <main className="flex-1 p-6">
        {activeTab === 'projects' && (
          <WorkspaceProjectsTab orgName={orgName} projects={projects} loading={loading} />
        )}
      </main>
    </div>
  )
}

interface RefreshResponse {
  accessToken?: string
  expiresAt?: string
  error?: { code?: string }
}

/**
 * 页面挂载时，若 store 里没有有效 token，先尝试从 sessionStorage 恢复（登录时写入），
 * 再尝试用 refresh cookie 换取新 accessToken。
 * 返回 ready=true 后再渲染 Apollo 请求，避免以空 token 发出请求。
 */
function useEndUserTokenReady(orgName: string): boolean {
  const setAccessToken = useEndUserAuthStore((s) => s.setAccessToken)
  const router = useRouter()

  // 用 getState() 同步读，避免 hook 订阅延迟导致初始值错误
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

    // token 缺失或已过期，尝试 silent refresh
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

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const { orgName } = params
  const ready = useEndUserTokenReady(orgName)

  const client = useMemo(
    () => createEndUserOrgScopedClient(orgName),
    [orgName]
  )

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">加载中...</p>
      </div>
    )
  }

  return (
    <ApolloProvider client={client}>
      <WorkspaceContent orgName={orgName} />
    </ApolloProvider>
  )
}
