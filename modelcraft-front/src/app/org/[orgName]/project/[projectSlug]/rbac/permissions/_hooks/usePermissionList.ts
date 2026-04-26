import { useQuery, useMutation } from '@apollo/client'
import { useCallback, useMemo } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  GET_END_USER_PERMISSIONS,
  DELETE_END_USER_PERMISSION,
} from '@/api-client/rbac'
import type { EndUserPermission } from '@/types'

interface UsePermissionListProps {
  orgName: string
  projectSlug: string
}

interface UsePermissionListReturn {
  allPermissions: EndUserPermission[]
  groupedPermissions: Record<string, EndUserPermission[]>
  loading: boolean
  error: Error | undefined
  deletePermission: (id: string) => Promise<{ success: boolean; errorMessage?: string }>
}

export function usePermissionList({
  orgName,
  projectSlug,
}: UsePermissionListProps): UsePermissionListReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  const { data, loading, error } = useQuery(GET_END_USER_PERMISSIONS, {
    client,
    skip: !projectSlug || !orgName,
  })

  const [deletePermissionMutation] = useMutation(DELETE_END_USER_PERMISSION, {
    client,
    refetchQueries: [GET_END_USER_PERMISSIONS],
  })

  const allPermissions: EndUserPermission[] = data?.endUserPermissions?.edges?.map((edge: any) => edge.node) ?? []

  // Group permissions by modelId
  const groupedPermissions = useMemo<Record<string, EndUserPermission[]>>(() => {
    return allPermissions.reduce<Record<string, EndUserPermission[]>>((acc, permission) => {
      const { modelId } = permission
      if (!acc[modelId]) {
        acc[modelId] = []
      }
      acc[modelId].push(permission)
      return acc
    }, {})
  }, [allPermissions])

  const deletePermission = useCallback(
    async (id: string) => {
      const result = await deletePermissionMutation({
        variables: { id },
      })

      const payload = result.data?.deleteEndUserPermission
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '删除权限点失败',
        }
      }
      return { success: true }
    },
    [deletePermissionMutation]
  )

  return {
    allPermissions,
    groupedPermissions,
    loading,
    error,
    deletePermission,
  }
}
