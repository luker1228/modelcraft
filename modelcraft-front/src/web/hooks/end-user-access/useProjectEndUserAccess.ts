'use client'

// src/web/hooks/end-user-access/useProjectEndUserAccess.ts
// Project 级终端用户访问控制 hook（EndUser v2）

import { useState, useEffect, useCallback } from 'react'

export interface EndUserProjectAccessEntry {
  accessId: string
  userId: string
  username: string
  displayName?: string
  permissionBundle?: string
  grantedAt: string
}

export interface GrantAccessPayload {
  userId: string
  permissionBundle?: string
}

interface UseProjectEndUserAccessReturn {
  accesses: EndUserProjectAccessEntry[]
  isLoading: boolean
  error: string | null
  reload: () => void
  grantAccess: (payload: GrantAccessPayload) => Promise<void>
  revokeAccess: (accessId: string) => Promise<void>
  updatePermissionBundle: (accessId: string, permissionBundle: string) => Promise<void>
}

export function useProjectEndUserAccess(
  orgName: string,
  projectSlug: string
): UseProjectEndUserAccessReturn {
  const [accesses, setAccesses] = useState<EndUserProjectAccessEntry[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName || !projectSlug) return
    setIsLoading(true)
    setError(null)

    fetch(`/api/bff/org/${orgName}/project/${projectSlug}/end-user-access`)
      .then(async (res) => {
        if (!res.ok) throw new Error((await res.json())?.error?.message ?? '加载失败')
        return res.json() as Promise<{ accesses: EndUserProjectAccessEntry[] }>
      })
      .then((data) => setAccesses(data.accesses))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : '加载访问控制列表失败'))
      .finally(() => setIsLoading(false))
  }, [orgName, projectSlug, version])

  const grantAccess = useCallback(
    async (payload: GrantAccessPayload) => {
      const res = await fetch(
        `/api/bff/org/${orgName}/project/${projectSlug}/end-user-access`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        }
      )
      if (!res.ok) {
        const data = await res.json()
        throw new Error(data?.error?.message ?? '授权失败')
      }
      reload()
    },
    [orgName, projectSlug, reload]
  )

  const revokeAccess = useCallback(
    async (accessId: string) => {
      const res = await fetch(
        `/api/bff/org/${orgName}/project/${projectSlug}/end-user-access/${accessId}`,
        { method: 'DELETE' }
      )
      if (!res.ok) {
        const data = await res.json()
        throw new Error(data?.error?.message ?? '撤销授权失败')
      }
      reload()
    },
    [orgName, projectSlug, reload]
  )

  const updatePermissionBundle = useCallback(
    async (accessId: string, permissionBundle: string) => {
      const res = await fetch(
        `/api/bff/org/${orgName}/project/${projectSlug}/end-user-access/${accessId}`,
        {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ permissionBundle }),
        }
      )
      if (!res.ok) {
        const data = await res.json()
        throw new Error(data?.error?.message ?? '更新权限失败')
      }
      reload()
    },
    [orgName, projectSlug, reload]
  )

  return { accesses, isLoading, error, reload, grantAccess, revokeAccess, updatePermissionBundle }
}
