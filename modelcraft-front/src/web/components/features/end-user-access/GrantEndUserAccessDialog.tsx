'use client'

// src/web/components/features/end-user-access/GrantEndUserAccessDialog.tsx
// 授权终端用户访问项目对话框（EndUser v2）

import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import type { OrgEndUser } from '@web/hooks/end-users/useOrgEndUsers'
import type { GrantAccessPayload } from '@web/hooks/end-user-access/useProjectEndUserAccess'

interface OrgEndUserListBffResponse {
  users?: OrgEndUser[]
}

const schema = z.object({
  userId: z.string().min(1, '请选择用户'),
  permissionBundle: z.string().optional(),
})

type FormValues = z.infer<typeof schema>

// Preset permission bundles — in production these would come from the API
const PERMISSION_BUNDLES = [
  { value: 'viewer', label: '查看者（只读）' },
  { value: 'editor', label: '编辑者（读写）' },
  { value: 'admin', label: '管理员（全部权限）' },
]

interface GrantEndUserAccessDialogProps {
  open: boolean
  orgName: string
  onClose: () => void
  onGrant: (payload: GrantAccessPayload) => Promise<void>
  /** Already-authorized user IDs to exclude from the list */
  existingUserIds: string[]
  /** Optional pre-selected user id from URL */
  preselectedUserId?: string
}

export function GrantEndUserAccessDialog({
  open,
  orgName,
  onClose,
  onGrant,
  existingUserIds,
  preselectedUserId,
}: GrantEndUserAccessDialogProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [users, setUsers] = useState<OrgEndUser[]>([])
  const [usersLoading, setUsersLoading] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState('')

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  // Load org users when dialog opens
  useEffect(() => {
    if (!open) return
    setUsersLoading(true)
    fetch(`/api/bff/org/${orgName}/end-user/users`)
      .then((r) => r.json() as Promise<OrgEndUserListBffResponse>)
      .then((d) => setUsers(d.users ?? []))
      .catch(() => setUsers([]))
      .finally(() => setUsersLoading(false))
  }, [open, orgName])

  useEffect(() => {
    if (!open || !preselectedUserId) return
    setSelectedUserId(preselectedUserId)
    setValue('userId', preselectedUserId)
  }, [open, preselectedUserId, setValue])

  const availableUsers = users.filter((u) => !existingUserIds.includes(u.id))

  const onSubmit = handleSubmit(async (values) => {
    setIsLoading(true)
    setError(null)
    try {
      await onGrant({
        userId: values.userId,
        permissionBundle: values.permissionBundle || undefined,
      })
      reset()
      onClose()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '授权失败，请重试')
    } finally {
      setIsLoading(false)
    }
  })

  const handleClose = () => {
    if (isLoading) return
    setSelectedUserId('')
    reset()
    setError(null)
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>授权终端用户访问</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="flex flex-col gap-2">
            <Label>选择用户 *</Label>
            {usersLoading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" /> 加载用户列表...
              </div>
            ) : (
              <Select
                onValueChange={(v) => {
                  setSelectedUserId(v)
                  setValue('userId', v)
                }}
                disabled={isLoading || availableUsers.length === 0}
                value={selectedUserId || undefined}
              >
                <SelectTrigger>
                  <SelectValue
                    placeholder={
                      availableUsers.length === 0
                        ? '所有 Org 用户均已有权限'
                        : '选择终端用户...'
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {availableUsers.map((u) => (
                    <SelectItem key={u.id} value={u.id}>
                      {u.username}
                      {u.displayName ? ` (${u.displayName})` : ''}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            {errors.userId && (
              <p className="text-sm text-destructive">{errors.userId.message}</p>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <Label>权限包</Label>
            <Select
              onValueChange={(v) => setValue('permissionBundle', v)}
              disabled={isLoading}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择权限包（可选）" />
              </SelectTrigger>
              <SelectContent>
                {PERMISSION_BUNDLES.map((b) => (
                  <SelectItem key={b.value} value={b.value}>
                    {b.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Hidden input for react-hook-form registration */}
          <input type="hidden" {...register('userId')} />
          <input type="hidden" {...register('permissionBundle')} />

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isLoading}>
              取消
            </Button>
            <Button type="submit" disabled={isLoading || availableUsers.length === 0}>
              {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
              授权
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
