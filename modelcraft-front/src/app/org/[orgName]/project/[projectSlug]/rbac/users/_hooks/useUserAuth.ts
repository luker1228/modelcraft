import { useQuery, useMutation } from '@apollo/client'
import { useCallback, useState } from 'react'
import { useProjectScopedClient } from '@bff/apollo/public'
import {
  GET_END_USER_ROLES,
  GET_END_USER_BUNDLES,
  GET_END_USER_EFFECTIVE_PERMISSIONS,
  ASSIGN_END_USER_ROLE_TO_USER,
  REVOKE_END_USER_ROLE_FROM_USER,
  ASSIGN_BUNDLE_TO_END_USER,
  REVOKE_BUNDLE_FROM_END_USER,
} from '@web/graphql'
import type {
  EndUserRole,
  EndUserPermissionBundle,
  EffectivePermissions,
} from '@/types'

// ── Mock ───────────────────────────────────────────────────────────────────────
// GET_END_USER_LIST 尚未在后端实现，Wave 1 阶段使用固定 Mock 数组跑通 MSW 流程

export interface MockEndUserUser {
  id: string
  username: string
  createdAt: string
  /** 已分配的显式角色（isImplicit === false）*/
  assignedRoles: EndUserRole[]
  /** 直接授权的权限包 */
  assignedBundles: EndUserPermissionBundle[]
}

const MOCK_USERS: MockEndUserUser[] = [
  {
    id: 'u-1',
    username: 'alice',
    createdAt: '2026-04-10T08:00:00Z',
    assignedRoles: [],
    assignedBundles: [],
  },
  {
    id: 'u-2',
    username: 'bob',
    createdAt: '2026-04-12T10:30:00Z',
    assignedRoles: [],
    assignedBundles: [],
  },
  {
    id: 'u-3',
    username: 'carol',
    createdAt: '2026-04-15T14:00:00Z',
    assignedRoles: [],
    assignedBundles: [],
  },
]

// ── Props / Return Types ───────────────────────────────────────────────────────

interface UseUserAuthProps {
  orgName: string
  projectSlug: string
}

interface MutationResult {
  success: boolean
  errorMessage?: string
}

export interface UseUserAuthReturn {
  /** Wave 1: mock 用户列表 */
  users: MockEndUserUser[]
  /** 当前选中的用户（右侧 Sheet 打开时非 null） */
  selectedUser: MockEndUserUser | null
  setSelectedUser: (user: MockEndUserUser | null) => void

  /** 角色列表（含隐式角色，UI 层负责过滤 isImplicit） */
  roles: EndUserRole[]
  rolesLoading: boolean

  /** 权限包列表 */
  bundles: EndUserPermissionBundle[]
  bundlesLoading: boolean

  /** 选中用户的有效权限（三通道合并），仅在 selectedUser 存在时加载 */
  effectivePermissions: EffectivePermissions[]
  effectiveLoading: boolean

  /** Mutation helpers */
  assignRole: (endUserId: string, roleId: string) => Promise<MutationResult>
  revokeRole: (endUserId: string, roleId: string) => Promise<MutationResult>
  assignBundle: (endUserId: string, bundleId: string) => Promise<MutationResult>
  revokeBundle: (endUserId: string, bundleId: string) => Promise<MutationResult>
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useUserAuth({ orgName, projectSlug }: UseUserAuthProps): UseUserAuthReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  const [selectedUser, setSelectedUser] = useState<MockEndUserUser | null>(null)

  // ── Queries ────────────────────────────────────────────────────────────────

  const { data: rolesData, loading: rolesLoading } = useQuery(GET_END_USER_ROLES, {
    client,
    skip: !projectSlug || !orgName,
  })

  const { data: bundlesData, loading: bundlesLoading } = useQuery(GET_END_USER_BUNDLES, {
    client,
    skip: !projectSlug || !orgName,
  })

  const { data: effectiveData, loading: effectiveLoading } = useQuery(
    GET_END_USER_EFFECTIVE_PERMISSIONS,
    {
      client,
      variables: { endUserId: selectedUser?.id ?? '', modelId: '' },
      skip: !selectedUser,
    }
  )

  // ── Mutations ──────────────────────────────────────────────────────────────

  const [assignRoleMutation] = useMutation(ASSIGN_END_USER_ROLE_TO_USER, { client })
  const [revokeRoleMutation] = useMutation(REVOKE_END_USER_ROLE_FROM_USER, { client })
  const [assignBundleMutation] = useMutation(ASSIGN_BUNDLE_TO_END_USER, { client })
  const [revokeBundleMutation] = useMutation(REVOKE_BUNDLE_FROM_END_USER, { client })

  // ── Derived Data ───────────────────────────────────────────────────────────

  const roles: EndUserRole[] = rolesData?.endUserRoles?.edges?.map((edge: any) => edge.node) ?? []
  const bundles: EndUserPermissionBundle[] = bundlesData?.endUserPermissionBundles?.edges?.map((edge: any) => edge.node) ?? []

  // GET_END_USER_EFFECTIVE_PERMISSIONS 返回的是单个对象（endUserId + modelId + grants），
  // 后续扩展时后端可能返回数组，这里统一包成数组方便 UI 迭代
  const rawEffective = effectiveData?.effectivePermissions?.effectivePermissions
  const effectivePermissions: EffectivePermissions[] = rawEffective
    ? [rawEffective]
    : []

  // ── Mutation Helpers ───────────────────────────────────────────────────────

  const assignRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      const result = await assignRoleMutation({
        variables: { endUserId, roleId },
      })
      const payload = result.data?.assignEndUserRole
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '分配角色失败' }
      }
      return { success: true }
    },
    [assignRoleMutation]
  )

  const revokeRole = useCallback(
    async (endUserId: string, roleId: string): Promise<MutationResult> => {
      const result = await revokeRoleMutation({
        variables: { endUserId, roleId },
      })
      const payload = result.data?.revokeEndUserRole
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '撤销角色失败' }
      }
      return { success: true }
    },
    [revokeRoleMutation]
  )

  const assignBundle = useCallback(
    async (endUserId: string, bundleId: string): Promise<MutationResult> => {
      const result = await assignBundleMutation({
        variables: { endUserId, bundleId },
      })
      const payload = result.data?.assignBundleToEndUser
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '授权权限包失败' }
      }
      return { success: true }
    },
    [assignBundleMutation]
  )

  const revokeBundle = useCallback(
    async (endUserId: string, bundleId: string): Promise<MutationResult> => {
      const result = await revokeBundleMutation({
        variables: { endUserId, bundleId },
      })
      const payload = result.data?.revokeBundleFromEndUser
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '撤销权限包失败' }
      }
      return { success: true }
    },
    [revokeBundleMutation]
  )

  return {
    users: MOCK_USERS,
    selectedUser,
    setSelectedUser,
    roles,
    rolesLoading,
    bundles,
    bundlesLoading,
    effectivePermissions,
    effectiveLoading,
    assignRole,
    revokeRole,
    assignBundle,
    revokeBundle,
  }
}
