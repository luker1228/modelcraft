import * as React from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { toast } from 'sonner'
import {
  Plus,
  Trash2,
  Users,
  KeyRound,
  ShieldOff,
  Loader2,
  PackagePlus,
  X,
} from 'lucide-react'

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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@web/components/ui/tooltip'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Skeleton } from '@web/components/ui/skeleton'
import { Checkbox } from '@web/components/ui/checkbox'
import { ScrollArea } from '@web/components/ui/scroll-area'

import { useRoleList } from '@/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList'
import { useRoleEdit } from '@/app/org/[orgName]/project/[projectSlug]/rbac/roles/[roleId]/_hooks/useRoleEdit'
import type { EndUserRole, EndUserPermissionBundle } from '@/types'

// ── Props ────────────────────────────────────────────────────────────────────

export interface RolesTabProps {
  orgName: string
  projectSlug: string
}

// ── Validation ───────────────────────────────────────────────────────────────

const createRoleSchema = z.object({
  name: z.string().min(1, '角色名称不能为空').max(64, '最多 64 个字符'),
  description: z.string().max(200, '最多 200 个字符').optional(),
})

type CreateRoleFormValues = z.infer<typeof createRoleSchema>

// ── CreateRoleDialog ─────────────────────────────────────────────────────────

interface CreateRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (values: CreateRoleFormValues) => Promise<void>
  submitting: boolean
}

