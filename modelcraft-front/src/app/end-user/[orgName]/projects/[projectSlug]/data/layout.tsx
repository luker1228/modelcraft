'use client'

import React from 'react'
import { EndUserAuthGuard } from '@web/components/features/end-user-auth/EndUserAuthGuard'

interface EndUserDataLayoutProps {
  children: React.ReactNode
}

export default function EndUserDataLayout({ children }: EndUserDataLayoutProps) {
  return (
    <EndUserAuthGuard loadingMessage="正在加载数据管理界面...">
      <div className="min-h-screen bg-background">{children}</div>
    </EndUserAuthGuard>
  )
}
