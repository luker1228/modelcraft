import type { Metadata } from 'next'
import { EndUserLoginCard } from '@web/components/features/end-user-auth/EndUserLoginCard'

export const metadata: Metadata = {
  title: '用户登录',
  robots: {
    index: false,
    follow: false,
  },
}

interface EndUserLoginPageProps {
  params: Promise<{
    orgName: string
    projectSlug: string
  }>
}

export default async function EndUserLoginPage({ params }: EndUserLoginPageProps) {
  const { orgName, projectSlug } = await params
  return <EndUserLoginCard orgName={orgName} projectSlug={projectSlug} />
}

