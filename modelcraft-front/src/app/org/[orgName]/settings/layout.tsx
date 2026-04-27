'use client'

import { useParams, usePathname, useRouter } from 'next/navigation'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { cn } from '@/shared/utils'
import { Shield, LogIn } from 'lucide-react'

const tabs = [
  { id: 'roles', label: 'Roles', icon: Shield },
  { id: 'login-settings', label: 'Login', icon: LogIn },
]

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { isLoading } = useRequireAuth()
  const params = useParams()
  const pathname = usePathname()
  const router = useRouter()
  const orgName = params.orgName as string

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  const activeTab = tabs.find((tab) =>
    pathname?.endsWith(`/settings/${tab.id}`)
  )?.id || 'roles'

  return (
    <AppLayout pageTitle="组织设置">
      <PageLayout maxWidth="5xl">
        <PageHeader
          title="组织设置"
          spacing="compact"
        />

        {/* Tab Navigation */}
        <div className="mb-6 border-b border-border">
          <nav className="flex gap-6" aria-label="Settings tabs">
            {tabs.map((tab) => {
              const Icon = tab.icon
              const isActive = activeTab === tab.id
              return (
                <button
                  key={tab.id}
                  onClick={() =>
                    router.push(`/org/${orgName}/settings/${tab.id}`)
                  }
                  className={cn(
                    'flex items-center gap-2 py-3 px-1 text-sm font-medium border-b-2 transition-colors',
                    isActive
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                  )}
                >
                  <Icon className="size-4" />
                  {tab.label}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Tab Content */}
        {children}
      </PageLayout>
    </AppLayout>
  )
}
