/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
import { useQuery, useMutation } from '@apollo/client'
import { useCallback, useMemo } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/public'
import {
  GET_END_USER_PERMISSIONS,
  DELETE_END_USER_PERMISSION,
  CREATE_END_USER_PERMISSION,
} from '@/api-client/rbac'
import { GET_MODELS, GET_MODEL_GROUPS } from '@/api-client/model'
import type { EndUserPermission, Model, ModelGroup } from '@/types'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ModelWithPermissions {
  model: Model
  permissions: EndUserPermission[]
  /** True if the model has an END_USER_REF field (owner), meaning RLS is applicable */
  hasOwnerField: boolean
}

export interface CreatePermissionInput {
  modelId: string
  action: import('@/types').EndUserPermissionAction
  rowScope: import('@/types').EndUserRowScope
  displayName?: string
  description?: string
  columnPolicy: import('@/types').ColumnPolicy
}

export interface UsePermissionsViewReturn {
  /** All unique database names derived from model groups */
  databaseNames: string[]
  /** Models for the currently selected database, each carrying its custom permissions */
  modelsForDb: ModelWithPermissions[]
  /** Whether the initial database list is loading */
  loadingDatabases: boolean
  /** Whether models for the selected database are loading */
  loadingModels: boolean
  /** Any loading/fetch error */
  error: Error | undefined
  deletePermission: (id: string) => Promise<{ success: boolean; errorMessage?: string }>
  createPermission: (input: CreatePermissionInput) => Promise<{ success: boolean; errorMessage?: string }>
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function usePermissionsView({
  orgName,
  projectSlug,
  selectedDatabaseName,
}: {
  orgName: string
  projectSlug: string
  /** The currently selected database name; when set, models for that DB are fetched */
  selectedDatabaseName: string
}): UsePermissionsViewReturn {
  const client = useProjectScopedClient(projectSlug, orgName)
  const skip = !projectSlug || !orgName

  // Phase 1: Fetch all model groups to derive available database names.
  // GET_MODEL_GROUPS returns models with databaseName but without fields —
  // this is intentionally lightweight for building the database selector.
  const {
    data: groupsData,
    loading: loadingDatabases,
    error: groupsError,
  } = useQuery(GET_MODEL_GROUPS, { client, skip })

  // Phase 2: Fetch full model details (including fields) for the selected database.
  const {
    data: modelsData,
    loading: loadingModels,
    error: modelsError,
  } = useQuery(GET_MODELS, {
    client,
    skip: skip || !selectedDatabaseName,
    variables: { input: { databaseName: selectedDatabaseName } },
  })

  // Fetch all custom permissions (project-wide, not per-database)
  const {
    data: permissionsData,
    error: permissionsError,
  } = useQuery(GET_END_USER_PERMISSIONS, { client, skip })

  const [deletePermissionMutation] = useMutation(DELETE_END_USER_PERMISSION, {
    client,
    refetchQueries: [GET_END_USER_PERMISSIONS],
  })

  const [createPermissionMutation] = useMutation(CREATE_END_USER_PERMISSION, {
    client,
    refetchQueries: [GET_END_USER_PERMISSIONS],
  })

  // ── Derived: databaseNames from model groups ───────────────────────────────

  const databaseNames = useMemo<string[]>(() => {
    const groups: ModelGroup[] = groupsData?.modelGroups ?? []
    const nameSet = new Set<string>()
    for (const group of groups) {
      for (const model of group.models ?? []) {
        if (model.databaseName) nameSet.add(model.databaseName)
      }
    }
    return Array.from(nameSet).sort((a, b) => a.localeCompare(b))
  }, [groupsData])

  // ── Derived: models for selected DB with permissions ──────────────────────

  const allPermissions: EndUserPermission[] = useMemo(
    () =>
      permissionsData?.endUserPermissions?.edges?.map(
        (e: { node: EndUserPermission }) => e.node,
      ) ?? [],
    [permissionsData],
  )

  const permissionsByModelId = useMemo<Record<string, EndUserPermission[]>>(() => {
    return allPermissions.reduce<Record<string, EndUserPermission[]>>((acc, perm) => {
      if (!acc[perm.modelId]) acc[perm.modelId] = []
      acc[perm.modelId].push(perm)
      return acc
    }, {})
  }, [allPermissions])

  const modelsForDb = useMemo<ModelWithPermissions[]>(() => {
    const models: Model[] =
      modelsData?.models?.edges?.map((e: { node: Model }) => e.node) ?? []

    return models
      .sort((a, b) => (a.title || a.name).localeCompare(b.title || b.name))
      .map((model) => ({
        model,
        permissions: permissionsByModelId[model.id] ?? [],
        hasOwnerField: model.fields?.some((f) => f.format === 'END_USER_REF') ?? false,
      }))
  }, [modelsData, permissionsByModelId])

  // ── Mutations ──────────────────────────────────────────────────────────────

  const deletePermission = useCallback(
    async (id: string) => {
      const result = await deletePermissionMutation({ variables: { id } })
      const payload = result.data?.deleteEndUserPermission
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '删除权限点失败' }
      }
      return { success: true }
    },
    [deletePermissionMutation],
  )

  const createPermission = useCallback(
    async (input: CreatePermissionInput) => {
      const result = await createPermissionMutation({
        variables: {
          projectSlug,
          input: {
            modelId: input.modelId,
            action: input.action,
            rowScope: input.rowScope,
            displayName: input.displayName || undefined,
            description: input.description || undefined,
            columnPolicy: {
              defaultMode: input.columnPolicy.defaultMode,
              rules: input.columnPolicy.rules.map((r) => ({
                fieldName: r.fieldName,
                mode: r.mode,
                maskPattern: r.maskPattern ?? null,
              })),
            },
          },
        },
      })
      const payload = result.data?.createEndUserPermission
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message ?? '创建权限点失败' }
      }
      return { success: true }
    },
    [createPermissionMutation, projectSlug],
  )

  return {
    databaseNames,
    modelsForDb,
    loadingDatabases,
    loadingModels,
    error: groupsError ?? modelsError ?? permissionsError,
    deletePermission,
    createPermission,
  }
}
