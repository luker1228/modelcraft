'use client'

// src/web/hooks/end-users/useOrgEndUsers.ts
// Org 级终端用户管理 hook（EndUser v2）
// 使用 Org-Scoped GraphQL（/api/bff/graphql/org/{orgName}/）

import { useState, useEffect, useCallback } from 'react'
import {
  LIST_END_USERS,
  CREATE_END_USER,
  UPDATE_END_USER_STATUS,
  DELETE_END_USER,
  RESET_END_USER_PASSWORD,
} from '@api-client/end-user/graphql-docs'
import { getOrgScopedClient } from '@api-client/apollo/develop-client'

// ── Types ──────────────────────────────────────────────────────────────────

export interface OrgEndUser {
  id: string
  username: string
  displayName?: string
  isForbidden: boolean
  createdBy?: string
  createdAt: string
  updatedAt: string
}

export interface CreateEndUserPayload {
  username: string
  password: string
  phone: string
}

interface UseOrgEndUsersReturn {
  users: OrgEndUser[]
  isLoading: boolean
  error: string | null
  search: string
  setSearch: (s: string) => void
  reload: () => void
  createUser: (payload: CreateEndUserPayload) => Promise<void>
  toggleUserStatus: (userId: string, isForbidden: boolean) => Promise<void>
  deleteUser: (userId: string) => Promise<void>
  resetPassword: (userId: string, newPassword: string) => Promise<void>
}

// ── GraphQL response shapes ─────────────────────────────────────────────────

interface ListEndUsersData {
  listEndUsers: {
    connection?: {
      nodes: OrgEndUser[]
      pageInfo: { hasNextPage: boolean; endCursor?: string }
      totalCount: number
    }
    error?: { __typename: string; message?: string }
  }
}

interface CreateEndUserData {
  createEndUser: {
    endUser?: OrgEndUser
    error?: { __typename: string; message?: string; suggestion?: string }
  }
}

interface UpdateEndUserStatusData {
  updateEndUserStatus: {
    endUser?: OrgEndUser
    error?: { __typename: string; message?: string }
  }
}

interface DeleteEndUserData {
  deleteEndUser: {
    success: boolean
    error?: { __typename: string; message?: string }
  }
}

interface ResetEndUserPasswordData {
  resetEndUserPassword: {
    success: boolean
    error?: { __typename: string; message?: string; suggestion?: string }
  }
}

// ── Hook ────────────────────────────────────────────────────────────────────

export function useOrgEndUsers(_orgName: string): UseOrgEndUsersReturn {
  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Debounce search by 300ms
  useEffect(() => {
    const id = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(id)
  }, [search])

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    let cancelled = false
    const client = getOrgScopedClient()

    setIsLoading(true)
    setError(null)

    const variables = {
      input: {
        first: 100,
        ...(debouncedSearch ? { search: debouncedSearch } : {}),
      },
    }

    client
      .query<ListEndUsersData>({
        query: LIST_END_USERS,
        variables,
        fetchPolicy: 'network-only',
      })
      .then(({ data }) => {
        if (cancelled) return
        const gqlError = data?.listEndUsers?.error
        if (gqlError?.message) {
          setError(gqlError.message)
          return
        }
        const nodes = data?.listEndUsers?.connection?.nodes ?? []
        // Sort by createdAt descending (newest first)
        const sorted = [...nodes].sort(
          (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        )
        setUsers(sorted)
      })
      .catch((e: unknown) => {
        if (cancelled) return
        setError(e instanceof Error ? e.message : '加载终端用户失败')
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [version, debouncedSearch])

  const createUser = useCallback(async (payload: CreateEndUserPayload) => {
    const client = getOrgScopedClient()
    const { data } = await client.mutate<CreateEndUserData>({
      mutation: CREATE_END_USER,
      variables: { input: { username: payload.username, password: payload.password, phone: payload.phone } },
    })
    const err = data?.createEndUser?.error
    if (err) {
      throw new Error(err.message ?? '创建用户失败')
    }
    reload()
  }, [reload])

  const toggleUserStatus = useCallback(async (userId: string, isForbidden: boolean) => {
    const client = getOrgScopedClient()
    const { data } = await client.mutate<UpdateEndUserStatusData>({
      mutation: UPDATE_END_USER_STATUS,
      variables: { input: { userId, isForbidden } },
    })
    const err = data?.updateEndUserStatus?.error
    if (err) {
      throw new Error(err.message ?? '更新用户状态失败')
    }
    reload()
  }, [reload])

  const deleteUser = useCallback(async (userId: string) => {
    const client = getOrgScopedClient()
    const { data } = await client.mutate<DeleteEndUserData>({
      mutation: DELETE_END_USER,
      variables: { input: { userId } },
    })
    const err = data?.deleteEndUser?.error
    if (err) {
      throw new Error(err.message ?? '删除用户失败')
    }
    reload()
  }, [reload])

  const resetPassword = useCallback(async (userId: string, newPassword: string) => {
    const client = getOrgScopedClient()
    const { data } = await client.mutate<ResetEndUserPasswordData>({
      mutation: RESET_END_USER_PASSWORD,
      variables: { input: { userId, newPassword } },
    })
    const err = data?.resetEndUserPassword?.error
    if (err) {
      throw new Error(err.message ?? '重置密码失败')
    }
  }, [])

  return { users, isLoading, error, search, setSearch, reload, createUser, toggleUserStatus, deleteUser, resetPassword }
}
