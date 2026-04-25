'use client'

import { useParams, usePathname, useRouter } from 'next/navigation'
import { Layers, Lock, Shield, Users } from 'lucide-react'
import { cn } from '@/shared/utils'
import { PageLayout, PageHeader } from '@web/components/features/layout'

const tabs = [
  { id: 'bundles',     label: '权限包',   icon: Layers },
  { id: 'permissions', label: '权限点',   icon: Lock   },
  { id: 'roles',       label: '角色',     icon: Shield },
  { id: 'users',       label: '用户授权', icon: Users  },
]

interface RBACLayoutProps {
  children: React.ReactNode
}

/**
 * RBAC 设置区域 Tab 导航 Layout
 *
 * Route: /org/[orgName]/project/[projectSlug]/rbac/*
 * Tabs: bundles / permissions / roles / users
 */
export default function RBACLayout({ children }: RBACLayoutProps) {
  const params = useParams()
  const pathname = usePathname()
  const router = useRouter()

  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  const activeTab =
    tabs.find((tab) => pathname?.includes(`/rbac/${tab.id}`))?.id ?? 'bundles'

  const handleTabChange = (tabId: string) => {
    router.push(`/org/${orgName}/project/${projectSlug}/rbac/${tabId}`)
  }

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader
        title="终端用户权限"
        spacing="compact"
      />

      {/* Tab Navigation */}
      <div className="mb-6 border-b border-border">
        <nav className="flex gap-6" aria-label="RBAC tabs">
          {tabs.map((tab) => {
            const Icon = tab.icon
            const isActive = activeTab === tab.id

            return (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={cn(
                  'flex items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition-colors',
                  isActive
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
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
  )
}
