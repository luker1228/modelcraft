'use client'

import { ReactNode, useCallback, useMemo, useState, useEffect } from 'react'
import { useRouter, useParams, usePathname } from 'next/navigation'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { UserMenu } from '@web/components/features/layout/UserMenu'
import { useOrganizationStore } from '@shared/stores/organization'
import { getCachedMemberships } from '@shared/cache/memberships-cache'
import { useProjectStore } from '@web/stores/project'
import { getToken, getUserInfoFromToken, removeToken } from '@bff/auth/public'
import {
  Sparkles,
  Search,
  HelpCircle,
  ChevronRight,
  FolderOpen,
  Users,
  Settings,
  Check,
  PanelLeftClose,
  PanelLeft,
  Table2,
  List,
  Shield,
  KeyRound,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/utils'
import { buildAppLayoutBreadcrumbs } from './app-layout-breadcrumbs'

interface AppLayoutProps {
  children: ReactNode
  /** Whether to show project-specific navigation in sidebar */
  showProjectNav?: boolean
  /** Kept for backward-compat, no longer used for breadcrumb */
  pageTitle?: string
}

interface MembershipInfo {
  orgId: string
  orgName: string
  displayName: string
  role: string
}

interface NavItem {
  label: string
  icon: LucideIcon
  href: string
}

interface NavSection {
  header?: string
  items: NavItem[]
}

/**
 * AppLayout — Global Application Layout
 *
 * Design specs (Stripe-inspired):
 * - Topbar height: 48px
 * - Sidebar width: 240px (expanded) / 64px (collapsed)
 * - Sidebar: section headers + indigo active state (left stripe)
 * - Content bg: bg-background (#F6F8FA)
 */
export function AppLayout({
  children,
  showProjectNav = false,
}: AppLayoutProps) {
  const router = useRouter()
  const params = useParams()
  const pathname = usePathname()

  const token = getToken()
  const userInfo = getUserInfoFromToken(token || '')
  const [storedUserName, setStoredUserName] = useState('')
  const displayName = userInfo?.name || userInfo?.userName || storedUserName || userInfo?.phone || 'User'

  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [orgSearchQuery, setOrgSearchQuery] = useState('')

  const storedMemberships = useOrganizationStore((state) => state.memberships)
  const loadMembershipsStore = useOrganizationStore((state) => state.loadMemberships)

  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  useEffect(() => {
    setStoredUserName(localStorage.getItem('defaultUserName') || '')
  }, [])

  const memberships = useMemo<MembershipInfo[]>(() => {
    if (storedMemberships.length > 0) {
      return storedMemberships as MembershipInfo[]
    }

    const cached = getCachedMemberships()
    return cached ?? []
  }, [storedMemberships])

  useEffect(() => {
    if (memberships.length > 0) return

    const token = getToken()
    if (!token) return

    loadMembershipsStore(token, false).catch((error) => {
      console.error('[AppLayout] Failed to fetch memberships:', error)
    })
  }, [memberships.length, loadMembershipsStore])

  const storeProjects = useProjectStore((state) => state.projects)
  const selectedProject = useProjectStore((state) => state.selectedProject)

  const filteredOrgs = useMemo(() => {
    if (!orgSearchQuery) return memberships
    return memberships.filter(
      (m) =>
        m.displayName.toLowerCase().includes(orgSearchQuery.toLowerCase()) ||
        m.orgName.toLowerCase().includes(orgSearchQuery.toLowerCase())
    )
  }, [memberships, orgSearchQuery])

  const currentOrg = useMemo(() => {
    return memberships.find((m) => m.orgName === orgName)
  }, [memberships, orgName])

  const currentProject = useMemo(() => {
    if (!projectSlug) return null

    if (selectedProject?.slug === projectSlug) {
      return selectedProject
    }

    return storeProjects.find((project) => project.slug === projectSlug) || null
  }, [projectSlug, selectedProject, storeProjects])

  const breadcrumbs = useMemo(() => {
    return buildAppLayoutBreadcrumbs({
      showProjectNav,
      orgName,
      orgDisplayName: currentOrg?.displayName || currentOrg?.orgName,
      projectSlug,
      projectDisplayName: currentProject?.title,
    })
  }, [showProjectNav, orgName, currentOrg?.displayName, currentOrg?.orgName, projectSlug, currentProject?.title])

  const handleLogout = useCallback(() => {
    removeToken()
    localStorage.removeItem('defaultUserName')
    localStorage.removeItem('defaultOrgName')
    useOrganizationStore.getState().clearOrganization()
    router.push('/login')
  }, [router])

  const handleOrgSelect = useCallback(
    (org: MembershipInfo) => {
      setOrgSearchQuery('')
      localStorage.setItem('defaultOrgName', org.orgName)
      router.push(`/org/${org.orgName}/workspace`)
    },
    [router]
  )

  const toggleSidebar = useCallback(() => {
    setSidebarCollapsed((prev) => !prev)
  }, [])

  const isNavActive = useCallback(
    (href: string) => {
      if (pathname === href || pathname?.startsWith(href + '/')) return true
      // rbac/* 路由高亮「访问控制」
      const rolesHref = `/org/${orgName}/project/${projectSlug}/roles`
      const rbacBase = `/org/${orgName}/project/${projectSlug}/rbac`
      if (href === rolesHref && pathname?.startsWith(rbacBase)) return true
      return false
    },
    [pathname, orgName, projectSlug]
  )

  // ── Navigation structure ──────────────────────────────────────────────────

  const workspaceNavSections: NavSection[] = [
    {
      header: '工作区',
      items: [
        { label: '项目', icon: FolderOpen, href: `/org/${orgName}/workspace` },
        { label: '团队', icon: Users, href: `/org/${orgName}/team` },
        { label: '终端用户', icon: KeyRound, href: `/org/${orgName}/end-users` },
      ],
    },
    {
      header: '设置',
      items: [
        { label: '组织设置', icon: Settings, href: `/org/${orgName}/settings` },
      ],
    },
  ]

  const projectNavSections: NavSection[] = [
    {
      header: '数据建模',
      items: [
        { label: '数据模型', icon: Table2, href: `/org/${orgName}/project/${projectSlug}/model-editor` },
        { label: '枚举管理', icon: List, href: `/org/${orgName}/project/${projectSlug}/enums` },
      ],
    },
    {
      header: '权限管理',
      items: [
        { label: '访问控制', icon: Shield, href: `/org/${orgName}/project/${projectSlug}/roles` },
        { label: '终端用户管理', icon: KeyRound, href: `/org/${orgName}/project/${projectSlug}/end-user-access` },
      ],
    },
    {
      header: '设置',
      items: [
        { label: '项目设置', icon: Settings, href: `/org/${orgName}/project/${projectSlug}/settings` },
      ],
    },
  ]

  const navSections = showProjectNav ? projectNavSections : workspaceNavSections

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div className="flex h-full flex-col overflow-hidden">

      {/* ── Topbar (48px) ─────────────────────────────────────────────────── */}
      <header className="flex h-12 flex-shrink-0 items-center gap-3 border-b border-border bg-card px-4">

        {/* Org selector */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-accent">
              <div className="flex size-6 flex-shrink-0 items-center justify-center rounded bg-primary">
                <Sparkles className="size-3.5 text-white" strokeWidth={1.5} />
              </div>
              {!sidebarCollapsed && (
                <span className="max-w-[120px] truncate text-[13px]">
                  {currentOrg?.displayName || currentOrg?.orgName || orgName}
                </span>
              )}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-64" align="start">
            <div className="p-2">
              <div className="relative">
                <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="搜索组织..."
                  value={orgSearchQuery}
                  onChange={(e) => setOrgSearchQuery(e.target.value)}
                  className="h-8 pl-8 text-sm"
                  onClick={(e) => e.stopPropagation()}
                />
              </div>
            </div>
            <DropdownMenuSeparator />
            <div className="max-h-[240px] overflow-y-auto">
              {filteredOrgs.map((org) => (
                <DropdownMenuItem
                  key={org.orgId}
                  onClick={() => handleOrgSelect(org)}
                  className="flex cursor-pointer items-center justify-between"
                >
                  <span className="text-[13px]">{org.displayName}</span>
                  {org.orgName === orgName && (
                    <Check className="size-3.5 text-primary" />
                  )}
                </DropdownMenuItem>
              ))}
              {filteredOrgs.length === 0 && (
                <div className="px-3 py-2 text-sm text-muted-foreground">未找到组织</div>
              )}
            </div>
          </DropdownMenuContent>
        </DropdownMenu>

        {breadcrumbs.length > 0 && (
          <nav className="flex min-w-0 items-center gap-1 text-xs" aria-label="breadcrumb">
            {breadcrumbs.map((item, index) => (
              <div key={`${item.label}-${index}`} className="flex min-w-0 items-center gap-1">
                {index > 0 && <ChevronRight className="size-3 text-muted-foreground" strokeWidth={1.5} />}
                {item.href ? (
                  <button
                    type="button"
                    onClick={() => router.push(item.href)}
                    className="max-w-[180px] truncate rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  >
                    {item.label}
                  </button>
                ) : (
                  <span
                    className={cn(
                      'max-w-[180px] truncate px-1 py-0.5 text-foreground',
                      item.isCurrent && 'font-medium'
                    )}
                  >
                    {item.label}
                  </span>
                )}
              </div>
            ))}
          </nav>
        )}

        {/* Spacer */}
        <div className="flex-1" />

        {/* Right actions: Help + User only */}
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="size-8 p-0 text-muted-foreground hover:bg-accent hover:text-foreground"
            title="帮助"
          >
            <HelpCircle className="size-4" strokeWidth={1.5} />
          </Button>

          <UserMenu
            userName={displayName}
            userEmail={userInfo?.phone}
            onLogout={handleLogout}
          />
        </div>
      </header>

      {/* ── Main Area (Sidebar + Content) ─────────────────────────────────── */}
      <div className="flex flex-1 overflow-hidden">

        {/* Sidebar */}
        <aside
          className={cn(
            'flex flex-shrink-0 flex-col overflow-hidden border-r border-border bg-card transition-all duration-200',
            sidebarCollapsed ? 'w-16' : 'w-[240px]'
          )}
        >
          {/* Nav sections */}
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-2 pt-3">
            {navSections.map((section, si) => (
              <div key={si} className={cn(si > 0 && 'mt-4')}>
                {/* Section header */}
                {section.header && !sidebarCollapsed && (
                  <div className="mb-1 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                    {section.header}
                  </div>
                )}
                {/* Section divider when collapsed */}
                {si > 0 && sidebarCollapsed && (
                  <div className="mb-2 border-t border-border" />
                )}
                <nav className="flex flex-col gap-0.5">
                  {section.items.map((item) => {
                    const active = isNavActive(item.href)
                    return (
                      <a
                        key={item.href}
                        href={item.href}
                        className={cn(
                          'flex items-center gap-2.5 rounded-md py-2 text-[13px] font-medium transition-colors duration-150',
                          sidebarCollapsed ? 'justify-center px-2' : 'px-3',
                          // Left border indicator — only when expanded
                          !sidebarCollapsed && 'border-l-[3px]',
                          active
                            ? [
                                'bg-primary/[0.08] text-primary',
                                !sidebarCollapsed && 'border-l-primary',
                              ]
                            : [
                                'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
                                !sidebarCollapsed && 'border-l-transparent',
                              ]
                        )}
                        title={sidebarCollapsed ? item.label : undefined}
                      >
                        <item.icon
                          className={cn('size-4 flex-shrink-0', active ? 'text-primary' : 'text-muted-foreground')}
                          strokeWidth={1.5}
                        />
                        {!sidebarCollapsed && <span>{item.label}</span>}
                      </a>
                    )
                  })}
                </nav>
              </div>
            ))}
          </div>

          {/* Sidebar footer — toggle button */}
          <div className="flex h-11 flex-shrink-0 items-center border-t border-border px-2">
            <button
              onClick={toggleSidebar}
              className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              title={sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'}
            >
              {sidebarCollapsed ? (
                <PanelLeft className="size-4" strokeWidth={1.5} />
              ) : (
                <PanelLeftClose className="size-4" strokeWidth={1.5} />
              )}
            </button>
          </div>
        </aside>

        {/* Content area */}
        <main className="flex-1 overflow-hidden bg-card">
          <div className="h-full overflow-y-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
