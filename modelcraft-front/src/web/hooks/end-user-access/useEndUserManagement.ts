'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

// src/web/hooks/end-user-access/useEndUserManagement.ts
// 用户管理页统一 hook：Org 用户列表 + Project 访问权限 + RBAC 角色/Bundle 分配
// 用户列表来源：GraphQL (Org scoped)
// 访问控制来源：GraphQL (Project scoped)
// 角色/Bundle 分配：GraphQL (Project scoped)

import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { getOrgScopedClient, useProjectScopedClient } from '@api-client/apollo/public'
import { LIST_END_USERS } from '@api-client/end-user/graphql-docs'
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
  LIST_PROJECT_END_USER_ACCESS,
  GRANT_PROJECT_END_USER_ACCESS,
  UPDATE_PROJECT_END_USER_ACCESS,
  REVOKE_PROJECT_END_USER_ACCESS,
} from '@/api-client/rbac'
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
  // User list (Org level, GraphQL)
  users: OrgEndUser[]
  usersLoading: boolean
  usersError: string | null
  reloadUsers: () => void

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

// ── GraphQL response types ───────────────────────────────────────────────────

interface OrgEndUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: OrgEndUser[]
    }
    error?: { message?: string }
  }
}

interface ProjectEndUserAccessNode {
  id: string
  endUser: {
    id: string
    username: string
  }
  permissionBundleId: string
  permissionBundleName: string
  grantedAt: string
}

interface ProjectEndUserAccessData {
  listProjectEndUserAccess?: {
    connection?: {
      nodes?: ProjectEndUserAccessNode[]
    }
    error?: { message?: string }
  }
}

interface GrantProjectEndUserAccessData {
  grantEndUserProjectAccess?: {
    error?: { message?: string }
  }
}

interface UpdateProjectEndUserAccessData {
  updateEndUserProjectAccess?: {
    error?: { message?: string }
  }
}

interface RevokeProjectEndUserAccessData {
  revokeEndUserProjectAccess?: {
    error?: { message?: string }
  }
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useEndUserManagement(
  orgName: string,
  projectSlug: string
): UseEndUserManagementReturn {
  const projectClient = useProjectScopedClient(projectSlug, orgName)

  // ── Org End Users (GraphQL) ────────────────────────────────────────────────

  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [usersLoading, setUsersLoading] = useState(false)
  const [usersError, setUsersError] = useState<string | null>(null)
  const [usersVersion, setUsersVersion] = useState(0)

  const reloadUsers = useCallback(() => setUsersVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName) return
    setUsersLoading(true)
    setUsersError(null)
    getOrgScopedClient()
      .query<OrgEndUsersData>({
        query: LIST_END_USERS,
        variables: { input: { first: 100 } },
        fetchPolicy: 'network-only',
      })
      .then(({ data }) => {
        const gqlError = data?.listEndUsers?.error
        if (gqlError?.message) throw new Error(gqlError.message)
        const nodes = data?.listEndUsers?.connection?.nodes ?? []
        const sorted = [...nodes].sort(
          (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        )
        setUsers(sorted)
      })
      .catch((e: unknown) => setUsersError(getErrorMessage(e, '加载用户失败')))
      .finally(() => setUsersLoading(false))
  }, [orgName, usersVersion])

  // ── Project Access (GraphQL) ───────────────────────────────────────────────

  const [accesses, setAccesses] = useState<EndUserProjectAccessEntry[]>([])
  const [accessLoading, setAccessLoading] = useState(false)
  const [accessError, setAccessError] = useState<string | null>(null)
  const [accessVersion, setAccessVersion] = useState(0)

  const reloadAccess = useCallback(() => setAccessVersion((v) => v + 1), [])

  useEffect(() => {
    if (!orgName || !projectSlug) return
    setAccessLoading(true)
    setAccessError(null)
    projectClient
      .query<ProjectEndUserAccessData>({
        query: LIST_PROJECT_END_USER_ACCESS,
        variables: { input: { first: 100 } },
        fetchPolicy: 'network-only',
      })
      .then(({ data }) => {
        const gqlError = data?.listProjectEndUserAccess?.error
        if (gqlError?.message) throw new Error(gqlError.message)
        const mapped: EndUserProjectAccessEntry[] =
          data?.listProjectEndUserAccess?.connection?.nodes?.map((node) => ({
            accessId: node.id,
            userId: node.endUser.id,
            username: node.endUser.username,
            permissionBundle: node.permissionBundleName,
            grantedAt: node.grantedAt,
          })) ?? []
        setAccesses(mapped)
      })
      .catch((e: unknown) => setAccessError(getErrorMessage(e, '加载访问列表失败')))
      .finally(() => setAccessLoading(false))
  }, [orgName, projectSlug, accessVersion, projectClient])

  const grantAccess = useCallback(
    async (userId: string, permissionBundleId: string) => {
      const { data } = await projectClient.mutate<GrantProjectEndUserAccessData>({
        mutation: GRANT_PROJECT_END_USER_ACCESS,
        variables: {
          input: {
            endUserId: userId,
            permissionBundleId,
          },
        },
      })
      const gqlError = data?.grantEndUserProjectAccess?.error
      if (gqlError?.message) throw new Error(gqlError.message)
      reloadAccess()
    },
    [projectClient, reloadAccess]
  )

  const revokeAccess = useCallback(
    async (accessId: string) => {
      const { data } = await projectClient.mutate<RevokeProjectEndUserAccessData>({
        mutation: REVOKE_PROJECT_END_USER_ACCESS,
        variables: {
          input: { accessId },
        },
      })
      const gqlError = data?.revokeEndUserProjectAccess?.error
      if (gqlError?.message) throw new Error(gqlError.message)
      reloadAccess()
    },
    [projectClient, reloadAccess]
  )

  const updateAccessBundle = useCallback(
    async (accessId: string, permissionBundleId: string) => {
      const { data } = await projectClient.mutate<UpdateProjectEndUserAccessData>({
        mutation: UPDATE_PROJECT_END_USER_ACCESS,
        variables: {
          input: {
            accessId,
            permissionBundleId,
          },
        },
      })
      const gqlError = data?.updateEndUserProjectAccess?.error
      if (gqlError?.message) throw new Error(gqlError.message)
      reloadAccess()
    },
    [projectClient, reloadAccess]
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
