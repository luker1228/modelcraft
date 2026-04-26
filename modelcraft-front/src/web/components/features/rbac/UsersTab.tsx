/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any */
import * as React from 'react'
import { toast } from 'sonner'
import {
  Users,
  ChevronDown,
  ChevronRight,
  ShieldCheck,
  Lock,
  X,
} from 'lucide-react'

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
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'

import {
  useUserAuth,
  type MockEndUserUser,
} from '@/app/org/[orgName]/project/[projectSlug]/rbac/users/_hooks/useUserAuth'
import type {
  EndUserRole,
  EndUserPermissionBundle,
  EndUserPermissionAction,
  EndUserRowScope,
  EffectivePermissions,
} from '@/types'

// ── Props ────────────────────────────────────────────────────────────────────

export interface UsersTabProps {
  orgName: string
  projectSlug: string
}

// ── Display helpers ───────────────────────────────────────────────────────────

const ACTION_LABEL: Record<EndUserPermissionAction, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '修改',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<EndUserRowScope, string> = {
  ALL: '全部',
  SELF: '本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '本部门及子部门',
}

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

// ── RoleBadges ────────────────────────────────────────────────────────────────

interface RoleBadgesProps {
  roles: EndUserRole[]
  maxVisible?: number
}

function RoleBadges({ roles, maxVisible = 2 }: RoleBadgesProps) {
  const explicitRoles = roles.filter((r) => !r.isImplicit)

  if (explicitRoles.length === 0) {
    return <span className="italic text-sm text-muted-foreground/60">暂无角色</span>
  }

  const visible = explicitRoles.slice(0, maxVisible)
  const overflow = explicitRoles.length - maxVisible

  return (
    <div className="flex flex-wrap items-center gap-1">
      {visible.map((role) => (
        <Badge key={role.id} variant="secondary" className="text-xs">
          {role.name}
        </Badge>
      ))}
      {overflow > 0 && (
        <span className="text-xs text-muted-foreground">+{overflow}</span>
      )}
    </div>
  )
}

// ── UserTableSkeleton ─────────────────────────────────────────────────────────

function UserTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[180px] text-xs font-medium text-muted-foreground">
              用户名
            </TableHead>
            <TableHead className="text-xs font-medium text-muted-foreground">
              已分配角色
            </TableHead>
            <TableHead className="w-[160px] text-xs font-medium text-muted-foreground">
              注册时间
            </TableHead>
            <TableHead className="w-[100px] text-right text-xs font-medium text-muted-foreground">
              操作
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-4 w-36" /></TableCell>
              <TableCell className="text-right">
                <Skeleton className="ml-auto h-7 w-16" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// ── RoleTab ───────────────────────────────────────────────────────────────────

interface RoleTabProps {
  user: MockEndUserUser
  allRoles: EndUserRole[]
  rolesLoading: boolean
  onAssignRole: (roleId: string) => Promise<void>
  onRevokeRole: (roleId: string) => Promise<void>
}

