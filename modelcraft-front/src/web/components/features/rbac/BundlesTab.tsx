import * as React from 'react'
import { useRouter } from 'next/navigation'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, PackageOpen } from 'lucide-react'

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

import { useBundleList } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleList'
import type { EndUserPermissionBundle } from '@/types'

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

            <DialogFooter className="pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={submitting}
              >
                取消
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? '创建中...' : '创建'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

// ── BundleTableSkeleton ──────────────────────────────────────────────────────

function BundleTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow className="h-9">
            <TableHead className="w-[200px] text-xs font-semibold text-muted-foreground">名称</TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">描述</TableHead>
            <TableHead className="w-[100px] text-xs font-semibold text-muted-foreground">权限点数量</TableHead>
            <TableHead className="w-[180px] text-xs font-semibold text-muted-foreground">创建时间</TableHead>
            <TableHead className="w-[120px] text-right text-xs font-semibold text-muted-foreground">操作</TableHead>
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
                <div className="flex justify-end gap-2">
                  <Skeleton className="h-7 w-14" />
                  <Skeleton className="h-7 w-14" />
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
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

// ── BundlesTab ───────────────────────────────────────────────────────────────

export function BundlesTab({ orgName, projectSlug }: BundlesTabProps) {
  const router = useRouter()

  const [createDialogOpen, setCreateDialogOpen] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)

  const { bundles, loading, error, createBundle, deleteBundle } = useBundleList({
    orgName,
    projectSlug,
  })

  const handleCreate = React.useCallback(
    async (values: CreateBundleFormValues) => {
      setSubmitting(true)
      try {
        const result = await createBundle({
          name: values.name,
          description: values.description,
        })
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
    [createBundle]
  )

  const handleDelete = React.useCallback(
    async (bundle: EndUserPermissionBundle) => {
      setDeletingId(bundle.id)
      try {
        const result = await deleteBundle(bundle.id)
        if (result.success) {
          toast.success(`权限包「${bundle.name}」已删除`)
        } else {
          toast.error(result.errorMessage ?? '删除失败，请重试')
        }
      } catch {
        toast.error('删除权限包时发生错误，请重试')
      } finally {
        setDeletingId(null)
      }
    },
    [deleteBundle]
  )

  const handleEdit = React.useCallback(
    (bundle: EndUserPermissionBundle) => {
      router.push(
        `/org/${orgName}/project/${projectSlug}/rbac/bundles/${bundle.id}`
      )
    },
    [router, orgName, projectSlug]
  )

  return (
    <div className="space-y-4">
      {/* Header */}
      <section className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold text-foreground">权限包</h2>
          <p className="mt-0.5 text-sm text-muted-foreground">
            管理项目的终端用户权限包，一个权限包包含一组权限点，可授予用户或角色。
          </p>
        </div>
        <Button
          size="sm"
          onClick={() => setCreateDialogOpen(true)}
          className="shrink-0"
        >
          <Plus className="mr-1.5 size-4" />
          创建权限包
        </Button>
      </section>

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
          <PackageOpen className="mb-3 size-10 text-muted-foreground/30" />
          <p className="text-sm font-semibold text-foreground">暂无权限包</p>
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
                <TableHead className="w-[200px] text-xs font-semibold text-muted-foreground">名称</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground">描述</TableHead>
                <TableHead className="w-[100px] text-xs font-semibold text-muted-foreground">权限点数量</TableHead>
                <TableHead className="w-[180px] text-xs font-semibold text-muted-foreground">创建时间</TableHead>
                <TableHead className="w-[120px] text-right text-xs font-semibold text-muted-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {bundles.map((bundle) => (
                <TableRow key={bundle.id}>
                  <TableCell className="font-semibold text-foreground">
                    {bundle.name}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {bundle.description || (
                      <span className="italic text-muted-foreground/50">无描述</span>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {bundle.permissions.length}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDateTime(bundle.createdAt)}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 px-2 text-xs"
                        onClick={() => handleEdit(bundle)}
                      >
                        <Pencil className="mr-1 size-3" />
                        编辑
                      </Button>

                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                            disabled={deletingId === bundle.id}
                          >
                            <Trash2 className="mr-1 size-3" />
                            {deletingId === bundle.id ? '删除中...' : '删除'}
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>确认删除权限包</AlertDialogTitle>
                            <AlertDialogDescription>
                              确定要删除权限包「{bundle.name}」吗？
                              删除后关联角色和用户将失去该权限包内的所有权限，此操作不可撤销。
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              onClick={() => handleDelete(bundle)}
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

      <CreateBundleDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreate}
        submitting={submitting}
      />
    </div>
  )
}
