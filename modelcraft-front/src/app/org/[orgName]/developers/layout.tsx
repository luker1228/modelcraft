'use client'

import { useParams, usePathname, useRouter } from 'next/navigation'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { cn } from '@/shared/utils'
import { Users, Shield } from 'lucide-react'

const tabs = [
  { id: 'members', label: '开发者', icon: Users },
  { id: 'roles', label: '角色', icon: Shield },
]

export default function DevelopersLayout({
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

  const activeTab =
    tabs.find((tab) => pathname?.endsWith(`/developers/${tab.id}`))?.id ?? 'members'

  return (
    <AppLayout pageTitle="开发者">
      <PageLayout maxWidth="7xl" background="card" padding="compact">
        <PageHeader title="开发者" spacing="compact" />

        {/* Tab Navigation */}
        <div className="mb-6 border-b border-border">
          <nav className="flex gap-6" aria-label="Developers tabs">
            {tabs.map((tab) => {
              const Icon = tab.icon
              const isActive = activeTab === tab.id
              return (
                <button
                  key={tab.id}
                  onClick={() => router.push(`/org/${orgName}/developers/${tab.id}`)}
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
