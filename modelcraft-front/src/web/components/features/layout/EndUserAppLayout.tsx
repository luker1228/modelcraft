'use client'

import { ReactNode, useMemo, useCallback, useState } from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { ApolloProvider, useQuery } from '@apollo/client'
import { ChevronRight, Terminal, ChevronsUpDown, Check, FolderOpen, KeyRound } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { UserMenu } from '@web/components/features/layout/UserMenu'
import { AppSidebarNav, type NavSection } from '@web/components/features/layout/AppSidebarNav'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { clearEndUserSessionArtifacts } from '@shared/auth/clear-end-user-session'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import { END_USER_PROJECTS } from '@api-client/end-user/graphql-docs'
import { cn } from '@/shared/utils'

type ActivePage = 'projects' | 'cli'

interface EndUserAppLayoutProps {
  children: ReactNode
  orgName: string
  /** Which top-level page is active — used to highlight sidebar nav items */
  activePage?: ActivePage
  /** Current project slug — when provided, shows project switcher in topbar */
  projectSlug?: string
  /** Current project title — displayed in breadcrumb */
  projectTitle?: string
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

/** Project switcher dropdown shown in topbar when on a project page. */
function ProjectSwitcher({
  orgName,
  currentSlug,
  currentTitle,
}: {
  orgName: string
  currentSlug: string
  currentTitle?: string
}) {
  const router = useRouter()
  const [open, setOpen] = useState(false)

  const { data, loading } = useQuery<{
    endUserProjects: Array<{ id: string; slug: string; title: string }>
  }>(END_USER_PROJECTS)
  const projects = data?.endUserProjects ?? []

  const displayName = currentTitle || currentSlug

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-1 rounded-md px-1.5 py-1 text-[13px] font-medium text-foreground transition-colors hover:bg-accent"
        >
          <span className="max-w-[160px] truncate">{displayName}</span>
          <ChevronsUpDown className="size-3 text-muted-foreground" strokeWidth={1.5} />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-[220px] border border-border p-1 shadow-lg" align="start">
        {loading ? (
          <div className="px-2.5 py-3 text-center text-xs text-muted-foreground">加载中...</div>
        ) : projects.length === 0 ? (
          <div className="px-2.5 py-3 text-center text-xs text-muted-foreground">暂无项目</div>
        ) : (
          <>
            <div className="px-2.5 pb-1 pt-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
              切换项目
            </div>
            {projects.map((p) => (
              <button
                key={p.id}
                type="button"
                className={cn(
                  'flex w-full items-center gap-2 rounded-sm px-2.5 py-1.5 text-left text-sm transition-colors',
                  p.slug === currentSlug
                    ? 'bg-accent text-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                )}
                onClick={() => {
                  setOpen(false)
                  if (p.slug !== currentSlug) {
                    router.push(`/end-user/${orgName}/projects/${p.slug}/data`)
                  }
                }}
              >
                <FolderOpen className="size-3.5 shrink-0 text-muted-foreground" strokeWidth={1.5} />
                <span className="flex-1 truncate">{p.title || p.slug}</span>
                {p.slug === currentSlug && (
                  <Check className="size-3.5 shrink-0 text-primary" strokeWidth={2} />
                )}
              </button>
            ))}
          </>
        )}
      </PopoverContent>
    </Popover>
  )
}

function EndUserAppLayoutInner({
  children,
  orgName,
  projectSlug,
  projectTitle,
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
            {projectSlug && (
              <>
                <ChevronRight className="size-3 text-muted-foreground" strokeWidth={1.5} />
                <ProjectSwitcher
                  orgName={orgName}
                  currentSlug={projectSlug}
                  currentTitle={projectTitle}
                />
              </>
            )}
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
        <AppSidebarNav
          navSections={[
            {
              header: '工作区',
              items: [
                { label: '项目', icon: '/icons/icon-folder-open.svg', href: `/end-user/${orgName}/dashboard` },
                { label: 'CLI 下载', icon: Terminal, href: `/end-user/${orgName}/dashboard/cli` },
              ],
            },
          ]}
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed((c) => !c)}
          collapsedWidth="w-[52px]"
          showCollapseToggle
        />

        {/* Content area */}
        <main className="flex-1 overflow-hidden">
          {children}
        </main>
      </div>
    </div>
  )
}
