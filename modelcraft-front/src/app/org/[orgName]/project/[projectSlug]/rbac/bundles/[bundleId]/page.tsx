'use client'

import * as React from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import { ArrowLeft, Plus, Trash2, ShieldCheck } from 'lucide-react'

import { Button } from '@web/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
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
import { Checkbox } from '@web/components/ui/checkbox'
import { Skeleton } from '@web/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'

import { useBundleEdit } from './_hooks/useBundleEdit'
import type { EndUserPermission } from '@/types'

// ── Validation Schema ────────────────────────────────────────────────────────

const editBundleSchema = z.object({
  name: z
    .string()
    .min(1, '权限包名称不能为空')
    .max(50, '权限包名称不能超过 50 个字符'),
  description: z
    .string()
    .max(200, '描述不能超过 200 个字符')
    .optional(),
})

type EditBundleFormValues = z.infer<typeof editBundleSchema>

// ── Action / RowScope Badge helpers ─────────────────────────────────────────

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

// ── Skeletons ────────────────────────────────────────────────────────────────

function InfoCardSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-5 w-24" />
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-9 w-full" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-20 w-full" />
        </div>
        <div className="flex justify-end">
          <Skeleton className="h-9 w-16" />
        </div>
      </CardContent>
    </Card>
  )
}

function PermissionsTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>模型</TableHead>
            <TableHead className="w-[100px]">操作</TableHead>
            <TableHead className="w-[160px]">行权限范围</TableHead>
            <TableHead className="w-[80px] text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
              <TableCell><Skeleton className="h-5 w-14" /></TableCell>
              <TableCell><Skeleton className="h-5 w-20" /></TableCell>
              <TableCell className="text-right"><Skeleton className="ml-auto h-7 w-14" /></TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// ── AddPermissionDialog ──────────────────────────────────────────────────────

interface AddPermissionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  allPermissions: EndUserPermission[]
  bundlePermissionIds: Set<string>
  onConfirm: (permissionIds: string[]) => Promise<void>
  submitting: boolean
}

