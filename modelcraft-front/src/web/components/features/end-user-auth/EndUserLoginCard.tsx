'use client'

// src/web/components/features/end-user-auth/EndUserLoginCard.tsx
// 终端用户登录卡片组件（对称 auth/auth-layout.tsx + login-form.tsx）
// 独立布局，不使用开发者侧的 AppLayout / AppSidebar

import React, { Suspense } from 'react'
import { AlertCircle, Loader2, Eye, EyeOff, Database } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import { useEndUserLoginForm, type EndUserLoginFormValues } from '@web/hooks/end-user-auth/useEndUserLoginForm'
import type { UseFormReturn } from 'react-hook-form'

// ============================================================================
// Sub Components
// ============================================================================

interface PasswordInputProps {
  id: string
  disabled?: boolean
  error?: string
  register: UseFormReturn<EndUserLoginFormValues>['register']
}

function PasswordInput({ id, disabled, error, register }: PasswordInputProps) {
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
        {...register('password')}
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
// Main Components
// ============================================================================

interface EndUserLoginFormContentProps {
  orgName: string
  projectSlug: string
}

function EndUserLoginFormContent({ orgName, projectSlug }: EndUserLoginFormContentProps) {
  const { form, onSubmit, isLoading, error } = useEndUserLoginForm(orgName, projectSlug)
  const {
    register,
    formState: { errors },
  } = form

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-5">
      {/* Error Banner */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="size-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Username Field */}
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

      {/* Password Field */}
      <div className="flex flex-col gap-2">
        <Label htmlFor="password">密码</Label>
        <PasswordInput
          id="password"
          disabled={isLoading}
          error={errors.password?.message}
          register={register}
        />
        {errors.password && (
          <p className="text-sm text-destructive">{errors.password.message}</p>
        )}
      </div>

      {/* Submit Button */}
      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
        登录
      </Button>
    </form>
  )
}

// ============================================================================
// Exported Component
// ============================================================================

interface EndUserLoginCardProps {
  orgName: string
  projectSlug: string
}

/**
 * 终端用户登录卡片组件。
 * 包含项目品牌展示 + 登录表单。
 */
export function EndUserLoginCard({ orgName, projectSlug }: EndUserLoginCardProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-md border bg-background shadow-sm">
        {/* Project Branding */}
        <CardHeader className="space-y-4 px-8 pt-8">
          <div className="flex items-center justify-center gap-3">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Database className="size-5 text-primary-foreground" strokeWidth={1.5} />
            </div>
            <span className="text-xl font-semibold text-foreground">
              {projectSlug}
            </span>
          </div>
          <div className="text-center">
            <CardTitle className="text-2xl">用户登录</CardTitle>
            <CardDescription className="mt-2">
              登录以访问 {orgName} / {projectSlug} 的数据
            </CardDescription>
          </div>
        </CardHeader>

        {/* Login Form */}
        <CardContent className="px-8 pb-8">
          <Suspense
            fallback={
              <div className="flex items-center justify-center py-8">
                <Loader2 className="size-6 animate-spin text-muted-foreground" />
              </div>
            }
          >
            <EndUserLoginFormContent orgName={orgName} projectSlug={projectSlug} />
          </Suspense>
        </CardContent>
      </Card>
    </div>
  )
}
