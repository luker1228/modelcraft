'use client'

// src/web/components/features/end-user-access/EndUserRoleAccessTable.tsx
// Project 级终端用户角色分配管理（EndUser Access Redesign）
// 数据来源：listProjectEndUserRoleUsers（新接口）
// 操作：assignEndUserRole / revokeEndUserRole
// 展示：一行 = 一个用户，角色列展示多个 Badge，可单独撤销

import { useState, useCallback, useMemo } from 'react'
import Link from 'next/link'
import { Plus, RefreshCw, Users, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/utils'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@web/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import { ScrollArea } from '@web/components/ui/scroll-area'
import {
  useProjectEndUserRoleUsers,
  type ProjectRoleUserEntry,
} from '@web/hooks/end-user-access/useProjectEndUserRoleUsers'

// ── Constants ─────────────────────────────────────────────────────────────────

/** 角色列最多内联展示的 Badge 数量，超出部分折叠为 "+N" */
const ROLES_INLINE_LIMIT = 3

// ── Types ─────────────────────────────────────────────────────────────────────

interface RoleItem {
  id: string
  name: string
  description?: string | null
  assignedAt: string
  assignmentId: string | undefined
}

interface GroupedRoleEntry {
  endUser: ProjectRoleUserEntry['endUser']
  roles: RoleItem[]
  earliestAssignedAt: string
}

interface RevokeTarget {
  endUser: ProjectRoleUserEntry['endUser']
  role: RoleItem
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmtDate(iso: string) {
  try { return new Date(iso).toLocaleDateString('zh-CN') } catch { return '-' }
}

function groupEntries(entries: ProjectRoleUserEntry[]): GroupedRoleEntry[] {
  const map = new Map<string, GroupedRoleEntry>()
  for (const e of entries) {
    const existing = map.get(e.endUser.id)
    if (existing) {
      existing.roles.push({
        id: e.role.id,
        name: e.role.name,
        description: e.role.description,
        assignedAt: e.assignedAt,
        assignmentId: e.assignmentId,
      })
      if (e.assignedAt < existing.earliestAssignedAt) {
        existing.earliestAssignedAt = e.assignedAt
      }
    } else {
      map.set(e.endUser.id, {
        endUser: e.endUser,
        roles: [{
          id: e.role.id,
          name: e.role.name,
          description: e.role.description,
          assignedAt: e.assignedAt,
          assignmentId: e.assignmentId,
        }],
        earliestAssignedAt: e.assignedAt,
      })
    }
  }
  return Array.from(map.values())
}

// ── RoleBadgeList ─────────────────────────────────────────────────────────────
// 内联展示前 N 个角色 Badge，超出部分通过 Popover 展开。
// 每个 Badge 右侧有 × 可单独撤销。

interface RoleBadgeListProps {
  roles: RoleItem[]
  onRevoke: (role: RoleItem) => void
}

function RoleBadge({ role, onRevoke }: { role: RoleItem; onRevoke: (role: RoleItem) => void }) {
  return (
    <Badge variant="secondary" className="flex shrink-0 items-center gap-1 pr-1">
      <span className="max-w-[120px] truncate">{role.name}</span>
      <button
        type="button"
        className="ml-0.5 rounded-sm opacity-60 transition-opacity hover:opacity-100 focus:outline-none"
        title={`撤销角色「${role.name}」`}
        onClick={() => onRevoke(role)}
      >
        <X className="size-3" />
      </button>
    </Badge>
  )
}

function RoleBadgeList({ roles, onRevoke }: RoleBadgeListProps) {
  const inline = roles.slice(0, ROLES_INLINE_LIMIT)
  const overflow = roles.slice(ROLES_INLINE_LIMIT)

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {inline.map((role) => (
        <RoleBadge key={role.id} role={role} onRevoke={onRevoke} />
      ))}

      {overflow.length > 0 && (
        <Popover>
          <PopoverTrigger asChild>
            <button
              type="button"
              className="inline-flex h-5 items-center rounded-full border border-border bg-muted px-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              +{overflow.length}
            </button>
          </PopoverTrigger>
          <PopoverContent className="w-64 p-0" align="start">
            <div className="border-b px-3 py-2">
              <p className="text-xs font-semibold text-muted-foreground">
                全部角色（{roles.length}）
              </p>
            </div>
            <ScrollArea className="max-h-64">
              <div className="space-y-px p-1">
                {roles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-muted"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium text-foreground">{role.name}</p>
                      {role.description && (
                        <p className="truncate text-xs text-muted-foreground">{role.description}</p>
                      )}
                    </div>
                    <button
                      type="button"
                      className="ml-2 shrink-0 rounded-sm p-0.5 text-muted-foreground opacity-60 transition-opacity hover:text-destructive hover:opacity-100"
                      title={`撤销角色「${role.name}」`}
                      onClick={() => onRevoke(role)}
                    >
                      <X className="size-3.5" />
                    </button>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </PopoverContent>
        </Popover>
      )}
    </div>
  )
}

// ── Assign Role Dialog ─────────────────────────────────────────────────────────
// 用于"添加用户"（新用户）和"添加角色"（已有用户追加角色）两种场景

interface AssignRoleDialogProps {
  open: boolean
  onClose: () => void
  orgUsers: Array<{ id: string; username: string; isForbidden: boolean }>
  availableRoles: Array<{ id: string; name: string; description?: string | null }>
  onConfirm: (endUserId: string, roleId: string) => Promise<void>
  // 若已有预选用户（追加角色场景），锁定用户选择器
  preselectedUser?: { id: string; username: string }
  // 该用户已拥有的角色 id，从可选角色中过滤掉
  assignedRoleIds?: string[]
  // 用于构造 RBAC 跳转链接
  orgName: string
  projectSlug: string
}

function AssignRoleDialog({
  open,
  onClose,
  orgUsers,
  availableRoles,
  onConfirm,
  preselectedUser,
  assignedRoleIds = [],
  orgName,
  projectSlug,
}: AssignRoleDialogProps) {
  const [selectedUserId, setSelectedUserId] = useState(preselectedUser?.id ?? '')
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [loading, setLoading] = useState(false)

  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setSelectedUserId(preselectedUser?.id ?? '')
      setSelectedRoleId('')
      onClose()
    }
  }

  // 每次打开时同步预选用户
  const filteredRoles = availableRoles.filter((r) => !assignedRoleIds.includes(r.id))
  const activeUsers = orgUsers.filter((u) => !u.isForbidden)
  const isAddingRole = !!preselectedUser
  const noRoles = filteredRoles.length === 0

  const handleConfirm = async () => {
    const userId = preselectedUser?.id ?? selectedUserId
    if (!userId || !selectedRoleId) return
    setLoading(true)
    try {
      await onConfirm(userId, selectedRoleId)
      setSelectedUserId(preselectedUser?.id ?? '')
      setSelectedRoleId('')
      onClose()
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{isAddingRole ? '追加角色' : '授权用户访问'}</DialogTitle>
          <DialogDescription>
            {isAddingRole
              ? `为 ${preselectedUser.username} 追加一个角色。`
              : '为 Org 用户分配本项目角色，授权后即可访问。'}
          </DialogDescription>
        </DialogHeader>

        {noRoles ? (
          <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
            {isAddingRole
              ? '该用户已拥有全部可用角色。'
              : (
                <>
                  当前项目尚无可用角色。请先前往{' '}
                  <Link
                    href={`/org/${orgName}/project/${projectSlug}/roles`}
                    className="font-medium underline underline-offset-2 hover:text-amber-900"
                    onClick={onClose}
                  >
                    RBAC 角色设置
                  </Link>
                  {' '}创建角色后再授权用户。
                </>
              )}
          </div>
        ) : (
          <div className="space-y-4 py-2">
            {/* 用户选择：追加角色场景锁定显示 */}
            {isAddingRole ? (
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">用户名</label>
                <div className="flex h-9 items-center rounded-md border bg-muted px-3 text-sm text-muted-foreground">
                  {preselectedUser.username}
                </div>
              </div>
            ) : (
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">用户名</label>
                <Select value={selectedUserId} onValueChange={setSelectedUserId}>
                  <SelectTrigger>
                    <SelectValue placeholder="请选择用户..." />
                  </SelectTrigger>
                  <SelectContent>
                    {activeUsers.length === 0 && (
                      <div className="px-3 py-2 text-sm text-muted-foreground">暂无可用用户</div>
                    )}
                    {activeUsers.map((u) => (
                      <SelectItem key={u.id} value={u.id}>{u.username}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">角色</label>
              <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
                <SelectTrigger>
                  <SelectValue placeholder="请选择角色..." />
                </SelectTrigger>
                <SelectContent>
                  {filteredRoles.map((r) => (
                    <SelectItem key={r.id} value={r.id}>
                      <div>
                        <span>{r.name}</span>
                        {r.description && (
                          <span className="ml-1.5 text-xs text-muted-foreground">{r.description}</span>
                        )}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>取消</Button>
          <Button
            disabled={noRoles || !(preselectedUser?.id ?? selectedUserId) || !selectedRoleId || loading}
            onClick={handleConfirm}
          >
            {loading ? '添加中...' : '确认'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Main Table Component ──────────────────────────────────────────────────────

interface EndUserRoleAccessTableProps {
  orgName: string
  projectSlug: string
}

export function EndUserRoleAccessTable({ orgName, projectSlug }: EndUserRoleAccessTableProps) {
  const {
    entries,
    loading,
    error,
    reload,
    orgUsers,
    availableRoles,
    assignRole,
    revokeRole,
  } = useProjectEndUserRoleUsers(orgName, projectSlug)

  const { pendingAction, setPendingAction } = useOnboarding()
  const highlightAssign = pendingAction === 'nav_assign_role'

  // 按用户聚合
  const groupedEntries = useMemo(() => groupEntries(entries), [entries])

  // 添加用户 / 添加角色 dialog
  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [addRoleForUser, setAddRoleForUser] = useState<{ id: string; username: string; assignedRoleIds: string[] } | null>(null)

  // 撤销单条角色 confirm
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget | null>(null)
  const [revoking, setRevoking] = useState(false)

  const handleAssignRole = useCallback(
    async (endUserId: string, roleId: string) => {
      const r = await assignRole(endUserId, roleId)
      if (r.success) {
        toast.success('角色已分配')
      } else {
        toast.error(r.errorMessage ?? '分配失败')
        throw new Error(r.errorMessage)
      }
    },
    [assignRole]
  )

  const handleRevoke = useCallback(async () => {
    if (!revokeTarget) return
    setRevoking(true)
    try {
      const r = await revokeRole(revokeTarget.endUser.id, revokeTarget.role.id)
      if (r.success) {
        toast.success('已撤销角色')
      } else {
        toast.error(r.errorMessage ?? '撤销失败')
      }
    } finally {
      setRevoking(false)
      setRevokeTarget(null)
    }
  }, [revokeTarget, revokeRole])

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground">
          {!loading && groupedEntries.length > 0 && (
            <span>共 {groupedEntries.length} 位用户</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => reload()} disabled={loading}>
            <RefreshCw className={`mr-1.5 size-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Button
            size="sm"
            className={cn(
              'bg-primary text-primary-foreground hover:bg-primary/90',
              highlightAssign && 'border-amber-400 bg-amber-50 text-amber-900 ring-2 ring-amber-400 ring-offset-1 animate-pulse hover:border-amber-500 hover:bg-amber-100'
            )}
            onClick={() => {
              if (highlightAssign) setPendingAction(null)
              setAddDialogOpen(true)
            }}
          >
            <Plus className="mr-1.5 size-4" />
            授权用户
          </Button>
        </div>
      </div>

      {/* Error state */}
      {error && !loading && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Users className="mb-3 size-10 text-muted-foreground/40" />
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      )}

      {/* Table */}
      {!error && (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">用户名</TableHead>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">角色</TableHead>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">首次授权</TableHead>
                <TableHead className="w-[120px] px-3 text-right text-sm font-semibold text-muted-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading && Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 4 }).map((__, j) => (
                    <TableCell key={j} className="px-3 py-2">
                      <Skeleton className="h-4 w-24" />
                    </TableCell>
                  ))}
                </TableRow>
              ))}

              {!loading && groupedEntries.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4}>
                    <div className="flex flex-col items-center justify-center py-14 text-center">
                      <Users className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1.5} />
                      <p className="text-sm font-semibold text-foreground">暂无授权记录</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        点击「授权用户」为 Org 用户分配角色，授权后用户即可访问本项目
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!loading && groupedEntries.map((group) => (
                <TableRow key={group.endUser.id} className="group">
                  {/* 用户名 */}
                  <TableCell className="px-3 py-2">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">{group.endUser.username}</span>
                      {group.endUser.isForbidden && (
                        <Badge variant="destructive" className="text-xs">已禁用</Badge>
                      )}
                    </div>
                  </TableCell>

                  {/* 角色列：最多内联 ROLES_INLINE_LIMIT 个，超出折叠为 "+N" Popover */}
                  <TableCell className="px-3 py-2">
                    <RoleBadgeList
                      roles={group.roles}
                      onRevoke={(role) => setRevokeTarget({ endUser: group.endUser, role })}
                    />
                  </TableCell>

                  {/* 首次授权时间 */}
                  <TableCell className="px-3 py-2 text-sm text-muted-foreground">
                    {fmtDate(group.earliestAssignedAt)}
                  </TableCell>

                  {/* 操作：追加角色 */}
                  <TableCell className="px-3 py-2 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 px-3 text-xs"
                      onClick={() =>
                        setAddRoleForUser({
                          id: group.endUser.id,
                          username: group.endUser.username,
                          assignedRoleIds: group.roles.map((r) => r.id),
                        })
                      }
                    >
                      <Plus className="mr-1 size-3" />
                      追加角色
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* 添加用户 dialog（新增用户场景） */}
      <AssignRoleDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        orgUsers={orgUsers}
        availableRoles={availableRoles}
        onConfirm={handleAssignRole}
        orgName={orgName}
        projectSlug={projectSlug}
      />

      {/* 追加角色 dialog（已有用户追加场景） */}
      <AssignRoleDialog
        open={!!addRoleForUser}
        onClose={() => setAddRoleForUser(null)}
        orgUsers={orgUsers}
        availableRoles={availableRoles}
        onConfirm={handleAssignRole}
        preselectedUser={addRoleForUser ?? undefined}
        assignedRoleIds={addRoleForUser?.assignedRoleIds}
        orgName={orgName}
        projectSlug={projectSlug}
      />

      {/* 撤销角色确认 */}
      <AlertDialog open={!!revokeTarget} onOpenChange={(o) => { if (!o) setRevokeTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认撤销角色</AlertDialogTitle>
            <AlertDialogDescription>
              确认撤销 <strong>{revokeTarget?.endUser.username}</strong> 的
              「<strong>{revokeTarget?.role.name}</strong>」角色？
              撤销后若该用户在本项目无其他角色，将无法访问本项目。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleRevoke} disabled={revoking}>
              {revoking ? '撤销中...' : '确认撤销'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
