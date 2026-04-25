'use client'

// src/web/components/features/end-user-auth/EndUserOrgLoginCard.tsx
// Org 级终端用户登录卡片组件（EndUser v2）
// 与旧的 EndUserLoginCard 结构相同，但调用 v2 BFF（Org 级登录）

import React from 'react'
import { AlertCircle, Loader2, Eye, EyeOff, Database } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import {
  useEndUserOrgLoginForm,
  type EndUserOrgLoginFormValues,
} from '@web/hooks/end-user-auth-v2/useEndUserOrgLoginForm'
import {
  useEndUserOrgRegisterForm,
  type EndUserOrgRegisterFormValues,
} from '@web/hooks/end-user-auth-v2/useEndUserOrgRegisterForm'
import type { Path, UseFormReturn, FieldValues } from 'react-hook-form'

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
          <EyeOff className="size-4" strokeWidth={1.5} />
        ) : (
          <Eye className="size-4" strokeWidth={1.5} />
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

function EndUserOrgRegisterFormContent({ orgName }: EndUserOrgLoginCardProps) {
  const { form, onSubmit, isLoading, error } = useEndUserOrgRegisterForm(orgName)
  const {
    register,
    formState: { errors },
  } = form

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-5">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="size-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col gap-2">
        <Label htmlFor="register-username">用户名</Label>
        <Input
          id="register-username"
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
        <Label htmlFor="register-password">密码</Label>
        <PasswordInput
          id="register-password"
          disabled={isLoading}
          error={errors.password?.message}
          register={register}
          fieldName={'password' as Path<EndUserOrgRegisterFormValues>}
        />
        {errors.password && (
          <p className="text-sm text-destructive">{errors.password.message}</p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="register-confirm-password">确认密码</Label>
        <PasswordInput
          id="register-confirm-password"
          disabled={isLoading}
          error={errors.confirmPassword?.message}
          register={register}
          fieldName={'confirmPassword' as Path<EndUserOrgRegisterFormValues>}
        />
        {errors.confirmPassword && (
          <p className="text-sm text-destructive">{errors.confirmPassword.message}</p>
        )}
      </div>

      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
        注册并登录
      </Button>
    </form>
  )
}

/**
 * Org 级终端用户登录卡片（EndUser v2）。
 * 登录分支由 useEndUserOrgLoginForm 处理：
 * - 1 个可访问 Project → 直接跳转数据页
 * - N 个可访问 Project → 跳转 select-project 页
 * - 0 个可访问 Project → 显示无权限错误
 */
export function EndUserOrgLoginCard({ orgName }: EndUserOrgLoginCardProps) {
  const [mode, setMode] = React.useState<'login' | 'register'>('login')
  const { form, onSubmit, isLoading, error } = useEndUserOrgLoginForm(orgName)
  const {
    register,
    formState: { errors },
  } = form

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
            <CardTitle className="text-2xl">
              {mode === 'login' ? '终端用户登录' : '终端用户注册'}
            </CardTitle>
            <CardDescription className="mt-2">
              {mode === 'login'
                ? `登录后访问 ${orgName} 的数据项目`
                : `注册后自动登录并进入 ${orgName} 的数据项目`}
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent className="px-8 pb-8">
          {mode === 'login' ? (
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
          ) : (
            <EndUserOrgRegisterFormContent orgName={orgName} />
          )}

          <div className="mt-4 text-center text-sm text-muted-foreground">
            {mode === 'login' ? '还没有账号？' : '已有账号？'}
            <Button
              type="button"
              variant="link"
              className="h-auto px-2 py-0 text-sm"
              onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            >
              {mode === 'login' ? '去注册' : '去登录'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