function RoleTab({
  user,
  allRoles,
  rolesLoading,
  onAssignRole,
  onRevokeRole,
}: RoleTabProps) {
  const [selectedRoleId, setSelectedRoleId] = React.useState<string>('')
  const [assigning, setAssigning] = React.useState(false)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)

  const assignedIds = new Set(user.assignedRoles.map((r) => r.id))
  const assignableRoles = allRoles.filter((r) => !r.isImplicit && !assignedIds.has(r.id))

  const handleAssign = async () => {
    if (!selectedRoleId) return
    setAssigning(true)
    try {
      await onAssignRole(selectedRoleId)
      setSelectedRoleId('')
    } finally {
      setAssigning(false)
    }
  }

  const handleRevoke = async (roleId: string) => {
    setRevokingId(roleId)
    try {
      await onRevokeRole(roleId)
    } finally {
      setRevokingId(null)
    }
  }

  if (rolesLoading) {
    return (
      <div className="space-y-3 py-2">
        {Array.from({ length: 2 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    )
  }

  const implicitRoles = allRoles.filter((r) => r.isImplicit)
  const explicitAssigned = user.assignedRoles.filter((r) => !r.isImplicit)

  return (
    <div className="space-y-3 py-1">
      <div className="space-y-1">
        {implicitRoles.map((role) => (
          <div
            key={role.id}
            className="flex items-center justify-between rounded-md border border-border px-3 py-2 opacity-75"
          >
            <div className="min-w-0 flex-1">
              <span className="truncate text-sm text-foreground">
                {role.name}
              </span>
              {role.description && (
                <p className="truncate text-xs text-muted-foreground">{role.description}</p>
              )}
            </div>
            <Badge variant="secondary" className="ml-3 shrink-0 text-xs">
              内置隐式
            </Badge>
          </div>
        ))}

        {explicitAssigned.map((role) => (
          <div
            key={role.id}
            className="flex items-center justify-between rounded-md border border-border px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="truncate text-sm text-foreground">
                {role.name}
              </span>
              {role.description && (
                <p className="truncate text-xs text-muted-foreground">{role.description}</p>
              )}
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="ml-2 size-7 shrink-0 p-0 text-muted-foreground hover:text-destructive"
              disabled={revokingId === role.id}
              onClick={() => handleRevoke(role.id)}
              aria-label="撤销角色"
            >
              <X className="size-3.5" />
            </Button>
          </div>
        ))}

        {implicitRoles.length === 0 && explicitAssigned.length === 0 && (
          <p className="py-4 text-center text-sm text-muted-foreground">
            该用户暂未分配任何角色
          </p>
        )}
      </div>

      {assignableRoles.length > 0 && (
        <div className="flex items-center gap-2 border-t border-border pt-3">
          <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
            <SelectTrigger className="h-8 flex-1 text-sm">
              <SelectValue placeholder="选择角色..." />
            </SelectTrigger>
            <SelectContent>
              {assignableRoles.map((role) => (
                <SelectItem key={role.id} value={role.id}>
                  {role.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            size="sm"
            className="h-8 shrink-0 bg-primary text-primary-foreground hover:bg-primary/90"
            disabled={!selectedRoleId || assigning}
            onClick={handleAssign}
          >
            {assigning ? '分配中...' : '分配角色'}
          </Button>
        </div>
      )}
    </div>
  )
}

// ── AssignBundleDialog ────────────────────────────────────────────────────────

interface AssignBundleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  availableBundles: EndUserPermissionBundle[]
  onAssign: (bundleId: string) => Promise<void>
}

function AssignBundleDialog({
  open,
  onOpenChange,
  availableBundles,
  onAssign,
}: AssignBundleDialogProps) {
  const [selectedId, setSelectedId] = React.useState<string>('')
  const [assigning, setAssigning] = React.useState(false)

  const handleConfirm = async () => {
    if (!selectedId) return
    setAssigning(true)
    try {
      await onAssign(selectedId)
      setSelectedId('')
      onOpenChange(false)
    } finally {
      setAssigning(false)
    }
  }

  const handleOpenChange = (next: boolean) => {
    if (!next) setSelectedId('')
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>授权权限包</DialogTitle>
        </DialogHeader>

        <div className="space-y-3 py-2">
          {availableBundles.length === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              所有权限包已授权，无可添加项
            </p>
          ) : (
            <div className="max-h-[300px] space-y-1.5 overflow-y-auto">
              {availableBundles.map((bundle) => (
                <button
                  key={bundle.id}
                  onClick={() => setSelectedId(bundle.id)}
                  className={`w-full rounded-md border px-3 py-2 text-left transition-colors ${
                    selectedId === bundle.id
                      ? 'border-primary bg-primary/5'
                      : 'border-border hover:border-primary/50 hover:bg-muted/40'
                  }`}
                >
                  <p className="text-sm font-semibold text-foreground">{bundle.name}</p>
                  {bundle.description && (
                    <p className="mt-0.5 text-xs text-muted-foreground">{bundle.description}</p>
                  )}
                  <p className="mt-0.5 text-xs text-muted-foreground/60">
                    {bundle.permissions.length} 个权限点
                  </p>
                </button>
              ))}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={assigning}>
            取消
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!selectedId || assigning || availableBundles.length === 0}
            className="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {assigning ? '授权中...' : '确认授权'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── BundleTab ─────────────────────────────────────────────────────────────────

interface BundleTabProps {
  user: MockEndUserUser
  allBundles: EndUserPermissionBundle[]
  bundlesLoading: boolean
  onAssignBundle: (bundleId: string) => Promise<void>
  onRevokeBundle: (bundleId: string) => Promise<void>
}

function BundleTab({
  user,
  allBundles,
  bundlesLoading,
  onAssignBundle,
  onRevokeBundle,
}: BundleTabProps) {
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)

  const assignedIds = new Set(user.assignedBundles.map((b) => b.id))
  const availableBundles = allBundles.filter((b) => !assignedIds.has(b.id))

  const handleRevoke = async (bundleId: string) => {
    setRevokingId(bundleId)
    try {
      await onRevokeBundle(bundleId)
    } finally {
      setRevokingId(null)
    }
  }

  if (bundlesLoading) {
    return (
      <div className="space-y-3 py-2">
        {Array.from({ length: 2 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-3 py-1">
      <div className="space-y-1">
        {user.assignedBundles.map((bundle) => (
          <div
            key={bundle.id}
            className="flex items-start justify-between rounded-md border border-border px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="block truncate text-sm text-foreground">
                {bundle.name}
              </span>
              {bundle.description && (
                <p className="truncate text-xs text-muted-foreground">{bundle.description}</p>
              )}
              <p className="mt-0.5 text-xs text-muted-foreground/70">
                {bundle.permissions.length} 个权限点
              </p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="ml-2 mt-0.5 size-7 shrink-0 p-0 text-muted-foreground hover:text-destructive"
              disabled={revokingId === bundle.id}
              onClick={() => handleRevoke(bundle.id)}
              aria-label="撤销权限包"
            >
              <X className="size-3.5" />
            </Button>
          </div>
        ))}

        {user.assignedBundles.length === 0 && (
          <p className="py-4 text-center text-sm text-muted-foreground">
            该用户暂未直接授权任何权限包
          </p>
        )}
      </div>

      <div className="border-t border-border pt-3">
        <Button
          variant="outline"
          size="sm"
          className="h-8 text-xs"
          onClick={() => setDialogOpen(true)}
          disabled={availableBundles.length === 0}
        >
          + 授权权限包
        </Button>
      </div>

      <AssignBundleDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        availableBundles={availableBundles}
        onAssign={onAssignBundle}
      />
    </div>
  )
}

// ── EffectivePermissionsTab ───────────────────────────────────────────────────

interface EffectivePermissionsTabProps {
  effectivePermissions: EffectivePermissions[]
  loading: boolean
}

interface ModelGrantGroupProps {
  modelId: string
  permissions: EffectivePermissions[]
}

function ModelGrantGroup({ modelId, permissions }: ModelGrantGroupProps) {
  const [open, setOpen] = React.useState(true)

  const grants = permissions.flatMap((ep) => ep.grants)

  return (
    <div className="overflow-hidden rounded-md border border-border">
      <button
        onClick={() => setOpen((prev) => !prev)}
        className="flex w-full items-center justify-between px-4 py-2.5 text-left transition-colors hover:bg-muted/40"
      >
        <div className="flex items-center gap-2">
          {open ? (
            <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          )}
          <span className="text-sm font-medium text-foreground">{modelId}</span>
          <span className="text-xs text-muted-foreground">{grants.length} 条授权</span>
        </div>
      </button>

      {open && (
        <div className="space-y-2 border-t border-border bg-muted/10 px-4 py-3">
          {grants.length === 0 ? (
            <p className="text-xs text-muted-foreground">该模型无有效授权</p>
          ) : (
            grants.map((grant, idx) => (
              <div
                key={idx}
                className="flex items-center gap-2"
              >
                <Badge variant="secondary" className="text-xs">
                  {ACTION_LABEL[grant.action] ?? grant.action}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {ROW_SCOPE_LABEL[grant.rowScope] ?? grant.rowScope}
                </span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}

function EffectivePermissionsTab({
  effectivePermissions,
  loading,
}: EffectivePermissionsTabProps) {
  if (loading) {
    return (
      <div className="space-y-3 py-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    )
  }

  const nonEmpty = effectivePermissions.filter(
    (ep) => ep.modelId && ep.grants.length > 0
  )

  if (nonEmpty.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-10 text-center">
        <Lock className="mb-3 size-8 text-muted-foreground/25" />
        <p className="text-sm font-medium text-foreground">该用户暂无任何权限</p>
        <p className="mt-1 text-xs text-muted-foreground">
          通过「角色」或「直接权限包」为该用户授权后，有效权限将在此展示
        </p>
      </div>
    )
  }

  const grouped = nonEmpty.reduce<Record<string, EffectivePermissions[]>>((acc, ep) => {
    if (!acc[ep.modelId]) acc[ep.modelId] = []
    acc[ep.modelId].push(ep)
    return acc
  }, {})

  return (
    <div className="space-y-2 py-1">
      {Object.entries(grouped).map(([modelId, eps]) => (
        <ModelGrantGroup key={modelId} modelId={modelId} permissions={eps} />
      ))}
    </div>
  )
}

// ── UserDetailSheet ───────────────────────────────────────────────────────────

interface UserDetailSheetProps {
  user: MockEndUserUser | null
  open: boolean
  onOpenChange: (open: boolean) => void
  allRoles: EndUserRole[]
  rolesLoading: boolean
  allBundles: EndUserPermissionBundle[]
  bundlesLoading: boolean
  effectivePermissions: EffectivePermissions[]
  effectiveLoading: boolean
  projectSlug: string
  onAssignRole: (endUserId: string, roleId: string) => Promise<void>
  onRevokeRole: (endUserId: string, roleId: string) => Promise<void>
  onAssignBundle: (endUserId: string, bundleId: string) => Promise<void>
  onRevokeBundle: (endUserId: string, bundleId: string) => Promise<void>
}

function UserDetailSheet({
  user,
  open,
  onOpenChange,
  allRoles,
  rolesLoading,
  allBundles,
  bundlesLoading,
  effectivePermissions,
  effectiveLoading,
  onAssignRole,
  onRevokeRole,
  onAssignBundle,
  onRevokeBundle,
}: UserDetailSheetProps) {
  if (!user) return null

  const handleAssignRole = async (roleId: string) => {
    await onAssignRole(user.id, roleId)
  }

  const handleRevokeRole = async (roleId: string) => {
    await onRevokeRole(user.id, roleId)
  }

  const handleAssignBundle = async (bundleId: string) => {
    await onAssignBundle(user.id, bundleId)
  }

  const handleRevokeBundle = async (bundleId: string) => {
    await onRevokeBundle(user.id, bundleId)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="flex flex-col sm:max-w-lg"
      >
        <SheetHeader className="border-b border-border pb-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-primary" />
            <SheetTitle className="font-semibold">
              {user.username}
            </SheetTitle>
          </div>
          <SheetDescription className="text-xs text-muted-foreground">
            注册时间：{formatDateTime(user.createdAt)}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1">
          <div className="py-4">
            <Tabs defaultValue="roles">
              <TabsList className="w-full">
                <TabsTrigger value="roles" className="flex-1">
                  角色
                </TabsTrigger>
                <TabsTrigger value="bundles" className="flex-1">
                  直接权限包
                </TabsTrigger>
                <TabsTrigger value="effective" className="flex-1">
                  有效权限
                </TabsTrigger>
              </TabsList>

              <TabsContent value="roles" className="mt-4">
                <RoleTab
                  user={user}
                  allRoles={allRoles}
                  rolesLoading={rolesLoading}
                  onAssignRole={handleAssignRole}
                  onRevokeRole={handleRevokeRole}
                />
              </TabsContent>

              <TabsContent value="bundles" className="mt-4">
                <BundleTab
                  user={user}
                  allBundles={allBundles}
                  bundlesLoading={bundlesLoading}
                  onAssignBundle={handleAssignBundle}
                  onRevokeBundle={handleRevokeBundle}
                />
              </TabsContent>

              <TabsContent value="effective" className="mt-4">
                <EffectivePermissionsTab
                  effectivePermissions={effectivePermissions}
                  loading={effectiveLoading}
                />
              </TabsContent>
            </Tabs>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

// ── UsersTab ──────────────────────────────────────────────────────────────────

export function UsersTab({ orgName, projectSlug }: UsersTabProps) {
  const {
    users,
    selectedUser,
    setSelectedUser,
    roles,
    rolesLoading,
    bundles,
    bundlesLoading,
    effectivePermissions,
    effectiveLoading,
    assignRole,
    revokeRole,
    assignBundle,
    revokeBundle,
  } = useUserAuth({ orgName, projectSlug })

  const [sheetOpen, setSheetOpen] = React.useState(false)

  const handleManage = React.useCallback(
    (user: MockEndUserUser) => {
      setSelectedUser(user)
      setSheetOpen(true)
    },
    [setSelectedUser]
  )

  const handleSheetOpenChange = React.useCallback(
    (open: boolean) => {
      setSheetOpen(open)
      if (!open) setSelectedUser(null)
    },
    [setSelectedUser]
  )

  const handleAssignRole = React.useCallback(
    async (endUserId: string, roleId: string) => {
      const result = await assignRole(endUserId, roleId)
      if (result.success) {
        toast.success('角色分配成功')
      } else {
        toast.error(result.errorMessage ?? '分配角色失败，请重试')
      }
    },
    [assignRole]
  )

  const handleRevokeRole = React.useCallback(
    async (endUserId: string, roleId: string) => {
      const result = await revokeRole(endUserId, roleId)
      if (result.success) {
        toast.success('角色已撤销')
      } else {
        toast.error(result.errorMessage ?? '撤销角色失败，请重试')
      }
    },
    [revokeRole]
  )

  const handleAssignBundle = React.useCallback(
    async (endUserId: string, bundleId: string) => {
      const result = await assignBundle(endUserId, bundleId)
      if (result.success) {
        toast.success('权限包授权成功')
      } else {
        toast.error(result.errorMessage ?? '授权权限包失败，请重试')
      }
    },
    [assignBundle]
  )

  const handleRevokeBundle = React.useCallback(
    async (endUserId: string, bundleId: string) => {
      const result = await revokeBundle(endUserId, bundleId)
      if (result.success) {
        toast.success('权限包已撤销')
      } else {
        toast.error(result.errorMessage ?? '撤销权限包失败，请重试')
      }
    },
    [revokeBundle]
  )

  return (
    <div className="space-y-4">
      {/* Table */}
      {users.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
          <Users className="mb-3 size-10 text-muted-foreground/25" />
          <p className="text-sm font-medium text-foreground">暂无终端用户</p>
          <p className="mt-1 text-sm text-muted-foreground">
            终端用户注册后将显示在此列表
          </p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[180px] text-xs font-medium text-muted-foreground">
                  用户名
                </TableHead>
                <TableHead className="text-xs font-medium text-muted-foreground">
                  已分配角色
                </TableHead>
                <TableHead className="w-[160px] text-xs font-medium text-muted-foreground">
                  注册时间
                </TableHead>
                <TableHead className="w-[100px] text-right text-xs font-medium text-muted-foreground">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((user) => (
                <TableRow
                  key={user.id}
                  className={
                    selectedUser?.id === user.id
                      ? 'bg-muted/30 hover:bg-muted/50'
                      : 'hover:bg-muted/50'
                  }
                >
                  <TableCell className="text-foreground">
                    {user.username}
                  </TableCell>
                  <TableCell>
                    <RoleBadges roles={user.assignedRoles} />
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDateTime(user.createdAt)}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 px-2 text-xs"
                      onClick={() => handleManage(user)}
                    >
                      管理授权
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Right-side Sheet */}
      <UserDetailSheet
        user={selectedUser}
        open={sheetOpen}
        onOpenChange={handleSheetOpenChange}
        allRoles={roles}
        rolesLoading={rolesLoading}
        allBundles={bundles}
        bundlesLoading={bundlesLoading}
        effectivePermissions={effectivePermissions}
        effectiveLoading={effectiveLoading}
        projectSlug={projectSlug}
        onAssignRole={handleAssignRole}
        onRevokeRole={handleRevokeRole}
        onAssignBundle={handleAssignBundle}
        onRevokeBundle={handleRevokeBundle}
      />
    </div>
  )
}
