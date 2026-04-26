import { useQuery, useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  GET_END_USER_BUNDLE,
  GET_END_USER_PERMISSIONS,
  UPDATE_END_USER_BUNDLE,
  ADD_PERMISSION_TO_BUNDLE,
  REMOVE_PERMISSION_FROM_BUNDLE,
} from '@/api-client/rbac'
import type { EndUserPermission, EndUserPermissionBundle } from '@/types'

// ── Props ────────────────────────────────────────────────────────────────────

interface UseBundleEditProps {
  orgName: string
  projectSlug: string
  bundleId: string
}

// ── Return ───────────────────────────────────────────────────────────────────

interface UpdateBundleInput {
  name: string
  description?: string
}

interface MutationResult {
  success: boolean
  errorMessage?: string
}

interface UseBundleEditReturn {
  bundle: EndUserPermissionBundle | null
  allPermissions: EndUserPermission[]
  loading: boolean
  error: Error | undefined
  updateBundle: (input: UpdateBundleInput) => Promise<MutationResult>
  addPermission: (permissionId: string) => Promise<MutationResult>
  removePermission: (permissionId: string) => Promise<MutationResult>
}

// ── Hook ─────────────────────────────────────────────────────────────────────

export function useBundleEdit({
  orgName,
  projectSlug,
  bundleId,
}: UseBundleEditProps): UseBundleEditReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  // ── Queries ────────────────────────────────────────────────────────────────

  const {
    data: bundleData,
    loading: bundleLoading,
    error: bundleError,
  } = useQuery(GET_END_USER_BUNDLE, {
    client,
    variables: { id: bundleId },
    skip: !projectSlug || !bundleId || !orgName,
  })

  const {
    data: permissionsData,
    loading: permissionsLoading,
    error: permissionsError,
  } = useQuery(GET_END_USER_PERMISSIONS, {
    client,
    skip: !projectSlug || !orgName,
  })

  // ── Mutations ──────────────────────────────────────────────────────────────

  const [updateBundleMutation] = useMutation(UPDATE_END_USER_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE],
  })

  const [addPermissionMutation] = useMutation(ADD_PERMISSION_TO_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE],
  })

  const [removePermissionMutation] = useMutation(REMOVE_PERMISSION_FROM_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLE],
  })

  // ── Callbacks ──────────────────────────────────────────────────────────────

  const updateBundle = useCallback(
    async (input: UpdateBundleInput): Promise<MutationResult> => {
      const result = await updateBundleMutation({
        variables: {
          id: bundleId,
          input: {
            name: input.name,
            description: input.description ?? '',
          },
        },
      })
      const payload = result.data?.updateEndUserPermissionBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '更新权限包失败' }
      }
      return { success: true }
    },
    [updateBundleMutation, bundleId]
  )

  const addPermission = useCallback(
    async (permissionId: string): Promise<MutationResult> => {
      const result = await addPermissionMutation({
        variables: {
          input: {
            bundleId,
            permissionId,
          },
        },
      })
      const payload = result.data?.addEndUserPermissionToBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '添加权限点失败' }
      }
      return { success: true }
    },
    [addPermissionMutation, bundleId]
  )

  const removePermission = useCallback(
    async (permissionId: string): Promise<MutationResult> => {
      const result = await removePermissionMutation({
        variables: {
          input: {
            bundleId,
            permissionId,
          },
        },
      })
      const payload = result.data?.removeEndUserPermissionFromBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '移除权限点失败' }
      }
      return { success: true }
    },
    [removePermissionMutation, bundleId]
  )

  // ── Return ─────────────────────────────────────────────────────────────────

  const bundle: EndUserPermissionBundle | null = bundleData?.endUserPermissionBundle ?? null
  const allPermissions: EndUserPermission[] = permissionsData?.endUserPermissions?.edges?.map((edge: any) => edge.node) ?? []
  const loading = bundleLoading || permissionsLoading
  const error = bundleError ?? permissionsError

  return {
    bundle,
    allPermissions,
    loading,
    error,
    updateBundle,
    addPermission,
    removePermission,
  }
}
