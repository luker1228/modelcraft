'use client'

// src/app/org/[orgName]/end-users/page.tsx
// Org 级终端用户管理页（EndUser v2）

import { useParams } from 'next/navigation'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { EndUsersManagementTable } from '@web/components/features/end-users/EndUsersManagementTable'

export default function OrgEndUsersPage() {
  const params = useParams()
  const orgName = params?.orgName as string

  return (
    <AppLayout pageTitle="终端用户">
      <div className="h-full overflow-auto bg-white">
        <div className="mx-auto max-w-7xl p-6">
          <div className="mb-6">
            <h1 className="font-heading text-xl font-semibold tracking-tight text-foreground">
              终端用户管理
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              管理 {orgName} 组织下的所有终端用户账号。终端用户访问权限在各项目中单独配置。
            </p>
          </div>
          <EndUsersManagementTable orgName={orgName} />
        </div>
      </div>
    </AppLayout>
  )
}