function AddPermissionDialog({
  open,
  onOpenChange,
  allPermissions,
  bundlePermissionIds,
  onConfirm,
  submitting,
}: AddPermissionDialogProps) {
  const [search, setSearch] = React.useState('')
  const [selected, setSelected] = React.useState<Set<string>>(new Set())

  // 关闭时重置状态
  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setSearch('')
      setSelected(new Set())
    }
    onOpenChange(nextOpen)
  }

  // 过滤出未关联的权限点
  const available = React.useMemo(
    () => allPermissions.filter((p) => !bundlePermissionIds.has(p.id)),
    [allPermissions, bundlePermissionIds]
  )

  // 搜索过滤
  const filtered = React.useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return available
    return available.filter(
      (p) =>
        (p.modelDisplayName ?? p.modelId).toLowerCase().includes(q) ||
        (ACTION_LABEL[p.action] ?? p.action).toLowerCase().includes(q)
    )
  }, [available, search])

  const toggleItem = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const handleConfirm = async () => {
    await onConfirm(Array.from(selected))
    handleOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>添加权限点</DialogTitle>
        </DialogHeader>

        {/* 搜索框 */}
        <Input
          placeholder="搜索模型名称或操作类型..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="mb-1"
        />

        {/* 权限点列表 */}
        <div className="max-h-80 overflow-y-auto rounded-md border">
          {filtered.length === 0 ? (
            <div className="flex items-center justify-center py-10 text-sm text-muted-foreground">
              {available.length === 0 ? '所有权限点已全部关联' : '未找到匹配的权限点'}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-8" />
                  <TableHead>模型</TableHead>
                  <TableHead className="w-[100px]">操作</TableHead>
                  <TableHead className="w-[140px]">行权限范围</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((p) => (
                  <TableRow
                    key={p.id}
                    className="cursor-pointer"
                    onClick={() => toggleItem(p.id)}
                  >
                    <TableCell>
                      <Checkbox
                        checked={selected.has(p.id)}
                        onCheckedChange={() => toggleItem(p.id)}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </TableCell>
                    <TableCell className="font-semibold text-foreground">
                      {p.modelDisplayName}
                    </TableCell>
                    <TableCell>
                      <Badge variant={ACTION_VARIANT[p.action] ?? 'secondary'}>
                        {ACTION_LABEL[p.action] ?? p.action}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {ROW_SCOPE_LABEL[p.rowScope] ?? p.rowScope}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>

        <DialogFooter className="mt-2">
          <span className="mr-auto text-sm text-muted-foreground">
            已选 {selected.size} 项
          </span>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={submitting}
          >
            取消
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={submitting || selected.size === 0}
          >
            {submitting ? '添加中...' : `确认添加`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── BundleEditPage ───────────────────────────────────────────────────────────

export default function BundleEditPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const { bundle, allPermissions, loading, error, updateBundle, addPermission, removePermission } =
    useBundleEdit({ orgName, projectSlug, bundleId })

  // ── Form ─────────────────────────────────────────────────────────────────

  const form = useForm<EditBundleFormValues>({
    resolver: zodResolver(editBundleSchema),
    defaultValues: { name: '', description: '' },
  })

  // bundle 加载完成后用实际数据回填表单
  React.useEffect(() => {
    if (bundle) {
      form.reset({
        name: bundle.name,
        description: bundle.description ?? '',
      })
    }
  }, [bundle, form])

  // ── Local state ───────────────────────────────────────────────────────────

  const [saving, setSaving] = React.useState(false)
  const [addDialogOpen, setAddDialogOpen] = React.useState(false)
  const [addingPermissions, setAddingPermissions] = React.useState(false)
  const [removingId, setRemovingId] = React.useState<string | null>(null)

  // ── Derived ───────────────────────────────────────────────────────────────

  const bundlePermissionIds = React.useMemo(
    () => new Set((bundle?.permissions ?? []).map((p) => p.id)),
    [bundle]
  )

  // ── Handlers ──────────────────────────────────────────────────────────────

  const handleSave = form.handleSubmit(async (values) => {
    // 只有名称变更才调用 mutation
    if (
      bundle &&
      values.name === bundle.name &&
      (values.description ?? '') === (bundle.description ?? '')
    ) {
      toast.info('没有可保存的更改')
      return
    }

    setSaving(true)
    try {
      const result = await updateBundle({
        name: values.name,
        description: values.description,
      })
      if (result.success) {
        toast.success('权限包已更新')
      } else {
        toast.error(result.errorMessage ?? '更新失败，请重试')
      }
    } catch {
      toast.error('更新权限包时发生错误，请重试')
    } finally {
      setSaving(false)
    }
  })

  const handleAddPermissions = React.useCallback(
    async (permissionIds: string[]) => {
      setAddingPermissions(true)
      let failCount = 0
      try {
        for (const id of permissionIds) {
          const result = await addPermission(id)
          if (!result.success) failCount++
        }
        if (failCount === 0) {
          toast.success(`已添加 ${permissionIds.length} 个权限点`)
        } else {
          toast.error(`${failCount} 个权限点添加失败，请重试`)
        }
      } catch {
        toast.error('添加权限点时发生错误，请重试')
      } finally {
        setAddingPermissions(false)
      }
    },
    [addPermission]
  )

  const handleRemovePermission = React.useCallback(
    async (permission: EndUserPermission) => {
      setRemovingId(permission.id)
      try {
        const result = await removePermission(permission.id)
        if (result.success) {
          toast.success(`已移除权限点「${permission.modelDisplayName} / ${ACTION_LABEL[permission.action] ?? permission.action}」`)
        } else {
          toast.error(result.errorMessage ?? '移除失败，请重试')
        }
      } catch {
        toast.error('移除权限点时发生错误，请重试')
      } finally {
        setRemovingId(null)
      }
    },
    [removePermission]
  )

  // ── Render ────────────────────────────────────────────────────────────────

  const bundlesListHref = `/org/${orgName}/project/${projectSlug}/rbac/bundles`

  return (
    <main className="size-full overflow-y-auto bg-background">
      <div className="mx-auto w-full max-w-[900px] px-6 pb-12 pt-10 xl:px-10">

        {/* 面包屑 */}
        <div className="mb-6">
          <Link
            href={bundlesListHref}
            className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="size-4" />
            权限包列表
          </Link>
        </div>

        {/* 页头 */}
        <div className="mb-8 space-y-1">
          <h1 className="text-2xl font-semibold tracking-tight">
            {loading ? <Skeleton className="inline-block h-7 w-48" /> : (bundle?.name ?? '权限包详情')}
          </h1>
          <p className="text-sm text-muted-foreground">编辑权限包基本信息，管理关联的权限点。</p>
        </div>

        {/* 错误提示 */}
        {error && (
          <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            加载失败：{error.message}
          </div>
        )}

        <div className="space-y-6">
          {/* ── 基本信息卡片 ── */}
          {loading ? (
            <InfoCardSkeleton />
          ) : (
            <Card>
              <CardHeader>
                <CardTitle className="text-base font-semibold">基本信息</CardTitle>
              </CardHeader>
              <CardContent>
                <Form {...form}>
                  <form onSubmit={handleSave} className="space-y-4">
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
                            <Input
                              placeholder="例如：订单管理包"
                              maxLength={50}
                              {...field}
                            />
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

                    <div className="flex justify-end">
                      <Button type="submit" size="sm" disabled={saving}>
                        {saving ? '保存中...' : '保存'}
                      </Button>
                    </div>
                  </form>
                </Form>
              </CardContent>
            </Card>
          )}

          {/* ── 权限点关联卡片 ── */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base font-semibold">已关联权限点</CardTitle>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={loading}
                  onClick={() => setAddDialogOpen(true)}
                >
                  <Plus className="mr-1.5 size-4" />
                  添加权限点
                </Button>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              {loading ? (
                <PermissionsTableSkeleton />
              ) : !bundle || bundle.permissions.length === 0 ? (
                <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-10">
                  <ShieldCheck className="mb-3 size-9 text-muted-foreground/40" />
                  <p className="text-sm font-semibold text-foreground">暂无关联权限点</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    点击「添加权限点」为该权限包关联权限
                  </p>
                </div>
              ) : (
                <div className="overflow-hidden rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>模型</TableHead>
                        <TableHead className="w-[100px]">操作</TableHead>
                        <TableHead className="w-[160px]">行权限范围</TableHead>
                        <TableHead className="w-[80px] text-right">移除</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {bundle.permissions.map((permission) => (
                        <TableRow key={permission.id}>
                          <TableCell className="font-semibold text-foreground">
                            {permission.modelDisplayName}
                          </TableCell>
                          <TableCell>
                            <Badge variant={ACTION_VARIANT[permission.action] ?? 'secondary'}>
                              {ACTION_LABEL[permission.action] ?? permission.action}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">
                              {ROW_SCOPE_LABEL[permission.rowScope] ?? permission.rowScope}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 px-2 text-xs text-destructive hover:bg-destructive/10 hover:text-destructive"
                              disabled={removingId === permission.id}
                              onClick={() => handleRemovePermission(permission)}
                            >
                              <Trash2 className="mr-1 size-3" />
                              {removingId === permission.id ? '移除中...' : '移除'}
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* 添加权限点弹窗 */}
      <AddPermissionDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        allPermissions={allPermissions}
        bundlePermissionIds={bundlePermissionIds}
        onConfirm={handleAddPermissions}
        submitting={addingPermissions}
      />
    </main>
  )
}
