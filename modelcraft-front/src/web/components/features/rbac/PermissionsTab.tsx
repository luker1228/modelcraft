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
} from 'lucide-react'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
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

import { usePermissionsView, type ModelWithPermissions } from '@/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView'
import type {
  EndUserPermission,
  EndUserPermissionAction,
  EndUserRowScope,
  ColumnPolicy,
} from '@/types'
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
  onConfirm: (action: EndUserPermissionAction, rowScope: EndUserRowScope) => Promise<void>
  onCancel: () => void
  saving: boolean
}

function InlineAddRow({ modelId: _modelId, onConfirm, onCancel, saving }: InlineAddRowProps) {
  const [action, setAction] = React.useState<EndUserPermissionAction>('SELECT')
  const [rowScope, setRowScope] = React.useState<EndUserRowScope>('ALL')

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

      {/* Name placeholder */}
      <TableCell className="py-2">
        <span className="text-xs text-muted-foreground/40">—</span>
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
            onClick={() => void onConfirm(action, rowScope)}
            disabled={saving}
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
        {permission.displayName && (
          <span className="text-xs text-muted-foreground">{permission.displayName}</span>
        )}
      </TableCell>

      <TableCell className="py-2.5 pr-4 text-right">
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-destructive"
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
  onCreate: (modelId: string, action: EndUserPermissionAction, rowScope: EndUserRowScope) => Promise<void>
  savingModelId: string | null
}

function ModelCard({
  item,
  projectSlug,
  deletingId,
  onDelete,
  onCreate,
  savingModelId,
}: ModelCardProps) {
  const [adding, setAdding] = React.useState(false)
  const { model, permissions, hasOwnerField } = item
  const saving = savingModelId === model.id

  const handleConfirmAdd = async (action: EndUserPermissionAction, rowScope: EndUserRowScope) => {
    await onCreate(model.id, action, rowScope)
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
  const [deletingId, setDeletingId] = React.useState<string | null>(null)
  const [savingModelId, setSavingModelId] = React.useState<string | null>(null)

  const {
    databaseNames,
    modelsForDb,
    loadingDatabases,
    loadingModels,
    error,
    deletePermission,
    createPermission,
  } = usePermissionsView({ orgName, projectSlug, selectedDatabaseName: selectedDb })

  // Auto-select first database when list loads
  React.useEffect(() => {
    if (!selectedDb && databaseNames.length > 0) {
      setSelectedDb(databaseNames[0]!)
    }
  }, [databaseNames, selectedDb])

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
    ) => {
      setSavingModelId(modelId)
      try {
        const result = await createPermission({
          modelId,
          action,
          rowScope,
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
      {/* Database selector */}
      <div className="flex items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">数据库</span>
        <Select value={selectedDb} onValueChange={setSelectedDb}>
          <SelectTrigger className="h-8 w-44 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {databaseNames.map((db) => (
              <SelectItem key={db} value={db} className="text-sm font-mono">
                {db}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {!loadingModels && modelsForDb.length > 0 && (
          <span className="text-xs text-muted-foreground">
            {modelsForDb.length} 个模型
          </span>
        )}
      </div>

      {/* Model cards */}
      {loadingModels ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="rounded-md border p-4 space-y-3">
              <div className="flex items-center gap-3">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-5 w-16" />
                <Skeleton className="ml-auto h-6 w-14" />
              </div>
            </div>
          ))}
        </div>
      ) : modelsForDb.length > 0 ? (
        <div className="space-y-3">
          {modelsForDb.map((item) => (
            <ModelCard
              key={item.model.id}
              item={item}
              projectSlug={projectSlug}
              deletingId={deletingId}
              onDelete={handleDelete}
              onCreate={handleCreate}
              savingModelId={savingModelId}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12">
          <p className="text-sm text-muted-foreground">该数据库下暂无模型</p>
        </div>
      )}
    </div>
  )
}
