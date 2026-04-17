// src/app/org/[orgName]/project/[projectSlug]/user/login/page.tsx
// 终端用户登录页（公开路由，无 auth 守卫）
// SEO 设置 robots: noindex

import type { Metadata } from 'next'
import { EndUserLoginCard } from '@web/components/features/end-user-auth/EndUserLoginCard'

// ============================================================================
// Metadata
// ============================================================================

export const metadata: Metadata = {
  title: '用户登录',
  robots: {
    index: false,
    follow: false,
  },
}

// ============================================================================
// Page Component
// ============================================================================

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
