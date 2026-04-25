'use client'

import * as React from 'react'
import { useParams, useRouter } from 'next/navigation'
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
import { PageHeader } from '@web/components/features/layout'

import { usePermissionList } from './_hooks/usePermissionList'
import type {
  EndUserPermission,
  EndUserPermissionAction,
  EndUserRowScope,
  ColumnPolicy,
} from '@/types'

// ── Action badge variant mapping ─────────────────────────────────────────────

type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline'

const ACTION_BADGE_VARIANT: Record<EndUserPermissionAction, BadgeVariant> = {
  SELECT: 'secondary',
  EXPORT: 'secondary',
  INSERT: 'outline',
  UPDATE: 'outline',
  DELETE: 'outline',
}

const ACTION_LABEL: Record<EndUserPermissionAction, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '修改',
  DELETE: '删除',
  EXPORT: '导出',
}

// ── Row scope badge variant mapping ─────────────────────────────────────────

const ROW_SCOPE_BADGE_VARIANT: Record<EndUserRowScope, BadgeVariant> = {
  ALL: 'secondary',
  SELF: 'outline',
  DEPT: 'outline',
  DEPT_AND_CHILDREN: 'outline',
}

const ROW_SCOPE_LABEL: Record<EndUserRowScope, string> = {
  ALL: '全部行',
  SELF: '本人行',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '本部门及下级',
}

// ── Column policy summary ────────────────────────────────────────────────────

function formatColumnPolicySummary(policy: ColumnPolicy | undefined | null): string {
  if (!policy) return '默认: 全部可见'
  const defaultLabel = policy.defaultMode === 'VISIBLE' ? '全部可见'
    : policy.defaultMode === 'HIDDEN' ? '全部隐藏'
    : policy.defaultMode === 'READONLY' ? '只读'
    : policy.defaultMode === 'MASKED' ? '脱敏'
    : policy.defaultMode
  const rulesCount = policy.rules?.length ?? 0
  if (rulesCount === 0) return `默认: ${defaultLabel}`
  return `默认: ${defaultLabel}，+${rulesCount} 条规则`
}

// ── Skeleton ─────────────────────────────────────────────────────────────────

function PermissionListSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="rounded-md border p-4 space-y-2">
          <Skeleton className="h-5 w-40" />
          <div className="space-y-2 pt-2">
            {Array.from({ length: 2 }).map((_, j) => (
              <div key={j} className="flex items-center gap-3">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-4 w-48" />
                <Skeleton className="ml-auto h-7 w-14" />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

// ── PermissionRow ────────────────────────────────────────────────────────────

interface PermissionRowProps {
  permission: EndUserPermission
  onDelete: (permission: EndUserPermission) => Promise<void>
  deletingId: string | null
}

function PermissionRow({ permission, onDelete, deletingId }: PermissionRowProps) {
  return (
    <div className="flex flex-wrap items-center gap-3 rounded-md px-3 py-2.5 hover:bg-muted/30">
      {/* Action badge */}
      <Badge variant={ACTION_BADGE_VARIANT[permission.action]} className="shrink-0 font-semibold">
        {ACTION_LABEL[permission.action]}
      </Badge>

      {/* Row scope badge */}
      <Badge
        variant={ROW_SCOPE_BADGE_VARIANT[permission.rowScope]}
        className="shrink-0"
      >
        {ROW_SCOPE_LABEL[permission.rowScope]}
      </Badge>

      {/* Column policy summary */}
      <span className="flex-1 text-xs text-muted-foreground">
        {formatColumnPolicySummary(
          permission.columnPolicy as ColumnPolicy | undefined
        )}
      </span>

      {/* Display name (if any) */}
      {permission.displayName && (
        <span className="text-sm text-foreground">{permission.displayName}</span>
      )}

      {/* Delete action */}
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto h-7 shrink-0 px-2 text-xs text-muted-foreground hover:text-destructive"
            disabled={deletingId === permission.id}
          >
            <Trash2 className="size-3.5" />
            {deletingId === permission.id ? '删除中...' : ''}
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
    </div>
  )
}

// ── ModelGroup ───────────────────────────────────────────────────────────────

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
      <CollapsibleTrigger asChild>
        <button
          type="button"
          className="flex w-full items-center gap-2 px-4 py-3 text-left hover:bg-muted/20 focus-visible:outline-none"
        >
          {open ? (
            <ChevronDown className="size-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-4 text-muted-foreground" />
          )}
          <span className="font-semibold text-foreground">{modelDisplayName}</span>
          <span className="ml-1 font-mono text-xs text-muted-foreground">({modelId})</span>
          <span className="ml-auto rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
            {permissions.length}
          </span>
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="border-t px-2 py-1">
          {permissions.map((permission) => (
            <PermissionRow
              key={permission.id}
              permission={permission}
              onDelete={onDelete}
              deletingId={deletingId}
            />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

// ── PermissionListPage ───────────────────────────────────────────────────────

export default function PermissionListPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const projectSlug = params.projectSlug

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
    <>
      <PageHeader
        title="权限点"
        spacing="compact"
        actions={
          <Button size="sm" onClick={handleCreateNew}>
            <Plus className="mr-1.5 size-4" />
            创建权限点
          </Button>
        }
      />

        {/* Error */}
        {error && (
          <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            加载权限点失败：{error.message}
          </div>
        )}

        {/* Content */}
        {loading ? (
          <PermissionListSkeleton />
        ) : modelEntries.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16">
            <ShieldAlert className="mb-4 size-10 text-muted-foreground/30" />
            <p className="text-[14px] font-medium text-foreground">暂无权限点</p>
            <p className="mt-1 text-[13px] text-muted-foreground">创建第一个权限点，开始配置访问控制</p>
            <Button size="sm" className="mt-5" onClick={handleCreateNew}>
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
    </>
  )
}
