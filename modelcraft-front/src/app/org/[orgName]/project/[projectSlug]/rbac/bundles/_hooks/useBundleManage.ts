/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
import { useQuery, useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  GET_END_USER_BUNDLE,
  GET_END_USER_PERMISSIONS,
  GET_END_USER_BUNDLES,
  ADD_PERMISSION_TO_BUNDLE,
  REMOVE_PERMISSION_FROM_BUNDLE,
} from '@/api-client/rbac'
import type { EndUserPermission, EndUserPermissionBundle } from '@/types'

interface UseBundleManageProps {
  orgName: string
  projectSlug: string
  /** Currently selected bundle ID. When null, queries are skipped. */
  bundleId: string | null
}

interface MutationResult {
  success: boolean
  errorMessage?: string
}

interface UseBundleManageReturn {
  bundle: EndUserPermissionBundle | null
  allPermissions: EndUserPermission[]
  loading: boolean
  error: Error | undefined
  addPermission: (permissionId: string) => Promise<MutationResult>
  removePermission: (permissionId: string) => Promise<MutationResult>
}

export function useBundleManage({
  orgName,
  projectSlug,
  bundleId,
}: UseBundleManageProps): UseBundleManageReturn {
  const client = useProjectScopedClient(projectSlug, orgName)
  const skip = !orgName || !projectSlug || !bundleId

  const {
    data: bundleData,
    loading: bundleLoading,
    error: bundleError,
  } = useQuery(GET_END_USER_BUNDLE, {
    client,
    variables: { id: bundleId ?? '' },
    skip,
  })

  const {
    data: permissionsData,
    loading: permissionsLoading,
    error: permissionsError,
  } = useQuery(GET_END_USER_PERMISSIONS, {
    client,
    skip: !orgName || !projectSlug,
  })

  const [addPermissionMutation] = useMutation(ADD_PERMISSION_TO_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES],
  })

  const [removePermissionMutation] = useMutation(REMOVE_PERMISSION_FROM_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES],
  })

  const addPermission = useCallback(
    async (permissionId: string): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await addPermissionMutation({
        variables: { input: { bundleId, permissionId } },
      })
      const payload = result.data?.addEndUserPermissionToBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '添加权限点失败' }
      }
      return { success: true }
    },
    [addPermissionMutation, bundleId],
  )

  const removePermission = useCallback(
    async (permissionId: string): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await removePermissionMutation({
        variables: { input: { bundleId, permissionId } },
      })
      const payload = result.data?.removeEndUserPermissionFromBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '移除权限点失败' }
      }
      return { success: true }
    },
    [removePermissionMutation, bundleId],
  )

  const bundle: EndUserPermissionBundle | null = bundleData?.endUserPermissionBundle ?? null
  const allPermissions: EndUserPermission[] = permissionsData?.endUserPermissions?.edges?.map(
    (edge: any) => edge.node,
  ) ?? []

  return {
    bundle,
    allPermissions,
    loading: bundleLoading || permissionsLoading,
    error: bundleError ?? permissionsError,
    addPermission,
    removePermission,
  }
}
