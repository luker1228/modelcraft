'use client'

// src/app/org/[orgName]/end-users/page.tsx
// Org 级终端用户管理页（EndUser v2）

import { useParams, useSearchParams } from 'next/navigation'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { EndUsersManagementTable } from '@web/components/features/end-users/EndUsersManagementTable'

export default function OrgEndUsersPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const orgName = params?.orgName as string
  const autoCreate = searchParams?.get('create') === 'true'

  return (
    <AppLayout pageTitle="终端用户">
      <PageLayout maxWidth="7xl">
        <PageHeader title="终端用户管理" bordered />
        <EndUsersManagementTable orgName={orgName} autoCreate={autoCreate} />
      </PageLayout>
    </AppLayout>
  )
}
