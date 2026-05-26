'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

import * as React from 'react'
import { toast } from 'sonner'
import {
  Plus,
  Trash2,
  ShieldAlert,
  Lock,
  Check,
  X,
  Sparkles,
  ChevronDown,
} from 'lucide-react'

import { Button } from '@web/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@web/components/ui/alert-dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { Skeleton } from '@web/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'

import { usePermissionsView, type ModelWithPermissions } from '@/app/org/[orgName]/project/[projectSlug]/roles/_hooks/usePermissionsView'
import type {
  EndUserPermission,
  EndUserPermissionAction,
  EndUserRowScope,
  ColumnPolicy,
} from '@/types'
import type { EndUserPermissionPreset } from '@/generated/graphql'
import { cn } from '@/shared/utils'

// ── Props ─────────────────────────────────────────────────────────────────────

export interface PermissionsTabProps {
  orgName: string
  projectSlug: string
}

// ── Constants ─────────────────────────────────────────────────────────────────

const ACTION_LABEL: Record<EndUserPermissionAction, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '修改',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<EndUserRowScope, string> = {
  ALL: '全部行',
  SELF: '本人行',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '本部门及下级',
}

const PRESET_OPTIONS: { value: EndUserPermissionPreset; label: string; description: string; requiresOwner: boolean }[] = [
  { value: 'READ_WRITE_ALL', label: '读写全部', description: '可查看、新增、修改、删除全部数据', requiresOwner: false },
  { value: 'READ_ALL', label: '只读全部', description: '可查看全部数据，不可修改', requiresOwner: false },
  { value: 'READ_WRITE_OWNER', label: '读写本人', description: '只能查看和操作自己创建的数据', requiresOwner: true },
  { value: 'READ_ALL_WRITE_OWNER', label: '读全部/写本人', description: '可查看全部，但只能操作自己的数据', requiresOwner: true },
]

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatColumnPolicySummary(policy: ColumnPolicy | undefined | null): string {
  if (!policy) return '全部可见'
  const defaultLabel =
    policy.defaultMode === 'VISIBLE' ? '全部可见'
    : policy.defaultMode === 'HIDDEN' ? '全部隐藏'
    : policy.defaultMode === 'READONLY' ? '只读'
    : policy.defaultMode === 'MASKED' ? '脱敏'
    : policy.defaultMode
  const rulesCount = policy.rules?.length ?? 0
  if (rulesCount === 0) return defaultLabel
  return `${defaultLabel} +${rulesCount} 规则`
}

// ── Inline Add Permission Row ─────────────────────────────────────────────────

interface InlineAddRowProps {
  modelId: string
  onConfirm: (action: EndUserPermissionAction, rowScope: EndUserRowScope, displayName: string) => Promise<void>
  onCancel: () => void
  saving: boolean
}

