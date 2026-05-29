'use client'

// src/web/components/features/end-user-auth/EndUserOrgLoginCard.tsx
// Org 级终端用户登录卡片组件（EndUser v2）
// 终端用户账号由管理员创建，不支持自注册。

import React from 'react'
import { AlertCircle, Loader2, Eye, EyeOff, Database } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import {
  useEndUserOrgLoginForm,
  type EndUserOrgLoginFormValues,
} from '@web/hooks/end-user-auth-v2/useEndUserOrgLoginForm'
import type { Path, UseFormReturn, FieldValues } from 'react-hook-form'
import { refreshEndUserAccessToken } from '@api-client/end-user/end-user-auth-client'

// ============================================================================
// Sub Components
// ============================================================================

interface PasswordInputProps<T extends FieldValues> {
  id: string
  disabled?: boolean
  error?: string
  register: UseFormReturn<T>['register']
  fieldName: Path<T>
}

function PasswordInput<T extends FieldValues>({
  id,
  disabled,
  error,
  register,
  fieldName,
}: PasswordInputProps<T>) {
  const [showPassword, setShowPassword] = React.useState(false)

  return (
    <div className="relative">
      <Input
        id={id}
        type={showPassword ? 'text' : 'password'}
        placeholder="请输入密码"
        autoComplete="current-password"
        disabled={disabled}
        aria-invalid={!!error}
        className="pr-10"
        {...register(fieldName)}
      />
      <button
        type="button"
        tabIndex={-1}
        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
        onClick={() => setShowPassword(!showPassword)}
      >
        {showPassword ? (
          <Eye className="size-4" strokeWidth={1.5} />
        ) : (
          <EyeOff className="size-4" strokeWidth={1.5} />
        )}
      </button>
    </div>
  )
}

// ============================================================================
// Main Component
// ============================================================================

interface EndUserOrgLoginCardProps {
  orgName: string
}

/**
 * Org 级终端用户登录卡片（EndUser v2）。
 * 终端用户账号由管理员在租户端创建，不支持自注册。
 * 登录分支由 useEndUserOrgLoginForm 处理：
 * - 有可访问 Project → 跳转 workspace
 * - 0 个可访问 Project → 跳转待授权页
 */
export function EndUserOrgLoginCard({ orgName }: EndUserOrgLoginCardProps) {
  const router = useRouter()
  const [checkingSession, setCheckingSession] = React.useState(true)
  const { form, onSubmit, isLoading, error } = useEndUserOrgLoginForm(orgName)
  const {
    register,
    formState: { errors },
  } = form

  React.useEffect(() => {
    let cancelled = false

    void (async () => {
      const token = await refreshEndUserAccessToken({ orgName })
      if (cancelled) return

      if (token) {
        router.replace(`/end-user/${orgName}/dashboard`)
        return
      }

      setCheckingSession(false)
    })()

    return () => {
      cancelled = true
    }
  }, [orgName, router])

  if (checkingSession) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
        <Card className="w-full max-w-md border bg-background shadow-sm">
          <CardContent className="flex items-center justify-center px-8 py-10 text-sm text-muted-foreground">
            <Loader2 className="mr-2 size-4 animate-spin" />
            正在恢复会话...
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-md border bg-background shadow-sm">
        {/* Org Branding */}
        <CardHeader className="space-y-4 px-8 pt-8">
          <div className="flex items-center justify-center gap-3">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Database className="size-5 text-primary-foreground" strokeWidth={1.5} />
            </div>
            <span className="text-xl font-semibold text-foreground">{orgName}</span>
          </div>
          <div className="text-center">
            <CardTitle className="text-2xl">终端用户登录</CardTitle>
            <CardDescription className="mt-2">
              登录后访问 {orgName} 的数据项目
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent className="px-8 pb-8">
          <form onSubmit={onSubmit} className="flex flex-col gap-5">
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="flex flex-col gap-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                type="text"
                placeholder="请输入用户名"
                autoComplete="username"
                disabled={isLoading}
                aria-invalid={!!errors.username}
                {...register('username')}
              />
              {errors.username && (
                <p className="text-sm text-destructive">{errors.username.message}</p>
              )}
            </div>

            <div className="flex flex-col gap-2">
              <Label htmlFor="password">密码</Label>
              <PasswordInput
                id="password"
                disabled={isLoading}
                error={errors.password?.message}
                register={register}
                fieldName={'password' as Path<EndUserOrgLoginFormValues>}
              />
              {errors.password && (
                <p className="text-sm text-destructive">{errors.password.message}</p>
              )}
            </div>

            <Button type="submit" disabled={isLoading} className="w-full">
              {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
              登录
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
