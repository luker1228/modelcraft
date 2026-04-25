'use client'

// src/web/hooks/end-user-access/useEndUserManagement.ts
// 用户管理页统一 hook：Org 用户列表 + Project 访问权限 + RBAC 角色/Bundle 分配
// 用户列表来源：REST BFF (Org 级)
// 访问控制来源：REST BFF (Project 级)
// 角色/Bundle 分配：GraphQL (Project scoped)

import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { useProjectScopedClient } from '@bff/apollo/public'
import {
  GET_END_USER_ROLES,
  GET_END_USER_BUNDLES,
  GET_END_USER_ROLE_ASSIGNMENTS,
  GET_END_USER_BUNDLE_ASSIGNMENTS,
  GET_END_USER_EFFECTIVE_PERMISSIONS,
  ASSIGN_END_USER_ROLE_TO_USER,
  REVOKE_END_USER_ROLE_FROM_USER,
  ASSIGN_BUNDLE_TO_END_USER,
  REVOKE_BUNDLE_FROM_END_USER,
} from '@web/graphql'
import type { EndUserRole, EndUserPermissionBundle, EffectivePermissions } from '@/types'
import type { OrgEndUser } from '@web/hooks/end-users/useOrgEndUsers'
import type { EndUserProjectAccessEntry } from './useProjectEndUserAccess'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface UserRoleAssignment {
  endUserId: string
  assignedAt: string
  role: EndUserRole
}

export interface UserBundleAssignment {
  endUserId: string
  assignedAt: string
  bundle: EndUserPermissionBundle
}

interface MutationResult {
  success: boolean
  errorMessage?: string
}

export interface UseEndUserManagementReturn {
  // User list (Org level)
  users: OrgEndUser[]
  usersLoading: boolean
  usersError: string | null
  reloadUsers: () => void
  createUser: (payload: { username: string; password: string; displayName?: string }) => Promise<void>
  toggleUserStatus: (userId: string, status: 'ACTIVE' | 'DISABLED') => Promise<void>
  deleteUser: (userId: string) => Promise<void>

  // Project access (Project level)
  accesses: EndUserProjectAccessEntry[]
  accessLoading: boolean
  accessError: string | null
  reloadAccess: () => void
  grantAccess: (userId: string, permissionBundleId: string) => Promise<void>
  revokeAccess: (accessId: string) => Promise<void>
  updateAccessBundle: (accessId: string, permissionBundleId: string) => Promise<void>

  // RBAC roles (all roles in project)
  allRoles: EndUserRole[]
  rolesLoading: boolean

  // RBAC bundles (all bundles in project)
  allBundles: EndUserPermissionBundle[]
  bundlesLoading: boolean

  // Per-user role/bundle assignments (loaded on demand when sheet opens)
  selectedUserId: string | null
  setSelectedUserId: (id: string | null) => void
  userRoleAssignments: UserRoleAssignment[]
  userRoleAssignmentsLoading: boolean
  userBundleAssignments: UserBundleAssignment[]
  userBundleAssignmentsLoading: boolean
  effectivePermissions: EffectivePermissions[]
  effectiveLoading: boolean

  // Role/bundle mutation helpers
  assignRole: (endUserId: string, roleId: string) => Promise<MutationResult>
  revokeRole: (endUserId: string, roleId: string) => Promise<MutationResult>
  assignBundle: (endUserId: string, bundleId: string) => Promise<MutationResult>
  revokeBundle: (endUserId: string, bundleId: string) => Promise<MutationResult>
}

// ── BFF response types ────────────────────────────────────────────────────────

interface AccessListResponse {
  accesses: EndUserProjectAccessEntry[]
  error?: { message?: string }
}

