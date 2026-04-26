import { useQuery, useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  GET_END_USER_ROLES,
  CREATE_END_USER_ROLE,
  DELETE_END_USER_ROLE,
} from '@web/graphql'
import type { EndUserRole } from '@/types'

interface UseRoleListProps {
  orgName: string
  projectSlug: string
}

interface CreateRoleInput {
  name: string
  description?: string
}

interface UseRoleListReturn {
  roles: EndUserRole[]
  loading: boolean
  error: Error | undefined
  createRole: (input: CreateRoleInput) => Promise<{ success: boolean; errorMessage?: string }>
  deleteRole: (role: EndUserRole) => Promise<{ success: boolean; errorMessage?: string }>
}

export function useRoleList({ orgName, projectSlug }: UseRoleListProps): UseRoleListReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  const { data, loading, error } = useQuery(GET_END_USER_ROLES, {
    client,
    skip: !projectSlug || !orgName,
  })

  const [createRoleMutation] = useMutation(CREATE_END_USER_ROLE, {
    client,
    refetchQueries: [GET_END_USER_ROLES],
  })

  const [deleteRoleMutation] = useMutation(DELETE_END_USER_ROLE, {
    client,
    refetchQueries: [GET_END_USER_ROLES],
  })

  const roles: EndUserRole[] = data?.endUserRoles?.edges?.map((edge: any) => edge.node) ?? []

  const createRole = useCallback(
    async (input: CreateRoleInput) => {
      const result = await createRoleMutation({
        variables: {
          input: {
            name: input.name,
            description: input.description ?? '',
          },
        },
      })

      const payload = result.data?.createEndUserRole
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '创建角色失败',
        }
      }
      return { success: true }
    },
    [createRoleMutation]
  )

  const deleteRole = useCallback(
    async (role: EndUserRole) => {
      // 隐式角色不允许删除
      if (role.isImplicit) {
        return {
          success: false,
          errorMessage: '内置隐式角色不可删除',
        }
      }

      const result = await deleteRoleMutation({
        variables: { id: role.id },
      })

      const payload = result.data?.deleteEndUserRole
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '删除角色失败',
        }
      }
      return { success: true }
    },
    [deleteRoleMutation]
  )

  return {
    roles,
    loading,
    error,
    createRole,
    deleteRole,
  }
}
