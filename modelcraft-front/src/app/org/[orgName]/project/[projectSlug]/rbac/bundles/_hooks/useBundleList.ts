import { useQuery, useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@bff/apollo/public'
import {
  GET_END_USER_BUNDLES,
  CREATE_END_USER_BUNDLE,
  DELETE_END_USER_BUNDLE,
} from '@web/graphql'
import type { EndUserPermissionBundle } from '@/types'

interface UseBundleListProps {
  orgName: string
  projectSlug: string
}

interface CreateBundleInput {
  name: string
  description?: string
}

interface UseBundleListReturn {
  bundles: EndUserPermissionBundle[]
  loading: boolean
  error: Error | undefined
  createBundle: (input: CreateBundleInput) => Promise<{ success: boolean; errorMessage?: string }>
  deleteBundle: (id: string) => Promise<{ success: boolean; errorMessage?: string }>
}

export function useBundleList({ orgName, projectSlug }: UseBundleListProps): UseBundleListReturn {
  const client = useProjectScopedClient(projectSlug, orgName)

  const { data, loading, error } = useQuery(GET_END_USER_BUNDLES, {
    client,
    variables: { projectSlug },
    skip: !projectSlug || !orgName,
  })

  const [createBundleMutation] = useMutation(CREATE_END_USER_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLES],
  })

  const [deleteBundleMutation] = useMutation(DELETE_END_USER_BUNDLE, {
    client,
    refetchQueries: [GET_END_USER_BUNDLES],
  })

  const bundles: EndUserPermissionBundle[] = data?.endUserBundles ?? []

  const createBundle = useCallback(
    async (input: CreateBundleInput) => {
      const result = await createBundleMutation({
        variables: {
          projectSlug,
          input: {
            name: input.name,
            description: input.description ?? '',
          },
        },
      })

      const payload = result.data?.createEndUserBundle
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '创建权限包失败',
        }
      }
      return { success: true }
    },
    [createBundleMutation, projectSlug]
  )

  const deleteBundle = useCallback(
    async (id: string) => {
      const result = await deleteBundleMutation({
        variables: { projectSlug, id },
      })

      const payload = result.data?.deleteEndUserBundle
      if (payload?.error) {
        return {
          success: false,
          errorMessage: payload.error.message ?? '删除权限包失败',
        }
      }
      return { success: true }
    },
    [deleteBundleMutation, projectSlug]
  )

  return {
    bundles,
    loading,
    error,
    createBundle,
    deleteBundle,
  }
}
