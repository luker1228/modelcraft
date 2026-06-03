'use client'

import { useParams } from 'next/navigation'
import Link from 'next/link'
import { UserPlus } from 'lucide-react'
import { EndUserRoleAccessTable } from '@web/components/features/end-user-access/EndUserRoleAccessTable'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { Button } from '@web/components/ui/button'

export default function ProjectEndUserAccessPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader
        title="用户授权"
        actions={
          <Button variant="outline" size="sm" asChild>
            <Link href={`/org/${orgName}/end-users`}>
              <UserPlus className="mr-1.5 size-4" />
              新建用户
            </Link>
          </Button>
        }
      />
      <EndUserRoleAccessTable orgName={orgName} projectSlug={projectSlug} />
    </PageLayout>
  )
}
