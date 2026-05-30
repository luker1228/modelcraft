'use client'

import { ReactNode, useMemo, useCallback } from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { ApolloProvider, useQuery } from '@apollo/client'
import { FolderOpen, ChevronRight } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { UserMenu } from '@web/components/features/layout/UserMenu'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import { cn } from '@/shared/utils'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

type ActiveSection = 'projects' | 'tokens'

interface EndUserAppLayoutProps {
  children: ReactNode
  orgName: string
  /** Current project slug — highlights the matching item in the sidebar */
  projectSlug?: string
  /** Which sidebar section is currently active */
  activeSection?: ActiveSection
  /** Called when a different section is selected */
  onSectionChange?: (section: ActiveSection) => void
}

/**
 * EndUserAppLayout — End-user facing application shell.
 *
 * Mirrors AppLayout visually (Topbar 48px + Sidebar 240px + Content),
 * but wired to the end-user auth store instead of the admin JWT.
 *
 * Reuses admin-stack primitives:
 * - `UserMenu`   — avatar + logout dropdown
 * - Same Topbar/Sidebar structure and design tokens as AppLayout
 *
 * Provides its own ApolloProvider (org-scoped) for sidebar project list.
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
  projectSlug,
  activeSection,
  onSectionChange,
}: EndUserAppLayoutProps) {
  const router = useRouter()
  const isAdmin = useEndUserAuthStore((s) => s.isAdmin)
  const userInfo = useEndUserAuthStore((s) => s.userInfo)

  // Sidebar: load accessible projects
  const { data, loading: projectsLoading } = useQuery<{
    endUserProjects: EndUserAccessibleProject[]
  }>(END_USER_PROJECTS)

  const projects = data?.endUserProjects ?? []

  const orgInitials = orgName
    .split(/[-_]/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('')

  const handleLogout = useCallback(async () => {
    await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    })
    useEndUserAuthStore.getState().clearSession()
    router.push(`/end-user/${orgName}/login`)
  }, [orgName, router])

  return (
    <div className="flex h-full flex-col overflow-hidden">

      {/* ── Topbar (48px) — mirrors AppLayout header ────────────────────── */}
      <header className="flex h-12 flex-shrink-0 items-center gap-3 border-b border-border bg-card px-4">

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
              管理端
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

        {/* Sidebar — mirrors AppLayout aside */}
        <aside className="flex w-[240px] flex-shrink-0 flex-col overflow-hidden border-r border-border bg-card">
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-2 pt-3">

            {/* Projects section */}
            <div>
              <div className="mb-1 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                项目
              </div>
              <nav className="flex flex-col gap-0.5">
                {projectsLoading ? (
                  <div className="flex flex-col gap-0.5 px-2 py-1">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="h-8 animate-pulse rounded-md bg-muted/60" />
                    ))}
                  </div>
                ) : projects.length === 0 ? (
                  <div className="flex flex-col items-center py-8 text-center text-muted-foreground">
                    <FolderOpen className="mb-2 size-5 opacity-20" strokeWidth={1.5} />
                    <p className="text-xs">暂无项目</p>
                  </div>
                ) : (
                  projects.map((project) => {
                    const active = projectSlug === project.slug
                    return (
                      <a
                        key={project.slug}
                        href={`/end-user/${orgName}/projects/${project.slug}/data`}
                        onClick={(e) => {
                          e.preventDefault()
                          router.push(`/end-user/${orgName}/projects/${project.slug}/data`)
                        }}
                        className={cn(
                          'flex items-center gap-2.5 rounded-md border-l-[3px] px-3 py-2 text-[13px] font-medium transition-colors duration-150',
                          active
                            ? 'border-l-primary bg-primary/[0.08] text-primary'
                            : 'border-l-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                        )}
                      >
                        <Image
                          src="/icons/icon-folder-open.svg"
                          alt={project.title}
                          width={16}
                          height={16}
                          className={cn('size-4 flex-shrink-0', active ? 'opacity-100' : 'opacity-50')}
                        />
                        <span className="min-w-0 flex-1 truncate">{project.title}</span>
                      </a>
                    )
                  })
                )}
              </nav>
            </div>

            {/* Settings section */}
            <div className="mt-4">
              <div className="mb-1 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                设置
              </div>
              <nav className="flex flex-col gap-0.5">
                <button
                  type="button"
                  onClick={() => onSectionChange?.('tokens')}
                  className={cn(
                    'flex w-full items-center gap-2.5 rounded-md border-l-[3px] px-3 py-2 text-[13px] font-medium transition-colors duration-150',
                    activeSection === 'tokens'
                      ? 'border-l-primary bg-primary/[0.08] text-primary'
                      : 'border-l-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                  )}
                >
                  <Image
                    src="/icons/icon-key-round.svg"
                    alt="Token 管理"
                    width={16}
                    height={16}
                    className={cn(
                      'size-4 flex-shrink-0',
                      activeSection === 'tokens' ? 'opacity-100' : 'opacity-50'
                    )}
                  />
                  <span className="flex-1 text-left">Token 管理</span>
                  <span className="rounded bg-muted px-1 py-0.5 text-[9px] text-muted-foreground">
                    待实现
                  </span>
                </button>
              </nav>
            </div>
          </div>
        </aside>

        {/* Content area — mirrors AppLayout main */}
        <main className="flex-1 overflow-hidden bg-card">
          <div className="h-full overflow-y-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
