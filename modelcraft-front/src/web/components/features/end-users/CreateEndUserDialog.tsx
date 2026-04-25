'use client'

// src/web/components/features/end-users/CreateEndUserDialog.tsx
// 创建终端用户对话框（EndUser v2）

import { useState } from 'react'
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
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import type { CreateEndUserPayload } from '@web/hooks/end-users/useOrgEndUsers'

const schema = z.object({
  username: z.string().min(1, '请输入用户名').max(64),
  password: z.string().min(6, '密码至少 6 位').max(128),
  displayName: z.string().max(64).optional(),
})

type FormValues = z.infer<typeof schema>

interface CreateEndUserDialogProps {
  open: boolean
  onClose: () => void
  onCreate: (payload: CreateEndUserPayload) => Promise<void>
}

export function CreateEndUserDialog({ open, onClose, onCreate }: CreateEndUserDialogProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const onSubmit = handleSubmit(async (values) => {
    setIsLoading(true)
    setError(null)
    try {
      await onCreate({
        username: values.username,
        password: values.password,
        displayName: values.displayName || undefined,
      })
      reset()
      onClose()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '创建失败，请重试')
    } finally {
      setIsLoading(false)
    }
  })

  const handleClose = () => {
    if (isLoading) return
    reset()
    setError(null)
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>新增终端用户</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="flex flex-col gap-2">
            <Label htmlFor="eu-username">用户名 *</Label>
            <Input id="eu-username" disabled={isLoading} {...register('username')} />
            {errors.username && (
              <p className="text-sm text-destructive">{errors.username.message}</p>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="eu-password">密码 *</Label>
            <Input id="eu-password" type="password" disabled={isLoading} {...register('password')} />
            {errors.password && (
              <p className="text-sm text-destructive">{errors.password.message}</p>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="eu-display-name">显示名称</Label>
            <Input id="eu-display-name" disabled={isLoading} placeholder="可选" {...register('displayName')} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isLoading}>
              取消
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
              创建
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
