'use client'

import { useParams, useSearchParams } from 'next/navigation'

import { PageHeader, PageLayout } from '@web/components/features/layout'

import { BundlesTab, PermissionsTab, RolesContent } from './_components'

export default function AccessControlPage() {
  const params = useParams()
  const searchParams = useSearchParams()

  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const rawTab = searchParams.get('tab')
  const activeTab = rawTab === 'bundles' ? 'bundles' : rawTab === 'permissions' ? 'permissions' : 'roles'

  const tabTitle = activeTab === 'bundles' ? '权限包' : activeTab === 'permissions' ? '权限点' : '角色'

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title={tabTitle} />

      <div className="mt-6">
        {activeTab === 'roles' && <RolesContent orgName={orgName} projectSlug={projectSlug} />}
        {activeTab === 'bundles' && <BundlesTab orgName={orgName} projectSlug={projectSlug} />}
        {activeTab === 'permissions' && <PermissionsTab orgName={orgName} projectSlug={projectSlug} />}
      </div>
    </PageLayout>
  )
}
