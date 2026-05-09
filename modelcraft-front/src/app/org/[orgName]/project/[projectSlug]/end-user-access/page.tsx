'use client'

import { useParams } from 'next/navigation'
import { EndUserRoleAccessTable } from '@web/components/features/end-user-access/EndUserRoleAccessTable'
import { PageLayout, PageHeader } from '@web/components/features/layout'

export default function ProjectEndUserAccessPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title="用户授权" />
      <EndUserRoleAccessTable orgName={orgName} projectSlug={projectSlug} />
    </PageLayout>
  )
}
