'use client'

// src/web/hooks/end-users/useOrgEndUsers.ts
// Org 级终端用户管理 hook（EndUser v2）

import { useState, useEffect, useCallback } from 'react'

export interface OrgEndUser {
  id: string
  username: string
  displayName?: string
  status: 'ACTIVE' | 'DISABLED'
  createdAt: string
}

export interface CreateEndUserPayload {
  username: string
  password: string
  displayName?: string
}

interface UseOrgEndUsersReturn {
  users: OrgEndUser[]
  isLoading: boolean
  error: string | null
  reload: () => void
  createUser: (payload: CreateEndUserPayload) => Promise<void>
  toggleUserStatus: (userId: string, status: 'ACTIVE' | 'DISABLED') => Promise<void>
  deleteUser: (userId: string) => Promise<void>
}

export function useOrgEndUsers(orgName: string): UseOrgEndUsersReturn {
  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName) return

    setIsLoading(true)
    setError(null)

    fetch(`/api/bff/org/${orgName}/end-user/users`, { cache: 'no-store' })
      .then(async (res) => {
        if (!res.ok) throw new Error(`加载失败: ${res.status}`)
        const data = (await res.json()) as { users?: OrgEndUser[] }
        setUsers(data.users ?? [])
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : '加载终端用户失败'))
      .finally(() => setIsLoading(false))
  }, [orgName, version])

  const createUser = useCallback(
    async (payload: CreateEndUserPayload) => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: payload.username, password: payload.password }),
      })
      if (!res.ok) {
        const data = (await res.json()) as { error?: { message?: string } }
        throw new Error(data.error?.message || '创建用户失败')
      }
      reload()
    },
    [orgName, reload]
  )

  const toggleUserStatus = useCallback(
    async (userId: string, status: 'ACTIVE' | 'DISABLED') => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users/${userId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      })
      if (!res.ok) {
        const data = (await res.json()) as { error?: { message?: string } }
        throw new Error(data.error?.message || '更新用户状态失败')
      }
      reload()
    },
    [orgName, reload]
  )

  const deleteUser = useCallback(
    async (userId: string) => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users/${userId}`, {
        method: 'DELETE',
      })
      if (!res.ok && res.status !== 204) {
        const data = (await res.json()) as { error?: { message?: string } }
        throw new Error(data.error?.message || '删除用户失败')
      }
      reload()
    },
    [orgName, reload]
  )

  return { users, isLoading, error, reload, createUser, toggleUserStatus, deleteUser }
}
