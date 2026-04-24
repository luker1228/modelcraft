import { useQuery, useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@bff/apollo/public'
import {
  GET_END_USER_ROLE,
  GET_END_USER_BUNDLES,
  ASSIGN_BUNDLE_TO_ROLE,
  REVOKE_BUNDLE_FROM_ROLE,
} from '@web/graphql'
import type { EndUserRole, EndUserPermissionBundle } from '@/types'

interface UseRoleEditProps {
  orgName: string
  projectSlug: string
  roleId: string
}

interface UseRoleEditReturn {
  role: EndUserRole | null
  allBundles: EndUserPermissionBundle[]
  loading: boolean
  error: Error | undefined
  assignBundle: (bundleId: string) => Promise<{ success: boolean; errorMessage?: string }>
  revokeBundle: (bundleId: string) => Promise<{ success: boolean; errorMessage?: string }>
}

export function useRoleEdit({ orgName, projectSlug, roleId }: UseRoleEditProps): UseRoleEditReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  const {
    data: roleData,
    loading: roleLoading,
    error: roleError,
  } = useQuery(GET_END_USER_ROLE, {
    client,
    variables: { projectSlug, id: roleId },
    skip: !projectSlug || !orgName || !roleId,
  })

  const {
    data: bundlesData,
    loading: bundlesLoading,
    error: bundlesError,
  } = useQuery(GET_END_USER_BUNDLES, {
    client,
    variables: { projectSlug },
    skip: !projectSlug || !orgName,
  })

  const [assignBundleMutation] = useMutation(ASSIGN_BUNDLE_TO_ROLE, {
    client,
    refetchQueries: [GET_END_USER_ROLE],
  })

  const [revokeBundleMutation] = useMutation(REVOKE_BUNDLE_FROM_ROLE, {
    client,
    refetchQueries: [GET_END_USER_ROLE],
  })

  const role: EndUserRole | null = roleData?.endUserRole ?? null
  const allBundles: EndUserPermissionBundle[] = bundlesData?.endUserBundles ?? []

  const assignBundle = useCallback(
    async (bundleId: string) => {
      const result = await assignBundleMutation({
        variables: { projectSlug, roleId, bundleId },
      })

      const payload = result.data?.assignBundleToRole
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '添加权限包失败',
        }
      }
      return { success: true }
    },
    [assignBundleMutation, projectSlug, roleId]
  )

  const revokeBundle = useCallback(
    async (bundleId: string) => {
      const result = await revokeBundleMutation({
        variables: { projectSlug, roleId, bundleId },
      })

      const payload = result.data?.revokeBundleFromRole
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '移除权限包失败',
        }
      }
      return { success: true }
    },
    [revokeBundleMutation, projectSlug, roleId]
  )

  return {
    role,
    allBundles,
    loading: roleLoading || bundlesLoading,
    error: roleError ?? bundlesError,
    assignBundle,
    revokeBundle,
  }
}
