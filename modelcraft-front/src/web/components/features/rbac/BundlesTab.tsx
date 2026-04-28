import * as React from 'react'
import Link from 'next/link'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import {
  Plus,
  Trash2,
  PackageOpen,
  Eye,
  Link2,
  Loader2,
  ShieldCheck,
  ArrowRight,
  X,
} from 'lucide-react'

import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
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
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@web/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@web/components/ui/form'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import { ScrollArea } from '@web/components/ui/scroll-area'

import { useBundleList } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleList'
import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type { EndUserPermission, EndUserPermissionBundle } from '@/types'

// ── Props ────────────────────────────────────────────────────────────────────

export interface BundlesTabProps {
  orgName: string
  projectSlug: string
}

// ── Validation Schema ────────────────────────────────────────────────────────

const createBundleSchema = z.object({
  name: z
    .string()
    .min(1, '权限包名称不能为空')
    .max(50, '权限包名称不能超过 50 个字符'),
  description: z
    .string()
    .max(200, '描述不能超过 200 个字符')
    .optional(),
})

type CreateBundleFormValues = z.infer<typeof createBundleSchema>

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

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatDateTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

function permissionLabel(p: EndUserPermission): string {
  const model = p.modelDisplayName ?? p.modelId
  const action = ACTION_LABEL[p.action] ?? p.action
  const scope = ROW_SCOPE_LABEL[p.rowScope] ?? p.rowScope
  return `${model} / ${action} / ${scope}`
}

// ── CreateBundleDialog ───────────────────────────────────────────────────────

interface CreateBundleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (values: CreateBundleFormValues) => Promise<void>
  submitting: boolean
}

