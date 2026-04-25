'use client'

import { useMemo, useState } from 'react'
import { useParams, useRouter, useSearchParams } from 'next/navigation'
import { Users, Plus, Trash2, Search, KeyRound, X, PackagePlus, Loader2, ShieldOff } from 'lucide-react'
import { toast } from 'sonner'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { usePermission } from '@web/hooks/auth/use-permission'
import { useRoleEdit } from '@/app/org/[orgName]/project/[projectSlug]/rbac/roles/[roleId]/_hooks/useRoleEdit'
import {
  BundlesTab,
  PermissionsTab,
} from '@web/components/features/rbac'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import type { Role } from '@/types'

// ── Types ────────────────────────────────────────────────────────────

interface MockUser {
  id: string
  name: string
  email: string
}

interface RoleMock extends Role {
  userIds: string[]
}

type TabValue = 'roles' | 'bundles' | 'permissions'

// ── Constants ────────────────────────────────────────────────────────

const MOCK_USERS: MockUser[] = [
  { id: 'u-1', name: 'Alice', email: 'alice@demo.io' },
  { id: 'u-2', name: 'Bob', email: 'bob@demo.io' },
  { id: 'u-3', name: 'Carol', email: 'carol@demo.io' },
  { id: 'u-4', name: 'David', email: 'david@demo.io' },
]

function nowISO(): string {
  return new Date().toISOString()
}

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
                    className="flex cursor-pointer items-start gap-3 rounded-md px-2 py-2 hover:bg-muted"
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

// ── LegacyRolesContent ───────────────────────────────────────────────

interface LegacyRolesContentProps {
  orgName: string
  projectSlug: string
}

