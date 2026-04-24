'use client'

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { useParams, useRouter } from 'next/navigation'
import { ArrowLeft, ChevronDown, ChevronRight, KeyRound, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import { getOrgScopedClient, useProjectScopedClient } from '@bff/apollo/public'
import { Badge } from '@web/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@web/components/ui/select'
import { Textarea } from '@web/components/ui/textarea'
import { DATABASE_CATALOG } from '@web/graphql/queries/cluster'
import { GET_MODELS_FOR_RELATION } from '@web/graphql/queries/model'
import {
  GET_PERMISSION_ROLES,
  GET_ROLE_PERMISSIONS_LIST,
} from '@web/graphql/queries/user'
import {
  ADD_PERMISSION_TO_ROLE,
  REMOVE_PERMISSION_FROM_ROLE,
} from '@web/graphql/mutations/user'
import { SET_MODEL_RLS_POLICY } from '@web/graphql/mutations/rls'

type Action = 'read' | 'create' | 'update' | 'delete'
type RLSPreset = 'READ_WRITE_OWNER' | 'READ_ALL_WRITE_OWNER' | 'READ_ALL' | 'READ_WRITE_ALL' | 'NO_ACCESS' | 'CUSTOM'

type RolePermissionDef = {
  obj: string
  act: string
}

type PermissionRole = {
  id: number
  name: string
  isSystem: boolean
}

type ModelDatabaseCatalogData = {
  modelDatabaseCatalog: {
    data: {
      databases: Array<{ name: string }>
    } | null
  }
}

type ModelsForRelationData = {
  models: {
    edges: Array<{
      node: {
        id: string
        name: string
      }
    }>
  }
}

type PermissionRolesData = {
  permissionRoles: PermissionRole[]
}

type RolePermissionsListData = {
  rolePermissionsList: RolePermissionDef[]
}

type RolePermissionMutationResult = {
  success: boolean
  error: {
    __typename: string
    message: string
    suggestion?: string
  } | null
}

type AddPermissionMutationData = {
  addPermissionToRole: RolePermissionMutationResult
}

type RemovePermissionMutationData = {
  removePermissionFromRole: RolePermissionMutationResult
}

type SetRLSPolicyMutationData = {
  setModelRLSPolicy: {
    policy: {
      preset: RLSPreset | null
    } | null
    error: {
      __typename: string
      message: string
      suggestion?: string
      path?: string
      variable?: string
    } | null
  }
}

const PRESET_LABELS: Record<RLSPreset, string> = {
  READ_WRITE_OWNER: '仅所有者可读写',
  READ_ALL_WRITE_OWNER: '全员可读，仅所有者可写',
  READ_ALL: '全员只读',
  READ_WRITE_ALL: '全员可读写',
  NO_ACCESS: '禁止访问',
  CUSTOM: '自定义策略',
}

const PRESET_VALUES: RLSPreset[] = [
  'READ_WRITE_OWNER',
  'READ_ALL_WRITE_OWNER',
  'READ_ALL',
  'READ_WRITE_ALL',
  'NO_ACCESS',
  'CUSTOM',
]

const DEFAULT_CUSTOM_EXPRESSIONS: Record<Action, string> = {
  read: '{"ownerId":{"_eq":{"_auth":"uid"}}}',
  create: '{"ownerId":{"_eq":{"_auth":"uid"}}}',
  update: '{"ownerId":{"_eq":{"_auth":"uid"}}}',
  delete: '{"ownerId":{"_eq":{"_auth":"uid"}}}',
}

function normalizeRoleKey(raw: string): string {
  return raw.toLowerCase().replace(/^r-/, '').replace(/^role-/, '')
}

function buildPermissionObj(projectSlug: string, database: string, modelName: string): string {
  return `model:${projectSlug}:${database}:${modelName}`
}

function parseModelFromObj(obj: string, projectSlug: string, database: string): string | null {
  const prefix = `model:${projectSlug}:${database}:`
  if (!obj.startsWith(prefix)) return null
  const modelName = obj.slice(prefix.length).trim()
  return modelName === '' ? null : modelName
}

export default function RolePermissionsPage() {
  const params = useParams()
  const router = useRouter()

  const orgName = String(params?.orgName ?? '')
  const projectSlug = String(params?.projectSlug ?? '')
  const roleIdParam = String(params?.roleId ?? '')

  const orgClient = getOrgScopedClient()
  const projectClient = useProjectScopedClient(projectSlug, orgName)

  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [selectedDatabase, setSelectedDatabase] = useState('')
  const [addOpen, setAddOpen] = useState(false)
  const [addDb, setAddDb] = useState('')
  const [addModel, setAddModel] = useState('')
  const [customEditorOpen, setCustomEditorOpen] = useState(false)
  const [customTarget, setCustomTarget] = useState<{ database: string; modelName: string; modelId: string | null } | null>(null)
  const [customDraft, setCustomDraft] = useState<Record<Action, string>>(DEFAULT_CUSTOM_EXPRESSIONS)

  const { data: rolesData, loading: roleLoading, refetch: refetchRoles } = useQuery<PermissionRolesData>(
    GET_PERMISSION_ROLES,
    {
      client: orgClient,
      variables: { orgName, includeSystem: true },
      skip: !orgName,
    }
  )

  const resolvedRole = useMemo(() => {
    const roles = rolesData?.permissionRoles ?? []
    if (roles.length === 0) return null

    if (/^\d+$/.test(roleIdParam)) {
      const id = Number(roleIdParam)
      return roles.find((r) => r.id === id) ?? null
    }

    const key = normalizeRoleKey(roleIdParam)
    return roles.find((r) => normalizeRoleKey(r.name) === key) ?? null
  }, [roleIdParam, rolesData])

  const roleId = resolvedRole?.id ?? null
  const canManage = Boolean(roleId)

  const { data: rolePermsData, loading: permsLoading, refetch: refetchPermissions } = useQuery<RolePermissionsListData>(
    GET_ROLE_PERMISSIONS_LIST,
    {
      client: orgClient,
      variables: { roleId: roleId ?? -1 },
      skip: !roleId,
    }
  )

  const { data: databaseData, loading: databasesLoading } = useQuery<ModelDatabaseCatalogData>(
    DATABASE_CATALOG,
    {
      client: projectClient,
      variables: {
        input: {
          page: 1,
          pageSize: 100,
        },
      },
      skip: !orgName || !projectSlug,
    }
  )

  const databaseOptions = useMemo(() => {
    return (databaseData?.modelDatabaseCatalog?.data?.databases ?? [])
      .map((item) => item.name)
      .filter((name) => name.trim() !== '')
  }, [databaseData])

  useEffect(() => {
    if (databaseOptions.length === 0) return
    if (databaseOptions.includes(selectedDatabase)) return
    setSelectedDatabase(databaseOptions[0])
  }, [databaseOptions, selectedDatabase])

  const { data: modelsData, loading: modelsLoading } = useQuery<ModelsForRelationData>(
    GET_MODELS_FOR_RELATION,
    {
      client: projectClient,
      variables: {
        input: {
          databaseName: selectedDatabase,
          limit: 200,
        },
      },
      skip: !selectedDatabase,
    }
  )

  const modelCatalog = useMemo(() => {
    const list = modelsData?.models?.edges?.map((edge) => edge.node) ?? []
    return list
  }, [modelsData])

  const modelIdByName = useMemo(() => {
    const map = new Map<string, string>()
    for (const model of modelCatalog) {
      map.set(model.name, model.id)
    }
    return map
  }, [modelCatalog])

  const rolePermissions = useMemo(
    () => rolePermsData?.rolePermissionsList ?? [],
    [rolePermsData]
  )

  const boundRows = useMemo(() => {
    const rows: Array<{ database: string; modelName: string; preset: RLSPreset; modelId: string | null }> = []
    for (const perm of rolePermissions) {
      const modelName = parseModelFromObj(perm.obj, projectSlug, selectedDatabase)
      if (!modelName) continue

      const normalizedAct = perm.act as RLSPreset
      const preset = PRESET_VALUES.includes(normalizedAct) ? normalizedAct : 'CUSTOM'
      rows.push({
        database: selectedDatabase,
        modelName,
        preset,
        modelId: modelIdByName.get(modelName) ?? null,
      })
    }
    rows.sort((a, b) => a.modelName.localeCompare(b.modelName))
    return rows
  }, [modelIdByName, projectSlug, rolePermissions, selectedDatabase])

  const addModelOptions = useMemo(() => {
    const existing = new Set(boundRows.map((row) => row.modelName))
    return modelCatalog
      .map((m) => m.name)
      .filter((name) => !existing.has(name))
  }, [boundRows, modelCatalog])

  const groupedPerms = useMemo(() => {
    if (!selectedDatabase) return []
    return [[selectedDatabase, boundRows]] as const
  }, [boundRows, selectedDatabase])

  const [addPermissionToRole, { loading: addingPermission }] = useMutation<AddPermissionMutationData>(
    ADD_PERMISSION_TO_ROLE,
    { client: orgClient }
  )
  const [removePermissionFromRole, { loading: removingPermission }] = useMutation<RemovePermissionMutationData>(
    REMOVE_PERMISSION_FROM_ROLE,
    { client: orgClient }
  )
  const [setModelRLSPolicy, { loading: savingRLS }] = useMutation<SetRLSPolicyMutationData>(
    SET_MODEL_RLS_POLICY,
    { client: projectClient }
  )

  const refreshPageData = async () => {
    await Promise.all([refetchRoles(), refetchPermissions()])
  }

  const toggleCollapse = (db: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(db)) next.delete(db)
      else next.add(db)
      return next
    })
  }

  const openAddPermission = () => {
    const initialDb = selectedDatabase || databaseOptions[0] || ''
    setAddDb(initialDb)
    setAddModel('')
    setAddOpen(true)
  }

  const getCurrentPermission = (database: string, modelName: string) => {
    const obj = buildPermissionObj(projectSlug, database, modelName)
    return rolePermissions.find((item) => item.obj === obj) ?? null
  }

  const persistPreset = async (database: string, modelName: string, preset: RLSPreset) => {
    if (!roleId) return false

    const obj = buildPermissionObj(projectSlug, database, modelName)
    const current = getCurrentPermission(database, modelName)

    if (current?.act === preset) return true

    if (current) {
      const removeRes = await removePermissionFromRole({
        variables: {
          roleId,
          obj,
          act: current.act,
        },
      })
      const removePayload = removeRes.data?.removePermissionFromRole
      if (!removePayload?.success) {
        toast.error(removePayload?.error?.message || '移除旧策略失败')
        return false
      }
    }

    const addRes = await addPermissionToRole({
      variables: {
        roleId,
        obj,
        act: preset,
      },
    })
    const addPayload = addRes.data?.addPermissionToRole
    if (!addPayload?.success) {
      toast.error(addPayload?.error?.message || '设置策略失败')
      return false
    }

    return true
  }

  const handleAddPermission = async () => {
    if (!roleId) {
      toast.error('角色不存在，无法新增权限')
      return
    }
    if (!addDb) {
      toast.error('请先选择数据库')
      return
    }
    if (!addModel) {
      toast.error('请选择模型')
      return
    }

    const ok = await persistPreset(addDb, addModel, 'NO_ACCESS')
    if (!ok) return

    await refreshPageData()
    setAddOpen(false)
    setAddModel('')
    toast.success('权限已新增')
  }

  const applyPreset = async (database: string, modelName: string, preset: RLSPreset, modelId: string | null) => {
    if (!canManage) return

    if (preset === 'CUSTOM') {
      setCustomTarget({ database, modelName, modelId })
      setCustomDraft(DEFAULT_CUSTOM_EXPRESSIONS)
      setCustomEditorOpen(true)
      return
    }

    const ok = await persistPreset(database, modelName, preset)
    if (!ok) return
    await refreshPageData()
    toast.success(`已应用预设策略：${PRESET_LABELS[preset]}`)
  }

  const handleDeletePermission = async (database: string, modelName: string) => {
    if (!roleId) return
    const current = getCurrentPermission(database, modelName)
    if (!current) return

    const res = await removePermissionFromRole({
      variables: {
        roleId,
        obj: current.obj,
        act: current.act,
      },
    })
    const payload = res.data?.removePermissionFromRole
    if (!payload?.success) {
      toast.error(payload?.error?.message || '删除权限失败')
      return
    }
    await refreshPageData()
    toast.success(`已删除 ${modelName} 权限`)
  }

  const handleSaveCustomStrategy = async () => {
    if (!customTarget || !roleId) return

    const persisted = await persistPreset(customTarget.database, customTarget.modelName, 'CUSTOM')
    if (!persisted) return

    if (!customTarget.modelId) {
      await refreshPageData()
      toast.success('已保存自定义策略标记')
      setCustomEditorOpen(false)
      return
    }

    const readExpr = customDraft.read.trim() || 'false'
    const createExpr = customDraft.create.trim() || 'false'
    const updateExpr = customDraft.update.trim() || 'false'
    const deleteExpr = customDraft.delete.trim() || 'false'

    const rlsRes = await setModelRLSPolicy({
      variables: {
        input: {
          modelId: customTarget.modelId,
          selectPredicate: readExpr,
          insertCheck: createExpr,
          updatePredicate: updateExpr,
          updateCheck: updateExpr,
          deletePredicate: deleteExpr,
        },
      },
    })

    const rlsPayload = rlsRes.data?.setModelRLSPolicy
    if (rlsPayload?.error) {
      toast.error(rlsPayload.error.message)
      return
    }

    await refreshPageData()
    setCustomEditorOpen(false)
    toast.success('已保存自定义策略')
  }

  const isBusy = addingPermission || removingPermission || savingRLS

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <button
          onClick={() => router.push(`/org/${orgName}/project/${projectSlug}/roles`)}
          className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 transition-colors hover:bg-slate-100 hover:text-foreground"
        >
          <ArrowLeft className="size-4" strokeWidth={1.5} />
          角色管理
        </button>
        <span>/</span>
        <span className="text-foreground">{resolvedRole?.name || roleIdParam}</span>
        <span>/</span>
        <span className="text-foreground">权限管理</span>
      </div>

      <section className="rounded-lg border border-border bg-background p-6 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="rounded-md bg-violet-100 p-2 text-violet-600">
            <KeyRound className="size-5" strokeWidth={1.5} />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-semibold text-foreground">{resolvedRole?.name || roleIdParam}</h1>
              {resolvedRole?.isSystem && <Badge variant="secondary">系统角色</Badge>}
            </div>
            <p className="text-sm text-muted-foreground">
              使用真实接口配置模型级策略，非 mock 数据
            </p>
          </div>
        </div>
      </section>

      <section className="rounded-lg border border-border bg-background p-4 shadow-sm">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">数据库</span>
            <Select value={selectedDatabase} onValueChange={setSelectedDatabase}>
              <SelectTrigger className="w-[220px]">
                <SelectValue placeholder={databasesLoading ? '加载中...' : '选择 database'} />
              </SelectTrigger>
              <SelectContent>
                {databaseOptions.map((db) => (
                  <SelectItem key={db} value={db}>
                    {db}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <button
            onClick={openAddPermission}
            disabled={!canManage || isBusy}
            className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-foreground transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
          >
            <Plus className="size-4" strokeWidth={1.5} />
            新增权限
          </button>
        </div>
      </section>

      {roleLoading || permsLoading || modelsLoading ? (
        <div className="rounded-lg border border-slate-200 bg-white px-4 py-6 text-sm text-muted-foreground">
          加载中...
        </div>
      ) : (
        <div className="space-y-3">
          {groupedPerms.map(([db, rows]) => {
            const isCollapsed = collapsed.has(db)
            return (
              <div key={db} className="overflow-hidden rounded-lg border border-slate-200 bg-white">
                <div
                  className="flex cursor-pointer select-none items-center gap-3 border-b border-slate-200 bg-slate-50 px-4 py-3"
                  onClick={() => toggleCollapse(db)}
                >
                  {isCollapsed
                    ? <ChevronRight className="size-4 text-muted-foreground" strokeWidth={2} />
                    : <ChevronDown className="size-4 text-muted-foreground" strokeWidth={2} />}
                  <code className="font-mono text-sm font-semibold text-foreground">{db}</code>
                  <span className="text-xs text-muted-foreground">{rows.length} 个模型</span>
                </div>

                {!isCollapsed && (
                  <table className="w-full border-collapse text-sm">
                    <thead className="border-b border-slate-100 bg-muted/20">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">模型</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">生效策略</th>
                        <th className="px-4 py-2 text-right text-xs font-medium text-muted-foreground">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((row, idx) => (
                        <tr
                          key={row.modelName}
                          className={`border-b border-slate-100 transition-colors last:border-0 hover:bg-slate-50 ${idx % 2 !== 0 ? 'bg-muted/10' : ''}`}
                        >
                          <td className="px-4 py-3">
                            <span className="font-medium text-foreground">{row.modelName}</span>
                          </td>
                          <td className="px-4 py-3">
                            <Badge variant="outline" className="text-xs">
                              {PRESET_LABELS[row.preset]}
                            </Badge>
                          </td>
                          <td className="w-56 p-3 text-right">
                            <div className="flex items-center justify-end gap-2">
                              <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                  <button
                                    type="button"
                                    disabled={!canManage || isBusy}
                                    className="inline-flex h-7 items-center rounded-md border border-slate-300 px-2 text-xs font-medium text-foreground hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
                                  >
                                    编辑策略
                                  </button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end" className="w-52">
                                  {PRESET_VALUES.map((presetKey) => (
                                    <DropdownMenuItem
                                      key={presetKey}
                                      className="flex items-center justify-between text-xs"
                                      onClick={() => applyPreset(db, row.modelName, presetKey, row.modelId)}
                                    >
                                      <span>{PRESET_LABELS[presetKey]}</span>
                                      {row.preset === presetKey && (
                                        <span className="rounded bg-blue-50 px-1.5 py-0.5 text-[10px] text-blue-700">
                                          当前
                                        </span>
                                      )}
                                    </DropdownMenuItem>
                                  ))}
                                </DropdownMenuContent>
                              </DropdownMenu>
                              <button
                                type="button"
                                onClick={() => handleDeletePermission(db, row.modelName)}
                                disabled={!canManage || isBusy}
                                className="inline-flex size-7 items-center justify-center rounded-md border border-red-200 text-red-600 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-40"
                                aria-label={`删除 ${row.modelName} 权限`}
                              >
                                <Trash2 className="size-3.5" strokeWidth={1.6} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                      {rows.length === 0 && (
                        <tr>
                          <td colSpan={3} className="px-4 py-6 text-center text-sm text-muted-foreground">
                            暂无模型权限，点击右上角“新增权限”
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                )}
              </div>
            )
          })}
        </div>
      )}

      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新增权限</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">第一步：选择 database</div>
              <Select
                value={addDb}
                onValueChange={(value) => {
                  setAddDb(value)
                  setAddModel('')
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择 database" />
                </SelectTrigger>
                <SelectContent>
                  {databaseOptions.map((db) => (
                    <SelectItem key={db} value={db}>
                      {db}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <div className="text-sm text-muted-foreground">第二步：选择模型</div>
              <Select value={addModel} onValueChange={setAddModel} disabled={!addDb || addModelOptions.length === 0}>
                <SelectTrigger>
                  <SelectValue placeholder={addDb ? '选择模型' : '请先选择 database'} />
                </SelectTrigger>
                <SelectContent>
                  {addModelOptions.map((model) => (
                    <SelectItem key={model} value={model}>
                      {model}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {addDb && addModelOptions.length === 0 && (
                <p className="text-xs text-muted-foreground">该 database 下所有模型都已添加权限</p>
              )}
            </div>
          </div>

          <DialogFooter>
            <button
              type="button"
              onClick={() => setAddOpen(false)}
              className="inline-flex h-9 items-center rounded-md border border-slate-300 px-3 text-sm text-foreground hover:bg-slate-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleAddPermission}
              disabled={!canManage || !addDb || !addModel || isBusy}
              className="inline-flex h-9 items-center rounded-md bg-blue-600 px-4 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
            >
              新增权限
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={customEditorOpen} onOpenChange={setCustomEditorOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              自定义策略
              {customTarget ? ` · ${customTarget.modelName}` : ''}
            </DialogTitle>
          </DialogHeader>

          <div className="grid grid-cols-1 gap-3 py-2 md:grid-cols-2">
            {([
              { key: 'read', label: '查看' },
              { key: 'create', label: '新增' },
              { key: 'update', label: '编辑' },
              { key: 'delete', label: '删除' },
            ] as Array<{ key: Action; label: string }>).map((item) => (
              <div key={item.key} className="space-y-1.5 rounded-md border border-border p-3">
                <div className="text-sm font-medium text-foreground">{item.label}</div>
                <div className="text-xs text-muted-foreground">填写表达式，`false` 表示禁止该动作</div>
                <Textarea
                  value={customDraft[item.key]}
                  onChange={(e) => setCustomDraft((prev) => ({ ...prev, [item.key]: e.target.value }))}
                  rows={4}
                  placeholder='例如: {"ownerId":{"_eq":{"_auth":"uid"}}}'
                  className="font-mono text-xs"
                />
              </div>
            ))}
          </div>

          <DialogFooter>
            <button
              type="button"
              onClick={() => setCustomEditorOpen(false)}
              className="inline-flex h-9 items-center rounded-md border border-slate-300 px-3 text-sm text-foreground hover:bg-slate-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSaveCustomStrategy}
              disabled={!canManage || isBusy}
              className="inline-flex h-9 items-center rounded-md bg-blue-600 px-4 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
            >
              保存自定义策略
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
