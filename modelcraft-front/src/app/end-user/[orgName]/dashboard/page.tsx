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
  const isAdmin = useEndUserAuthStore((s) => s.isAdmin)

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

  const orgInitials = orgName
    .split(/[-_]/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('')

  return (
    <div className="flex min-h-dvh flex-col bg-[#FBFBFA]">
      {/* Topbar */}
      <header className="sticky top-0 z-20 flex h-12 items-center justify-between border-b border-[#EAEAEA] bg-white px-6">
        <div className="flex items-center gap-3">
          <span className="text-[13px] font-semibold tracking-tight text-[#111111]">ModelCraft</span>
          <span className="text-[#D8DDE5]">/</span>
          <div className="flex items-center gap-1.5">
            <div className="flex h-5 w-5 items-center justify-center rounded bg-[#F0EFF9] text-[9px] font-semibold text-[#4F46E5]">
              {orgInitials || orgName[0]?.toUpperCase()}
            </div>
            <span className="text-[13px] font-medium text-[#111111]">{orgName}</span>
          </div>
        </div>

        <div className="flex items-center gap-1">
          {isAdmin === true && (
            <button
              onClick={() => router.push(`/org/${orgName}/dashboard`)}
              className="inline-flex h-7 items-center rounded border border-[#EAEAEA] bg-white px-3 text-xs font-medium text-[#787774] transition-colors hover:border-[#D0CECC] hover:text-[#111111]"
            >
              管理端
            </button>
          )}
          <button
            onClick={() => void handleLogout()}
            className="inline-flex h-7 items-center rounded px-3 text-xs font-medium text-[#787774] transition-colors hover:text-[#111111]"
          >
            登出
          </button>
        </div>
      </header>

      {/* Page shell */}
      <div className="mx-auto w-full max-w-5xl flex-1 px-6 py-10">
        {/* Page header */}
        <div className="mb-10">
          <h1 className="text-[20px] font-semibold leading-tight tracking-[-0.02em] text-[#111111]">
            工作台
          </h1>
          <p className="mt-1.5 text-[13px] text-[#787774]">选择一个项目开始数据操作</p>
        </div>

        {/* Projects */}
        <WorkspaceProjectsTab orgName={orgName} projects={projects} loading={loading} />
      </div>
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

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const { orgName } = params
  const ready = useEndUserTokenReady(orgName)

  const client = useMemo(() => createEndUserOrgScopedClient(orgName), [orgName])

  if (!ready) {
    return (
      <div className="flex min-h-dvh items-center justify-center bg-[#FBFBFA]">
        <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#EAEAEA] border-t-[#111111]" />
      </div>
    )
  }

  return (
    <ApolloProvider client={client}>
      <WorkspaceContent orgName={orgName} />
    </ApolloProvider>
  )
}
