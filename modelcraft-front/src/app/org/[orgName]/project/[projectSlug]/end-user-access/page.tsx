'use client'

// src/app/org/[orgName]/project/[projectSlug]/end-user-access/page.tsx
// Project 级终端用户访问控制页（EndUser v2）

import { useParams } from 'next/navigation'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { EndUserAccessTable } from '@web/components/features/end-user-access/EndUserAccessTable'

export default function ProjectEndUserAccessPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <AppLayout pageTitle="终端用户访问">
      <div className="h-full overflow-auto bg-white">
        <div className="mx-auto max-w-7xl p-6">
          <div className="mb-6">
            <h1 className="font-heading text-xl font-semibold tracking-tight text-foreground">
              终端用户访问控制
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              管理有权访问此项目的终端用户及其权限。终端用户账号在 Org 层统一管理。
            </p>
          </div>
          <EndUserAccessTable orgName={orgName} projectSlug={projectSlug} />
        </div>
      </div>
    </AppLayout>
  )
}
