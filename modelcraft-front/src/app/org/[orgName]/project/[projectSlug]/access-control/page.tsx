'use client'

import { useParams } from 'next/navigation'

import { PageHeader, PageLayout } from '@web/components/features/layout'
import { RlsPolicyContent } from './_components'

export default function AccessControlPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title="RLS 策略" />

      <div className="mt-6">
        <RlsPolicyContent orgName={orgName} projectSlug={projectSlug} />
      </div>
    </PageLayout>
  )
}
