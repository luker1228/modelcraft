'use client'

import { usePathname } from 'next/navigation'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isLoading } = useRequireAuth()
  const pathname = usePathname()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  const title = pathname?.endsWith('/settings/api-tokens')
    ? 'API Token'
    : '组织设置'

  return (
    <AppLayout pageTitle={title}>
      <PageLayout maxWidth="5xl">
        <PageHeader
          title={title}
          spacing="compact"
        />

        {children}
      </PageLayout>
    </AppLayout>
  )
}
