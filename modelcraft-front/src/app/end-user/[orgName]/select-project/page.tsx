import type { Metadata } from 'next'
import { EndUserProjectSelector } from '@web/components/features/end-user-auth/EndUserProjectSelector'

export const metadata: Metadata = {
  title: '选择项目',
  robots: {
    index: false,
    follow: false,
  },
}

interface SelectProjectPageProps {
  params: Promise<{
    orgName: string
  }>
}

export default async function SelectProjectPage({ params }: SelectProjectPageProps) {
  const { orgName } = await params
  return <EndUserProjectSelector orgName={orgName} />
}
