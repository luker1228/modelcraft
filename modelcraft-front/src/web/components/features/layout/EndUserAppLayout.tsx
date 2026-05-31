'use client'

import { ReactNode, useMemo, useCallback, useState } from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { ApolloProvider } from '@apollo/client'
import { ChevronRight, Terminal, PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { UserMenu } from '@web/components/features/layout/UserMenu'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { clearEndUserSessionArtifacts } from '@shared/auth/clear-end-user-session'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import { cn } from '@/shared/utils'

type ActivePage = 'projects' | 'cli'

interface EndUserAppLayoutProps {
  children: ReactNode
  orgName: string
  /** Which top-level page is active — used to highlight sidebar nav items */
  activePage?: ActivePage
}

/**
 * EndUserAppLayout — End-user facing application shell.
 *
 * Provides topbar + collapsible sidebar nav for end-user pages.
 */
export function EndUserAppLayout(props: EndUserAppLayoutProps) {
  const client = useMemo(
    () => createEndUserOrgScopedClient(props.orgName),
    [props.orgName]
  )

  return (
    <ApolloProvider client={client}>
      <EndUserAppLayoutInner {...props} />
    </ApolloProvider>
  )
}

function EndUserAppLayoutInner({
  children,
  orgName,
  activePage,
}: EndUserAppLayoutProps) {
  const router = useRouter()
  const isAdmin = useEndUserAuthStore((s) => s.isAdmin)
  const userInfo = useEndUserAuthStore((s) => s.userInfo)
  const [collapsed, setCollapsed] = useState(false)

  const orgInitials = orgName
    .split(/[-_]/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('')

  const handleLogout = useCallback(async () => {
    try {
      await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
        method: 'POST',
        credentials: 'same-origin',
      })
    } catch {
      // ignore errors, always clear session and redirect
    }
    clearEndUserSessionArtifacts()
    router.push('/')
  }, [orgName, router])

  return (
    <div className="flex h-full flex-col overflow-hidden">

      {/* ── Topbar (48px) ────────────────────────────────────────────────── */}
      <header className="flex h-12 flex-shrink-0 items-center gap-2 border-b border-border bg-card px-3">

        {/* Collapse toggle */}
        <button
          type="button"
          onClick={() => setCollapsed((c) => !c)}
          className="flex size-7 flex-shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          title={collapsed ? '展开侧边栏' : '收起侧边栏'}
        >
          {collapsed ? (
            <PanelLeftOpen className="size-4" strokeWidth={1.5} />
          ) : (
            <PanelLeftClose className="size-4" strokeWidth={1.5} />
          )}
        </button>

        {/* Brand + org breadcrumb */}
        <div className="flex items-center gap-2">
          <div className="flex size-6 flex-shrink-0 items-center justify-center rounded bg-primary">
            <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={14} height={14} />
          </div>
          <div className="flex items-center gap-1 text-[13px]">
            <span className="font-semibold text-foreground">ModelCraft</span>
            <ChevronRight className="size-3 text-muted-foreground" strokeWidth={1.5} />
            <div className="flex items-center gap-1.5">
              <div className="flex size-5 items-center justify-center rounded bg-primary/10 text-[9px] font-semibold text-primary">
                {orgInitials || orgName[0]?.toUpperCase()}
              </div>
              <span className="font-medium text-foreground">{orgName}</span>
            </div>
          </div>
        </div>

        {/* Spacer */}
        <div className="flex-1" />

        {/* Right actions */}
        <div className="flex items-center gap-1">
          {isAdmin === true && (
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-3 text-xs"
              onClick={() => router.push(`/org/${orgName}/dashboard`)}
            >
              切换到管理视图
            </Button>
          )}
          <UserMenu
            userName={userInfo?.username || orgName}
            onLogout={() => void handleLogout()}
            onProfile={() => {}}
            onSettings={() => {}}
          />
        </div>
      </header>

      {/* ── Main Area (Sidebar + Content) ───────────────────────────────── */}
      <div className="flex flex-1 overflow-hidden">

        {/* Sidebar */}
        <aside
          className={cn(
            'flex flex-shrink-0 flex-col overflow-hidden border-r border-border bg-card transition-all duration-200',
            collapsed ? 'w-[52px]' : 'w-[240px]'
          )}
        >
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-2 pt-3">
            <nav className="flex flex-col gap-0.5">

              {/* Projects */}
              <button
                type="button"
                onClick={() => router.push(`/end-user/${orgName}/dashboard`)}
                title={collapsed ? '项目' : undefined}
                className={cn(
                  'flex w-full items-center rounded-md border-l-[3px] py-2 text-[13px] font-medium transition-colors duration-150',
                  collapsed ? 'justify-center px-2' : 'gap-2.5 px-3',
                  activePage === 'projects'
                    ? 'border-l-primary bg-primary/[0.08] text-primary'
                    : 'border-l-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                )}
              >
                <Image
                  src="/icons/icon-folder-open.svg"
                  alt="项目"
                  width={16}
                  height={16}
                  className={cn(
                    'size-4 flex-shrink-0',
                    activePage === 'projects' ? 'opacity-100' : 'opacity-50'
                  )}
                />
                {!collapsed && <span className="flex-1 text-left">项目</span>}
              </button>

              {/* CLI */}
              <button
                type="button"
                onClick={() => router.push(`/end-user/${orgName}/dashboard/cli`)}
                title={collapsed ? 'CLI 下载' : undefined}
                className={cn(
                  'flex w-full items-center rounded-md border-l-[3px] py-2 text-[13px] font-medium transition-colors duration-150',
                  collapsed ? 'justify-center px-2' : 'gap-2.5 px-3',
                  activePage === 'cli'
                    ? 'border-l-primary bg-primary/[0.08] text-primary'
                    : 'border-l-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                )}
              >
                <Terminal
                  className={cn(
                    'size-4 flex-shrink-0',
                    activePage === 'cli' ? 'opacity-100' : 'opacity-50'
                  )}
                  strokeWidth={1.5}
                />
                {!collapsed && <span className="flex-1 text-left">CLI 下载</span>}
              </button>

            </nav>
          </div>
        </aside>

        {/* Content area */}
        <main className="flex-1 overflow-hidden">
          {children}
        </main>
      </div>
    </div>
  )
}
