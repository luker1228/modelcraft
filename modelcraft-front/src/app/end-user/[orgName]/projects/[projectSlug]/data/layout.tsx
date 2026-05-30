'use client'

import React from 'react'
import { useParams } from 'next/navigation'
import { EndUserAuthGuard } from '@web/components/features/end-user-auth/EndUserAuthGuard'
import { EndUserCopilotWrapper } from '@web/components/features/copilot/CopilotProvider'
import "@copilotkit/react-ui/styles.css"

interface EndUserDataLayoutProps {
  children: React.ReactNode
}

export default function EndUserDataLayout({ children }: EndUserDataLayoutProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const orgName = params?.orgName ?? ''
  const projectSlug = params?.projectSlug ?? ''

  return (
    <EndUserAuthGuard loadingMessage="正在加载数据管理界面...">
      <EndUserCopilotWrapper orgName={orgName} projectSlug={projectSlug}>
        <div className="h-screen overflow-hidden">{children}</div>
      </EndUserCopilotWrapper>
    </EndUserAuthGuard>
  )
}
