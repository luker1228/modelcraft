'use client'

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

const schema = z.object({
  newPassword: z.string().min(8, '密码至少 8 位').max(128),
})

type FormValues = z.infer<typeof schema>

interface ResetEndUserPasswordDialogProps {
  open: boolean
  onClose: () => void
  onReset: (newPassword: string) => Promise<void>
  username: string
}

export function ResetEndUserPasswordDialog({
  open,
  onClose,
  onReset,
  username,
}: ResetEndUserPasswordDialogProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
  })

  const onSubmit = handleSubmit(async (values) => {
    setIsLoading(true)
    setError(null)
    try {
      await onReset(values.newPassword)
      reset()
      onClose()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '重置密码失败，请重试')
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
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>修改密码</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          <p className="text-sm text-muted-foreground">
            为用户 <span className="font-medium text-foreground">@{username}</span> 设置新密码
          </p>
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="new-password">新密码</Label>
            <Input
              id="new-password"
              type="password"
              disabled={isLoading}
              autoComplete="new-password"
              placeholder="至少 8 位，包含字母和数字"
              {...register('newPassword')}
            />
            {errors.newPassword && (
              <p className="text-xs text-destructive">{errors.newPassword.message}</p>
            )}
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isLoading}>
              取消
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
              确认修改
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
