'use client'

import { useMemo, useState, useCallback } from 'react'
import { useParams, useRouter, useSearchParams } from 'next/navigation'
import { Users, Plus, Trash2, Search, KeyRound, X, PackagePlus, Loader2, ShieldOff, UserPlus } from 'lucide-react'
import { toast } from 'sonner'
import { useQuery, useMutation } from '@apollo/client'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
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
} from '@web/components/ui/alert-dialog'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@web/components/ui/sheet'
import {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@web/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { ScrollArea } from '@web/components/ui/scroll-area'
import { Checkbox } from '@web/components/ui/checkbox'
import { Skeleton } from '@web/components/ui/skeleton'
import { usePermission } from '@web/hooks/auth/use-permission'
import { useRoleEdit } from '@/app/org/[orgName]/project/[projectSlug]/rbac/roles/[roleId]/_hooks/useRoleEdit'
import { useRoleList } from '@/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList'
import {
  BundlesTab,
  PermissionsTab,
} from '@web/components/features/rbac'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { useProjectScopedClient } from '@bff/apollo/public'
import {
  LIST_END_USERS,
  GET_END_USER_ROLE_ASSIGNMENTS,
  ASSIGN_END_USER_ROLE_TO_USER,
  REVOKE_END_USER_ROLE_FROM_USER,
} from '@web/graphql'
import type { EndUserRole } from '@/types'

// ── Types ────────────────────────────────────────────────────────────

type TabValue = 'roles' | 'bundles' | 'permissions'

// ── LegacyBundlesSheet ───────────────────────────────────────────────

interface LegacyBundlesSheetProps {
  roleId: string
  roleName: string
  orgName: string
  projectSlug: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

function LegacyBundlesSheet({ roleId, roleName, orgName, projectSlug, open, onOpenChange }: LegacyBundlesSheetProps) {
  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [selectedBundleIds, setSelectedBundleIds] = useState<string[]>([])
  const [adding, setAdding] = useState(false)
  const [revokingId, setRevokingId] = useState<string | null>(null)

  const { role, allBundles, loading, assignBundle, revokeBundle } = useRoleEdit({ orgName, projectSlug, roleId })

  const assignedBundleIds = new Set(role?.permissionBundles.map((b) => b.id) ?? [])
  const assignedBundles = role?.permissionBundles ?? []
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
    setAddDialogOpen(false)
    if (failCount > 0) toast.error(`${failCount} 个权限包添加失败`)
    else toast.success(`已添加 ${selectedBundleIds.length} 个权限包`)
  }

  const handleRevoke = async (bundleId: string, bundleName: string) => {
    setRevokingId(bundleId)
    const result = await revokeBundle(bundleId)
    setRevokingId(null)
    if (result.success) toast.success(`已移除权限包「${bundleName}」`)
    else toast.error(result.errorMessage ?? '移除失败')
  }

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="flex w-full flex-col sm:max-w-lg">
          <SheetHeader className="shrink-0">
            <SheetTitle className="flex items-center gap-2">
              <KeyRound className="size-4 text-muted-foreground" strokeWidth={1.5} />
              {roleName} · 权限管理
            </SheetTitle>
            <SheetDescription>管理此角色关联的权限包</SheetDescription>
          </SheetHeader>

          <div className="mt-4 shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => { setSelectedBundleIds([]); setAddDialogOpen(true) }}
              disabled={loading || unassignedBundles.length === 0}
            >
              <PackagePlus className="size-4" strokeWidth={1.5} />
              添加权限包
            </Button>
          </div>

          <ScrollArea className="mt-3 flex-1">
            {loading ? (
              <div className="space-y-2">
                {[1, 2, 3].map((i) => <Skeleton key={i} className="h-14 w-full rounded-md" />)}
              </div>
            ) : assignedBundles.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <KeyRound className="mb-3 size-10 text-muted-foreground/30" strokeWidth={1} />
                <p className="text-sm font-semibold text-foreground">尚未关联权限包</p>
                <p className="mt-1 text-xs text-muted-foreground">点击「添加权限包」为此角色授予权限</p>
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
                      <p className="mt-1 text-xs text-muted-foreground">{bundle.permissions.length} 个权限点</p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="mt-0.5 size-7 shrink-0 text-muted-foreground hover:text-destructive"
                      disabled={revokingId === bundle.id}
                      onClick={() => handleRevoke(bundle.id, bundle.name)}
                    >
                      {revokingId === bundle.id
                        ? <Loader2 className="size-4 animate-spin" />
                        : <X className="size-4" />
                      }
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </ScrollArea>
        </SheetContent>
      </Sheet>

      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>添加权限包到「{roleName}」</DialogTitle>
          </DialogHeader>
          <ScrollArea className="max-h-60 py-2">
            {unassignedBundles.length === 0 ? (
              <p className="text-sm text-muted-foreground">所有权限包已关联</p>
            ) : (
              <div className="space-y-1">
                {unassignedBundles.map((bundle) => (
                  <label
                    key={bundle.id}
                    className="flex cursor-pointer items-start gap-3 rounded-md p-2 hover:bg-muted"
                  >
                    <Checkbox
                      className="mt-0.5"
                      checked={selectedBundleIds.includes(bundle.id)}
                      onCheckedChange={(checked) =>
                        setSelectedBundleIds((prev) =>
                          checked ? [...prev, bundle.id] : prev.filter((id) => id !== bundle.id)
                        )
                      }
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
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>取消</Button>
            <Button onClick={handleAddConfirm} disabled={selectedBundleIds.length === 0 || adding} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {adding
                ? <><Loader2 className="mr-2 size-4 animate-spin" />添加中...</>
                : `添加${selectedBundleIds.length > 0 ? ` (${selectedBundleIds.length})` : ''}`
              }
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ── RoleUsersSheet ────────────────────────────────────────────────────

interface RoleUsersSheetProps {
  role: EndUserRole
  orgName: string
  projectSlug: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

function RoleUsersSheet({ role, orgName, projectSlug, open, onOpenChange }: RoleUsersSheetProps) {
  const canManageRoles = usePermission('*')
  const client = useProjectScopedClient(projectSlug, orgName)

  const [search, setSearch] = useState('')
  const [assigningId, setAssigningId] = useState<string | null>(null)
  const [revokingId, setRevokingId] = useState<string | null>(null)

  // 获取所有 EndUser
  const { data: usersData, loading: usersLoading } = useQuery<{
    listEndUsers?: { connection?: { nodes?: EndUserNode[] } }
  }>(LIST_END_USERS, {
    client,
    variables: { input: { first: 100 } },
    skip: !open,
  })

  type EndUserNode = { id: string; username: string; isForbidden: boolean }

  const filteredUsers = useMemo(() => {
    const nodes = (usersData as { listEndUsers?: { connection?: { nodes?: EndUserNode[] } } } | undefined)
      ?.listEndUsers?.connection?.nodes ?? []
    return nodes.filter((u) =>
      u.username.toLowerCase().includes(search.toLowerCase())
    )
  }, [usersData, search])

  const [assignedUserIds, setAssignedUserIds] = useState<Set<string>>(new Set())

  // 使用 GET_END_USER_ROLE_ASSIGNMENTS 需要逐个用户查，不实用
  // 改为在分配/撤销时本地维护已分配状态（乐观更新），初始从 API 获取用户的角色列表

  type AssignPayload = { assignEndUserRole?: { error?: { message?: string } | null } }
  type RevokePayload = { revokeEndUserRole?: { error?: { message?: string } | null; success?: boolean } }

  const [assignRoleMutation] = useMutation(ASSIGN_END_USER_ROLE_TO_USER, { client })
  const [revokeRoleMutation] = useMutation(REVOKE_END_USER_ROLE_FROM_USER, { client })

  const handleAssign = useCallback(async (userId: string) => {
    setAssigningId(userId)
    try {
      const result = await assignRoleMutation({
        variables: { input: { endUserId: userId, roleId: role.id } },
      })
      const payload = (result.data as AssignPayload | undefined)?.assignEndUserRole
      if (payload?.error) {
        toast.error(payload.error.message ?? '分配失败')
      } else {
        setAssignedUserIds((prev) => new Set([...prev, userId]))
        toast.success('角色已分配')
      }
    } catch {
      toast.error('分配失败')
    } finally {
      setAssigningId(null)
    }
  }, [assignRoleMutation, role.id])

  const handleRevoke = useCallback(async (userId: string, username: string) => {
    setRevokingId(userId)
    try {
      const result = await revokeRoleMutation({
        variables: { input: { endUserId: userId, roleId: role.id } },
      })
      const payload = (result.data as RevokePayload | undefined)?.revokeEndUserRole
      if (payload?.error) {
        toast.error(payload.error.message ?? '撤销失败')
      } else {
        setAssignedUserIds((prev) => {
          const next = new Set(prev)
          next.delete(userId)
          return next
        })
        toast.success(`已撤销用户「${username}」的角色`)
      }
    } catch {
      toast.error('撤销失败')
    } finally {
      setRevokingId(null)
    }
  }, [revokeRoleMutation, role.id])

  return (
    <Sheet open={open} onOpenChange={(v) => { if (!v) setSearch(''); onOpenChange(v) }}>
      <SheetContent className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader className="shrink-0">
          <SheetTitle className="flex items-center gap-2">
            <Users className="size-4 text-muted-foreground" strokeWidth={1.5} />
            {role.name} · 用户管理
            {role.isImplicit && <Badge variant="secondary">内置</Badge>}
          </SheetTitle>
          <SheetDescription>{role.description || '为终端用户分配或撤销此角色'}</SheetDescription>
        </SheetHeader>

        <div className="relative mt-4 shrink-0">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
          <Input
            placeholder="搜索用户..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        <ScrollArea className="mt-3 flex-1">
          {usersLoading ? (
            <div className="space-y-2">
              {[1, 2, 3, 4].map((i) => <Skeleton key={i} className="h-12 w-full rounded-md" />)}
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Users className="mb-3 size-10 text-muted-foreground/30" strokeWidth={1} />
              <p className="text-sm font-semibold text-foreground">
                {search ? '未找到匹配用户' : '暂无终端用户'}
              </p>
            </div>
          ) : (
            <div className="space-y-1">
              {filteredUsers.map((user) => {
                const assigned = assignedUserIds.has(user.id)
                const isAssigning = assigningId === user.id
                const isRevoking = revokingId === user.id
                return (
                  <div
                    key={user.id}
                    className="flex items-center justify-between gap-3 rounded-md border bg-card px-4 py-3"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold text-foreground">{user.username}</p>
                      {user.isForbidden && (
                        <p className="mt-0.5 text-xs text-destructive">已禁用</p>
                      )}
                    </div>
                    {canManageRoles && !role.isImplicit && (
                      assigned ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 shrink-0 gap-1 text-xs text-muted-foreground hover:text-destructive"
                          disabled={isRevoking}
                          onClick={() => handleRevoke(user.id, user.username)}
                        >
                          {isRevoking
                            ? <Loader2 className="size-3.5 animate-spin" />
                            : <X className="size-3.5" />
                          }
                          撤销
                        </Button>
                      ) : (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 shrink-0 gap-1 text-xs text-muted-foreground hover:text-foreground"
                          disabled={isAssigning || user.isForbidden}
                          onClick={() => handleAssign(user.id)}
                        >
                          {isAssigning
                            ? <Loader2 className="size-3.5 animate-spin" />
                            : <UserPlus className="size-3.5" />
                          }
                          分配
                        </Button>
                      )
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

// ── RolesContent ─────────────────────────────────────────────────────

interface RolesContentProps {
  orgName: string
  projectSlug: string
}

function RolesContent({ orgName, projectSlug }: RolesContentProps) {
  const canManageRoles = usePermission('*')

  const { roles, loading, createRole, deleteRole } = useRoleList({ orgName, projectSlug })

  const [search, setSearch] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDescription, setNewDescription] = useState('')
  const [creating, setCreating] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<EndUserRole | null>(null)
  const [deleting, setDeleting] = useState(false)

  const [usersSheetRole, setUsersSheetRole] = useState<EndUserRole | null>(null)
  const [bundlesSheetRole, setBundlesSheetRole] = useState<EndUserRole | null>(null)

  // ── Derived ───────────────────────────────────────────────────────

  const filteredRoles = useMemo(
    () => roles.filter(
      (r) =>
        r.name.toLowerCase().includes(search.toLowerCase()) ||
        (r.description ?? '').toLowerCase().includes(search.toLowerCase())
    ),
    [roles, search]
  )

  // ── Handlers ─────────────────────────────────────────────────────

  const handleCreate = async () => {
    if (!newName.trim()) { toast.error('请输入角色名称'); return }
    setCreating(true)
    const result = await createRole({ name: newName.trim(), description: newDescription.trim() || undefined })
    setCreating(false)
    if (result.success) {
      toast.success('角色已创建')
      setCreateOpen(false)
      setNewName('')
      setNewDescription('')
    } else {
      toast.error(result.errorMessage ?? '创建角色失败')
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    const result = await deleteRole(deleteTarget)
    setDeleting(false)
    if (result.success) {
      toast.success(`已删除角色「${deleteTarget.name}」`)
      setDeleteTarget(null)
    } else {
      toast.error(result.errorMessage ?? '删除失败')
    }
  }

  if (loading) {
    return (
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <Skeleton className="h-9 w-48" />
          <Skeleton className="h-9 w-24" />
        </div>
        <div className="overflow-hidden rounded-lg border border-border">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-12 w-full border-b last:border-0" />)}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="relative max-w-xs flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
          <Input
            placeholder="搜索角色..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button onClick={() => setCreateOpen(true)} disabled={!canManageRoles} size="sm" className="bg-primary text-primary-foreground hover:bg-primary/90">
          <Plus className="size-4" strokeWidth={1.5} />
          新建角色
        </Button>
      </div>

      {/* Table */}
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow className="border-b-2 border-border bg-card hover:bg-card">
              <TableHead className="h-10 w-[200px] text-[11px] font-medium uppercase tracking-wider text-foreground">角色名称</TableHead>
              <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">描述</TableHead>
              <TableHead className="h-10 w-[220px] text-right text-[11px] font-medium uppercase tracking-wider text-foreground">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredRoles.map((role) => (
              <TableRow key={role.id} className="group border-b border-border last:border-0 hover:bg-foreground/[0.015]">
                <TableCell className="h-12 text-[13px]">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-foreground">{role.name}</span>
                    {role.isImplicit && (
                      <Badge variant="secondary" className="text-xs">内置</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="h-12 text-[13px] text-muted-foreground">
                  {role.description || <span className="text-muted-foreground/40">—</span>}
                </TableCell>
                <TableCell className="h-12 text-right">
                  <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-foreground"
                      onClick={() => setUsersSheetRole(role)}
                    >
                      <Users className="size-3.5" strokeWidth={1.5} />
                      用户管理
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-foreground"
                      onClick={() => setBundlesSheetRole(role)}
                    >
                      <KeyRound className="size-3.5" strokeWidth={1.5} />
                      权限管理
                    </Button>
                    {!role.isImplicit && canManageRoles && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTarget(role)}
                      >
                        <Trash2 className="size-3.5" strokeWidth={1.5} />
                        删除
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        {filteredRoles.length === 0 && (
          <div className="flex flex-col items-center justify-center py-14 text-center">
            <ShieldOff className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm font-semibold text-foreground">
              {search ? '未找到匹配的角色' : '暂无角色'}
            </p>
            {!search && (
              <p className="mt-1 text-xs text-muted-foreground">
                点击「新建角色」开始配置权限
              </p>
            )}
            {!search && canManageRoles && (
              <Button variant="outline" size="sm" className="mt-4" onClick={() => setCreateOpen(true)}>
                <Plus className="size-4" strokeWidth={1.5} />
                新建角色
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>新建角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">
                角色名称 <span className="text-destructive">*</span>
              </label>
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="例如：Editor"
                onKeyDown={(e) => e.key === 'Enter' && !creating && handleCreate()}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">描述</label>
              <Textarea
                value={newDescription}
                onChange={(e) => setNewDescription(e.target.value)}
                placeholder="可选，描述该角色的用途"
                rows={2}
                className="resize-none"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
            <Button onClick={handleCreate} disabled={creating} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {creating ? <><Loader2 className="mr-2 size-4 animate-spin" />创建中...</> : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Alert */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除角色</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除角色「{deleteTarget?.name}」吗？此操作不可撤销，该角色的所有用户将失去对应权限。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={handleDeleteConfirm}
              disabled={deleting}
            >
              {deleting ? <><Loader2 className="mr-2 size-4 animate-spin" />删除中...</> : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Users Sheet */}
      {usersSheetRole && (
        <RoleUsersSheet
          role={usersSheetRole}
          orgName={orgName}
          projectSlug={projectSlug}
          open={!!usersSheetRole}
          onOpenChange={(open) => { if (!open) setUsersSheetRole(null) }}
        />
      )}

      {/* Bundles Sheet */}
      {bundlesSheetRole && (
        <LegacyBundlesSheet
          roleId={bundlesSheetRole.id}
          roleName={bundlesSheetRole.name}
          orgName={orgName}
          projectSlug={projectSlug}
          open={!!bundlesSheetRole}
          onOpenChange={(open) => { if (!open) setBundlesSheetRole(null) }}
        />
      )}
    </div>
  )
}

// ── RolesPage ─────────────────────────────────────────────────────────────────

export default function RolesPage() {
  const params = useParams()
  const router = useRouter()
  const searchParams = useSearchParams()

  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const rawTab = searchParams.get('tab') as TabValue | null
  const VALID_TABS: TabValue[] = ['roles', 'bundles', 'permissions']
  const activeTab: TabValue = rawTab && VALID_TABS.includes(rawTab) ? rawTab : 'roles'

  const handleTabChange = (value: string) => {
    const url = new URL(window.location.href)
    url.searchParams.set('tab', value)
    router.replace(url.pathname + url.search)
  }

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title="权限管理" />

      {/* Tab navigation — underline style */}
      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="h-auto w-full justify-start gap-0 rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            value="roles"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:bg-transparent hover:text-foreground aria-selected:border-primary aria-selected:bg-transparent aria-selected:text-primary aria-selected:shadow-none"
          >
            角色
          </TabsTrigger>
          <TabsTrigger
            value="bundles"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:bg-transparent hover:text-foreground aria-selected:border-primary aria-selected:bg-transparent aria-selected:text-primary aria-selected:shadow-none"
          >
            权限包
          </TabsTrigger>
          <TabsTrigger
            value="permissions"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:bg-transparent hover:text-foreground aria-selected:border-primary aria-selected:bg-transparent aria-selected:text-primary aria-selected:shadow-none"
          >
            权限点
          </TabsTrigger>
        </TabsList>

        <TabsContent value="roles" className="mt-6">
          <RolesContent orgName={orgName} projectSlug={projectSlug} />
        </TabsContent>

        <TabsContent value="bundles" className="mt-6">
          <BundlesTab orgName={orgName} projectSlug={projectSlug} />
        </TabsContent>

        <TabsContent value="permissions" className="mt-6">
          <PermissionsTab orgName={orgName} projectSlug={projectSlug} />
        </TabsContent>
      </Tabs>
    </PageLayout>
  )
}
