'use client'

import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isLoading } = useRequireAuth()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  return (
    <AppLayout pageTitle="组织设置">
      <PageLayout maxWidth="5xl">
        <PageHeader
          title="组织设置"
          spacing="compact"
        />

        {children}
      </PageLayout>
    </AppLayout>
  )
}
