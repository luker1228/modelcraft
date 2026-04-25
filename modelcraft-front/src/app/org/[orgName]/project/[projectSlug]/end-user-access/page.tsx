'use client'

import { useParams } from 'next/navigation'
import { EndUserManagementTable } from '@web/components/features/end-user-access/EndUserManagementTable'
import { PageLayout, PageHeader } from '@web/components/features/layout'

export default function ProjectEndUserAccessPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title="用户管理" />
      <EndUserManagementTable orgName={orgName} projectSlug={projectSlug} />
    </PageLayout>
  )
}
