'use client'

// src/web/components/features/end-user-access/EndUserRoleAccessTable.tsx
// Project 级终端用户角色分配管理（EndUser Access Redesign）

import { useState, useCallback, useMemo, useEffect } from 'react'
import Link from 'next/link'
import { Plus, RefreshCw, Users, X, UserPlus, ChevronLeft, Loader2, Eye, EyeOff } from 'lucide-react'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { cn } from '@/shared/utils'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
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
  type CreateOrgEndUserPayload,
} from '@web/hooks/end-user-access/useProjectEndUserRoleUsers'

// ── Constants ─────────────────────────────────────────────────────────────────

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
  try {
    return new Date(iso).toLocaleDateString('zh-CN')
  } catch {
    return '-'
  }
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
        roles: [
          {
            id: e.role.id,
            name: e.role.name,
            description: e.role.description,
            assignedAt: e.assignedAt,
            assignmentId: e.assignmentId,
          },
        ],
        earliestAssignedAt: e.assignedAt,
      })
    }
  }
  return Array.from(map.values())
}

// ── RoleBadge & RoleBadgeList ─────────────────────────────────────────────────

function RoleBadge({
  role,
  onRevoke,
}: {
  role: RoleItem
  onRevoke: (role: RoleItem) => void
}) {
  return (
    <Badge variant="secondary" className="flex shrink-0 items-center gap-1 pr-1">
      <span className="max-w-[120px] truncate">{role.name}</span>
      <button
        type="button"
        className="ml-0.5 rounded-sm opacity-60 transition-opacity hover:opacity-100 focus:outline-none"
        title={'撤销角色' + role.name}
        onClick={() => onRevoke(role)}
      >
        <X className="size-3" />
      </button>
    </Badge>
  )
}

function RoleBadgeList({
  roles,
  onRevoke,
}: {
  roles: RoleItem[]
  onRevoke: (role: RoleItem) => void
}) {
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
                        <p className="truncate text-xs text-muted-foreground">
                          {role.description}
                        </p>
                      )}
                    </div>
                    <button
                      type="button"
                      className="ml-2 shrink-0 rounded-sm p-0.5 text-muted-foreground opacity-60 transition-opacity hover:text-destructive hover:opacity-100"
                      title={'撤销角色' + role.name}
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

// ── CreateAndAssignDialog ─────────────────────────────────────────────────────
// Step 1: 选择已有用户 OR 新建用户
// Step 2: 选角色并确认授权
// preselectedUser: 追加角色场景，直接跳 Step 2

const createUserSchema = z
  .object({
    username: z.string().min(1, '请输入用户名').max(64),
    phone: z.string().regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号'),
    password: z.string().min(8, '密码至少 8 位').max(128),
    confirmPassword: z.string().min(1, '请再次输入密码'),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })

type CreateUserFormValues = z.infer<typeof createUserSchema>
type DialogMode = 'select' | 'create'

interface CreateAndAssignDialogProps {
  open: boolean
  onClose: () => void
  orgUsers: Array<{ id: string; username: string; isForbidden: boolean }>
  availableRoles: Array<{ id: string; name: string; description?: string | null }>
  onConfirm: (endUserId: string, roleId: string) => Promise<void>
  onCreateUser: (
    payload: CreateOrgEndUserPayload
  ) => Promise<{
    success: boolean
    endUser?: { id: string; username: string }
    errorMessage?: string
  }>
  preselectedUser?: { id: string; username: string; assignedRoleIds: string[] }
  orgName: string
  projectSlug: string
}

