'use client'

import * as React from 'react'
import { useParams, useRouter } from 'next/navigation'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import { ArrowLeft, Plus, Trash2, PackageOpen } from 'lucide-react'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import { Checkbox } from '@web/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
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
import { Skeleton } from '@web/components/ui/skeleton'

import { useRoleEdit } from './_hooks/useRoleEdit'
import type { EndUserPermissionBundle } from '@/types'

// ── Validation Schema ────────────────────────────────────────────────────────

const roleInfoSchema = z.object({
  name: z
    .string()
    .min(1, '角色名称不能为空')
    .max(64, '角色名称不能超过 64 个字符'),
  description: z
    .string()
    .max(200, '描述不能超过 200 个字符')
    .optional(),
})

type RoleInfoFormValues = z.infer<typeof roleInfoSchema>

// ── AddBundleDialog ──────────────────────────────────────────────────────────

interface AddBundleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  availableBundles: EndUserPermissionBundle[]
  onConfirm: (bundleIds: string[]) => Promise<void>
  submitting: boolean
}

function AddBundleDialog({
  open,
  onOpenChange,
  availableBundles,
  onConfirm,
  submitting,
}: AddBundleDialogProps) {
  const [selected, setSelected] = React.useState<Set<string>>(new Set())

  const handleToggle = (id: string) => {
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

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) setSelected(new Set())
    onOpenChange(nextOpen)
  }

  const handleConfirm = async () => {
    await onConfirm(Array.from(selected))
    setSelected(new Set())
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>添加权限包</DialogTitle>
        </DialogHeader>

        {availableBundles.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10">
            <PackageOpen className="mb-3 size-8 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">所有权限包已关联，暂无可添加项</p>
          </div>
        ) : (
          <div className="max-h-72 overflow-y-auto rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-10" />
                  <TableHead>权限包名称</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead className="w-[80px]">权限点</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {availableBundles.map((bundle) => (
                  <TableRow
                    key={bundle.id}
                    className="cursor-pointer"
                    onClick={() => handleToggle(bundle.id)}
                  >
                    <TableCell>
                      <Checkbox
                        checked={selected.has(bundle.id)}
                        onCheckedChange={() => handleToggle(bundle.id)}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </TableCell>
                    <TableCell className="font-semibold text-foreground">
                      {bundle.name}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {bundle.description || (
                        <span className="italic text-muted-foreground/50">无描述</span>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {bundle.permissions.length}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        <DialogFooter>
          <Button
            type="button"
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
            {submitting ? '添加中...' : `确认（已选 ${selected.size} 项）`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── PageSkeleton ─────────────────────────────────────────────────────────────

function PageSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-24" />
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-9 w-24" />
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    </div>
  )
}

// ── RoleEditPage ─────────────────────────────────────────────────────────────

export default function RoleEditPage() {
  const params = useParams<{ orgName: string; projectSlug: string; roleId: string }>()
  const router = useRouter()
  const { orgName, projectSlug, roleId } = params

  const [addBundleDialogOpen, setAddBundleDialogOpen] = React.useState(false)
  const [addingBundles, setAddingBundles] = React.useState(false)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)

  const { role, allBundles, loading, error, assignBundle, revokeBundle } = useRoleEdit({
    orgName,
    projectSlug,
    roleId,
  })

  const form = useForm<RoleInfoFormValues>({
    resolver: zodResolver(roleInfoSchema),
    defaultValues: { name: '', description: '' },
  })

  // Sync form when role data loads
  React.useEffect(() => {
    if (role) {
      form.reset({
        name: role.name,
        description: role.description ?? '',
      })
    }
  }, [role, form])

  // Bundles not yet associated
  const assignedBundleIds = React.useMemo(
    () => new Set(role?.permissionBundles.map((b) => b.id) ?? []),
    [role]
  )
  const availableBundles = React.useMemo(
    () => allBundles.filter((b) => !assignedBundleIds.has(b.id)),
    [allBundles, assignedBundleIds]
  )

  const handleSaveInfo = form.handleSubmit(async (_values) => {
    // NOTE: UPDATE_END_USER_ROLE mutation not yet in spec — placeholder
    toast.info('角色信息保存功能待后端实现')
  })

  const handleAddBundles = React.useCallback(
    async (bundleIds: string[]) => {
      setAddingBundles(true)
      try {
        const results = await Promise.all(bundleIds.map((id) => assignBundle(id)))
        const failed = results.filter((r) => !r.success)
        if (failed.length === 0) {
          toast.success(`已成功添加 ${bundleIds.length} 个权限包`)
          setAddBundleDialogOpen(false)
        } else {
          toast.error(failed[0]?.errorMessage ?? '部分权限包添加失败，请重试')
        }
      } catch {
        toast.error('添加权限包时发生错误，请重试')
      } finally {
        setAddingBundles(false)
      }
    },
    [assignBundle]
  )

  const handleRevokeBundle = React.useCallback(
    async (bundle: EndUserPermissionBundle) => {
      setRevokingId(bundle.id)
      try {
        const result = await revokeBundle(bundle.id)
        if (result.success) {
          toast.success(`已移除权限包「${bundle.name}」`)
        } else {
          toast.error(result.errorMessage ?? '移除失败，请重试')
        }
      } catch {
        toast.error('移除权限包时发生错误，请重试')
      } finally {
        setRevokingId(null)
      }
    },
    [revokeBundle]
  )

  return (
    <main className="size-full overflow-y-auto bg-background">
      <div className="mx-auto w-full max-w-[900px] px-6 pb-12 pt-10 xl:px-10">
        {/* Breadcrumb */}
        <Button
          variant="ghost"
          size="sm"
          className="-ml-2 mb-6 gap-1.5 text-muted-foreground hover:text-foreground"
          onClick={() =>
            router.push(`/org/${orgName}/project/${projectSlug}/rbac/roles`)
          }
        >
          <ArrowLeft className="size-4" />
          角色列表
        </Button>

        {/* Error */}
        {error && (
          <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            加载角色详情失败：{error.message}
          </div>
        )}

        {loading || !role ? (
          <PageSkeleton />
        ) : (
          <div className="space-y-6">
            {/* Page title */}
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-semibold tracking-tight">{role.name}</h1>
              {role.isImplicit && (
                <Badge variant="secondary">内置隐式</Badge>
              )}
            </div>

            {/* isImplicit banner */}
            {role.isImplicit && (
              <div className="rounded bg-muted p-3 text-sm text-muted-foreground">
                内置隐式角色由系统管理，基本信息不可修改，但可以调整关联的权限包。
              </div>
            )}

            {/* Basic Info Card */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">基本信息</CardTitle>
              </CardHeader>
              <CardContent>
                <Form {...form}>
                  <form onSubmit={handleSaveInfo} className="space-y-4">
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
                              placeholder="角色名称"
                              maxLength={64}
                              disabled={role.isImplicit}
                              {...field}
                            />
                          </FormControl>
                          {role.isImplicit && (
                            <p className="text-xs text-muted-foreground">
                              内置隐式角色，名称不可修改
                            </p>
                          )}
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="description"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>描述</FormLabel>
                          <FormControl>
                            <Textarea
                              placeholder="简要描述该角色的用途..."
                              maxLength={200}
                              rows={3}
                              disabled={role.isImplicit}
                              {...field}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <div className="flex justify-end">
                      <Button
                        type="submit"
                        size="sm"
                        disabled={role.isImplicit}
                      >
                        保存
                      </Button>
                    </div>
                  </form>
                </Form>
              </CardContent>
            </Card>

            {/* Permission Bundles Card */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">已关联权限包</CardTitle>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setAddBundleDialogOpen(true)}
                  >
                    <Plus className="mr-1.5 size-4" />
                    添加权限包
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {role.permissionBundles.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12">
                    <PackageOpen className="mb-3 size-8 text-muted-foreground/30" />
                    <p className="text-[13px] text-muted-foreground">暂未关联任何权限包</p>
                    <Button size="sm" className="mt-4" onClick={() => setAddBundleDialogOpen(true)}>
                      <Plus className="mr-1.5 size-4" />
                      添加权限包
                    </Button>
                  </div>
                ) : (
                  <div className="overflow-hidden rounded-lg border border-border bg-card">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[200px]">名称</TableHead>
                          <TableHead>描述</TableHead>
                          <TableHead className="w-[80px]">权限点数量</TableHead>
                          <TableHead className="w-[80px] text-right">操作</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {role.permissionBundles.map((bundle) => (
                          <TableRow key={bundle.id}>
                            <TableCell className="font-semibold text-foreground">
                              {bundle.name}
                            </TableCell>
                            <TableCell className="text-muted-foreground">
                              {bundle.description || (
                                <span className="italic text-muted-foreground/50">无描述</span>
                              )}
                            </TableCell>
                            <TableCell className="text-muted-foreground">
                              {bundle.permissions.length}
                            </TableCell>
                            <TableCell className="text-right">
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                                disabled={revokingId === bundle.id}
                                onClick={() => handleRevokeBundle(bundle)}
                              >
                                <Trash2 className="mr-1 size-3" />
                                {revokingId === bundle.id ? '移除中...' : '移除'}
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
        )}
      </div>

      <AddBundleDialog
        open={addBundleDialogOpen}
        onOpenChange={setAddBundleDialogOpen}
        availableBundles={availableBundles}
        onConfirm={handleAddBundles}
        submitting={addingBundles}
      />
    </main>
  )
}
