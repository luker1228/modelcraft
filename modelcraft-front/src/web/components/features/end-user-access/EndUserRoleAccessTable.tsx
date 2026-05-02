'use client'

// src/web/components/features/end-user-access/EndUserRoleAccessTable.tsx
// Project 级终端用户角色分配管理（EndUser Access Redesign）
// 数据来源：listProjectEndUserRoleUsers（新接口）
// 操作：assignEndUserRole / revokeEndUserRole

import { useState, useCallback } from 'react'
import { Plus, RefreshCw, Users } from 'lucide-react'
import { toast } from 'sonner'
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
  useProjectEndUserRoleUsers,
  type ProjectRoleUserEntry,
} from '@web/hooks/end-user-access/useProjectEndUserRoleUsers'

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmtDate(iso: string) {
  try { return new Date(iso).toLocaleDateString('zh-CN') } catch { return '-' }
}

// ── Add User Dialog ──────────────────────────────────────────────────────────

interface AddUserDialogProps {
  open: boolean
  onClose: () => void
  orgUsers: Array<{ id: string; username: string; isForbidden: boolean }>
  availableRoles: Array<{ id: string; name: string; description?: string | null }>
  onConfirm: (endUserId: string, roleId: string) => Promise<void>
}

function AddUserDialog({ open, onClose, orgUsers, availableRoles, onConfirm }: AddUserDialogProps) {
  const [selectedUserId, setSelectedUserId] = useState('')
  const [selectedRoleId, setSelectedRoleId] = useState(() => availableRoles[0]?.id ?? '')
  const [loading, setLoading] = useState(false)

  // Reset when dialog opens
  const handleOpenChange = (o: boolean) => {
    if (!o) {
      setSelectedUserId('')
      setSelectedRoleId(availableRoles[0]?.id ?? '')
      onClose()
    }
  }

  const handleConfirm = async () => {
    if (!selectedUserId || !selectedRoleId) return
    setLoading(true)
    try {
      await onConfirm(selectedUserId, selectedRoleId)
      setSelectedUserId('')
      setSelectedRoleId(availableRoles[0]?.id ?? '')
      onClose()
    } finally {
      setLoading(false)
    }
  }

  const noRoles = availableRoles.length === 0
  const activeUsers = orgUsers.filter((u) => !u.isForbidden)

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>添加终端用户</DialogTitle>
          <DialogDescription>
            从 Org 用户列表中选择用户并分配角色。分配成功后用户即可访问本项目。
          </DialogDescription>
        </DialogHeader>

        {noRoles ? (
          <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
            当前项目尚无可用角色。请先在 RBAC 设置页创建角色后再添加用户。
          </div>
        ) : (
          <div className="space-y-4 py-2">
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

            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">角色</label>
              <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
                <SelectTrigger>
                  <SelectValue placeholder="请选择角色..." />
                </SelectTrigger>
                <SelectContent>
                  {availableRoles.map((r) => (
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
            disabled={noRoles || !selectedUserId || !selectedRoleId || loading}
            onClick={handleConfirm}
          >
            {loading ? '添加中...' : '确认'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Change Role Dialog ────────────────────────────────────────────────────────

interface ChangeRoleDialogProps {
  open: boolean
  entry: ProjectRoleUserEntry | null
  availableRoles: Array<{ id: string; name: string }>
  onClose: () => void
  onConfirm: (endUserId: string, oldRoleId: string, newRoleId: string) => Promise<void>
}

function ChangeRoleDialog({ open, entry, availableRoles, onClose, onConfirm }: ChangeRoleDialogProps) {
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [loading, setLoading] = useState(false)

  const handleConfirm = async () => {
    if (!entry || !selectedRoleId) return
    setLoading(true)
    try {
      await onConfirm(entry.endUser.id, entry.role.id, selectedRoleId)
      setSelectedRoleId('')
      onClose()
    } finally {
      setLoading(false)
    }
  }

  const otherRoles = availableRoles.filter((r) => r.id !== entry?.role.id)

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) { setSelectedRoleId(''); onClose() } }}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>修改角色</DialogTitle>
          <DialogDescription>
            为 <strong>{entry?.endUser.username}</strong> 修改角色（将撤销旧角色并分配新角色）
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-1.5 py-2">
          <label className="text-sm font-medium text-foreground">新角色</label>
          <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
            <SelectTrigger>
              <SelectValue placeholder="选择新角色..." />
            </SelectTrigger>
            <SelectContent>
              {otherRoles.map((r) => (
                <SelectItem key={r.id} value={r.id}>{r.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>取消</Button>
          <Button disabled={!selectedRoleId || loading} onClick={handleConfirm}>
            {loading ? '修改中...' : '确认'}
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

  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [changeRoleEntry, setChangeRoleEntry] = useState<ProjectRoleUserEntry | null>(null)
  const [revokeEntry, setRevokeEntry] = useState<ProjectRoleUserEntry | null>(null)
  const [revoking, setRevoking] = useState(false)

  const handleAddUser = useCallback(
    async (endUserId: string, roleId: string) => {
      const r = await assignRole(endUserId, roleId)
      if (r.success) {
        toast.success('用户已添加')
      } else {
        toast.error(r.errorMessage ?? '添加失败')
        throw new Error(r.errorMessage)
      }
    },
    [assignRole]
  )

  const handleChangeRole = useCallback(
    async (endUserId: string, oldRoleId: string, newRoleId: string) => {
      const revokeResult = await revokeRole(endUserId, oldRoleId)
      if (!revokeResult.success) {
        toast.error(revokeResult.errorMessage ?? '撤销旧角色失败')
        throw new Error(revokeResult.errorMessage)
      }
      const assignResult = await assignRole(endUserId, newRoleId)
      if (!assignResult.success) {
        toast.error(assignResult.errorMessage ?? '分配新角色失败')
        throw new Error(assignResult.errorMessage)
      }
      toast.success('角色已修改')
    },
    [revokeRole, assignRole]
  )

  const handleRevoke = useCallback(async () => {
    if (!revokeEntry) return
    setRevoking(true)
    try {
      const r = await revokeRole(revokeEntry.endUser.id, revokeEntry.role.id)
      if (r.success) {
        toast.success('已撤销角色分配')
      } else {
        toast.error(r.errorMessage ?? '撤销失败')
      }
    } finally {
      setRevoking(false)
      setRevokeEntry(null)
    }
  }, [revokeEntry, revokeRole])

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground">
          {!loading && entries.length > 0 && (
            <span>共 {entries.length} 条角色分配</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => reload()} disabled={loading}>
            <RefreshCw className={`mr-1.5 size-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Button
            size="sm"
            className="bg-primary text-primary-foreground hover:bg-primary/90"
            onClick={() => setAddDialogOpen(true)}
          >
            <Plus className="mr-1.5 size-4" />
            添加用户
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
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">用户名</TableHead>
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">角色</TableHead>
                <TableHead className="px-3 font-semibold text-sm text-muted-foreground">授权时间</TableHead>
                <TableHead className="w-[180px] px-3 text-right font-semibold text-sm text-muted-foreground">操作</TableHead>
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

              {!loading && entries.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4}>
                    <div className="flex flex-col items-center justify-center py-14 text-center">
                      <Users className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1.5} />
                      <p className="text-sm font-semibold text-foreground">暂无角色分配</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        点击「添加用户」为 Org 用户分配角色，授权后用户即可访问本项目
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!loading && entries.map((entry) => (
                <TableRow key={entry.assignmentId}>
                  {/* 用户名 */}
                  <TableCell className="px-3 py-2">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">{entry.endUser.username}</span>
                      {entry.endUser.isForbidden && (
                        <Badge variant="destructive" className="text-xs">已禁用</Badge>
                      )}
                    </div>
                  </TableCell>

                  {/* 角色 */}
                  <TableCell className="px-3 py-2">
                    <div>
                      <span className="text-sm text-foreground">{entry.role.name}</span>
                      {entry.role.description && (
                        <span className="ml-1.5 text-xs text-muted-foreground">{entry.role.description}</span>
                      )}
                    </div>
                  </TableCell>

                  {/* 授权时间 */}
                  <TableCell className="px-3 py-2 text-sm text-muted-foreground">
                    {fmtDate(entry.assignedAt)}
                  </TableCell>

                  {/* 操作 */}
                  <TableCell className="px-3 py-2 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 px-3 text-xs"
                        onClick={() => setChangeRoleEntry(entry)}
                      >
                        修改角色
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 px-3 text-xs text-destructive hover:text-destructive"
                        onClick={() => setRevokeEntry(entry)}
                      >
                        撤销
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Dialogs */}
      <AddUserDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        orgUsers={orgUsers}
        availableRoles={availableRoles}
        onConfirm={handleAddUser}
      />

      <ChangeRoleDialog
        open={!!changeRoleEntry}
        entry={changeRoleEntry}
        availableRoles={availableRoles}
        onClose={() => setChangeRoleEntry(null)}
        onConfirm={handleChangeRole}
      />

      <AlertDialog open={!!revokeEntry} onOpenChange={(o) => { if (!o) setRevokeEntry(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认撤销</AlertDialogTitle>
            <AlertDialogDescription>
              确认撤销 <strong>{revokeEntry?.endUser.username}</strong> 的
              <strong>{revokeEntry?.role.name}</strong> 角色分配？
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
