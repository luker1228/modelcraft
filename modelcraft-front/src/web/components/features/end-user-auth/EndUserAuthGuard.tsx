'use client'

// src/web/components/features/end-user-auth/EndUserAuthGuard.tsx
// 终端用户认证守卫组件
// 封装 useRequireEndUserAuth hook，在 loading 期间显示加载状态

import React from 'react'
import { Loader2 } from 'lucide-react'
import { useRequireEndUserAuth } from '@web/hooks/end-user-auth/useRequireEndUserAuth'

interface EndUserAuthGuardProps {
  /** 认证成功后渲染的子组件 */
  children: React.ReactNode
  /** 自定义加载状态消息 */
  loadingMessage?: string
}

/**
 * 终端用户认证守卫组件。
 * 在 silent refresh 恢复 session 期间显示加载状态，
 * refresh 失败时由 hook 内部处理重定向到登录页。
 */
export function EndUserAuthGuard({
  children,
  loadingMessage = '正在验证身份...',
}: EndUserAuthGuardProps) {
  const { isLoading } = useRequireEndUserAuth()

  if (isLoading) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-background">
        <Loader2 className="size-8 animate-spin text-primary" strokeWidth={1.5} />
        <p className="text-sm text-muted-foreground">{loadingMessage}</p>
      </div>
    )
  }

  return <>{children}</>
}
