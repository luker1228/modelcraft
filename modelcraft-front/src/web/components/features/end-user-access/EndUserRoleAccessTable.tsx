'use client'

// src/web/components/features/end-user-access/EndUserRoleAccessTable.tsx
// Project 级终端用户角色分配管理（EndUser Access Redesign）
// 数据来源：listProjectEndUserRoleUsers（新接口）
// 操作：assignEndUserRole / revokeEndUserRole
// 展示：一行 = 一个用户，角色列展示多个 Badge，可单独撤销

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
              <p className="text-xs font-semibold text-muted-foreground">全部角色（{roles.length}）</p>
            </div>
            <ScrollArea className="max-h-64">
              <div className="space-y-px p-1">
                {roles.map((role) => (
                  <div key={role.id} className="flex items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-muted">
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

// ── CreateAndAssignDialog ──────────────────────────────────────────────────────
// 两步弹窗：
//   Step 1 — 选择已有用户 OR 新建用户
//   Step 2 — 选角色并确认授权
//
// preselectedUser: 追加角色场景，锁定用户跳过 Step 1

const createUserSchema = z.object({
  username: z.string().min(1, '请输入用户名').max(64),
  phone: z.string().regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号'),
  password: z.string().min(8, '密码至少 8 位').max(128),
  confirmPassword: z.string().min(1, '请再次输入密码'),
}).refine((d) => d.password === d.confirmPassword, {
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
  onCreateUser: (payload: CreateOrgEndUserPayload) => Promise<{ success: boolean; endUser?: { id: string; username: string }; errorMessage?: string }>
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
  // Step 1 state
  const [mode, setMode] = useState<DialogMode>('select')
  const [selectedUserId, setSelectedUserId] = useState(preselectedUser?.id ?? '')
  const [createError, setCreateError] = useState<string | null>(null)
  const [creatingUser, setCreatingUser] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  // Step 2 state
  const [step, setStep] = useState<1 | 2>(1)
  const [resolvedUser, setResolvedUser] = useState<{ id: string; username: string } | null>(
    preselectedUser ? { id: preselectedUser.id, username: preselectedUser.username } : null
  )
  const [selectedRoleId, setSelectedRoleId] = useState('')
  const [assigning, setAssigning] = useState(false)

  const { register, handleSubmit, reset, formState: { errors } } = useForm<CreateUserFormValues>({
    resolver: zodResolver(createUserSchema),
  })

  // 追加角色场景：直接进入 Step 2
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
    setResolvedUser(preselectedUser ? { id: preselectedUser.id, username: preselectedUser.username } : null)
    setSelectedRoleId('')
    setCreateError(null)
    setCreatingUser(false)
    setAssigning(false)
    setShowPassword(false)
    setShowConfirmPassword(false)
    reset()
    onClose()
  }

  // Step 1 → Step 2（选已有用户）
  const handleSelectExistingNext = () => {
    if (!selectedUserId) return
    const found = orgUsers.find((u) => u.id === selectedUserId)
    if (!found) return
    setResolvedUser({ id: found.id, username: found.username })
    setStep(2)
  }

  // Step 1 → Step 2（新建用户）
  const handleCreateUserSubmit = handleSubmit(async (values) => {
    setCreatingUser(true)
    setCreateError(null)
    try {
      const result = await onCreateUser({ username: values.username, phone: values.phone, password: values.password })
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

  // Step 2 → 授权
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

  // ── 标题 ─────────────────────────────────────────────────────────────────────
  const title = preselectedUser
    ? '追加角色'
    : step === 1
      ? '授权用户访问'
      : `为 ${resolvedUser?.username ?? ''} 分配角色`

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {step === 2 && !preselectedUser && (
              <button
                type="button"
                onClick={() => { setStep(1); setResolvedUser(null); setSelectedRoleId('') }}
                className="rounded-sm p-0.5 text-muted-foreground hover:text-foreground"
                aria-label="返回上一步"
              >
                <ChevronLeft className="size-4" />
              </button>
            )}
            {title}
          </DialogTitle>
          {step === 1 && (
            <DialogDescription>
              选择已有用户，或新建一个用户后立即授权。
            </DialogDescription>
          )}
        </DialogHeader>

        {/* ── Step 1：选择用户 ─────────────────────────────────────────────── */}
        {step === 1 && (
          <div className="flex flex-col gap-4">
            {/* 模式切换 Tab */}
            <div className="flex rounded-md border border-border bg-muted/40 p-0.5">
              <button
                type="button"
                className={cn(
                  'flex-1 rounded-sm px-3 py-1.5 text-sm font-medium transition-colors',
                  mode === 'select'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
    // ... 984 lines omitted
}
    // ... 983 lines omitted
interface EndUserRoleAccessTableProps {
    // ... 982 lines omitted
}
    // ... 981 lines omitted
export function EndUserRoleAccessTable({ orgName, projectSlug, autoAssign }: EndUserRoleAccessTableProps) {
    // ... 980 lines omitted
      }
    // ... 979 lines omitted
      }
    // ... 978 lines omitted
    }
    // ... 977 lines omitted
}
    // ... 976 lines omitted
import { useState, useCallback, useMemo, useEffect } from 'react'
import Link from 'next/link'
import { Plus, RefreshCw, Users, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/utils'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import {
    // ... 966 lines omitted
import {
    // ... 965 lines omitted
import {
    // ... 964 lines omitted
import {
    // ... 963 lines omitted
import {
    // ... 962 lines omitted
import { ScrollArea } from '@web/components/ui/scroll-area'
import {
    // ... 960 lines omitted
  type ProjectRoleUserEntry,
    // ... 959 lines omitted
interface RoleItem {
    // ... 958 lines omitted
}
    // ... 957 lines omitted
interface GroupedRoleEntry {
    // ... 956 lines omitted
}
    // ... 955 lines omitted
interface RevokeTarget {
    // ... 954 lines omitted
}
    // ... 953 lines omitted
function fmtDate(iso: string) {
    // ... 952 lines omitted
}
    // ... 951 lines omitted
function groupEntries(entries: ProjectRoleUserEntry[]): GroupedRoleEntry[] {
    // ... 950 lines omitted
      }
    // ... 949 lines omitted
    }
  }
    // ... 947 lines omitted
}
    // ... 946 lines omitted
interface RoleBadgeListProps {
    // ... 945 lines omitted
}
    // ... 944 lines omitted
function RoleBadge({ role, onRevoke }: { role: RoleItem; onRevoke: (role: RoleItem) => void }) {
    // ... 943 lines omitted
}
    // ... 942 lines omitted
function RoleBadgeList({ roles, onRevoke }: RoleBadgeListProps) {
    // ... 941 lines omitted
}
    // ... 940 lines omitted
interface AssignRoleDialogProps {
    // ... 939 lines omitted
}
    // ... 938 lines omitted
function AssignRoleDialog({
    // ... 937 lines omitted
    }
  }
    // ... 935 lines omitted
    }
  }
    // ... 933 lines omitted
}
    // ... 932 lines omitted
interface EndUserRoleAccessTableProps {
    // ... 931 lines omitted
}
    // ... 930 lines omitted
export function EndUserRoleAccessTable({ orgName, projectSlug, autoAssign }: EndUserRoleAccessTableProps) {
    // ... 929 lines omitted
    }
    // ... 928 lines omitted
      }
    // ... 927 lines omitted
      }
    // ... 926 lines omitted
    }
    // ... 925 lines omitted
                      }
    // ... 924 lines omitted
}
