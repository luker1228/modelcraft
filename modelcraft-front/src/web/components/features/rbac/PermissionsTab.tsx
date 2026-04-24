import * as React from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Plus, ChevronDown, ChevronRight, Trash2, ShieldAlert } from 'lucide-react'

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
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@web/components/ui/collapsible'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'

import { usePermissionList } from '@/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionList'
import type {
  EndUserPermission,
  EndUserPermissionAction,
  EndUserRowScope,
  ColumnPolicy,
} from '@/types'

// ── Props ────────────────────────────────────────────────────────────────────

export interface PermissionsTabProps {
  orgName: string
  projectSlug: string
}

// ── Action badge config ───────────────────────────────────────────────────────

type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline'

const ACTION_BADGE_VARIANT: Record<EndUserPermissionAction, BadgeVariant> = {
  SELECT: 'secondary',
  EXPORT: 'secondary',
  INSERT: 'outline',
  UPDATE: 'outline',
  DELETE: 'outline',
}

// DELETE gets a subdued destructive tint — visible but not alarming in a list
const ACTION_BADGE_EXTRA_CLASS: Partial<Record<EndUserPermissionAction, string>> = {
  DELETE: 'text-destructive border-destructive/30',
}

const ACTION_LABEL: Record<EndUserPermissionAction, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '修改',
  DELETE: '删除',
  EXPORT: '导出',
}

// ── Row scope labels (plain text, no badge) ───────────────────────────────────

const ROW_SCOPE_LABEL: Record<EndUserRowScope, string> = {
  ALL: '全部行',
  SELF: '本人行',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '本部门及下级',
}

// ── Column policy summary ─────────────────────────────────────────────────────

function formatColumnPolicySummary(policy: ColumnPolicy | undefined | null): string {
  if (!policy) return '默认: 全部可见'
  const defaultLabel =
    policy.defaultMode === 'VISIBLE' ? '全部可见'
    : policy.defaultMode === 'HIDDEN' ? '全部隐藏'
    : policy.defaultMode === 'READONLY' ? '只读'
    : policy.defaultMode === 'MASKED' ? '脱敏'
    : policy.defaultMode
  const rulesCount = policy.rules?.length ?? 0
  if (rulesCount === 0) return `默认: ${defaultLabel}`
  return `默认: ${defaultLabel}  +${rulesCount} 条规则`
}

// ── Skeleton ──────────────────────────────────────────────────────────────────

function PermissionListSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="rounded-md border p-4 space-y-2">
          <Skeleton className="h-5 w-40" />
          <div className="space-y-2 pt-2">
            {Array.from({ length: 2 }).map((_, j) => (
              <div key={j} className="flex items-center gap-3">
                <Skeleton className="h-5 w-14" />
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-4 w-48" />
                <Skeleton className="ml-auto h-7 w-7" />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

// ── PermissionRow (TableRow) ──────────────────────────────────────────────────

interface PermissionRowProps {
  permission: EndUserPermission
  onDelete: (permission: EndUserPermission) => Promise<void>
  deletingId: string | null
}

function PermissionRow({ permission, onDelete, deletingId }: PermissionRowProps) {
  return (
    <TableRow className="hover:bg-muted/50">
      {/* Action */}
      <TableCell className="py-3 pl-4">
        <Badge
          variant={ACTION_BADGE_VARIANT[permission.action]}
          className={ACTION_BADGE_EXTRA_CLASS[permission.action]}
        >
          {ACTION_LABEL[permission.action]}
        </Badge>
      </TableCell>

      {/* Row scope — inline text, intentionally not a badge */}
      <TableCell className="py-3">
        <span className="text-xs text-muted-foreground">
          {ROW_SCOPE_LABEL[permission.rowScope]}
        </span>
      </TableCell>

      {/* Column policy summary */}
      <TableCell className="py-3">
        <span className="text-xs text-muted-foreground">
          {formatColumnPolicySummary(
            permission.columnPolicy as ColumnPolicy | undefined
          )}
        </span>
      </TableCell>

      {/* Display name (if any) */}
      <TableCell className="py-3">
        {permission.displayName && (
          <span className="text-sm text-foreground">{permission.displayName}</span>
        )}
      </TableCell>

      {/* Delete — icon-only ghost, hover red */}
      <TableCell className="py-3 pr-4 text-right">
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 text-muted-foreground hover:text-destructive"
              disabled={deletingId === permission.id}
            >
              <Trash2 className="size-3.5" />
              <span className="sr-only">删除</span>
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>确认删除权限点</AlertDialogTitle>
              <AlertDialogDescription>
                确定要删除该权限点（
                {ACTION_LABEL[permission.action]} · {ROW_SCOPE_LABEL[permission.rowScope]}
                ）吗？已包含该权限点的权限包将自动失去此能力，此操作不可撤销。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                onClick={() => onDelete(permission)}
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

// ── ModelGroup ────────────────────────────────────────────────────────────────

interface ModelGroupProps {
  modelId: string
  modelDisplayName: string
  permissions: EndUserPermission[]
  onDelete: (permission: EndUserPermission) => Promise<void>
  deletingId: string | null
}

function ModelGroup({
  modelId,
  modelDisplayName,
  permissions,
  onDelete,
  deletingId,
}: ModelGroupProps) {
  const [open, setOpen] = React.useState(true)

  return (
    <Collapsible open={open} onOpenChange={setOpen} className="rounded-md border">
      {/* Group trigger */}
      <CollapsibleTrigger asChild>
        <button
          type="button"
          className="flex w-full items-center gap-2 px-4 py-3 text-left hover:bg-muted/50 focus-visible:outline-none"
        >
          {open ? (
            <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          )}
          <span className="text-sm font-medium text-foreground">{modelDisplayName}</span>
          <span className="font-mono text-xs text-muted-foreground">({modelId})</span>
          <span className="ml-auto text-xs text-muted-foreground">
            {permissions.length} 个权限点
          </span>
        </button>
      </CollapsibleTrigger>

      {/* Expanded: flat table, no extra padding wrapper */}
      <CollapsibleContent>
        <div className="border-t">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="h-8 pl-4 text-xs font-medium text-muted-foreground">
                  动作
                </TableHead>
                <TableHead className="h-8 text-xs font-medium text-muted-foreground">
                  行范围
                </TableHead>
                <TableHead className="h-8 text-xs font-medium text-muted-foreground">
                  列策略
                </TableHead>
                <TableHead className="h-8 text-xs font-medium text-muted-foreground">
                  名称
                </TableHead>
                <TableHead className="h-8 w-10 pr-4" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {permissions.map((permission) => (
                <PermissionRow
                  key={permission.id}
                  permission={permission}
                  onDelete={onDelete}
                  deletingId={deletingId}
                />
              ))}
            </TableBody>
          </Table>
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

// ── PermissionsTab ────────────────────────────────────────────────────────────

export function PermissionsTab({ orgName, projectSlug }: PermissionsTabProps) {
  const router = useRouter()

  const [deletingId, setDeletingId] = React.useState<string | null>(null)

  const { groupedPermissions, loading, error, deletePermission } = usePermissionList({
    orgName,
    projectSlug,
  })

  const modelEntries = Object.entries(groupedPermissions)

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
        toast.error('删除权限点时发生错误，请重试')
      } finally {
        setDeletingId(null)
      }
    },
    [deletePermission]
  )

  const handleCreateNew = React.useCallback(() => {
    router.push(`/org/${orgName}/project/${projectSlug}/rbac/permissions/new`)
  }, [router, orgName, projectSlug])

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex justify-end">
        <Button size="sm" onClick={handleCreateNew} className="shrink-0">
          <Plus className="mr-1.5 size-4" />
          创建权限点
        </Button>
      </div>

      {/* Error banner */}
      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          加载权限点失败：{error.message}
        </div>
      )}

      {/* Content */}
      {loading ? (
        <PermissionListSkeleton />
      ) : modelEntries.length === 0 ? (
        /* Empty state — icon + title + hint + action */
        <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
          <ShieldAlert className="mb-3 size-10 text-muted-foreground/25" />
          <p className="text-sm font-medium text-foreground">暂无权限点</p>
          <p className="mt-1 text-xs text-muted-foreground">
            点击「创建权限点」添加第一个权限点
          </p>
          <Button
            size="sm"
            variant="outline"
            className="mt-4"
            onClick={handleCreateNew}
          >
            <Plus className="mr-1.5 size-4" />
            创建权限点
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {modelEntries.map(([modelId, permissions]) => {
            const modelDisplayName = permissions[0]?.modelDisplayName ?? modelId
            return (
              <ModelGroup
                key={modelId}
                modelId={modelId}
                modelDisplayName={modelDisplayName}
                permissions={permissions}
                onDelete={handleDelete}
                deletingId={deletingId}
              />
            )
          })}
        </div>
      )}
    </div>
  )
}
