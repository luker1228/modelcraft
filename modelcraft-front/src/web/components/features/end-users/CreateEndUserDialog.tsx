'use client'

// src/web/components/features/end-users/CreateEndUserDialog.tsx
// 创建终端用户对话框（EndUser v2）

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Loader2, CheckCircle2, ArrowRight, Eye, EyeOff } from 'lucide-react'
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
  phone: z.string().regex(/^1[3-9]\d{9}$/, '请输入有效的 11 位手机号'),
  password: z.string().min(8, '密码至少 8 位').max(128),
  confirmPassword: z.string().min(1, '请再次输入密码'),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
})

type FormValues = z.infer<typeof schema>

interface CreateEndUserDialogProps {
  open: boolean
  onClose: () => void
  onCreate: (payload: CreateEndUserPayload) => Promise<void>
  /** 所属组织，用于生成项目授权跳转链接 */
  orgName: string
}

export function CreateEndUserDialog({ open, onClose, onCreate, orgName }: CreateEndUserDialogProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [createdUsername, setCreatedUsername] = useState<string | null>(null)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

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
      await onCreate({ username: values.username, phone: values.phone, password: values.password })
      // 创建成功 → 进入引导状态而非直接关闭
      setCreatedUsername(values.username)
      reset()
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
    setCreatedUsername(null)
    setShowPassword(false)
    setShowConfirmPassword(false)
    onClose()
  }

  const handleGoToProjectAuth = () => {
    handleClose()
    router.push(`/org/${orgName}/dashboard`)
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{createdUsername ? '用户创建成功' : '新增终端用户'}</DialogTitle>
        </DialogHeader>

        {/* 创建成功状态：引导授权 */}
        {createdUsername ? (
          <div className="flex flex-col gap-5 py-2">
            <div className="flex items-start gap-3 rounded-lg border border-emerald-200 bg-emerald-50 p-4">
              <CheckCircle2 className="mt-0.5 size-5 shrink-0 text-emerald-600" strokeWidth={1.5} />
              <div className="space-y-1">
                <p className="text-sm font-medium text-foreground">
                  终端用户 <span className="font-semibold text-emerald-700">@{createdUsername}</span> 已创建
                </p>
                <p className="text-sm text-muted-foreground">
                  新用户默认没有任何项目访问权限，需要在项目的「用户授权」页面为其分配角色后，用户才能登录使用。
                </p>
              </div>
            </div>

            <div className="rounded-lg border border-border bg-muted/30 p-4">
              <p className="mb-3 text-sm font-medium text-foreground">下一步：分配项目权限</p>
              <ol className="space-y-1.5 text-sm text-muted-foreground">
                <li className="flex gap-2"><span className="font-medium text-foreground">1.</span> 进入目标项目</li>
                <li className="flex gap-2"><span className="font-medium text-foreground">2.</span> 左侧菜单 → 「用户授权」</li>
                <li className="flex gap-2"><span className="font-medium text-foreground">3.</span> 搜索 <span className="font-mono font-medium text-foreground">@{createdUsername}</span>，分配角色</li>
              </ol>
            </div>

            <DialogFooter className="flex-col gap-2 sm:flex-row">
              <Button variant="outline" onClick={handleClose} className="w-full sm:w-auto">
                稍后处理
              </Button>
              <Button onClick={handleGoToProjectAuth} className="w-full sm:w-auto">
                前往项目列表
                <ArrowRight className="ml-1.5 size-4" />
              </Button>
            </DialogFooter>
          </div>
        ) : (
          /* 创建表单状态 */
          <form onSubmit={onSubmit} className="flex flex-col gap-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="eu-username">用户名</Label>
              <Input id="eu-username" disabled={isLoading} autoComplete="username" {...register('username')} />
              {errors.username && (
                <p className="text-xs text-destructive">{errors.username.message}</p>
              )}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="eu-phone">手机号</Label>
              <Input
                id="eu-phone"
                type="tel"
                placeholder="请输入 11 位手机号"
                disabled={isLoading}
                autoComplete="tel"
                {...register('phone')}
              />
              {errors.phone && (
                <p className="text-xs text-destructive">{errors.phone.message}</p>
              )}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="eu-password">密码</Label>
              <div className="relative">
                <Input
                  id="eu-password"
                  type={showPassword ? 'text' : 'password'}
                  disabled={isLoading}
                  autoComplete="new-password"
                  className="pr-10"
                  {...register('password')}
                />
                <button
                  type="button"
                  tabIndex={-1}
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  aria-label={showPassword ? '隐藏密码' : '显示密码'}
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
              <Label htmlFor="eu-confirm-password">确认密码</Label>
              <div className="relative">
                <Input
                  id="eu-confirm-password"
                  type={showConfirmPassword ? 'text' : 'password'}
                  disabled={isLoading}
                  autoComplete="new-password"
                  className="pr-10"
                  {...register('confirmPassword')}
                />
                <button
                  type="button"
                  tabIndex={-1}
                  onClick={() => setShowConfirmPassword((v) => !v)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  aria-label={showConfirmPassword ? '隐藏密码' : '显示密码'}
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
            <p className="text-xs text-muted-foreground">
              创建后需在项目「用户授权」页面为该用户分配角色，用户才能登录。
            </p>
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
        )}
      </DialogContent>
    </Dialog>
  )
}
