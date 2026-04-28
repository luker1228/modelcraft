'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */

// src/web/components/features/end-user-access/EndUserManagementTable.tsx
// 用户管理：Org 用户列表 × 项目访问 × RBAC 角色/Bundle
// 数据全部接真实接口，不使用 Mock

import { useState, useCallback } from 'react'
import {
  Plus,
  MoreHorizontal,
  RefreshCw,
  Users,
  UserCheck,
  ShieldCheck,
  Lock,
  X,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import { ScrollArea } from '@web/components/ui/scroll-area'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { CreateEndUserDialog } from '../end-users/CreateEndUserDialog'
import {
  useEndUserManagement,
  type UserRoleAssignment,
  type UserBundleAssignment,
} from '@web/hooks/end-user-access/useEndUserManagement'
import type { OrgEndUser } from '@web/hooks/end-users/useOrgEndUsers'
import type { EndUserRole, EndUserPermissionBundle, EffectivePermissions } from '@/types'

// ── Constants ─────────────────────────────────────────────────────────────────

const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询', INSERT: '新增', UPDATE: '修改', DELETE: '删除', EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部', SELF: '本人', DEPT: '本部门', DEPT_AND_CHILDREN: '本部门及子部门',
}

const BUNDLE_LABEL: Record<string, string> = {
  viewer: '查看者', editor: '编辑者', admin: '管理员',
}

// Deterministic avatar color from username initial
const AVATAR_COLORS: { bg: string; text: string }[] = [
  { bg: 'bg-blue-100', text: 'text-blue-700' },
  { bg: 'bg-violet-100', text: 'text-violet-700' },
  { bg: 'bg-amber-100', text: 'text-amber-700' },
  { bg: 'bg-emerald-100', text: 'text-emerald-700' },
  { bg: 'bg-rose-100', text: 'text-rose-700' },
  { bg: 'bg-sky-100', text: 'text-sky-700' },
  { bg: 'bg-orange-100', text: 'text-orange-700' },
]

function avatarColor(username: string) {
  const idx = username.charCodeAt(0) % AVATAR_COLORS.length
  return AVATAR_COLORS[idx]
}

function fmtDate(iso: string) {
  try { return new Date(iso).toLocaleDateString('zh-CN') } catch { return '-' }
}

// ── Sheet: Roles Tab ──────────────────────────────────────────────────────────

interface RolesTabProps {
  userId: string
  assignments: UserRoleAssignment[]
  loading: boolean
  allRoles: EndUserRole[]
  onAssign: (userId: string, roleId: string) => Promise<void>
  onRevoke: (userId: string, roleId: string) => Promise<void>
}

