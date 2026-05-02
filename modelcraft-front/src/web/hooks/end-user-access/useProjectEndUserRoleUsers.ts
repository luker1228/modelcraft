'use client'

// src/web/hooks/end-user-access/useProjectEndUserRoleUsers.ts
// Project 级终端用户角色分配列表 hook（EndUser Access Redesign）
// 数据来源：GraphQL listProjectEndUserRoleUsers

import { useCallback } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  LIST_PROJECT_END_USER_ROLE_USERS,
  ASSIGN_END_USER_ROLE_TO_USER,
  REVOKE_END_USER_ROLE_FROM_USER,
  GET_END_USER_ROLES,
} from '@/api-client/rbac'
import { LIST_END_USERS } from '@api-client/end-user/graphql-docs'
import { getOrgScopedClient } from '@api-client/apollo/public'
import type { EndUserRole } from '@/types'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ProjectRoleUserEntry {
  assignmentId?: string
  endUser: {
    id: string
    username: string
    isForbidden: boolean
  }
  role: {
    id: string
    name: string
    description?: string | null
  }
  assignedAt: string
}

export interface OrgEndUserOption {
  id: string
  username: string
  isForbidden: boolean
}

interface MutationResult {
  success: boolean
  errorMessage?: string
}

// ── GraphQL response types ────────────────────────────────────────────────────

interface ListProjectEndUserRoleUsersData {
  listProjectEndUserRoleUsers?: {
    connection?: {
      nodes?: Array<{
        endUser: { id: string; username: string; isForbidden: boolean }
        role: { id: string; name: string; description?: string | null }
        assignedAt: string
      }>
      pageInfo?: { hasNextPage: boolean; endCursor?: string | null }
      totalCount?: number
    }
    error?: { __typename?: string; message?: string }
  }
}

interface ListEndUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: OrgEndUserOption[]
    }
    error?: { message?: string }
  }
}

interface GetEndUserRolesData {
  endUserRoles?: {
    edges?: Array<{
      node: EndUserRole
    }>
  }
}

// ── Mutation response types ───────────────────────────────────────────────────

interface AssignEndUserRoleData {
  assignEndUserRole?: {
    endUserId?: string
    role?: { id: string; name: string }
    error?: { __typename?: string; message?: string }
  }
}

interface RevokeEndUserRoleData {
  revokeEndUserRole?: {
    success?: boolean
    error?: { __typename?: string; message?: string }
  }
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useProjectEndUserRoleUsers(orgName: string, projectSlug: string) {
  const projectClient = useProjectScopedClient(projectSlug, orgName)
  const orgClient = getOrgScopedClient(orgName)

  // 列出该 Project 下的所有角色分配
  const {
    data: roleUsersData,
    loading: roleUsersLoading,
    error: roleUsersError,
    refetch: refetchRoleUsers,
  } = useQuery<ListProjectEndUserRoleUsersData>(LIST_PROJECT_END_USER_ROLE_USERS, {
    client: projectClient,
    variables: { input: {} },
    skip: !orgName || !projectSlug,
    fetchPolicy: 'cache-and-network',
  })

  // 列出 Org 内所有 EndUser（用于「添加用户」下拉）
  const { data: orgUsersData } = useQuery<ListEndUsersData>(LIST_END_USERS, {
    client: orgClient,
    variables: { input: { first: 100 } },
    skip: !orgName,
    fetchPolicy: 'cache-and-network',
  })

  // 列出 Project 下所有 Role（用于「添加用户」下拉）
  const { data: rolesData } = useQuery<GetEndUserRolesData>(GET_END_USER_ROLES, {
    client: projectClient,
    variables: { input: { includeImplicit: false, first: 100 } },
    skip: !orgName || !projectSlug,
    fetchPolicy: 'cache-and-network',
  })

  const [assignRoleMutation] = useMutation<AssignEndUserRoleData>(ASSIGN_END_USER_ROLE_TO_USER, {
    client: projectClient,
  })
  const [revokeRoleMutation] = useMutation<RevokeEndUserRoleData>(REVOKE_END_USER_ROLE_FROM_USER, {
    client: projectClient,
  })

  // ── Derived data ──────────────────────────────────────────────────────────

  const entries: ProjectRoleUserEntry[] =
    roleUsersData?.listProjectEndUserRoleUsers?.connection?.nodes?.map((n, idx) => ({
      assignmentId: `${n.endUser.id}-${n.role.id}-${idx}`,
      endUser: n.endUser,
      role: n.role,
      assignedAt: n.assignedAt,
    })) ?? []

  const orgUsers: OrgEndUserOption[] =
    orgUsersData?.listEndUsers?.connection?.nodes ?? []

  const allRoles: EndUserRole[] =
    rolesData?.endUserRoles?.edges?.map((e) => e.node) ?? []

  const availableRoles = allRoles.filter((r) => !r.isImplicit)

  // ── Mutations ─────────────────────────────────────────────────────────────

  const assignRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      try {
        const { data } = await assignRoleMutation({
          variables: { input: { endUserId, roleId } },
        })
        const err = data?.assignEndUserRole?.error
        if (err) {
          return { success: false, errorMessage: err.message ?? '分配失败' }
        }
        await refetchRoleUsers()
        return { success: true }
      } catch (e) {
        return { success: false, errorMessage: e instanceof Error ? e.message : '分配失败' }
      }
    },
    [assignRoleMutation, refetchRoleUsers]
  )

  const revokeRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      try {
        const { data } = await revokeRoleMutation({
          variables: { input: { endUserId, roleId } },
        })
        const err = data?.revokeEndUserRole?.error
        if (err) {
          return { success: false, errorMessage: err.message ?? '撤销失败' }
        }
        await refetchRoleUsers()
        return { success: true }
      } catch (e) {
        return { success: false, errorMessage: e instanceof Error ? e.message : '撤销失败' }
      }
    },
    [revokeRoleMutation, refetchRoleUsers]
  )

  return {
    entries,
    loading: roleUsersLoading,
    error: roleUsersError?.message ?? roleUsersData?.listProjectEndUserRoleUsers?.error?.message ?? null,
    reload: refetchRoleUsers,
    orgUsers,
    availableRoles,
    assignRole,
    revokeRole,
  }
}
