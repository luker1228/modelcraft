'use client'

// src/web/hooks/end-user-auth-v2/useEndUserProjectSelector.ts
// 选择 Project hook（EndUser v2）
// 从 sessionStorage 读取可访问的 Project 列表，选择后调用 select-project BFF

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

// ============================================================================
// BFF Response Types
// ============================================================================

interface SelectProjectBffResponse {
  error?: { code?: string; message?: string }
  selectedProject?: string
}

export interface UseEndUserProjectSelectorReturn {
  projects: EndUserAccessibleProject[]
  selectedSlug: string | null
  isLoading: boolean
  error: string | null
  selectProject: (projectSlug: string) => void
  confirmSelection: () => Promise<void>
}

/**
 * 终端用户选择 Project hook（EndUser v2）。
 *
 * projects 来源：从 sessionStorage 读取登录时写入的 `eu_accessible_projects_{orgName}`。
 * 若 sessionStorage 为空（刷新或直接访问此页），显示"请先登录"提示。
 */
export function useEndUserProjectSelector(orgName: string): UseEndUserProjectSelectorReturn {
  const router = useRouter()
  const setEndUserToken = useEndUserAuthStore((s) => s.setAccessToken)
  const [projects, setProjects] = useState<EndUserAccessibleProject[]>([])
  const [selectedSlug, setSelectedSlug] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Load projects from sessionStorage on mount
  useEffect(() => {
    const key = `eu_accessible_projects_${orgName}`
    const raw = sessionStorage.getItem(key)
    if (!raw) {
      setError('会话已过期，请重新登录')
      return
    }
    try {
      const parsed = JSON.parse(raw) as EndUserAccessibleProject[]
      setProjects(parsed)
      if (parsed.length > 0) {
        setSelectedSlug(parsed[0].slug)
      }
    } catch {
      setError('会话数据损坏，请重新登录')
    }
  }, [orgName])

  const selectProject = useCallback((projectSlug: string) => {
    setSelectedSlug(projectSlug)
    setError(null)
  }, [])

  const confirmSelection = useCallback(async () => {
    if (!selectedSlug) {
      setError('请选择一个项目')
      return
    }
    setIsLoading(true)
    setError(null)

    try {
      const refreshToken = sessionStorage.getItem(`eu_refresh_token_${orgName}`)
      if (!refreshToken) {
        setError('会话已过期，请重新登录')
        setTimeout(() => router.push(`/end-user/${orgName}/login`), 1500)
        return
      }

      const res = await fetch(`/api/end-user/auth/select-project`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          orgName,
          refreshToken,
          projectSlug: selectedSlug,
        }),
      })

      const data = (await res.json()) as SelectProjectBffResponse

      if (!res.ok) {
        if (res.status === 401) {
          setError('会话已过期，请重新登录')
          setTimeout(() => router.push(`/end-user/${orgName}/login`), 1500)
          return
        }
        setError(data.error?.message ?? '选择项目失败，请重试')
        return
      }

      // Clean up sessionStorage
      sessionStorage.removeItem(`eu_accessible_projects_${orgName}`)
      sessionStorage.removeItem(`eu_refresh_token_${orgName}`)
      sessionStorage.setItem(`eu_selected_project_${orgName}`, data.selectedProject ?? selectedSlug)

      // Select-project 只返回 selectedProject；真正 access token 通过 refresh 获取
      const refreshRes = await fetch('/api/end-user/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgName, refreshToken }),
      })

      if (!refreshRes.ok) {
        setError('会话已过期，请重新登录')
        setTimeout(() => router.push(`/end-user/${orgName}/login`), 1500)
        return
      }

      const refreshData = (await refreshRes.json()) as {
        accessToken?: string
        expiresAt?: string
      }
      const expiresIn = refreshData.expiresAt
        ? Math.max(1, Math.floor((new Date(refreshData.expiresAt).getTime() - Date.now()) / 1000))
        : 3600
      setEndUserToken(refreshData.accessToken ?? '', expiresIn)
      router.push(`/end-user/${orgName}/${data.selectedProject ?? selectedSlug}/data`)
    } catch {
      setError('网络错误，请检查连接后重试')
    } finally {
      setIsLoading(false)
    }
  }, [selectedSlug, orgName, router, setEndUserToken])

  return { projects, selectedSlug, isLoading, error, selectProject, confirmSelection }
}
