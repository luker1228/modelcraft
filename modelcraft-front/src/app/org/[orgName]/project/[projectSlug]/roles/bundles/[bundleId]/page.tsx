'use client'

import * as React from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ArrowLeft, ShieldCheck, Loader2, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
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
import { PageLayout } from '@web/components/features/layout'

import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type { EndUserPermission } from '@/types'

// ── Label maps ───────────────────────────────────────────────────────────────

const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '更新',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部',
  SELF: '本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '部门及子部门',
}

const ACTION_VARIANT: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  SELECT: 'secondary',
  INSERT: 'default',
  UPDATE: 'outline',
  DELETE: 'destructive',
  EXPORT: 'outline',
}

// ── BundleDetailPage ─────────────────────────────────────────────────────────

export default function BundleDetailPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`

  const { bundle, loading, error, removePermission } = useBundleManage({
    orgName,
    projectSlug,
    bundleId,
  })

  const [removingId, setRemovingId] = React.useState<string | null>(null)

  const handleRemove = async (perm: EndUserPermission) => {
    setRemovingId(perm.id)
    try {
      const result = await removePermission(perm.id)
      if (result.success) {
        toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
      } else {
        toast.error(result.errorMessage ?? '移除失败，请重试')
      }
    } catch {
      toast.error('移除权限点时发生错误')
    } finally {
      setRemovingId(null)
    }
  }

  const permissions = bundle?.permissions ?? []

  return (
    <PageLayout maxWidth="5xl">
      {/* Back nav */}
      <div className="mb-6">
        <Link
          href={backHref}
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          返回权限包列表
        </Link>
      </div>

      {/* Header */}
      <div className="mb-8">
        {loading ? (
          <>
            <Skeleton className="h-7 w-48" />
            <Skeleton className="mt-2 h-4 w-64" />
          </>
        ) : (
          <>
            <h1 className="text-xl font-semibold tracking-tight text-foreground">
              {bundle?.name ?? '权限包详情'}
            </h1>
            {bundle?.description && (
              <p className="mt-1 text-sm text-muted-foreground">{bundle.description}</p>
            )}
          </>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          加载失败：{error.message}
        </div>
      )}

      {/* Permissions list */}
      <div className="rounded-md border border-border">
        {/* Header row */}
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            关联的权限点
          </span>
          {!loading && (
            <span className="ml-2 rounded-full bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
              {permissions.length}
            </span>
          )}
        </div>

        {loading ? (
          <div className="space-y-px">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-4 w-28" />
                <Skeleton className="h-5 w-14" />
                <Skeleton className="h-5 w-16" />
              </div>
            ))}
          </div>
        ) : permissions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <ShieldCheck className="mb-3 size-8 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">暂无关联权限点</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {permissions.map((perm) => (
              <div
                key={perm.id}
                className="group flex items-center gap-3 px-4 py-3 hover:bg-muted/20"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-foreground">
                      {perm.displayName || perm.modelDisplayName || perm.modelId}
                    </span>
                    {perm.displayName && (perm.modelDisplayName || perm.modelId) && (
                      <span className="font-mono text-[11px] text-muted-foreground/60">
                        {perm.modelDisplayName ?? perm.modelId}
                      </span>
                    )}
                  </div>
                  <div className="mt-1 flex items-center gap-1.5">
                    <Badge
                      variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
                      className="h-4 px-1.5 py-0 text-[10px]"
                    >
                      {ACTION_LABEL[perm.action] ?? perm.action}
                    </Badge>
                    <Badge variant="outline" className="h-4 px-1.5 py-0 text-[10px]">
                      {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                    </Badge>
                  </div>
                </div>

                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-destructive"
                      disabled={removingId === perm.id}
                    >
                      {removingId === perm.id ? (
                        <Loader2 className="size-3.5 animate-spin" />
                      ) : (
                        <Trash2 className="size-3.5" />
                      )}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>确认解绑权限点</AlertDialogTitle>
                      <AlertDialogDescription>
                        确定要从该权限包中移除「{perm.displayName ?? perm.modelId}」吗？
                        此操作不会删除权限点本身。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction
                        className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        onClick={() => void handleRemove(perm)}
                      >
                        确认移除
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            ))}
          </div>
        )}
      </div>
    </PageLayout>
  )
}
