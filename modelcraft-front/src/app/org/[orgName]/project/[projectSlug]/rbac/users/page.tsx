'use client'

import * as React from 'react'
import { useParams } from 'next/navigation'
import { toast } from 'sonner'
import {
  Users,
  ChevronDown,
  ChevronRight,
  ShieldCheck,
  Lock,
} from 'lucide-react'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
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
} from './_hooks/useUserAuth'
import type {
  EndUserRole,
  EndUserPermissionBundle,
  EndUserPermissionAction,
  EndUserRowScope,
  EffectivePermissions,
} from '@/types'

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
  /** 最多展示几个，超出显示 "+N" */
  maxVisible?: number
}

function RoleBadges({ roles, maxVisible = 2 }: RoleBadgesProps) {
  const explicitRoles = roles.filter((r) => !r.isImplicit)

  if (explicitRoles.length === 0) {
    return <span className="text-sm text-muted-foreground/60 italic">暂无角色</span>
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
            <TableHead className="w-[180px]">用户名</TableHead>
            <TableHead>已分配角色</TableHead>
            <TableHead className="w-[160px]">注册时间</TableHead>
            <TableHead className="w-[100px] text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 3 }).map((_, i) => (
            <TableRow key={i}>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-4 w-36" /></TableCell>
              <TableCell className="text-right"><Skeleton className="h-7 w-16 ml-auto" /></TableCell>
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

  // 可分配的角色：排除隐式角色 + 已分配角色
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

  // 合并展示：隐式角色（不可撤销）+ 显式角色
  const implicitRoles = allRoles.filter((r) => r.isImplicit)
  const explicitAssigned = user.assignedRoles.filter((r) => !r.isImplicit)

  return (
    <div className="space-y-4 py-1">
      {/* 已分配角色列表 */}
      <div className="space-y-2">
        {implicitRoles.map((role) => (
          <div
            key={role.id}
            className="flex items-center justify-between rounded-md border border-border bg-muted/30 px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="truncate text-sm font-semibold text-foreground">
                {role.name}
              </span>
              {role.description && (
                <p className="truncate text-xs text-muted-foreground">{role.description}</p>
              )}
            </div>
            <Badge variant="outline" className="ml-3 shrink-0 text-xs text-muted-foreground">
              隐式角色
            </Badge>
          </div>
        ))}

        {explicitAssigned.map((role) => (
          <div
            key={role.id}
            className="flex items-center justify-between rounded-md border border-border px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="truncate text-sm font-semibold text-foreground">
                {role.name}
              </span>
              {role.description && (
                <p className="truncate text-xs text-muted-foreground">{role.description}</p>
              )}
            </div>
            <Button
              variant="outline"
              size="sm"
              className="ml-3 h-7 shrink-0 px-2 text-xs text-destructive hover:text-destructive"
              disabled={revokingId === role.id}
              onClick={() => handleRevoke(role.id)}
            >
              {revokingId === role.id ? '撤销中...' : '撤销'}
            </Button>
          </div>
        ))}

        {implicitRoles.length === 0 && explicitAssigned.length === 0 && (
          <p className="text-center text-sm text-muted-foreground py-4">
            该用户暂未分配任何角色
          </p>
        )}
      </div>

      {/* 分配角色行 */}
      {assignableRoles.length > 0 && (
        <div className="flex items-center gap-2 border-t border-border pt-4">
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
            className="h-8 shrink-0"
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
            <p className="text-center text-sm text-muted-foreground py-4">
              所有权限包已授权，无可添加项
            </p>
          ) : (
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
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
    <div className="space-y-4 py-1">
      <div className="space-y-2">
        {user.assignedBundles.map((bundle) => (
          <div
            key={bundle.id}
            className="flex items-center justify-between rounded-md border border-border px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="truncate text-sm font-semibold text-foreground">
                {bundle.name}
              </span>
              {bundle.description && (
                <p className="truncate text-xs text-muted-foreground">{bundle.description}</p>
              )}
            </div>
            <Button
              variant="outline"
              size="sm"
              className="ml-3 h-7 shrink-0 px-2 text-xs text-destructive hover:text-destructive"
              disabled={revokingId === bundle.id}
              onClick={() => handleRevoke(bundle.id)}
            >
              {revokingId === bundle.id ? '撤销中...' : '撤销'}
            </Button>
          </div>
        ))}

        {user.assignedBundles.length === 0 && (
          <p className="text-center text-sm text-muted-foreground py-4">
            该用户暂未直接授权任何权限包
          </p>
        )}
      </div>

      <div className="border-t border-border pt-4">
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

  // Flatten all grants from this modelId (normally one entry per modelId but normalise just in case)
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
          <span className="text-sm font-semibold text-foreground">{modelId}</span>
          <Badge variant="secondary" className="text-xs">
            {grants.length} 条授权
          </Badge>
        </div>
      </button>

      {open && (
        <div className="border-t border-border bg-muted/10 px-4 py-3 space-y-2">
          {grants.length === 0 ? (
            <p className="text-xs text-muted-foreground">该模型无有效授权</p>
          ) : (
            grants.map((grant, idx) => (
              <div
                key={idx}
                className="flex flex-wrap items-center gap-2"
              >
                <Badge variant="outline" className="text-xs">
                  {ACTION_LABEL[grant.action] ?? grant.action}
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  {ROW_SCOPE_LABEL[grant.rowScope] ?? grant.rowScope}
                </Badge>
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

  // Filter out "empty" entries (backend can return a placeholder with empty modelId)
  const nonEmpty = effectivePermissions.filter(
    (ep) => ep.modelId && ep.grants.length > 0
  )

  if (nonEmpty.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-10 text-center">
        <Lock className="mb-3 size-8 text-muted-foreground/40" />
        <p className="text-sm font-semibold text-foreground">该用户暂无任何权限</p>
        <p className="mt-1 text-xs text-muted-foreground">
          通过「角色」或「直接权限包」为该用户授权后，有效权限将在此展示
        </p>
      </div>
    )
  }

  // Group by modelId
  const grouped = nonEmpty.reduce<Record<string, EffectivePermissions[]>>((acc, ep) => {
    if (!acc[ep.modelId]) acc[ep.modelId] = []
    acc[ep.modelId].push(ep)
    return acc
  }, {})

  return (
    <div className="space-y-3 py-1">
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
        className="flex w-[480px] flex-col sm:max-w-[480px]"
      >
        <SheetHeader className="border-b border-border pb-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-primary" />
            <SheetTitle className="text-lg font-semibold">
              {user.username}
            </SheetTitle>
          </div>
          <p className="text-xs text-muted-foreground">
            注册时间：{formatDateTime(user.createdAt)}
          </p>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto py-4">
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

            {/* ── Tab 1: 角色 ── */}
            <TabsContent value="roles" className="mt-4">
              <RoleTab
                user={user}
                allRoles={allRoles}
                rolesLoading={rolesLoading}
                onAssignRole={handleAssignRole}
                onRevokeRole={handleRevokeRole}
              />
            </TabsContent>

            {/* ── Tab 2: 直接授权权限包 ── */}
            <TabsContent value="bundles" className="mt-4">
              <BundleTab
                user={user}
                allBundles={allBundles}
                bundlesLoading={bundlesLoading}
                onAssignBundle={handleAssignBundle}
                onRevokeBundle={handleRevokeBundle}
              />
            </TabsContent>

            {/* ── Tab 3: 有效权限 ── */}
            <TabsContent value="effective" className="mt-4">
              <EffectivePermissionsTab
                effectivePermissions={effectivePermissions}
                loading={effectiveLoading}
              />
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  )
}

// ── UserAuthPage ──────────────────────────────────────────────────────────────

export default function UserAuthPage() {
  const { orgName, projectSlug } = useParams<{
    orgName: string
    projectSlug: string
  }>()

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

  // 打开 Sheet
  const handleManage = React.useCallback(
    (user: MockEndUserUser) => {
      setSelectedUser(user)
      setSheetOpen(true)
    },
    [setSelectedUser]
  )

  // Sheet 关闭时清空选中用户
  const handleSheetOpenChange = React.useCallback(
    (open: boolean) => {
      setSheetOpen(open)
      if (!open) setSelectedUser(null)
    },
    [setSelectedUser]
  )

  // Mutation wrappers with toast feedback
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
    <>

      {/* Header */}
      <section className="mb-8 space-y-1">
          <h2 className="text-2xl font-semibold tracking-tight">用户授权</h2>
          <p className="text-sm text-muted-foreground">
            为终端用户分配角色或直接授予权限包，管理其在该项目中的访问权限。
          </p>
        </section>

        {/* Table */}
        {users.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-16">
            <Users className="mb-3 size-10 text-muted-foreground/40" />
            <p className="text-sm font-semibold text-foreground">暂无终端用户</p>
            <p className="mt-1 text-sm text-muted-foreground">
              终端用户注册后将显示在此列表
            </p>
          </div>
        ) : (
          <div className="overflow-hidden rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[180px]">用户名</TableHead>
                  <TableHead>已分配角色</TableHead>
                  <TableHead className="w-[160px]">注册时间</TableHead>
                  <TableHead className="w-[100px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow
                    key={user.id}
                    className={
                      selectedUser?.id === user.id
                        ? 'bg-muted/30'
                        : undefined
                    }
                  >
                    <TableCell className="font-semibold text-foreground">
                      {user.username}
                    </TableCell>
                    <TableCell>
                      <RoleBadges roles={user.assignedRoles} />
                    </TableCell>
                    <TableCell className="text-muted-foreground">
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
    </>
  )
}