function CreateAndAssignDialog({
  open,
  onClose,
  orgUsers,
  availableRoles,
  onConfirm,
  onCreateUser,
  preselectedUser,
  orgName,
  projectSlug,
}: CreateAndAssignDialogProps) {
  const [mode, setMode] = useState<DialogMode>('select')
  const [selectedUserId, setSelectedUserId] = useState(preselectedUser?.id ?? '')
  const [createError, setCreateError] = useState<string | null>(null)
  const [creatingUser, setCreatingUser] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  const [step, setStep] = useState<1 | 2>(1)
  const [resolvedUser, setResolvedUser] = useState<{ id: string; username: string } | null>(
    preselectedUser ? { id: preselectedUser.id, username: preselectedUser.username } : null
  )
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [assigning, setAssigning] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateUserFormValues>({ resolver: zodResolver(createUserSchema) })

  useEffect(() => {
    if (open && preselectedUser) {
      setStep(2)
      setResolvedUser({ id: preselectedUser.id, username: preselectedUser.username })
    }
  }, [open, preselectedUser])

  const handleClose = () => {
    setMode('select')
    setStep(1)
    setSelectedUserId(preselectedUser?.id ?? '')
    setResolvedUser(
      preselectedUser
        ? { id: preselectedUser.id, username: preselectedUser.username }
        : null
    )
    setSelectedRoleId('')
    setCreateError(null)
    setCreatingUser(false)
    setAssigning(false)
    setShowPassword(false)
    setShowConfirmPassword(false)
    reset()
    onClose()
  }

  const handleSelectExistingNext = () => {
    if (!selectedUserId) return
    const found = orgUsers.find((u) => u.id === selectedUserId)
    if (!found) return
    setResolvedUser({ id: found.id, username: found.username })
    setStep(2)
  }

  const handleCreateUserSubmit = handleSubmit(async (values) => {
    setCreatingUser(true)
    setCreateError(null)
    try {
      const result = await onCreateUser({
        username: values.username,
        phone: values.phone,
        password: values.password,
      })
      if (!result.success || !result.endUser) {
        setCreateError(result.errorMessage ?? '创建失败，请重试')
        return
      }
      setResolvedUser(result.endUser)
      reset()
      setStep(2)
    } finally {
      setCreatingUser(false)
    }
  })

  const handleAssign = async () => {
    if (!resolvedUser || !selectedRoleId) return
    setAssigning(true)
    try {
      await onConfirm(resolvedUser.id, selectedRoleId)
      handleClose()
    } finally {
      setAssigning(false)
    }
  }

  const assignedRoleIds = preselectedUser?.assignedRoleIds ?? []
  const filteredRoles = availableRoles.filter((r) => !assignedRoleIds.includes(r.id))
  const activeUsers = orgUsers.filter((u) => !u.isForbidden)
  const noRoles = filteredRoles.length === 0

  const stepTitle =
    preselectedUser
      ? '追加角色'
      : step === 1
        ? '授权用户访问'
        : '为 ' + (resolvedUser?.username ?? '') + ' 分配角色'

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose() }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {step === 2 && !preselectedUser && (
              <button
                type="button"
                onClick={() => {
                  setStep(1)
                  setResolvedUser(null)
                  setSelectedRoleId('')
                }}
                className="rounded-sm p-0.5 text-muted-foreground hover:text-foreground"
                aria-label="返回上一步"
              >
                <ChevronLeft className="size-4" />
              </button>
            )}
            {stepTitle}
          </DialogTitle>
          {step === 1 && (
            <DialogDescription>
              选择已有用户，或新建一个用户后立即授权。
            </DialogDescription>
          )}
        </DialogHeader>

        {step === 1 && (
          <div className="flex flex-col gap-4">
            <div className="flex rounded-md border border-border bg-muted/40 p-0.5">
              <button
                type="button"
                className={cn(
                  'flex-1 rounded-sm px-3 py-1.5 text-sm font-medium transition-colors',
                  mode === 'select'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                )}
                onClick={() => { setMode('select'); setCreateError(null) }}
              >
                选择已有用户
              </button>
              <button
                type="button"
                className={cn(
                  'flex flex-1 items-center justify-center gap-1.5 rounded-sm px-3 py-1.5 text-sm font-medium transition-colors',
                  mode === 'create'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                )}
                onClick={() => { setMode('create'); setCreateError(null) }}
              >
                <UserPlus className="size-3.5" />
                新建用户
              </button>
            </div>

            {mode === 'select' && (
              <>
                <div className="space-y-1.5">
                  <Label>用户名</Label>
                  <Select value={selectedUserId} onValueChange={setSelectedUserId}>
                    <SelectTrigger>
                      <SelectValue placeholder="请选择用户..." />
                    </SelectTrigger>
                    <SelectContent>
                      {activeUsers.length === 0 && (
                        <div className="px-3 py-2 text-sm text-muted-foreground">暂无可用用户</div>
                      )}
                      {activeUsers.map((u) => (
                        <SelectItem key={u.id} value={u.id}>
                          {u.username}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={handleClose}>
                    取消
                  </Button>
                  <Button disabled={!selectedUserId} onClick={handleSelectExistingNext}>
                    下一步
                  </Button>
                </DialogFooter>
              </>
            )}

            {mode === 'create' && (
              <form onSubmit={handleCreateUserSubmit} className="flex flex-col gap-3">
                {createError && (
                  <Alert variant="destructive">
                    <AlertDescription>{createError}</AlertDescription>
                  </Alert>
                )}
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="new-eu-username">用户名</Label>
                  <Input
                    id="new-eu-username"
                    disabled={creatingUser}
                    autoComplete="username"
                    {...register('username')}
                  />
                  {errors.username && (
                    <p className="text-xs text-destructive">{errors.username.message}</p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="new-eu-phone">手机号</Label>
                  <Input
                    id="new-eu-phone"
                    type="tel"
                    placeholder="11 位手机号"
                    disabled={creatingUser}
                    {...register('phone')}
                  />
                  {errors.phone && (
                    <p className="text-xs text-destructive">{errors.phone.message}</p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="new-eu-password">密码</Label>
                  <div className="relative">
                    <Input
                      id="new-eu-password"
                      type={showPassword ? 'text' : 'password'}
                      disabled={creatingUser}
                      className="pr-10"
                      autoComplete="new-password"
                      {...register('password')}
                    />
                    <button
                      type="button"
                      tabIndex={-1}
                      onClick={() => setShowPassword((v) => !v)}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showPassword ? (
                        <Eye className="size-4" strokeWidth={1.5} />
                      ) : (
                        <EyeOff className="size-4" strokeWidth={1.5} />
                      )}
                    </button>
                  </div>
                  {errors.password && (
                    <p className="text-xs text-destructive">{errors.password.message}</p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="new-eu-confirm">确认密码</Label>
                  <div className="relative">
                    <Input
                      id="new-eu-confirm"
                      type={showConfirmPassword ? 'text' : 'password'}
                      disabled={creatingUser}
                      className="pr-10"
                      autoComplete="new-password"
                      {...register('confirmPassword')}
                    />
                    <button
                      type="button"
                      tabIndex={-1}
                      onClick={() => setShowConfirmPassword((v) => !v)}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showConfirmPassword ? (
                        <Eye className="size-4" strokeWidth={1.5} />
                      ) : (
                        <EyeOff className="size-4" strokeWidth={1.5} />
                      )}
                    </button>
                  </div>
                  {errors.confirmPassword && (
                    <p className="text-xs text-destructive">{errors.confirmPassword.message}</p>
                  )}
                </div>
                <DialogFooter className="mt-1">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleClose}
                    disabled={creatingUser}
                  >
                    取消
                  </Button>
                  <Button type="submit" disabled={creatingUser}>
                    {creatingUser && <Loader2 className="mr-2 size-4 animate-spin" />}
                    创建并继续
                  </Button>
                </DialogFooter>
              </form>
            )}
          </div>
        )}

        {step === 2 && (
          <div className="flex flex-col gap-4">
            {noRoles ? (
              <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
                当前项目尚无可用角色。请先前往{' '}
                <Link
                  href={'/org/' + orgName + '/project/' + projectSlug + '/access-control?tab=roles'}
                  className="font-medium underline underline-offset-2 hover:text-amber-900"
                  onClick={handleClose}
                >
                  RBAC 角色设置
                </Link>
                {' '}创建角色后再授权用户。
              </div>
            ) : (
              <div className="space-y-1.5">
                <Label>分配角色</Label>
                <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
                  <SelectTrigger>
                    <SelectValue placeholder="请选择角色..." />
                  </SelectTrigger>
                  <SelectContent>
                    {filteredRoles.map((r) => (
                      <SelectItem key={r.id} value={r.id}>
                        <span>{r.name}</span>
                        {r.description && (
                          <span className="ml-1.5 text-xs text-muted-foreground">
                            {r.description}
                          </span>
                        )}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <DialogFooter>
              <Button variant="outline" onClick={handleClose}>
                取消
              </Button>
              <Button
                disabled={noRoles || !selectedRoleId || assigning}
                onClick={handleAssign}
              >
                {assigning && <Loader2 className="mr-2 size-4 animate-spin" />}
                {assigning ? '授权中...' : '确认授权'}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

// ── Main Table Component ──────────────────────────────────────────────────────

interface EndUserRoleAccessTableProps {
  orgName: string
  projectSlug: string
  autoAssign?: boolean
}

export function EndUserRoleAccessTable({
  orgName,
  projectSlug,
  autoAssign,
}: EndUserRoleAccessTableProps) {
  const {
    entries,
    loading,
    error,
    reload,
    orgUsers,
    availableRoles,
    assignRole,
    revokeRole,
    createOrgEndUser,
  } = useProjectEndUserRoleUsers(orgName, projectSlug)

  const { pendingAction, setPendingAction } = useOnboarding()
  const highlightAssign = pendingAction === 'nav_assign_role'

  const groupedEntries = useMemo(() => groupEntries(entries), [entries])

  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [addRoleForUser, setAddRoleForUser] = useState<{
    id: string
    username: string
    assignedRoleIds: string[]
  } | null>(null)
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget | null>(null)
  const [revoking, setRevoking] = useState(false)

  useEffect(() => {
    if (autoAssign) setAddDialogOpen(true)
  }, [autoAssign])

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
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground">
          {!loading && groupedEntries.length > 0 && (
            <span>共 {groupedEntries.length} 位用户</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => reload()} disabled={loading}>
            <RefreshCw className={cn('mr-1.5 size-4', loading && 'animate-spin')} />
            刷新
          </Button>
          <Button
            size="sm"
            className={cn(
              'bg-primary text-primary-foreground hover:bg-primary/90',
              highlightAssign &&
                'border-amber-400 bg-amber-50 text-amber-900 ring-2 ring-amber-400 ring-offset-1 animate-pulse hover:border-amber-500 hover:bg-amber-100'
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

      {error && !loading && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Users className="mb-3 size-10 text-muted-foreground/40" />
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      )}

      {!error && (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">
                  用户名
                </TableHead>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">
                  角色
                </TableHead>
                <TableHead className="px-3 text-sm font-semibold text-muted-foreground">
                  首次授权
                </TableHead>
                <TableHead className="w-[120px] px-3 text-right text-sm font-semibold text-muted-foreground">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading &&
                Array.from({ length: 3 }).map((_, i) => (
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
                      <Users
                        className="mb-3 size-9 text-muted-foreground/30"
                        strokeWidth={1.5}
                      />
                      <p className="text-sm font-semibold text-foreground">暂无授权记录</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        点击「授权用户」为 Org 用户分配角色，授权后用户即可访问本项目
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!loading &&
                groupedEntries.map((group) => (
                  <TableRow key={group.endUser.id} className="group">
                    <TableCell className="px-3 py-2">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-foreground">
                          {group.endUser.username}
                        </span>
                        {group.endUser.isForbidden && (
                          <Badge variant="destructive" className="text-xs">
                            已禁用
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="px-3 py-2">
                      <RoleBadgeList
                        roles={group.roles}
                        onRevoke={(role) =>
                          setRevokeTarget({ endUser: group.endUser, role })
                        }
                      />
                    </TableCell>
                    <TableCell className="px-3 py-2 text-sm text-muted-foreground">
                      {fmtDate(group.earliestAssignedAt)}
                    </TableCell>
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

      <CreateAndAssignDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        orgUsers={orgUsers}
        availableRoles={availableRoles}
        onConfirm={handleAssignRole}
        onCreateUser={createOrgEndUser}
        orgName={orgName}
        projectSlug={projectSlug}
      />

      <CreateAndAssignDialog
        open={!!addRoleForUser}
        onClose={() => setAddRoleForUser(null)}
        orgUsers={orgUsers}
        availableRoles={availableRoles}
        onConfirm={handleAssignRole}
        onCreateUser={createOrgEndUser}
        preselectedUser={addRoleForUser ?? undefined}
        orgName={orgName}
        projectSlug={projectSlug}
      />

      <AlertDialog
        open={!!revokeTarget}
        onOpenChange={(o) => { if (!o) setRevokeTarget(null) }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认撤销角色</AlertDialogTitle>
            <AlertDialogDescription>
              确认撤销 <strong>{revokeTarget?.endUser.username}</strong> 的「
              <strong>{revokeTarget?.role.name}</strong>
              」角色？撤销后若该用户在本项目无其他角色，将无法访问本项目。
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
