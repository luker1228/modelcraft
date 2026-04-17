'use client'

// src/app/org/[orgName]/project/[projectSlug]/data/layout.tsx
// 数据管理页布局（守卫：end_user_refresh_token）
// 使用 EndUserAuthGuard 包裹子页面，在认证完成前显示加载状态

import React from 'react'
import { EndUserAuthGuard } from '@web/components/features/end-user-auth/EndUserAuthGuard'

interface DataLayoutProps {
  children: React.ReactNode
}

/**
 * 终端用户数据管理路由布局。
 * 使用 EndUserAuthGuard 进行认证守卫：
 * - middleware 已保证 end_user_refresh_token cookie 存在才能进入
 * - 页面刷新后 in-memory token 丢失时，自动执行 silent refresh
 * - refresh 失败时重定向到 /user/login
 */
export default function DataLayout({ children }: DataLayoutProps) {
  return (
    <EndUserAuthGuard loadingMessage="正在加载数据管理界面...">
      <div className="min-h-screen bg-background">
        {children}
      </div>
    </EndUserAuthGuard>
  )
}