interface BffErrorResponse {
  error?: { message?: string }
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useEndUserManagement(
  orgName: string,
  projectSlug: string
): UseEndUserManagementReturn {
  const projectClient = useProjectScopedClient(projectSlug, orgName)

  // ── Org User List (REST BFF) ───────────────────────────────────────────────

  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [usersLoading, setUsersLoading] = useState(false)
  const [usersError, setUsersError] = useState<string | null>(null)
  const [usersVersion, setUsersVersion] = useState(0)

  const reloadUsers = useCallback(() => setUsersVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName) return
    setUsersLoading(true)
    setUsersError(null)
    fetch(`/api/bff/org/${orgName}/end-user/users`)
      .then(async (res) => {
        if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '加载失败')
        return res.json() as Promise<{ users: OrgEndUser[] }>
      })
      .then((d) => setUsers(d.users))
      .catch((e: unknown) => setUsersError(e instanceof Error ? e.message : '加载用户失败'))
      .finally(() => setUsersLoading(false))
  }, [orgName, usersVersion])

  const createUser = useCallback(
    async (payload: { username: string; password: string; displayName?: string }) => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '创建失败')
      reloadUsers()
    },
    [orgName, reloadUsers]
  )

  const toggleUserStatus = useCallback(
    async (userId: string, status: 'ACTIVE' | 'DISABLED') => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users/${userId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      })
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '更新状态失败')
      reloadUsers()
    },
    [orgName, reloadUsers]
  )

  const deleteUser = useCallback(
    async (userId: string) => {
      const res = await fetch(`/api/bff/org/${orgName}/end-user/users/${userId}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '删除失败')
      reloadUsers()
    },
    [orgName, reloadUsers]
  )

  // ── Project Access (REST BFF) ──────────────────────────────────────────────

  const [accesses, setAccesses] = useState<EndUserProjectAccessEntry[]>([])
  const [accessLoading, setAccessLoading] = useState(false)
  const [accessError, setAccessError] = useState<string | null>(null)
  const [accessVersion, setAccessVersion] = useState(0)

  const reloadAccess = useCallback(() => setAccessVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName || !projectSlug) return
    setAccessLoading(true)
    setAccessError(null)
    fetch(`/api/bff/org/${orgName}/project/${projectSlug}/end-user-access`)
      .then(async (res) => {
        if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '加载失败')
        return res.json() as Promise<AccessListResponse>
      })
      .then((d) => setAccesses(d.accesses))
      .catch((e: unknown) => setAccessError(e instanceof Error ? e.message : '加载访问列表失败'))
      .finally(() => setAccessLoading(false))
  }, [orgName, projectSlug, accessVersion])

  const grantAccess = useCallback(
    async (userId: string, permissionBundleId: string) => {
      const res = await fetch(`/api/bff/org/${orgName}/project/${projectSlug}/end-user-access`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ userId, permissionBundle: permissionBundleId }),
      })
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '授权失败')
      reloadAccess()
    },
    [orgName, projectSlug, reloadAccess]
  )

  const revokeAccess = useCallback(
    async (accessId: string) => {
      const res = await fetch(
        `/api/bff/org/${orgName}/project/${projectSlug}/end-user-access/${accessId}`,
        { method: 'DELETE' }
      )
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '撤销失败')
      reloadAccess()
    },
    [orgName, projectSlug, reloadAccess]
  )

  const updateAccessBundle = useCallback(
    async (accessId: string, permissionBundleId: string) => {
      const res = await fetch(
        `/api/bff/org/${orgName}/project/${projectSlug}/end-user-access/${accessId}`,
        {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ permissionBundle: permissionBundleId }),
        }
      )
      if (!res.ok) throw new Error(((await res.json()) as BffErrorResponse)?.error?.message ?? '更新权限失败')
      reloadAccess()
    },
    [orgName, projectSlug, reloadAccess]
  )

  // ── All Roles & Bundles (GraphQL, project-scoped) ─────────────────────────

  const { data: rolesData, loading: rolesLoading } = useQuery(GET_END_USER_ROLES, {
    client: projectClient,
    skip: !projectSlug || !orgName,
  })

  const { data: bundlesData, loading: bundlesLoading } = useQuery(GET_END_USER_BUNDLES, {
    client: projectClient,
    skip: !projectSlug || !orgName,
  })

  const allRoles: EndUserRole[] =
    rolesData?.endUserRoles?.edges?.map((e: { node: EndUserRole }) => e.node) ?? []
  const allBundles: EndUserPermissionBundle[] =
    bundlesData?.endUserPermissionBundles?.edges?.map((e: { node: EndUserPermissionBundle }) => e.node) ?? []

  // ── Per-user assignments (loaded on demand) ───────────────────────────────

  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)

  const { data: roleAssignData, loading: userRoleAssignmentsLoading } = useQuery(
    GET_END_USER_ROLE_ASSIGNMENTS,
    {
      client: projectClient,
      variables: { endUserId: selectedUserId ?? '' },
      skip: !selectedUserId,
    }
  )

  const { data: bundleAssignData, loading: userBundleAssignmentsLoading } = useQuery(
    GET_END_USER_BUNDLE_ASSIGNMENTS,
    {
      client: projectClient,
      variables: { endUserId: selectedUserId ?? '' },
      skip: !selectedUserId,
    }
  )

  const { data: effectiveData, loading: effectiveLoading } = useQuery(
    GET_END_USER_EFFECTIVE_PERMISSIONS,
    {
      client: projectClient,
      variables: { endUserId: selectedUserId ?? '', modelId: '' },
      skip: !selectedUserId,
    }
  )

  const userRoleAssignments: UserRoleAssignment[] = roleAssignData?.endUserRoleAssignments ?? []
  const userBundleAssignments: UserBundleAssignment[] = bundleAssignData?.endUserBundleAssignments ?? []
  const rawEffective = effectiveData?.effectivePermissions?.effectivePermissions
  const effectivePermissions: EffectivePermissions[] = rawEffective ? [rawEffective] : []

  // ── RBAC Mutations ────────────────────────────────────────────────────────

  const [assignRoleMutation] = useMutation(ASSIGN_END_USER_ROLE_TO_USER, {
    client: projectClient,
    refetchQueries: [{ query: GET_END_USER_ROLE_ASSIGNMENTS, variables: { endUserId: selectedUserId ?? '' } }],
  })

  const [revokeRoleMutation] = useMutation(REVOKE_END_USER_ROLE_FROM_USER, {
    client: projectClient,
    refetchQueries: [{ query: GET_END_USER_ROLE_ASSIGNMENTS, variables: { endUserId: selectedUserId ?? '' } }],
  })

  const [assignBundleMutation] = useMutation(ASSIGN_BUNDLE_TO_END_USER, {
    client: projectClient,
    refetchQueries: [{ query: GET_END_USER_BUNDLE_ASSIGNMENTS, variables: { endUserId: selectedUserId ?? '' } }],
  })

  const [revokeBundleMutation] = useMutation(REVOKE_BUNDLE_FROM_END_USER, {
    client: projectClient,
    refetchQueries: [{ query: GET_END_USER_BUNDLE_ASSIGNMENTS, variables: { endUserId: selectedUserId ?? '' } }],
  })

  const assignRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      const result = await assignRoleMutation({ variables: { input: { endUserId, roleId } } })
      const payload = result.data?.assignEndUserRole
      if (payload?.error) return { success: false, errorMessage: payload.error.message ?? '分配角色失败' }
      return { success: true }
    },
    [assignRoleMutation]
  )

  const revokeRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      const result = await revokeRoleMutation({ variables: { input: { endUserId, roleId } } })
      const payload = result.data?.revokeEndUserRole
      if (payload?.error) return { success: false, errorMessage: payload.error.message ?? '撤销角色失败' }
      return { success: true }
    },
    [revokeRoleMutation]
  )

  const assignBundle = useCallback(
    async (endUserId: string, bundleId: string): Promise<MutationResult> => {
      const result = await assignBundleMutation({ variables: { input: { endUserId, bundleId } } })
      const payload = result.data?.assignBundleToEndUser
      if (payload?.error) return { success: false, errorMessage: payload.error.message ?? '授权权限包失败' }
      return { success: true }
    },
    [assignBundleMutation]
  )

  const revokeBundle = useCallback(
    async (endUserId: string, bundleId: string): Promise<MutationResult> => {
      const result = await revokeBundleMutation({ variables: { input: { endUserId, bundleId } } })
      const payload = result.data?.revokeBundleFromEndUser
      if (payload?.error) return { success: false, errorMessage: payload.error.message ?? '撤销权限包失败' }
      return { success: true }
    },
    [revokeBundleMutation]
  )

  return {
    users,
    usersLoading,
    usersError,
    reloadUsers,
    createUser,
    toggleUserStatus,
    deleteUser,

    accesses,
    accessLoading,
    accessError,
    reloadAccess,
    grantAccess,
    revokeAccess,
    updateAccessBundle,

    allRoles,
    rolesLoading,
    allBundles,
    bundlesLoading,

    selectedUserId,
    setSelectedUserId,
    userRoleAssignments,
    userRoleAssignmentsLoading,
    userBundleAssignments,
    userBundleAssignmentsLoading,
    effectivePermissions,
    effectiveLoading,

    assignRole,
    revokeRole,
    assignBundle,
    revokeBundle,
  }
}
