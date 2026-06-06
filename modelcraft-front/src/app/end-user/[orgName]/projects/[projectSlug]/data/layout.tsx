'use client'

import React from 'react'
import { EndUserAuthGuard } from '@web/components/features/end-user-auth/EndUserAuthGuard'

interface EndUserDataLayoutProps {
  children: React.ReactNode
}

export default function EndUserDataLayout({ children }: EndUserDataLayoutProps) {
  return (
    <EndUserAuthGuard loadingMessage="正在加载数据管理界面...">
      {/* End-user data workspace intentionally does not mount Copilot for now.
          This page is table-first and the floating trigger is too intrusive.
          Keep the dedicated wrapper implementation in place so we can re-enable
          it later without reworking the route structure. */}
      <div className="h-screen overflow-hidden">{children}</div>
    </EndUserAuthGuard>
  )
}
