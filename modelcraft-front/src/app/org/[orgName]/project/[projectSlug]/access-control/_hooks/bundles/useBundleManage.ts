/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
import { useQuery, useMutation } from '@apollo/client'
import { useCallback, useMemo } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import {
  GET_END_USER_BUNDLE,
  GET_END_USER_PERMISSIONS,
  GET_END_USER_BUNDLES,
  GET_END_USER_ROLES,
  ADD_PERMISSION_TO_BUNDLE,
  REMOVE_DATA_PERMISSION_ITEM_FROM_BUNDLE,
  BIND_PRESET_ITEM_TO_BUNDLE,
  BIND_CUSTOM_ITEM_TO_BUNDLE,
  UPDATE_END_USER_BUNDLE,
  RESTORE_END_USER_BUNDLE,
} from '@/api-client/rbac'
import { DATABASE_CATALOG } from '@/api-client/cluster'
import type {
  EndUserPermission,
  EndUserPermissionBundle,
  EndUserRole,
  EndUserPermissionPreset,
} from '@/types'

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
  databaseNames: string[]
  associatedRoles: EndUserRole[]
  loading: boolean
  rolesLoading: boolean
  databasesLoading: boolean
  error: Error | undefined
  /** 旧接口：绑定 custom permission（兼容旧页面） */
  addPermission: (permissionId: string) => Promise<MutationResult>
  /** 新接口：按 modelId 删除 item */
  removeItemByModelId: (modelId: string) => Promise<MutationResult>
  /** 新接口：绑定 preset item */
  bindPresetItem: (modelId: string, preset: EndUserPermissionPreset, sortOrder?: number) => Promise<MutationResult>
  /** 新接口：绑定 custom item */
  bindCustomItem: (modelId: string, customPermissionId: string, sortOrder?: number) => Promise<MutationResult>
  updateBundle: (name: string, description?: string) => Promise<MutationResult>
  restoreBundle: (targetVersion: number) => Promise<MutationResult>
}

export function useBundleManage({
  orgName,
  projectSlug,
  bundleId,
}: UseBundleManageProps): UseBundleManageReturn {
  const client = useProjectScopedClient(projectSlug)
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

  const {
    data: rolesData,
    loading: rolesLoading,
  } = useQuery(GET_END_USER_ROLES, {
    client,
    skip: !orgName || !projectSlug,
  })

  const {
    data: databaseCatalogData,
    loading: databaseCatalogLoading,
  } = useQuery(DATABASE_CATALOG, {
    client,
    variables: {
      input: {
        page: 1,
        pageSize: 100,
      },
    },
    skip: !orgName || !projectSlug,
  })

  const databaseNames = useMemo(
    () =>
      (databaseCatalogData?.modelDatabaseCatalog?.data?.databases ?? [])
        .map((item: any) => item?.name)
        .filter((name: string | undefined): name is string => Boolean(name)),
    [databaseCatalogData],
  )

  const refetchQueries = [GET_END_USER_BUNDLE, GET_END_USER_BUNDLES]

  const [addPermissionMutation] = useMutation(ADD_PERMISSION_TO_BUNDLE, {
    client,
    refetchQueries,
  })

  const [removeItemMutation] = useMutation(REMOVE_DATA_PERMISSION_ITEM_FROM_BUNDLE, {
    client,
    refetchQueries,
  })

  const [bindPresetMutation] = useMutation(BIND_PRESET_ITEM_TO_BUNDLE, {
    client,
    refetchQueries,
  })

  const [bindCustomMutation] = useMutation(BIND_CUSTOM_ITEM_TO_BUNDLE, {
    client,
    refetchQueries,
  })

  const [updateBundleMutation] = useMutation(UPDATE_END_USER_BUNDLE, {
    client,
    refetchQueries,
  })

  const [restoreBundleMutation] = useMutation(RESTORE_END_USER_BUNDLE, {
    client,
    refetchQueries,
  })

  // 旧接口：绑定 custom permission（兼容旧 AddStrategyDialog）
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

  // 新接口：按 modelId 移除 item
  const removeItemByModelId = useCallback(
    async (modelId: string): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await removeItemMutation({
        variables: { input: { bundleId, modelId } },
      })
      const payload = result.data?.removeDataPermissionItemFromBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '移除权限配置失败' }
      }
      return { success: true }
    },
    [removeItemMutation, bundleId],
  )

  // 新接口：绑定 preset item
  const bindPresetItem = useCallback(
    async (
      modelId: string,
      preset: EndUserPermissionPreset,
      sortOrder = 0,
    ): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await bindPresetMutation({
        variables: { input: { bundleId, modelId, preset, sortOrder } },
      })
      const payload = result.data?.bindPresetItemToBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '绑定预设失败' }
      }
      return { success: true }
    },
    [bindPresetMutation, bundleId],
  )

  // 新接口：绑定 custom item
  const bindCustomItem = useCallback(
    async (
      modelId: string,
      customPermissionId: string,
      sortOrder = 0,
    ): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await bindCustomMutation({
        variables: { input: { bundleId, modelId, customPermissionId, sortOrder } },
      })
      const payload = result.data?.bindCustomItemToBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '绑定自定义权限失败' }
      }
      return { success: true }
    },
    [bindCustomMutation, bundleId],
  )

  const updateBundle = useCallback(
    async (name: string, description?: string): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await updateBundleMutation({
        variables: { id: bundleId, input: { name, description } },
      })
      const payload = result.data?.updateEndUserPermissionBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '更新权限包失败' }
      }
      return { success: true }
    },
    [updateBundleMutation, bundleId],
  )

  const restoreBundle = useCallback(
    async (targetVersion: number): Promise<MutationResult> => {
      if (!bundleId) return { success: false, errorMessage: '未选择权限包' }
      const result = await restoreBundleMutation({
        variables: { input: { bundleId, targetVersion } },
      })
      const payload = result.data?.restoreEndUserPermissionBundle
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '还原版本失败' }
      }
      return { success: true }
    },
    [restoreBundleMutation, bundleId],
  )

  const bundle: EndUserPermissionBundle | null = bundleData?.endUserPermissionBundle ?? null
  const allPermissions: EndUserPermission[] = permissionsData?.endUserPermissions?.edges?.map(
    (edge: any) => edge.node,
  ) ?? []

  const allRoles: EndUserRole[] = rolesData?.endUserRoles?.edges?.map(
    (edge: any) => edge.node,
  ) ?? []

  const associatedRoles = bundleId
    ? allRoles.filter((role) =>
        role.permissionBundles.some((pb: any) => pb.bundle?.id === bundleId),
      )
    : []

  return {
    bundle,
    allPermissions,
    databaseNames,
    associatedRoles,
    loading: bundleLoading || permissionsLoading,
    rolesLoading,
    databasesLoading: databaseCatalogLoading,
    error: bundleError ?? permissionsError,
    addPermission,
    removeItemByModelId,
    bindPresetItem,
    bindCustomItem,
    updateBundle,
    restoreBundle,
  }
}
