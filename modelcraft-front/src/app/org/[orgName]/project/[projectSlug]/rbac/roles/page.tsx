'use client'

import * as React from 'react'
import { useParams, useRouter } from 'next/navigation'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, ShieldOff } from 'lucide-react'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@web/components/ui/tooltip'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Skeleton } from '@web/components/ui/skeleton'

import { useRoleList } from './_hooks/useRoleList'
import type { EndUserRole } from '@/types'

// ── Validation Schema ────────────────────────────────────────────────────────

const createRoleSchema = z.object({
  name: z
    .string()
    .min(1, '角色名称不能为空')
    .max(64, '角色名称不能超过 64 个字符'),
  description: z
    .string()
    .max(200, '描述不能超过 200 个字符')
    .optional(),
})

type CreateRoleFormValues = z.infer<typeof createRoleSchema>

// ── CreateRoleDialog ─────────────────────────────────────────────────────────

interface CreateRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (values: CreateRoleFormValues) => Promise<void>
  submitting: boolean
}

function CreateRoleDialog({
  open,
  onOpenChange,
  onSubmit,
  submitting,
}: CreateRoleDialogProps) {
  const form = useForm<CreateRoleFormValues>({
    resolver: zodResolver(createRoleSchema),
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
          <DialogTitle>创建角色</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={handleSubmit} className="space-y-4">
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
                      placeholder="例如：编辑员"
                      maxLength={64}
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
                      placeholder="简要描述该角色的用途..."
                      maxLength={200}
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
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

// ── RoleTableSkeleton ────────────────────────────────────────────────────────

function RoleTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[220px]">角色名</TableHead>
            <TableHead>描述</TableHead>
            <TableHead className="w-[100px]">权限包数量</TableHead>
            <TableHead className="w-[140px] text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-36" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-4 w-10" /></TableCell>
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

// ── RoleListPage ─────────────────────────────────────────────────────────────

export default function RoleListPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const projectSlug = params.projectSlug

  const [createDialogOpen, setCreateDialogOpen] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)

  const { roles, loading, error, createRole, deleteRole } = useRoleList({
    orgName,
    projectSlug,
  })

  const handleCreate = React.useCallback(
    async (values: CreateRoleFormValues) => {
      setSubmitting(true)
      try {
        const result = await createRole({
          name: values.name,
          description: values.description,
        })
        if (result.success) {
          toast.success(`角色「${values.name}」已创建`)
          setCreateDialogOpen(false)
        } else {
          toast.error(result.errorMessage ?? '创建失败，请重试')
        }
      } catch {
        toast.error('创建角色时发生错误，请重试')
      } finally {
        setSubmitting(false)
      }
    },
    [createRole]
  )

  const handleDelete = React.useCallback(
    async (role: EndUserRole) => {
      setDeletingId(role.id)
      try {
        const result = await deleteRole(role)
        if (result.success) {
          toast.success(`角色「${role.name}」已删除`)
        } else {
          toast.error(result.errorMessage ?? '删除失败，请重试')
        }
      } catch {
        toast.error('删除角色时发生错误，请重试')
      } finally {
        setDeletingId(null)
      }
    },
    [deleteRole]
  )

  const handleEdit = React.useCallback(
    (role: EndUserRole) => {
      router.push(
        `/org/${orgName}/project/${projectSlug}/rbac/roles/${role.id}`
      )
    },
    [router, orgName, projectSlug]
  )

  return (
    <TooltipProvider>
      <main className="size-full overflow-y-auto bg-background">
        <div className="mx-auto w-full max-w-[1200px] px-6 pb-12 pt-10 xl:px-10">
          {/* Header */}
          <section className="mb-8 flex items-start justify-between gap-4">
            <div className="space-y-1">
              <h1 className="text-2xl font-semibold tracking-tight">角色</h1>
              <p className="text-sm text-muted-foreground">
                管理项目的终端用户角色，每个角色可关联多个权限包，授予用户后生效。
              </p>
            </div>
            <Button
              size="sm"
              onClick={() => setCreateDialogOpen(true)}
              className="shrink-0"
            >
              <Plus className="mr-1.5 size-4" />
              创建角色
            </Button>
          </section>

          {/* Error */}
          {error && (
            <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
              加载角色失败：{error.message}
            </div>
          )}

          {/* Table */}
          {loading ? (
            <RoleTableSkeleton />
          ) : roles.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
              <ShieldOff className="mb-3 size-10 text-muted-foreground/40" />
              <p className="text-sm font-semibold text-foreground">暂无角色</p>
              <p className="mt-1 text-sm text-muted-foreground">
                点击「创建角色」添加第一个角色
              </p>
              <Button
                size="sm"
                variant="outline"
                className="mt-4"
                onClick={() => setCreateDialogOpen(true)}
              >
                <Plus className="mr-1.5 size-4" />
                创建角色
              </Button>
            </div>
          ) : (
            <div className="overflow-hidden rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[220px]">角色名</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead className="w-[100px]">权限包数量</TableHead>
                    <TableHead className="w-[140px] text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {roles.map((role) => (
                    <TableRow
                      key={role.id}
                      className={role.isImplicit ? 'opacity-75' : undefined}
                    >
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-foreground">
                            {role.name}
                          </span>
                          {role.isImplicit && (
                            <Badge variant="secondary" className="text-xs">
                              内置隐式
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {role.description || (
                          <span className="italic text-muted-foreground/50">无描述</span>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {role.permissionBundles.length}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 px-2 text-xs"
                            onClick={() => handleEdit(role)}
                          >
                            <Pencil className="mr-1 size-3" />
                            编辑
                          </Button>

                          {role.isImplicit ? (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span>
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                                    disabled
                                  >
                                    <Trash2 className="mr-1 size-3" />
                                    删除
                                  </Button>
                                </span>
                              </TooltipTrigger>
                              <TooltipContent>
                                内置隐式角色不可删除
                              </TooltipContent>
                            </Tooltip>
                          ) : (
                            <AlertDialog>
                              <AlertDialogTrigger asChild>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                                  disabled={deletingId === role.id}
                                >
                                  <Trash2 className="mr-1 size-3" />
                                  {deletingId === role.id ? '删除中...' : '删除'}
                                </Button>
                              </AlertDialogTrigger>
                              <AlertDialogContent>
                                <AlertDialogHeader>
                                  <AlertDialogTitle>确认删除角色</AlertDialogTitle>
                                  <AlertDialogDescription>
                                    确定要删除角色「{role.name}」吗？
                                    删除后已分配该角色的用户将失去相应权限，此操作不可撤销。
                                  </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                  <AlertDialogCancel>取消</AlertDialogCancel>
                                  <AlertDialogAction
                                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                    onClick={() => handleDelete(role)}
                                  >
                                    确认删除
                                  </AlertDialogAction>
                                </AlertDialogFooter>
                              </AlertDialogContent>
                            </AlertDialog>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>

        <CreateRoleDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          onSubmit={handleCreate}
          submitting={submitting}
        />
      </main>
    </TooltipProvider>
  )
}