function LegacyRolesContent({ orgName, projectSlug }: LegacyRolesContentProps) {
  const canManageRoles = usePermission('*')

  const [roles, setRoles] = useState<RoleMock[]>([
    {
      id: 'r-admin',
      name: 'Admin',
      description: '系统管理员，拥有全部权限',
      permissions: [],
      isSystem: true,
      createdAt: nowISO(),
      updatedAt: nowISO(),
      userIds: ['u-1'],
    },
    {
      id: 'r-ops',
      name: 'Ops',
      description: '运营角色，只读权限',
      permissions: [],
      isSystem: false,
      createdAt: nowISO(),
      updatedAt: nowISO(),
      userIds: ['u-2', 'u-3'],
    },
  ])

  const [search, setSearch] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDescription, setNewDescription] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<RoleMock | null>(null)

  const [sheetRole, setSheetRole] = useState<RoleMock | null>(null)
  const [userToAssign, setUserToAssign] = useState('')

  const [bundlesSheetRoleId, setBundlesSheetRoleId] = useState<string | null>(null)
  const bundlesSheetRole = useMemo(
    () => roles.find((r) => r.id === bundlesSheetRoleId) ?? null,
    [roles, bundlesSheetRoleId]
  )

  // ── Derived ───────────────────────────────────────────────────────

  const filteredRoles = useMemo(
    () => roles.filter(
      (r) =>
        r.name.toLowerCase().includes(search.toLowerCase()) ||
        (r.description ?? '').toLowerCase().includes(search.toLowerCase())
    ),
    [roles, search]
  )

  const liveSheetRole = useMemo(
    () => sheetRole ? roles.find((r) => r.id === sheetRole.id) ?? null : null,
    [roles, sheetRole]
  )

  const sheetUsers = useMemo(
    () => liveSheetRole ? MOCK_USERS.filter((u) => liveSheetRole.userIds.includes(u.id)) : [],
    [liveSheetRole]
  )

  const assignableUsers = useMemo(
    () => liveSheetRole ? MOCK_USERS.filter((u) => !liveSheetRole.userIds.includes(u.id)) : [],
    [liveSheetRole]
  )

  // ── Handlers ─────────────────────────────────────────────────────

  const handleCreate = () => {
    if (!newName.trim()) { toast.error('请输入角色名称'); return }
    if (roles.some((r) => r.name.toLowerCase() === newName.trim().toLowerCase())) {
      toast.error('角色名称已存在'); return
    }
    setRoles((prev) => [
      {
        id: `r-${Date.now()}`,
        name: newName.trim(),
        description: newDescription.trim() || undefined,
        permissions: [],
        isSystem: false,
        createdAt: nowISO(),
        updatedAt: nowISO(),
        userIds: [],
      },
      ...prev,
    ])
    toast.success('角色已创建')
    setCreateOpen(false)
    setNewName('')
    setNewDescription('')
  }

  const handleDeleteConfirm = () => {
    if (!deleteTarget) return
    setRoles((prev) => prev.filter((r) => r.id !== deleteTarget.id))
    toast.success(`已删除角色「${deleteTarget.name}」`)
    setDeleteTarget(null)
  }

  const assignUser = () => {
    if (!liveSheetRole || !userToAssign) return
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? { ...r, userIds: [...r.userIds, userToAssign], updatedAt: nowISO() }
          : r
      )
    )
    setUserToAssign('')
    toast.success('用户已分配')
  }

  const removeUser = (userId: string) => {
    if (!liveSheetRole) return
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? { ...r, userIds: r.userIds.filter((id) => id !== userId), updatedAt: nowISO() }
          : r
      )
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
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[200px]">角色名称</TableHead>
              <TableHead>描述</TableHead>
              <TableHead className="w-[220px] text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredRoles.map((role) => (
              <TableRow key={role.id}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-foreground">{role.name}</span>
                    {role.isSystem && (
                      <Badge variant="secondary" className="text-xs">系统</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {role.description || <span className="text-muted-foreground/40">—</span>}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 gap-1.5 text-xs"
                      onClick={() => { setSheetRole(role); setUserToAssign('') }}
                    >
                      <Users className="size-3.5" strokeWidth={1.5} />
                      用户管理
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 gap-1.5 text-xs"
                      onClick={() => setBundlesSheetRoleId(role.id)}
                    >
                      <KeyRound className="size-3.5" strokeWidth={1.5} />
                      权限管理
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 gap-1.5 text-xs text-muted-foreground hover:text-destructive"
                      disabled={role.isSystem || !canManageRoles}
                      onClick={() => setDeleteTarget(role)}
                    >
                      <Trash2 className="size-3.5" strokeWidth={1.5} />
                      删除
                    </Button>
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
                onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
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
            <Button onClick={handleCreate} className="bg-primary text-primary-foreground hover:bg-primary/90">创建</Button>
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
            >
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Users Sheet */}
      <Sheet open={!!sheetRole} onOpenChange={(open) => { if (!open) setSheetRole(null) }}>
        <SheetContent className="w-full overflow-y-auto sm:max-w-lg">
          {liveSheetRole && (
            <>
              <SheetHeader className="mb-6">
                <SheetTitle className="flex items-center gap-2">
                  <Users className="size-4 text-muted-foreground" strokeWidth={1.5} />
                  {liveSheetRole.name} · 用户管理
                  {liveSheetRole.isSystem && <Badge variant="secondary">系统</Badge>}
                </SheetTitle>
                <SheetDescription>{liveSheetRole.description || '无描述'}</SheetDescription>
              </SheetHeader>

              <div className="mb-4 flex gap-2">
                <Select value={userToAssign} onValueChange={setUserToAssign}>
                  <SelectTrigger className="flex-1 text-sm">
                    <SelectValue placeholder="选择用户分配到此角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {assignableUsers.map((u) => (
                      <SelectItem key={u.id} value={u.id}>
                        {u.name} · {u.email}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button onClick={assignUser} disabled={!userToAssign || !canManageRoles}>
                  分配
                </Button>
              </div>

              <div className="overflow-hidden rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="text-xs">用户</TableHead>
                      <TableHead className="text-xs">邮箱</TableHead>
                      <TableHead className="w-16 text-right text-xs">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {sheetUsers.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-medium">{user.name}</TableCell>
                        <TableCell className="text-muted-foreground">{user.email}</TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 px-2 text-xs text-muted-foreground hover:text-destructive"
                            disabled={!canManageRoles}
                            onClick={() => removeUser(user.id)}
                          >
                            移除
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                    {sheetUsers.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={3} className="py-6 text-center text-sm text-muted-foreground">
                          暂无绑定用户
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </>
          )}
        </SheetContent>
      </Sheet>

      {/* Bundles Sheet */}
      {bundlesSheetRole && (
        <LegacyBundlesSheet
          roleId={bundlesSheetRole.id}
          roleName={bundlesSheetRole.name}
          orgName={orgName}
          projectSlug={projectSlug}
          open={!!bundlesSheetRoleId}
          onOpenChange={(open) => { if (!open) setBundlesSheetRoleId(null) }}
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
    <PageLayout maxWidth="6xl">
      <PageHeader title="权限管理" />

      {/* Tab navigation — underline style */}
      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="h-auto w-full justify-start gap-0 rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            value="roles"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:text-foreground data-[state=active]:border-foreground data-[state=active]:text-foreground data-[state=active]:shadow-none"
          >
            角色
          </TabsTrigger>
          <TabsTrigger
            value="bundles"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:text-foreground data-[state=active]:border-foreground data-[state=active]:text-foreground data-[state=active]:shadow-none"
          >
            权限包
          </TabsTrigger>
          <TabsTrigger
            value="permissions"
            className="rounded-none border-b-2 border-transparent bg-transparent px-4 pb-3 pt-0 text-sm font-medium text-muted-foreground shadow-none transition-none hover:text-foreground data-[state=active]:border-foreground data-[state=active]:text-foreground data-[state=active]:shadow-none"
          >
            权限点
          </TabsTrigger>
        </TabsList>

        <TabsContent value="roles" className="mt-6">
          <LegacyRolesContent orgName={orgName} projectSlug={projectSlug} />
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