function InlineAddRow({ modelId: _modelId, onConfirm, onCancel, saving }: InlineAddRowProps) {
  const [action, setAction] = React.useState<EndUserPermissionAction>('SELECT')
  const [rowScope, setRowScope] = React.useState<EndUserRowScope>('ALL')
  const [displayName, setDisplayName] = React.useState('')

  const canConfirm = displayName.trim().length > 0

  return (
    <TableRow className="bg-primary/[0.03]">
      {/* Action */}
      <TableCell className="py-2 pl-4">
        <Select
          value={action}
          onValueChange={(v) => setAction(v as EndUserPermissionAction)}
          disabled={saving}
        >
          <SelectTrigger className="h-7 w-24 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(Object.keys(ACTION_LABEL) as EndUserPermissionAction[]).map((a) => (
              <SelectItem key={a} value={a} className="text-xs">
                {ACTION_LABEL[a]}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </TableCell>

      {/* Row scope */}
      <TableCell className="py-2">
        <Select
          value={rowScope}
          onValueChange={(v) => setRowScope(v as EndUserRowScope)}
          disabled={saving}
        >
          <SelectTrigger className="h-7 w-28 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(Object.keys(ROW_SCOPE_LABEL) as EndUserRowScope[]).map((s) => (
              <SelectItem key={s} value={s} className="text-xs">
                {ROW_SCOPE_LABEL[s]}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </TableCell>

      {/* Column policy — default VISIBLE, configurable later */}
      <TableCell className="py-2">
        <span className="text-xs text-muted-foreground">全部可见</span>
      </TableCell>

      {/* Name — required */}
      <TableCell className="py-2">
        <input
          type="text"
          placeholder="名称（必填）"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          disabled={saving}
          className="h-7 w-full rounded border border-input bg-transparent px-2 text-xs placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50"
        />
      </TableCell>

      {/* Confirm / cancel */}
      <TableCell className="py-2 pr-4 text-right">
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="size-7 text-muted-foreground hover:text-destructive"
            onClick={onCancel}
            disabled={saving}
          >
            <X className="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-7 text-primary hover:bg-primary/10"
            onClick={() => void onConfirm(action, rowScope, displayName.trim())}
            disabled={saving || !canConfirm}
          >
            <Check className="size-3.5" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}

// ── Permission Row ────────────────────────────────────────────────────────────

interface PermissionRowProps {
  permission: EndUserPermission
  onDelete: (permission: EndUserPermission) => Promise<void>
  deletingId: string | null
}

function PermissionRow({ permission, onDelete, deletingId }: PermissionRowProps) {
  const isDeleting = deletingId === permission.id
  return (
    <TableRow className={cn('group hover:bg-muted/30', isDeleting && 'opacity-50')}>
      <TableCell className="py-2.5 pl-4">
        <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-foreground">
          {ACTION_LABEL[permission.action]}
        </span>
      </TableCell>

      <TableCell className="py-2.5">
        <span className="text-xs text-muted-foreground">
          {ROW_SCOPE_LABEL[permission.rowScope]}
        </span>
      </TableCell>

      <TableCell className="py-2.5">
        <span className="text-xs text-muted-foreground">
          {formatColumnPolicySummary(permission.columnPolicy as ColumnPolicy | undefined)}
        </span>
      </TableCell>

      <TableCell className="py-2.5">
        {permission.displayName ? (
          <span className="text-xs text-foreground">{permission.displayName}</span>
        ) : (
          <span className="text-xs text-muted-foreground/40">—</span>
        )}
      </TableCell>

      <TableCell className="py-2.5 pr-4 text-right">
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
              disabled={isDeleting}
            >
              <Trash2 className="size-3.5" />
              <span className="sr-only">删除</span>
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>确认删除权限点</AlertDialogTitle>
              <AlertDialogDescription>
                确定要删除「{ACTION_LABEL[permission.action]} · {ROW_SCOPE_LABEL[permission.rowScope]}」吗？
                已包含该权限点的权限包将自动失去此能力，此操作不可撤销。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                onClick={() => void onDelete(permission)}
              >
                确认删除
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </TableCell>
    </TableRow>
  )
}

// ── Model Card ────────────────────────────────────────────────────────────────

interface ModelCardProps {
  item: ModelWithPermissions
  projectSlug: string
  deletingId: string | null
  onDelete: (permission: EndUserPermission) => Promise<void>
  onCreate: (modelId: string, action: EndUserPermissionAction, rowScope: EndUserRowScope, displayName: string) => Promise<void>
  onApplyPreset: (modelId: string, preset: EndUserPermissionPreset) => Promise<void>
  savingModelId: string | null
  applyingPresetModelId: string | null
}

function ModelCard({
  item,
  projectSlug,
  deletingId,
  onDelete,
  onCreate,
  onApplyPreset,
  savingModelId,
  applyingPresetModelId,
}: ModelCardProps) {
  const [adding, setAdding] = React.useState(false)
  const { model, permissions, hasOwnerField } = item
  const saving = savingModelId === model.id
  const applyingPreset = applyingPresetModelId === model.id

  const handleConfirmAdd = async (action: EndUserPermissionAction, rowScope: EndUserRowScope, displayName: string) => {
    await onCreate(model.id, action, rowScope, displayName)
    setAdding(false)
  }

  return (
    <div className="rounded-md border border-border bg-card">
      {/* Card header */}
      <div className="flex items-center gap-3 border-b border-border px-4 py-3">
        <Lock className="size-3.5 shrink-0 text-muted-foreground/60" />
        <div className="flex min-w-0 flex-1 items-center gap-2">
          <span className="text-sm font-semibold text-foreground">
            {model.title || model.name}
          </span>
          <span className="font-mono text-[11px] text-muted-foreground">{model.name}</span>
        </div>

        {/* Apply preset policy dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 shrink-0 px-2 text-xs text-muted-foreground hover:text-foreground"
              disabled={applyingPreset || saving}
            >
              <Sparkles className="mr-1 size-3" />
              预设策略
              <ChevronDown className="ml-1 size-3 opacity-60" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
              应用后将替换该模型下现有预设权限点
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            {PRESET_OPTIONS.map((opt) => (
              <DropdownMenuItem
                key={opt.value}
                disabled={opt.requiresOwner && !hasOwnerField}
                onClick={() => void onApplyPreset(model.id, opt.value)}
                className="flex flex-col items-start gap-0.5 py-2"
              >
                <span className="text-xs font-medium">{opt.label}</span>
                <span className="text-[11px] text-muted-foreground">{opt.description}</span>
                {opt.requiresOwner && !hasOwnerField && (
                  <span className="text-[11px] text-destructive/70">需要 END_USER_REF 字段</span>
                )}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        <Button
          variant="ghost"
          size="sm"
          className="h-7 shrink-0 px-2 text-xs text-muted-foreground hover:text-foreground"
          onClick={() => setAdding(true)}
          disabled={adding || saving}
        >
          <Plus className="mr-1 size-3" />
          新增
        </Button>
      </div>

      {/* Permissions table */}
      {(permissions.length > 0 || adding) ? (
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="h-7 pl-4 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                动作
              </TableHead>
              <TableHead className="h-7 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                行范围
              </TableHead>
              <TableHead className="h-7 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                列策略
              </TableHead>
              <TableHead className="h-7 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                名称
              </TableHead>
              <TableHead className="h-7 w-10 pr-4" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {permissions.map((perm) => (
              <PermissionRow
                key={perm.id}
                permission={perm}
                onDelete={onDelete}
                deletingId={deletingId}
              />
            ))}
            {adding && (
              <InlineAddRow
                modelId={model.id}
                onConfirm={handleConfirmAdd}
                onCancel={() => setAdding(false)}
                saving={saving}
              />
            )}
          </TableBody>
        </Table>
      ) : (
        /* No permissions yet, not in adding mode */
        <div className="flex items-center gap-2 px-4 py-3">
          <span className="text-xs text-muted-foreground/50">暂无自定义权限点</span>
          <button
            type="button"
            className="text-xs text-primary hover:underline"
            onClick={() => setAdding(true)}
          >
            添加一个
          </button>
        </div>
      )}
    </div>
  )
}

// ── PermissionsTab ────────────────────────────────────────────────────────────

export function PermissionsTab({ orgName, projectSlug }: PermissionsTabProps) {
  const [selectedDb, setSelectedDb] = React.useState<string>('')
  const [selectedModelId, setSelectedModelId] = React.useState<string>('')
  const [deletingId, setDeletingId] = React.useState<string | null>(null)
  const [savingModelId, setSavingModelId] = React.useState<string | null>(null)
  const [applyingPresetModelId, setApplyingPresetModelId] = React.useState<string | null>(null)

  const {
    databaseNames,
    modelsForDb,
    loadingDatabases,
    loadingModels,
    error,
    deletePermission,
    createPermission,
    applyPresetPolicy,
  } = usePermissionsView({ orgName, projectSlug, selectedDatabaseName: selectedDb })

  // Auto-select first database when list loads
  React.useEffect(() => {
    if (!selectedDb && databaseNames.length > 0) {
      setSelectedDb(databaseNames[0]!)
    }
  }, [databaseNames, selectedDb])

  // Auto-select first model when db changes
  React.useEffect(() => {
    if (modelsForDb.length > 0) {
      const ids = modelsForDb.map((m) => m.model.id)
      if (!selectedModelId || !ids.includes(selectedModelId)) {
        setSelectedModelId(ids[0]!)
      }
    } else {
      setSelectedModelId('')
    }
  }, [modelsForDb, selectedModelId])

  const selectedModelItem = modelsForDb.find((m) => m.model.id === selectedModelId) ?? null

  const handleDelete = React.useCallback(
    async (permission: EndUserPermission) => {
      setDeletingId(permission.id)
      try {
        const result = await deletePermission(permission.id)
        if (result.success) {
          toast.success('权限点已删除')
        } else {
          toast.error(result.errorMessage ?? '删除失败，请重试')
        }
      } catch {
        toast.error('删除权限点时发生错误')
      } finally {
        setDeletingId(null)
      }
    },
    [deletePermission],
  )

  const handleCreate = React.useCallback(
    async (
      modelId: string,
      action: EndUserPermissionAction,
      rowScope: EndUserRowScope,
      displayName: string,
    ) => {
      setSavingModelId(modelId)
      try {
        const result = await createPermission({
          modelId,
          action,
          rowScope,
          displayName,
          columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
        })
        if (result.success) {
          toast.success('权限点已添加')
        } else {
          toast.error(result.errorMessage ?? '添加失败，请重试')
        }
      } catch {
        toast.error('添加权限点时发生错误')
      } finally {
        setSavingModelId(null)
      }
    },
    [createPermission],
  )

  const handleApplyPreset = React.useCallback(
    async (modelId: string, preset: EndUserPermissionPreset) => {
      setApplyingPresetModelId(modelId)
      try {
        const result = await applyPresetPolicy(modelId, preset)
        if (result.success) {
          const presetLabel = PRESET_OPTIONS.find((o) => o.value === preset)?.label ?? preset
          toast.success(`已应用预设策略：${presetLabel}`)
        } else {
          toast.error(result.errorMessage ?? '应用预设策略失败，请重试')
        }
      } catch {
        toast.error('应用预设策略时发生错误')
      } finally {
        setApplyingPresetModelId(null)
      }
    },
    [applyPresetPolicy],
  )

  if (loadingDatabases) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-8 w-44" />
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="rounded-md border p-4">
            <div className="flex items-center gap-3">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-5 w-16" />
              <Skeleton className="ml-auto h-6 w-14" />
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
        加载失败：{error.message}
      </div>
    )
  }

  if (databaseNames.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
        <ShieldAlert className="mb-3 size-10 text-muted-foreground/25" />
        <p className="text-sm font-medium text-foreground">暂无模型</p>
        <p className="mt-1 text-xs text-muted-foreground">请先在数据模型页面创建模型</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Selectors row */}
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">数据库</span>
        <Select value={selectedDb} onValueChange={(v) => { setSelectedDb(v); setSelectedModelId('') }}>
          <SelectTrigger className="h-8 w-44 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {databaseNames.map((db) => (
              <SelectItem key={db} value={db} className="font-mono text-sm">
                {db}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {!loadingModels && modelsForDb.length > 0 && (
          <>
            <span className="text-xs text-muted-foreground/40">/</span>
            <span className="text-xs font-medium text-muted-foreground">模型</span>
            <Select value={selectedModelId} onValueChange={setSelectedModelId}>
              <SelectTrigger className="h-8 w-48 text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {modelsForDb.map(({ model }) => (
                  <SelectItem key={model.id} value={model.id} className="text-sm">
                    <span className="font-medium">{model.title || model.name}</span>
                    <span className="ml-1.5 font-mono text-[11px] text-muted-foreground">{model.name}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </>
        )}
      </div>

      {/* Single model card */}
      {loadingModels ? (
        <div className="space-y-3 rounded-md border p-4">
          <div className="flex items-center gap-3">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-5 w-16" />
            <Skeleton className="ml-auto h-6 w-14" />
          </div>
        </div>
      ) : selectedModelItem ? (
        <ModelCard
          item={selectedModelItem}
          projectSlug={projectSlug}
          deletingId={deletingId}
          onDelete={handleDelete}
          onCreate={handleCreate}
          onApplyPreset={handleApplyPreset}
          savingModelId={savingModelId}
          applyingPresetModelId={applyingPresetModelId}
        />
      ) : modelsForDb.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12">
          <p className="text-sm text-muted-foreground">该数据库下暂无模型</p>
        </div>
      ) : null}
    </div>
  )
}