function CreateRoleDialog({ open, onOpenChange, onSubmit, submitting }: CreateRoleDialogProps) {
  const form = useForm<CreateRoleFormValues>({
    resolver: zodResolver(createRoleSchema),
    defaultValues: { name: '', description: '' },
  })

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit(values)
    form.reset()
  })

  return (
    <Dialog open={open} onOpenChange={(next) => { if (!next) form.reset(); onOpenChange(next) }}>
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
                  <FormLabel>名称 <span className="text-destructive">*</span></FormLabel>
                  <FormControl>
                    <Input placeholder="例如：编辑员" maxLength={64} {...field} />
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
                    <Textarea placeholder="简要描述该角色的用途..." maxLength={200} rows={3} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>取消</Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? <><Loader2 className="mr-2 size-4 animate-spin" />创建中...</> : '创建'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

// ── UsersSheet ───────────────────────────────────────────────────────────────
// 授权用户 Sheet：查看该角色关联的终端用户，支持添加/移除

const MOCK_ALL_USERS = [
  { id: 'u-1', username: 'alice', createdAt: '2025-01-10T00:00:00Z' },
  { id: 'u-2', username: 'bob',   createdAt: '2025-02-01T00:00:00Z' },
  { id: 'u-3', username: 'carol', createdAt: '2025-03-05T00:00:00Z' },
  { id: 'u-4', username: 'david', createdAt: '2025-04-01T00:00:00Z' },
]

interface UsersSheetProps {
  role: EndUserRole | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function UsersSheet({ role, open, onOpenChange }: UsersSheetProps) {
  // Wave 1 mock：用 state 模拟已分配用户
  const [assignedUserIds, setAssignedUserIds] = React.useState<string[]>([])
  const [addDialogOpen, setAddDialogOpen] = React.useState(false)
  const [selectedToAdd, setSelectedToAdd] = React.useState<string[]>([])

  // 每次角色切换时重置
  React.useEffect(() => {
    setAssignedUserIds([])
  }, [role?.id])

  const assignedUsers = MOCK_ALL_USERS.filter((u) => assignedUserIds.includes(u.id))
  const unassignedUsers = MOCK_ALL_USERS.filter((u) => !assignedUserIds.includes(u.id))

  const handleAddConfirm = () => {
    setAssignedUserIds((prev) => [...prev, ...selectedToAdd])
    setSelectedToAdd([])
    setAddDialogOpen(false)
    toast.success(`已添加 ${selectedToAdd.length} 位用户`)
  }

  const handleRemove = (userId: string, username: string) => {
    setAssignedUserIds((prev) => prev.filter((id) => id !== userId))
    toast.success(`已移除用户 ${username}`)
  }

  if (!role) return null

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="flex w-full flex-col sm:max-w-lg">
          <SheetHeader className="shrink-0">
            <SheetTitle className="flex items-center gap-2">
              <Users className="size-4 text-muted-foreground" />
              授权用户
            </SheetTitle>
            <SheetDescription>
              <span className="font-medium text-foreground">{role.name}</span>
              {role.isImplicit && (
                <Badge variant="secondary" className="ml-2 text-xs">内置隐式</Badge>
              )}
              <span className="ml-1 text-muted-foreground">· 查看并管理持有此角色的终端用户</span>
            </SheetDescription>
          </SheetHeader>

          <div className="mt-4 shrink-0">
            <Button
              size="sm"
              variant="outline"
              onClick={() => { setSelectedToAdd([]); setAddDialogOpen(true) }}
              disabled={unassignedUsers.length === 0}
            >
              <Plus className="mr-1.5 size-4" />
              添加用户
            </Button>
          </div>

          <ScrollArea className="mt-3 flex-1">
            {assignedUsers.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <Users className="mb-3 size-10 text-muted-foreground/30" />
                <p className="text-sm font-semibold text-foreground">暂无授权用户</p>
                <p className="mt-1 text-xs text-muted-foreground">点击「添加用户」将终端用户分配到此角色</p>
              </div>
            ) : (
              <div className="overflow-hidden rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>用户名</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead className="w-16 text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assignedUsers.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-medium text-foreground">{user.username}</TableCell>
                        <TableCell className="text-muted-foreground text-sm">
                          {new Date(user.createdAt).toLocaleDateString('zh-CN')}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
                            onClick={() => handleRemove(user.id, user.username)}
                          >
                            <X className="size-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </ScrollArea>
        </SheetContent>
      </Sheet>

      {/* 添加用户 Dialog */}
      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>添加用户到「{role.name}」</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 py-2">
            {unassignedUsers.length === 0 ? (
              <p className="text-sm text-muted-foreground">所有用户已在此角色中</p>
            ) : (
              unassignedUsers.map((user) => (
                <label
                  key={user.id}
                  className="flex cursor-pointer items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted"
                >
                  <Checkbox
                    checked={selectedToAdd.includes(user.id)}
                    onCheckedChange={(checked) => {
                      setSelectedToAdd((prev) =>
                        checked ? [...prev, user.id] : prev.filter((id) => id !== user.id)
                      )
                    }}
                  />
                  <span className="text-sm text-foreground">{user.username}</span>
                </label>
              ))
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>取消</Button>
            <Button
              onClick={handleAddConfirm}
              disabled={selectedToAdd.length === 0}
            >
              添加 {selectedToAdd.length > 0 && `(${selectedToAdd.length})`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ── BundlesSheet ─────────────────────────────────────────────────────────────
// 权限管理 Sheet：查看并管理该角色关联的权限包

interface BundlesSheetProps {
  role: EndUserRole | null
  orgName: string
  projectSlug: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

function BundlesSheet({ role, orgName, projectSlug, open, onOpenChange }: BundlesSheetProps) {
  const [addBundleDialogOpen, setAddBundleDialogOpen] = React.useState(false)
  const [selectedBundleIds, setSelectedBundleIds] = React.useState<string[]>([])
  const [adding, setAdding] = React.useState(false)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)

  const { role: roleDetail, allBundles, loading, assignBundle, revokeBundle } = useRoleEdit({
    orgName,
    projectSlug,
    roleId: role?.id ?? '',
  })

  const assignedBundleIds = new Set(roleDetail?.permissionBundles.map((b) => b.id) ?? [])
  const unassignedBundles = allBundles.filter((b) => !assignedBundleIds.has(b.id))

  const handleAddConfirm = async () => {
    setAdding(true)
    let failCount = 0
    for (const bundleId of selectedBundleIds) {
      const result = await assignBundle(bundleId)
      if (!result.success) failCount++
    }
    setAdding(false)
    setSelectedBundleIds([])
    setAddBundleDialogOpen(false)
    if (failCount > 0) {
      toast.error(`${failCount} 个权限包添加失败`)
    } else {
      toast.success(`已添加 ${selectedBundleIds.length} 个权限包`)
    }
  }

  const handleRevoke = async (bundle: EndUserPermissionBundle) => {
    setRevokingId(bundle.id)
    const result = await revokeBundle(bundle.id)
    setRevokingId(null)
    if (result.success) {
      toast.success(`已移除权限包「${bundle.name}」`)
    } else {
      toast.error(result.errorMessage ?? '移除失败')
    }
  }

  if (!role) return null

  const assignedBundles = roleDetail?.permissionBundles ?? []

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="flex w-full flex-col sm:max-w-lg">
          <SheetHeader className="shrink-0">
            <SheetTitle className="flex items-center gap-2">
              <KeyRound className="size-4 text-muted-foreground" />
              权限管理
            </SheetTitle>
            <SheetDescription>
              <span className="font-medium text-foreground">{role.name}</span>
              {role.isImplicit && (
                <Badge variant="secondary" className="ml-2 text-xs">内置隐式</Badge>
              )}
              <span className="ml-1 text-muted-foreground">· 管理此角色关联的权限包</span>
            </SheetDescription>
          </SheetHeader>

          <div className="mt-4 shrink-0">
            <Button
              size="sm"
              variant="outline"
              onClick={() => { setSelectedBundleIds([]); setAddBundleDialogOpen(true) }}
              disabled={loading || unassignedBundles.length === 0}
            >
              <PackagePlus className="mr-1.5 size-4" />
              添加权限包
            </Button>
          </div>

          <ScrollArea className="mt-3 flex-1">
            {loading ? (
              <div className="space-y-2">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-14 w-full rounded-md" />
                ))}
              </div>
            ) : assignedBundles.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <KeyRound className="mb-3 size-10 text-muted-foreground/30" />
                <p className="text-sm font-semibold text-foreground">尚未关联权限包</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  点击「添加权限包」为此角色授予权限
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                {assignedBundles.map((bundle) => (
                  <div
                    key={bundle.id}
                    className="flex items-start justify-between gap-3 rounded-md border bg-card px-4 py-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold text-foreground">{bundle.name}</p>
                      {bundle.description && (
                        <p className="mt-0.5 truncate text-xs text-muted-foreground">{bundle.description}</p>
                      )}
                      <p className="mt-1 text-xs text-muted-foreground">
                        {bundle.permissions.length} 个权限点
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 shrink-0 px-2 text-muted-foreground hover:text-destructive"
                      disabled={revokingId === bundle.id}
                      onClick={() => handleRevoke(bundle)}
                    >
                      {revokingId === bundle.id ? (
                        <Loader2 className="size-4 animate-spin" />
                      ) : (
                        <X className="size-4" />
                      )}
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </ScrollArea>
        </SheetContent>
      </Sheet>

      {/* 添加权限包 Dialog */}
      <Dialog open={addBundleDialogOpen} onOpenChange={setAddBundleDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>添加权限包到「{role.name}」</DialogTitle>
          </DialogHeader>
          <ScrollArea className="max-h-60 py-2">
            {unassignedBundles.length === 0 ? (
              <p className="text-sm text-muted-foreground">所有权限包已关联</p>
            ) : (
              <div className="space-y-1">
                {unassignedBundles.map((bundle) => (
                  <label
                    key={bundle.id}
                    className="flex cursor-pointer items-start gap-3 rounded-md px-2 py-2 hover:bg-muted"
                  >
                    <Checkbox
                      className="mt-0.5"
                      checked={selectedBundleIds.includes(bundle.id)}
                      onCheckedChange={(checked) => {
                        setSelectedBundleIds((prev) =>
                          checked ? [...prev, bundle.id] : prev.filter((id) => id !== bundle.id)
                        )
                      }}
                    />
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-semibold text-foreground">{bundle.name}</p>
                      {bundle.description && (
                        <p className="text-xs text-muted-foreground">{bundle.description}</p>
                      )}
                    </div>
                  </label>
                ))}
              </div>
            )}
          </ScrollArea>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddBundleDialogOpen(false)}>取消</Button>
            <Button onClick={handleAddConfirm} disabled={selectedBundleIds.length === 0 || adding}>
              {adding ? <><Loader2 className="mr-2 size-4 animate-spin" />添加中...</> : `添加${selectedBundleIds.length > 0 ? ` (${selectedBundleIds.length})` : ''}`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ── RoleTableSkeleton ────────────────────────────────────────────────────────

function RoleTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[200px]">角色名</TableHead>
            <TableHead>描述</TableHead>
            <TableHead className="w-[80px]">权限包</TableHead>
            <TableHead className="w-[200px] text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {[1, 2, 3].map((i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-4 w-8" /></TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Skeleton className="h-7 w-20" />
                  <Skeleton className="h-7 w-20" />
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// ── RolesTab (主组件) ────────────────────────────────────────────────────────

export function RolesTab({ orgName, projectSlug }: RolesTabProps) {
  const [createDialogOpen, setCreateDialogOpen] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)

  // Sheet 状态
  const [usersSheetRole, setUsersSheetRole] = React.useState<EndUserRole | null>(null)
  const [bundlesSheetRole, setBundlesSheetRole] = React.useState<EndUserRole | null>(null)

  const { roles, loading, error, createRole, deleteRole } = useRoleList({ orgName, projectSlug })

  const handleCreate = React.useCallback(async (values: CreateRoleFormValues) => {
    setSubmitting(true)
    try {
      const result = await createRole({ name: values.name, description: values.description })
      if (result.success) {
        toast.success(`角色「${values.name}」已创建`)
        setCreateDialogOpen(false)
      } else {
        toast.error(result.errorMessage ?? '创建失败，请重试')
      }
    } catch {
      toast.error('创建角色时发生错误')
    } finally {
      setSubmitting(false)
    }
  }, [createRole])

  const handleDelete = React.useCallback(async (role: EndUserRole) => {
    setDeletingId(role.id)
    try {
      const result = await deleteRole(role)
      if (result.success) {
        toast.success(`角色「${role.name}」已删除`)
      } else {
        toast.error(result.errorMessage ?? '删除失败，请重试')
      }
    } catch {
      toast.error('删除角色时发生错误')
    } finally {
      setDeletingId(null)
    }
  }, [deleteRole])

  return (
    <TooltipProvider>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold tracking-tight text-foreground">终端用户角色</h2>
            <p className="text-sm text-muted-foreground">
              为每个角色配置权限包，再将角色分配给终端用户。
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateDialogOpen(true)} className="shrink-0">
            <Plus className="mr-1.5 size-4" />
            创建角色
          </Button>
        </div>

        {/* Error */}
        {error && (
          <div className="rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
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
            <p className="mt-1 text-sm text-muted-foreground">点击「创建角色」开始配置权限</p>
            <Button size="sm" variant="outline" className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-1.5 size-4" />
              创建角色
            </Button>
          </div>
        ) : (
          <div className="overflow-hidden rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[200px]">角色名</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead className="w-[80px]">权限包</TableHead>
                  <TableHead className="w-[220px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {roles.map((role) => (
                  <TableRow key={role.id} className={role.isImplicit ? 'opacity-75' : undefined}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground">{role.name}</span>
                        {role.isImplicit && (
                          <Badge variant="secondary" className="text-xs">内置隐式</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {role.description || <span className="italic text-muted-foreground/40">无描述</span>}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {role.permissionBundles.length}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        {/* 授权用户 */}
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 gap-1.5 px-2.5 text-xs"
                          onClick={() => setUsersSheetRole(role)}
                        >
                          <Users className="size-3.5" />
                          授权用户
                        </Button>

                        {/* 权限管理 */}
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 gap-1.5 px-2.5 text-xs"
                          onClick={() => setBundlesSheetRole(role)}
                        >
                          <KeyRound className="size-3.5" />
                          权限管理
                        </Button>

                        {/* 删除 */}
                        {role.isImplicit ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-7 w-7 p-0 text-muted-foreground"
                                  disabled
                                >
                                  <Trash2 className="size-3.5" />
                                </Button>
                              </span>
                            </TooltipTrigger>
                            <TooltipContent>内置隐式角色不可删除</TooltipContent>
                          </Tooltip>
                        ) : (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
                                disabled={deletingId === role.id}
                              >
                                {deletingId === role.id ? (
                                  <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                  <Trash2 className="size-3.5" />
                                )}
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

        {/* Dialogs & Sheets */}
        <CreateRoleDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          onSubmit={handleCreate}
          submitting={submitting}
        />

        <UsersSheet
          role={usersSheetRole}
          open={!!usersSheetRole}
          onOpenChange={(open) => { if (!open) setUsersSheetRole(null) }}
        />

        <BundlesSheet
          role={bundlesSheetRole}
          orgName={orgName}
          projectSlug={projectSlug}
          open={!!bundlesSheetRole}
          onOpenChange={(open) => { if (!open) setBundlesSheetRole(null) }}
        />
      </div>
    </TooltipProvider>
  )
}