function CreateBundleDialog({
  open,
  onOpenChange,
  onSubmit,
  submitting,
}: CreateBundleDialogProps) {
  const form = useForm<CreateBundleFormValues>({
    resolver: zodResolver(createBundleSchema),
    defaultValues: { name: '', description: '' },
  })

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit(values)
    form.reset()
  })

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) form.reset()
    onOpenChange(nextOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>创建权限包</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={handleSubmit} className="space-y-3 pt-1">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    名称
                    <span className="ml-1 text-destructive">*</span>
                  </FormLabel>
                  <FormControl>
                    <Input placeholder="例如：订单管理包" maxLength={50} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>描述（可选）</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="简要描述该权限包的用途..."
                      maxLength={200}
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className="pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={submitting}
              >
                取消
              </Button>
              <Button type="submit" disabled={submitting} className="bg-primary text-primary-foreground hover:bg-primary/90">
                {submitting ? '创建中...' : '创建'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

// ── ViewBundleSheet ──────────────────────────────────────────────────────────

interface ViewBundleSheetProps {
  bundle: EndUserPermissionBundle | null
  loading: boolean
  open: boolean
  onOpenChange: (open: boolean) => void
  onRemove: (permission: EndUserPermission) => Promise<void>
  removingId: string | null
}

function ViewBundleSheet({
  bundle,
  loading,
  open,
  onOpenChange,
  onRemove,
  removingId,
}: ViewBundleSheetProps) {
  const permissions = bundle?.permissions ?? []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{bundle?.name ?? '权限包详情'}</SheetTitle>
          <SheetDescription>
            {bundle?.description ?? '查看该权限包关联的权限点，可直接解绑。'}
          </SheetDescription>
        </SheetHeader>

        <div className="mt-4 flex-1 overflow-hidden">
          {loading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : permissions.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12">
              <ShieldCheck className="mb-3 size-8 text-muted-foreground/30" />
              <p className="text-sm text-muted-foreground">暂无关联权限点</p>
            </div>
          ) : (
            <ScrollArea className="h-full pr-1">
              <div className="space-y-1.5">
                {permissions.map((perm) => (
                  <div
                    key={perm.id}
                    className="flex items-center justify-between rounded-md border bg-card px-3 py-2"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-foreground">
                        {perm.modelDisplayName ?? perm.modelId}
                      </p>
                      <div className="mt-0.5 flex items-center gap-1.5">
                        <Badge
                          variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
                          className="h-4 px-1 py-0 text-[10px]"
                        >
                          {ACTION_LABEL[perm.action] ?? perm.action}
                        </Badge>
                        <Badge variant="outline" className="h-4 px-1 py-0 text-[10px]">
                          {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                        </Badge>
                      </div>
                    </div>
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="ml-2 size-7 shrink-0 text-muted-foreground hover:text-destructive"
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
                            确定要从该权限包中移除「{permissionLabel(perm)}」吗？
                            此操作不会删除权限点本身。
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>取消</AlertDialogCancel>
                          <AlertDialogAction
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            onClick={() => void onRemove(perm)}
                          >
                            确认移除
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}

// ── ManagePoliciesSheet ──────────────────────────────────────────────────────

interface ManagePoliciesSheetProps {
  bundle: EndUserPermissionBundle | null
  allPermissions: EndUserPermission[]
  loading: boolean
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (toAdd: string[], toRemove: string[]) => Promise<void>
}

function ManagePoliciesSheet({
  bundle,
  allPermissions,
  loading,
  open,
  onOpenChange,
  onSave,
}: ManagePoliciesSheetProps) {
  const [search, setSearch] = React.useState('')
  const [pendingAdd, setPendingAdd] = React.useState<Set<string>>(new Set())
  const [pendingRemove, setPendingRemove] = React.useState<Set<string>>(new Set())
  const [saving, setSaving] = React.useState(false)

  // Reset local state when sheet closes
  React.useEffect(() => {
    if (!open) {
      setSearch('')
      setPendingAdd(new Set())
      setPendingRemove(new Set())
    }
  }, [open])

  const originalIds = React.useMemo(
    () => new Set((bundle?.permissions ?? []).map((p) => p.id)),
    [bundle],
  )

  const permById = React.useMemo(() => {
    const map = new Map<string, EndUserPermission>()
    allPermissions.forEach((p) => map.set(p.id, p))
    return map
  }, [allPermissions])

  // Left: permissions not already in bundle and not pending-add, filtered by search
  const leftPermissions = React.useMemo(() => {
    const q = search.trim().toLowerCase()
    return allPermissions
      .filter((p) => !originalIds.has(p.id) && !pendingAdd.has(p.id))
      .filter((p) => {
        if (!q) return true
        return (
          (p.modelDisplayName ?? p.modelId).toLowerCase().includes(q) ||
          (ACTION_LABEL[p.action] ?? p.action).toLowerCase().includes(q)
        )
      })
  }, [allPermissions, originalIds, pendingAdd, search])

  // Right: original (not pending-remove) + pending-add
  const rightPermissions = React.useMemo(() => {
    const original = (bundle?.permissions ?? []).filter((p) => !pendingRemove.has(p.id))
    const added = Array.from(pendingAdd)
      .map((id) => permById.get(id))
      .filter((p): p is EndUserPermission => !!p)
    return [...original, ...added]
  }, [bundle, pendingRemove, pendingAdd, permById])

  const totalChanges = pendingAdd.size + pendingRemove.size

  const handleMoveRight = (permId: string) => {
    setPendingAdd((prev) => new Set(prev).add(permId))
  }

  const handleMoveLeft = (perm: EndUserPermission) => {
    if (pendingAdd.has(perm.id)) {
      // New addition — just cancel it
      setPendingAdd((prev) => {
        const next = new Set(prev)
        next.delete(perm.id)
        return next
      })
    } else {
      // Original item — mark for removal
      setPendingRemove((prev) => new Set(prev).add(perm.id))
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await onSave(Array.from(pendingAdd), Array.from(pendingRemove))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>关联策略 — {bundle?.name ?? ''}</SheetTitle>
          <SheetDescription>
            点击 → 将权限点移入权限包，点击 × 可移出。确认无误后点「保存」提交。
          </SheetDescription>
        </SheetHeader>

        {loading ? (
          <div className="mt-4 space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : (
          <div className="mt-4 flex flex-1 gap-4 overflow-hidden">
            {/* ── 左侧：可添加 ── */}
            <div className="flex flex-1 flex-col overflow-hidden rounded-md border bg-muted/20">
              <div className="border-b px-3 py-2.5">
                <p className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  可添加的权限点
                  <span className="ml-1 text-foreground/60">({leftPermissions.length})</span>
                </p>
                <Input
                  placeholder="搜索模型名称..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="h-7 text-xs"
                />
              </div>
              <ScrollArea className="flex-1">
                {leftPermissions.length === 0 ? (
                  <p className="py-8 text-center text-xs text-muted-foreground">
                    {allPermissions.filter((p) => !originalIds.has(p.id)).length === 0
                      ? '所有权限点已全部关联'
                      : '无匹配结果'}
                  </p>
                ) : (
                  <div className="space-y-1 p-2">
                    {leftPermissions.map((perm) => (
                      <div
                        key={perm.id}
                        className="flex items-center justify-between rounded px-2 py-1.5 hover:bg-accent"
                      >
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-xs font-medium text-foreground">
                            {perm.modelDisplayName ?? perm.modelId}
                          </p>
                          <div className="mt-0.5 flex items-center gap-1">
                            <Badge
                              variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
                              className="h-4 px-1 py-0 text-[10px]"
                            >
                              {ACTION_LABEL[perm.action] ?? perm.action}
                            </Badge>
                            <Badge variant="outline" className="h-4 px-1 py-0 text-[10px]">
                              {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                            </Badge>
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="ml-1 size-6 shrink-0 text-primary hover:bg-primary/10"
                          disabled={saving}
                          onClick={() => handleMoveRight(perm.id)}
                        >
                          <ArrowRight className="size-3" />
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </ScrollArea>
            </div>

            {/* ── 右侧：已关联 ── */}
            <div className="flex flex-1 flex-col overflow-hidden rounded-md border bg-muted/20">
              <div className="border-b px-3 py-2.5">
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  已关联
                  <span className="ml-1 text-foreground/60">({rightPermissions.length})</span>
                </p>
              </div>
              <ScrollArea className="flex-1">
                {rightPermissions.length === 0 ? (
                  <p className="py-8 text-center text-xs text-muted-foreground">暂无关联权限点</p>
                ) : (
                  <div className="space-y-1 p-2">
                    {rightPermissions.map((perm) => {
                      const isNew = pendingAdd.has(perm.id)
                      return (
                        <div
                          key={perm.id}
                          className="flex items-center justify-between rounded px-2 py-1.5 hover:bg-accent"
                        >
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-1.5">
                              <p className="truncate text-xs font-medium text-foreground">
                                {perm.modelDisplayName ?? perm.modelId}
                              </p>
                              {isNew && (
                                <span className="shrink-0 rounded bg-primary/10 px-1 py-0 text-[10px] font-medium text-primary">
                                  待添加
                                </span>
                              )}
                            </div>
                            <div className="mt-0.5 flex items-center gap-1">
                              <Badge
                                variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
                                className="h-4 px-1 py-0 text-[10px]"
                              >
                                {ACTION_LABEL[perm.action] ?? perm.action}
                              </Badge>
                              <Badge variant="outline" className="h-4 px-1 py-0 text-[10px]">
                                {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                              </Badge>
                            </div>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="ml-1 size-6 shrink-0 text-muted-foreground hover:text-destructive"
                            disabled={saving}
                            onClick={() => handleMoveLeft(perm)}
                          >
                            <X className="size-3" />
                          </Button>
                        </div>
                      )
                    })}
                  </div>
                )}
              </ScrollArea>
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="mt-4 flex items-center justify-end gap-2 border-t pt-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button
            size="sm"
            disabled={saving || totalChanges === 0}
            onClick={() => void handleSave()}
          >
            {saving ? (
              <>
                <Loader2 className="mr-1.5 size-3.5 animate-spin" />
                保存中...
              </>
            ) : totalChanges > 0 ? (
              `保存（${pendingAdd.size > 0 ? `+${pendingAdd.size}` : ''}${pendingRemove.size > 0 ? ` -${pendingRemove.size}` : ''}）`
            ) : (
              '保存'
            )}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}

// ── BundleTableSkeleton ──────────────────────────────────────────────────────

function BundleTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow className="h-9">
            <TableHead className="w-[200px] text-xs font-medium text-muted-foreground">名称</TableHead>
            <TableHead className="text-xs font-medium text-muted-foreground">描述</TableHead>
            <TableHead className="w-[100px] text-xs font-medium text-muted-foreground">权限点数量</TableHead>
            <TableHead className="w-[180px] text-xs font-medium text-muted-foreground">创建时间</TableHead>
            <TableHead className="w-[180px] text-right text-xs font-medium text-muted-foreground">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-4 w-10" /></TableCell>
              <TableCell><Skeleton className="h-4 w-36" /></TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-1.5">
                  <Skeleton className="h-7 w-12" />
                  <Skeleton className="h-7 w-16" />
                  <Skeleton className="h-7 w-12" />
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// ── BundlesTab ───────────────────────────────────────────────────────────────

type SheetMode = 'view' | 'manage' | null

export function BundlesTab({ orgName, projectSlug }: BundlesTabProps) {
  const [createDialogOpen, setCreateDialogOpen] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)

  const [sheetMode, setSheetMode] = React.useState<SheetMode>(null)
  const [selectedBundleId, setSelectedBundleId] = React.useState<string | null>(null)
  const [removingId, setRemovingId] = React.useState<string | null>(null)

  const { bundles, loading, error, createBundle, deleteBundle } = useBundleList({
    orgName,
    projectSlug,
  })

  const { bundle, allPermissions, loading: bundleLoading, addPermission, removePermission } = useBundleManage({
    orgName,
    projectSlug,
    bundleId: selectedBundleId,
  })

  // ── Handlers ──────────────────────────────────────────────────────────────

  const handleCreate = React.useCallback(
    async (values: CreateBundleFormValues) => {
      setSubmitting(true)
      try {
        const result = await createBundle({ name: values.name, description: values.description })
        if (result.success) {
          toast.success(`权限包「${values.name}」已创建`)
          setCreateDialogOpen(false)
        } else {
          toast.error(result.errorMessage ?? '创建失败，请重试')
        }
      } catch {
        toast.error('创建权限包时发生错误，请重试')
      } finally {
        setSubmitting(false)
      }
    },
    [createBundle],
  )

  const handleDelete = React.useCallback(
    async (b: EndUserPermissionBundle) => {
      setDeletingId(b.id)
      try {
        const result = await deleteBundle(b.id)
        if (result.success) {
          toast.success(`权限包「${b.name}」已删除`)
        } else {
          toast.error(result.errorMessage ?? '删除失败，请重试')
        }
      } catch {
        toast.error('删除权限包时发生错误，请重试')
      } finally {
        setDeletingId(null)
      }
    },
    [deleteBundle],
  )

  const openSheet = (bundleId: string, mode: SheetMode) => {
    setSelectedBundleId(bundleId)
    setSheetMode(mode)
  }

  const closeSheet = () => setSheetMode(null)

  const handleRemovePermission = React.useCallback(
    async (perm: EndUserPermission) => {
      setRemovingId(perm.id)
      try {
        const result = await removePermission(perm.id)
        if (result.success) {
          toast.success(`已移除「${permissionLabel(perm)}」`)
        } else {
          toast.error(result.errorMessage ?? '移除失败，请重试')
        }
      } catch {
        toast.error('移除权限点时发生错误')
      } finally {
        setRemovingId(null)
      }
    },
    [removePermission],
  )

  const handleSaveManage = React.useCallback(
    async (toAdd: string[], toRemove: string[]) => {
      let failCount = 0
      for (const id of toAdd) {
        const result = await addPermission(id)
        if (!result.success) failCount++
      }
      for (const id of toRemove) {
        const result = await removePermission(id)
        if (!result.success) failCount++
      }
      if (failCount === 0) {
        toast.success(`已保存 ${toAdd.length + toRemove.length} 项变更`)
        closeSheet()
      } else {
        toast.error(`${failCount} 项操作失败，请重试`)
      }
    },
    [addPermission, removePermission],
  )

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex justify-end">
        <Button size="sm" onClick={() => setCreateDialogOpen(true)} className="shrink-0">
          <Plus className="mr-1.5 size-4" />
          创建权限包
        </Button>
      </div>

      {/* Error */}
      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          加载权限包失败：{error.message}
        </div>
      )}

      {/* Table */}
      {loading ? (
        <BundleTableSkeleton />
      ) : bundles.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
          <PackageOpen className="mb-3 size-10 text-muted-foreground/25" />
          <p className="text-sm font-medium text-foreground">暂无权限包</p>
          <p className="mt-1 text-xs text-muted-foreground">
            点击「创建权限包」添加第一个权限包
          </p>
          <Button
            size="sm"
            variant="outline"
            className="mt-4"
            onClick={() => setCreateDialogOpen(true)}
          >
            <Plus className="mr-1.5 size-4" />
            创建权限包
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="h-9">
                <TableHead className="w-[200px] text-xs font-medium text-muted-foreground">名称</TableHead>
                <TableHead className="text-xs font-medium text-muted-foreground">描述</TableHead>
                <TableHead className="w-[100px] text-xs font-medium text-muted-foreground">权限点数量</TableHead>
                <TableHead className="w-[180px] text-xs font-medium text-muted-foreground">创建时间</TableHead>
                <TableHead className="w-[180px] text-right text-xs font-medium text-muted-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {bundles.map((b) => (
                <TableRow key={b.id} className="hover:bg-muted/50">
                  <TableCell className="font-semibold text-foreground">
                    <Link
                      href={`/org/${orgName}/project/${projectSlug}/roles/bundles/${b.id}`}
                      className="hover:underline"
                    >
                      {b.name}
                    </Link>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {b.description || (
                      <span className="italic text-muted-foreground/50">无描述</span>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {b.permissions?.length ?? 0}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDateTime(b.createdAt)}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      {/* 查看 */}
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 px-2 text-xs"
                        onClick={() => openSheet(b.id, 'view')}
                      >
                        <Eye className="mr-1 size-3" />
                        查看
                      </Button>

                      {/* 关联策略 */}
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 px-2 text-xs"
                        onClick={() => openSheet(b.id, 'manage')}
                      >
                        <Link2 className="mr-1 size-3" />
                        关联策略
                      </Button>

                      {/* 删除 */}
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                            disabled={deletingId === b.id}
                          >
                            {deletingId === b.id ? (
                              <Loader2 className="mr-1 size-3 animate-spin" />
                            ) : (
                              <Trash2 className="mr-1 size-3" />
                            )}
                            {deletingId === b.id ? '删除中...' : '删除'}
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>确认删除权限包</AlertDialogTitle>
                            <AlertDialogDescription>
                              确定要删除权限包「{b.name}」吗？
                              删除后关联角色和用户将失去该权限包内的所有权限，此操作不可撤销。
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              onClick={() => void handleDelete(b)}
                            >
                              确认删除
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Dialogs & Sheets */}
      <CreateBundleDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreate}
        submitting={submitting}
      />

      <ViewBundleSheet
        bundle={sheetMode === 'view' ? bundle : null}
        loading={bundleLoading && sheetMode === 'view'}
        open={sheetMode === 'view'}
        onOpenChange={(open) => { if (!open) closeSheet() }}
        onRemove={handleRemovePermission}
        removingId={removingId}
      />

      <ManagePoliciesSheet
        bundle={sheetMode === 'manage' ? bundle : null}
        allPermissions={allPermissions}
        loading={bundleLoading && sheetMode === 'manage'}
        open={sheetMode === 'manage'}
        onOpenChange={(open) => { if (!open) closeSheet() }}
        onSave={handleSaveManage}
      />
    </div>
  )
}