function RolesTab({ userId, assignments, loading, allRoles, onAssign, onRevoke }: RolesTabProps) {
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [assigning, setAssigning] = useState(false)
  const [revokingId, setRevokingId] = useState<string | null>(null)

  const assignedIds = new Set(assignments.map((a) => a.role.id))
  const assignableRoles = allRoles.filter((r) => !r.isImplicit && !assignedIds.has(r.id))
  const implicitRoles = allRoles.filter((r) => r.isImplicit)
  const explicitAssigned = assignments.filter((a) => !a.role.isImplicit)

  const handleAssign = async () => {
    if (!selectedRoleId) return
    setAssigning(true)
    try {
      await onAssign(userId, selectedRoleId)
      setSelectedRoleId('')
    } finally {
      setAssigning(false)
    }
  }

  const handleRevoke = async (roleId: string) => {
    setRevokingId(roleId)
    try { await onRevoke(userId, roleId) }
    finally { setRevokingId(null) }
  }

  if (loading) return (
    <div className="space-y-2 py-2">
      {[1, 2].map((i) => <Skeleton key={i} className="h-10 w-full" />)}
    </div>
  )

  return (
    <div className="space-y-3 py-1">
      <div className="space-y-1.5">
        {implicitRoles.map((role) => (
          <div key={role.id} className="flex items-center justify-between rounded-md border px-3 py-2 opacity-60">
            <div>
              <p className="text-sm text-foreground">{role.name}</p>
              {role.description && <p className="text-xs text-muted-foreground">{role.description}</p>}
            </div>
            <Badge variant="secondary" className="text-xs">内置隐式</Badge>
          </div>
        ))}
        {explicitAssigned.map((a) => (
          <div key={a.role.id} className="flex items-center justify-between rounded-md border px-3 py-2">
            <div>
              <p className="text-sm text-foreground">{a.role.name}</p>
              {a.role.description && <p className="text-xs text-muted-foreground">{a.role.description}</p>}
            </div>
            <Button
              variant="ghost" size="icon"
              className="size-7 text-muted-foreground hover:text-destructive"
              disabled={revokingId === a.role.id}
              onClick={() => handleRevoke(a.role.id)}
            >
              <X className="size-3.5" />
            </Button>
          </div>
        ))}
        {implicitRoles.length === 0 && explicitAssigned.length === 0 && (
          <p className="py-6 text-center text-sm text-muted-foreground">暂未分配任何角色</p>
        )}
      </div>

      {assignableRoles.length > 0 && (
        <div className="flex items-center gap-2 border-t pt-3">
          <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
            <SelectTrigger className="h-8 flex-1 text-sm">
              <SelectValue placeholder="选择角色..." />
            </SelectTrigger>
            <SelectContent>
              {assignableRoles.map((r) => (
                <SelectItem key={r.id} value={r.id}>{r.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            size="sm" className="h-8 shrink-0"
            disabled={!selectedRoleId || assigning}
            onClick={handleAssign}
          >
            {assigning ? '分配中...' : '分配'}
          </Button>
        </div>
      )}
    </div>
  )
}

// ── Sheet: Bundles Tab ────────────────────────────────────────────────────────

interface BundlesTabProps {
  userId: string
  assignments: UserBundleAssignment[]
  loading: boolean
  allBundles: EndUserPermissionBundle[]
  onAssign: (userId: string, bundleId: string) => Promise<void>
  onRevoke: (userId: string, bundleId: string) => Promise<void>
}

function BundlesTab({ userId, assignments, loading, allBundles, onAssign, onRevoke }: BundlesTabProps) {
  const [selectedBundleId, setSelectedBundleId] = useState('')
  const [assigning, setAssigning] = useState(false)
  const [revokingId, setRevokingId] = useState<string | null>(null)

  const assignedIds = new Set(assignments.map((a) => a.bundle.id))
  const availableBundles = allBundles.filter((b) => !assignedIds.has(b.id))

  const handleAssign = async () => {
    if (!selectedBundleId) return
    setAssigning(true)
    try {
      await onAssign(userId, selectedBundleId)
      setSelectedBundleId('')
    } finally {
      setAssigning(false)
    }
  }

  const handleRevoke = async (bundleId: string) => {
    setRevokingId(bundleId)
    try { await onRevoke(userId, bundleId) }
    finally { setRevokingId(null) }
  }

  if (loading) return (
    <div className="space-y-2 py-2">
      {[1, 2].map((i) => <Skeleton key={i} className="h-10 w-full" />)}
    </div>
  )

  return (
    <div className="space-y-3 py-1">
      <div className="space-y-1.5">
        {assignments.map((a) => (
          <div key={a.bundle.id} className="flex items-start justify-between rounded-md border px-3 py-2">
            <div>
              <p className="text-sm text-foreground">{a.bundle.name}</p>
              {a.bundle.description && <p className="text-xs text-muted-foreground">{a.bundle.description}</p>}
              <p className="mt-0.5 text-xs text-muted-foreground/60">
                {a.bundle.permissions.length} 个权限点
              </p>
            </div>
            <Button
              variant="ghost" size="icon"
              className="mt-0.5 size-7 text-muted-foreground hover:text-destructive"
              disabled={revokingId === a.bundle.id}
              onClick={() => handleRevoke(a.bundle.id)}
            >
              <X className="size-3.5" />
            </Button>
          </div>
        ))}
        {assignments.length === 0 && (
          <p className="py-6 text-center text-sm text-muted-foreground">暂无直接授权的权限包</p>
        )}
      </div>

      {availableBundles.length > 0 && (
        <div className="flex items-center gap-2 border-t pt-3">
          <Select value={selectedBundleId} onValueChange={setSelectedBundleId}>
            <SelectTrigger className="h-8 flex-1 text-sm">
              <SelectValue placeholder="选择权限包..." />
            </SelectTrigger>
            <SelectContent>
              {availableBundles.map((b) => (
                <SelectItem key={b.id} value={b.id}>{b.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            size="sm" className="h-8 shrink-0"
            disabled={!selectedBundleId || assigning}
            onClick={handleAssign}
          >
            {assigning ? '授权中...' : '授权'}
          </Button>
        </div>
      )}
    </div>
  )
}

// ── Sheet: Effective Permissions Tab ─────────────────────────────────────────

function EffectiveTab({ permissions, loading }: { permissions: EffectivePermissions[]; loading: boolean }) {
  const [openModelId, setOpenModelId] = useState<string | null>(null)

  if (loading) return (
    <div className="space-y-2 py-2">
      {[1, 2, 3].map((i) => <Skeleton key={i} className="h-10 w-full" />)}
    </div>
  )

  const nonEmpty = permissions.filter((ep) => ep.grants.length > 0)

  if (nonEmpty.length === 0) return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <Lock className="mb-3 size-8 text-muted-foreground/25" />
      <p className="text-sm font-medium text-foreground">暂无有效权限</p>
      <p className="mt-1 text-xs text-muted-foreground">通过角色或直接权限包授权后将在此展示</p>
    </div>
  )

  return (
    <div className="space-y-2 py-1">
      {nonEmpty.map((ep) => {
        const isOpen = openModelId === ep.modelId
        return (
          <div key={ep.modelId} className="overflow-hidden rounded-md border">
            <button
              onClick={() => setOpenModelId(isOpen ? null : ep.modelId)}
              className="flex w-full items-center justify-between px-3 py-2.5 text-left hover:bg-muted/40"
            >
              <div className="flex items-center gap-2">
                {isOpen ? <ChevronDown className="size-3.5 text-muted-foreground" /> : <ChevronRight className="size-3.5 text-muted-foreground" />}
                <span className="text-sm font-medium">{ep.modelId}</span>
                <span className="text-xs text-muted-foreground">{ep.grants.length} 条</span>
              </div>
            </button>
            {isOpen && (
              <div className="space-y-1.5 border-t bg-muted/10 px-3 py-2.5">
                {ep.grants.map((g, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <Badge variant="secondary" className="text-xs">{ACTION_LABEL[g.action] ?? g.action}</Badge>
                    <span className="text-xs text-muted-foreground">{ROW_SCOPE_LABEL[g.rowScope] ?? g.rowScope}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

// ── User Detail Sheet ─────────────────────────────────────────────────────────

interface UserDetailSheetProps {
  user: OrgEndUser | null
  open: boolean
  onOpenChange: (open: boolean) => void
  // access
  hasAccess: boolean
  accessEntry?: { accessId: string; permissionBundle?: string }
  allBundles: EndUserPermissionBundle[]
  onGrantAccess: (userId: string, bundleId: string) => Promise<void>
  onRevokeAccess: (accessId: string) => Promise<void>
  onUpdateAccessBundle: (accessId: string, bundleId: string) => Promise<void>
  // roles
  roleAssignments: UserRoleAssignment[]
  roleAssignmentsLoading: boolean
  allRoles: EndUserRole[]
  onAssignRole: (userId: string, roleId: string) => Promise<void>
  onRevokeRole: (userId: string, roleId: string) => Promise<void>
  // bundles
  bundleAssignments: UserBundleAssignment[]
  bundleAssignmentsLoading: boolean
  onAssignBundle: (userId: string, bundleId: string) => Promise<void>
  onRevokeBundle: (userId: string, bundleId: string) => Promise<void>
  // effective
  effectivePermissions: EffectivePermissions[]
  effectiveLoading: boolean
}

function UserDetailSheet({
  user, open, onOpenChange,
  hasAccess, accessEntry, allBundles,
  onGrantAccess, onRevokeAccess, onUpdateAccessBundle,
  roleAssignments, roleAssignmentsLoading, allRoles, onAssignRole, onRevokeRole,
  bundleAssignments, bundleAssignmentsLoading, onAssignBundle, onRevokeBundle,
  effectivePermissions, effectiveLoading,
}: UserDetailSheetProps) {
  const [accessUpdating, setAccessUpdating] = useState(false)

  if (!user) return null

  const handleGrantAccess = async (bundleId: string) => {
    setAccessUpdating(true)
    try { await onGrantAccess(user.id, bundleId) }
    catch (e: unknown) { toast.error(e instanceof Error ? e.message : '授权失败') }
    finally { setAccessUpdating(false) }
  }

  const handleRevokeAccess = async () => {
    if (!accessEntry) return
    setAccessUpdating(true)
    try { await onRevokeAccess(accessEntry.accessId) }
    catch (e: unknown) { toast.error(e instanceof Error ? e.message : '撤销失败') }
    finally { setAccessUpdating(false) }
  }

  const handleUpdateBundle = async (bundleId: string) => {
    if (!accessEntry) return
    setAccessUpdating(true)
    try { await onUpdateAccessBundle(accessEntry.accessId, bundleId) }
    catch (e: unknown) { toast.error(e instanceof Error ? e.message : '更新失败') }
    finally { setAccessUpdating(false) }
  }

  const BUNDLE_LABELS: Record<string, string> = {
    viewer: '查看者（只读）', editor: '编辑者（读写）', admin: '管理员（全权限）',
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex flex-col sm:max-w-lg">
        <SheetHeader className="border-b pb-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-primary" />
            <SheetTitle className="font-semibold">{user.username}</SheetTitle>
            <Badge
              variant={!user.isForbidden ? 'outline' : 'destructive'}
              className={!user.isForbidden ? 'border-emerald-200 bg-emerald-50 text-emerald-700 text-xs' : 'text-xs'}
            >
              {!user.isForbidden ? '正常' : '已禁用'}
            </Badge>
          </div>
          <SheetDescription>
            {user.displayName && <span className="mr-2">{user.displayName}</span>}
            注册于 {fmtDate(user.createdAt)}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1">
          <div className="py-4">
            <Tabs defaultValue="access">
              <TabsList className="w-full">
                <TabsTrigger value="access" className="flex-1">项目访问</TabsTrigger>
                <TabsTrigger value="roles" className="flex-1">角色</TabsTrigger>
                <TabsTrigger value="bundles" className="flex-1">直接权限包</TabsTrigger>
                <TabsTrigger value="effective" className="flex-1">有效权限</TabsTrigger>
              </TabsList>

              {/* ── 项目访问 tab ── */}
              <TabsContent value="access" className="mt-4 space-y-4">
                <div className="rounded-md border p-4">
                  <div className="mb-3 flex items-center justify-between">
                    <span className="text-sm font-medium text-foreground">项目访问权限</span>
                    {hasAccess ? (
                      <Badge variant="secondary" className="gap-1 text-xs">
                        <UserCheck className="size-3" />已授权
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-xs text-muted-foreground">未授权</Badge>
                    )}
                  </div>

                  {hasAccess ? (
                    <div className="space-y-3">
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-muted-foreground">权限包：</span>
                        <Select
                          value={accessEntry?.permissionBundle ?? ''}
                          onValueChange={handleUpdateBundle}
                          disabled={accessUpdating}
                        >
                          <SelectTrigger className="h-7 w-44 text-xs">
                            <SelectValue placeholder="选择权限包" />
                          </SelectTrigger>
                          <SelectContent>
                            {Object.entries(BUNDLE_LABELS).map(([v, label]) => (
                              <SelectItem key={v} value={v} className="text-xs">{label}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <Button
                        variant="outline" size="sm"
                        className="h-7 text-xs text-destructive hover:text-destructive"
                        disabled={accessUpdating}
                        onClick={handleRevokeAccess}
                      >
                        撤销项目访问权限
                      </Button>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      <p className="text-xs text-muted-foreground">选择权限包后授权该用户访问本项目</p>
                      <div className="flex items-center gap-2">
                        {Object.entries(BUNDLE_LABELS).map(([v, label]) => (
                          <Button
                            key={v} variant="outline" size="sm"
                            className="h-7 text-xs"
                            disabled={accessUpdating}
                            onClick={() => handleGrantAccess(v)}
                          >
                            {label.split('（')[0]}
                          </Button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </TabsContent>

              {/* ── 角色 tab ── */}
              <TabsContent value="roles" className="mt-4">
                <RolesTab
                  userId={user.id}
                  assignments={roleAssignments}
                  loading={roleAssignmentsLoading}
                  allRoles={allRoles}
                  onAssign={onAssignRole}
                  onRevoke={onRevokeRole}
                />
              </TabsContent>

              {/* ── 直接权限包 tab ── */}
              <TabsContent value="bundles" className="mt-4">
                <BundlesTab
                  userId={user.id}
                  assignments={bundleAssignments}
                  loading={bundleAssignmentsLoading}
                  allBundles={allBundles}
                  onAssign={onAssignBundle}
                  onRevoke={onRevokeBundle}
                />
              </TabsContent>

              {/* ── 有效权限 tab ── */}
              <TabsContent value="effective" className="mt-4">
                <EffectiveTab permissions={effectivePermissions} loading={effectiveLoading} />
              </TabsContent>
            </Tabs>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

// ── Main Table ────────────────────────────────────────────────────────────────

interface EndUserManagementTableProps {
  orgName: string
  projectSlug: string
}

export function EndUserManagementTable({ orgName, projectSlug }: EndUserManagementTableProps) {
  const mgmt = useEndUserManagement(orgName, projectSlug)
  const [createOpen, setCreateOpen] = useState(false)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const isLoading = mgmt.usersLoading || mgmt.accessLoading

  // Build accessMap: userId → access entry
  const accessMap = new Map(mgmt.accesses.map((a) => [a.userId, a]))

  // Selected user object
  const selectedUser = mgmt.users.find((u) => u.id === mgmt.selectedUserId) ?? null

  const handleOpenSheet = useCallback(
    (user: OrgEndUser) => {
      mgmt.setSelectedUserId(user.id)
      setSheetOpen(true)
    },
    [mgmt]
  )

  const handleCloseSheet = useCallback(() => {
    setSheetOpen(false)
    mgmt.setSelectedUserId(null)
  }, [mgmt])

  const handleToggleStatus = async (user: OrgEndUser) => {
    setActionError(null)
    const newStatus = user.isForbidden ? 'ACTIVE' : 'DISABLED'
    try {
      await mgmt.toggleUserStatus(user.id, newStatus)
      toast.success(newStatus === 'ACTIVE' ? `已启用 ${user.username}` : `已禁用 ${user.username}`)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '操作失败')
    }
  }

  const handleDelete = async (user: OrgEndUser) => {
    if (!confirm(`确认删除用户 ${user.username}？此操作不可恢复。`)) return
    setActionError(null)
    try {
      await mgmt.deleteUser(user.id)
      toast.success(`已删除用户 ${user.username}`)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '删除失败')
    }
  }

  const handleAssignRole = async (userId: string, roleId: string) => {
    const r = await mgmt.assignRole(userId, roleId)
    if (r.success) toast.success('角色已分配')
    else toast.error(r.errorMessage ?? '分配角色失败')
  }

  const handleRevokeRole = async (userId: string, roleId: string) => {
    const r = await mgmt.revokeRole(userId, roleId)
    if (r.success) toast.success('角色已撤销')
    else toast.error(r.errorMessage ?? '撤销角色失败')
  }

  const handleAssignBundle = async (userId: string, bundleId: string) => {
    const r = await mgmt.assignBundle(userId, bundleId)
    if (r.success) toast.success('权限包已授权')
    else toast.error(r.errorMessage ?? '授权失败')
  }

  const handleRevokeBundle = async (userId: string, bundleId: string) => {
    const r = await mgmt.revokeBundle(userId, bundleId)
    if (r.success) toast.success('权限包已撤销')
    else toast.error(r.errorMessage ?? '撤销失败')
  }

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground">
          {!isLoading && mgmt.users.length > 0 && (
            <span>{mgmt.users.length} 名用户，{mgmt.accesses.length} 名已授权</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {actionError && <p className="text-xs text-destructive">{actionError}</p>}
          <Button size="sm" variant="outline" onClick={() => { mgmt.reloadUsers(); mgmt.reloadAccess() }} disabled={isLoading}>
            <RefreshCw className={`mr-1.5 size-4 ${isLoading ? 'animate-spin' : ''}`} />刷新
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)} className="bg-primary text-primary-foreground hover:bg-primary/90">
            <Plus className="size-4" />新增用户
          </Button>
        </div>
      </div>

      {/* Error state */}
      {(mgmt.usersError ?? mgmt.accessError) && !isLoading && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Users className="mb-3 size-10 text-muted-foreground/40" />
          <p className="text-sm text-muted-foreground">{mgmt.usersError ?? mgmt.accessError}</p>
        </div>
      )}

      {/* Table */}
      {!(mgmt.usersError ?? mgmt.accessError) && (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[220px] px-3 font-semibold text-sm text-muted-foreground">用户</TableHead>
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">账号状态</TableHead>
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">项目权限</TableHead>
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">加入时间</TableHead>
                <TableHead className="w-[160px] px-3 text-right font-semibold text-sm text-muted-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && Array.from({ length: 4 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 5 }).map((__, j) => (
                    <TableCell key={j} className="px-3 py-2"><Skeleton className="h-4 w-24" /></TableCell>
                  ))}
                </TableRow>
              ))}

              {!isLoading && mgmt.users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className="flex flex-col items-center justify-center py-14 text-center">
                      <Users className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1.5} />
                      <p className="text-sm font-semibold text-foreground">暂无终端用户</p>
                      <p className="mt-1 text-xs text-muted-foreground">新增用户后可在此管理其项目访问权限</p>
                      <Button size="sm" variant="outline" className="mt-4" onClick={() => setCreateOpen(true)}>
                        <Plus className="mr-1.5 size-3.5" />新增第一个用户
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!isLoading && mgmt.users.map((user) => {
                const access = accessMap.get(user.id)
                const isSelected = mgmt.selectedUserId === user.id
                const av = avatarColor(user.username)

                return (
                  <TableRow
                    key={user.id}
                    className={isSelected ? 'bg-muted/30' : 'hover:bg-muted/50'}
                  >
                    {/* 用户 */}
                    <TableCell className="px-3 py-2">
                      <div className="flex items-center gap-2">
                        <div className={`flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold ${av.bg} ${av.text}`}>
                          {user.username.charAt(0).toUpperCase()}
                        </div>
                        <div className="min-w-0">
                          <span className="block font-medium text-foreground">{user.username}</span>
                          {user.displayName && (
                            <span className="block text-xs text-muted-foreground">{user.displayName}</span>
                          )}
                        </div>
                      </div>
                    </TableCell>

                    {/* 账号状态 */}
                    <TableCell className="px-3 py-2">
                      {!user.isForbidden ? (
                        <Badge variant="secondary" className="text-xs">正常</Badge>
                      ) : (
                        <Badge variant="destructive" className="text-xs">已禁用</Badge>
                      )}
                    </TableCell>

                    {/* 项目权限 */}
                    <TableCell className="px-3 py-2 text-muted-foreground">
                      {access ? (
                        <Badge variant="secondary" className="gap-1 text-xs">
                          <UserCheck className="size-3" />
                          {BUNDLE_LABEL[access.permissionBundle ?? ''] ?? access.permissionBundle ?? '已授权'}
                        </Badge>
                      ) : (
                        <span className="text-sm text-muted-foreground/60">—</span>
                      )}
                    </TableCell>

                    {/* 加入时间 */}
                    <TableCell className="px-3 py-2 text-muted-foreground">
                      {fmtDate(user.createdAt)}
                    </TableCell>

                    {/* 操作 */}
                    <TableCell className="px-3 py-2 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 gap-1.5 px-3 text-xs"
                          onClick={() => handleOpenSheet(user)}
                        >
                          <ShieldCheck className="size-3.5" strokeWidth={1.5} />
                          权限管理
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button size="icon" variant="ghost" className="size-8">
                              <MoreHorizontal className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => handleToggleStatus(user)}>
                              {user.isForbidden ? '启用账号' : '禁用账号'}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => handleDelete(user)}
                            >
                              删除用户
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Create dialog */}
      <CreateEndUserDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={mgmt.createUser}
      />

      {/* User detail sheet */}
      <UserDetailSheet
        user={selectedUser}
        open={sheetOpen}
        onOpenChange={(o) => { if (!o) handleCloseSheet() }}
        hasAccess={!!accessMap.get(selectedUser?.id ?? '')}
        accessEntry={accessMap.get(selectedUser?.id ?? '')}
        allBundles={mgmt.allBundles}
        onGrantAccess={mgmt.grantAccess}
        onRevokeAccess={mgmt.revokeAccess}
        onUpdateAccessBundle={mgmt.updateAccessBundle}
        roleAssignments={mgmt.userRoleAssignments}
        roleAssignmentsLoading={mgmt.userRoleAssignmentsLoading}
        allRoles={mgmt.allRoles}
        onAssignRole={handleAssignRole}
        onRevokeRole={handleRevokeRole}
        bundleAssignments={mgmt.userBundleAssignments}
        bundleAssignmentsLoading={mgmt.userBundleAssignmentsLoading}
        onAssignBundle={handleAssignBundle}
        onRevokeBundle={handleRevokeBundle}
        effectivePermissions={mgmt.effectivePermissions}
        effectiveLoading={mgmt.effectiveLoading}
      />
    </div>
  )
}
