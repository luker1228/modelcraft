'use client'

import { ReactNode, useCallback, useMemo, useState, useEffect } from 'react'
import Image from 'next/image'
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
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'
import { getCachedMemberships } from '@shared/cache/memberships-cache'
import { useProjectStore } from '@web/stores/project'
import { getToken, getUserInfoFromToken, removeToken } from '@api-client/auth/public'
import {
  refreshEndUserAccessToken,
} from '@api-client/end-user/end-user-auth-client'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import {
  Search,
  HelpCircle,
  ChevronRight,
  ChevronDown,
  Check,
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

interface NavSubItem {
  label: string
  href: string
  /** URL search param to match for active state, e.g. "tab=bundles" */
  tabParam?: string
}

interface NavItem {
  label: string
  /** LucideIcon component or path string to a custom SVG (e.g. "/icons/icon-foo.svg") */
  icon: LucideIcon | string
  href: string
  /** Optional sub-items rendered as expandable children when sidebar is expanded */
  children?: NavSubItem[]
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

  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  const token = getToken()
  const userInfo = getUserInfoFromToken(token || '')
  const [storedUserName, setStoredUserName] = useState('')
  const displayName = userInfo?.name || userInfo?.userName || storedUserName || userInfo?.phone || 'User'

  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [orgSearchQuery, setOrgSearchQuery] = useState('')
  const [hasEndUserSession, setHasEndUserSession] = useState(() => {
    const s = useEndUserAuthStore.getState()
    return !!s.accessToken && !s.isTokenExpired()
  })

  // Attempt silent refresh only when no valid session was found synchronously
  useEffect(() => {
    if (hasEndUserSession) return
    // Attempt silent refresh — fire-and-forget, no redirect
    refreshEndUserAccessToken({ orgName }).then((token) => {
      if (token) setHasEndUserSession(true)
    }).catch(() => { /* ignore */ })
  }, [orgName, hasEndUserSession])

  const storedMemberships = useOrganizationStore((state) => state.memberships)
  const loadMembershipsStore = useOrganizationStore((state) => state.loadMemberships)

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
    router.push(TENANT_LOGIN_PATH)
  }, [router])

  const handleOrgSelect = useCallback(
    (org: MembershipInfo) => {
      setOrgSearchQuery('')
      localStorage.setItem('defaultOrgName', org.orgName)
      router.push(`/org/${org.orgName}/dashboard`)
    },
    [router]
  )

  const toggleSidebar = useCallback(() => {
    setSidebarCollapsed((prev) => !prev)
  }, [])

  const isNavActive = useCallback(
    (href: string) => {
      if (pathname === href || pathname?.startsWith(href + '/')) return true
      return false
    },
    [pathname]
  )

  const isSubItemActive = useCallback(
    (sub: NavSubItem) => {
      // Match pathname first
      const pathMatch = pathname === sub.href.split('?')[0]
      if (!pathMatch) return false
      if (!sub.tabParam) return true
      // Check search param
      const paramKey = sub.tabParam.split('=')[0]
      const paramVal = sub.tabParam.split('=')[1]
      if (typeof window === 'undefined') return false
      const currentVal = new URLSearchParams(window.location.search).get(paramKey ?? '')
      // Default tab (roles) is active when no tab param or tab=roles
      if (paramVal === 'roles') return !currentVal || currentVal === 'roles'
      // Default view (schema) is active when no view param or view=schema
      if (paramVal === 'schema') return !currentVal || currentVal === 'schema'
      return currentVal === paramVal
    },
    [pathname]
  )

  // Track which nav items with children are expanded
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set())

  // Auto-expand the "访问控制" item when its route is active
  useEffect(() => {
    const rolesHref = `/org/${orgName}/project/${projectSlug}/access-control`
    if (isNavActive(rolesHref)) {
      setExpandedItems((prev) => new Set([...prev, rolesHref]))
    }
  }, [pathname, orgName, projectSlug, isNavActive])

  // Auto-expand the "数据模型" item when its route is active
  useEffect(() => {
    const modelEditorHref = `/org/${orgName}/project/${projectSlug}/model-editor`
    if (isNavActive(modelEditorHref)) {
      setExpandedItems((prev) => new Set([...prev, modelEditorHref]))
    }
  }, [pathname, orgName, projectSlug, isNavActive])

  // ── Navigation structure ──────────────────────────────────────────────────

  const workspaceNavSections: NavSection[] = [
    {
      header: '工作区',
      items: [
        { label: '项目', icon: '/icons/icon-folder-open.svg', href: `/org/${orgName}/dashboard` },
        { label: '开发者', icon: '/icons/icon-users.svg', href: `/org/${orgName}/developers` },
        { label: '终端用户', icon: '/icons/icon-key-round.svg', href: `/org/${orgName}/end-users` },
      ],
    },
    {
      header: '设置',
      items: [
        { label: '组织设置', icon: '/icons/icon-settings.svg', href: `/org/${orgName}/settings` },
      ],
    },
  ]

  const projectNavSections: NavSection[] = [
    {
      header: '数据建模',
      items: [
        { label: '数据模型', icon: '/icons/icon-table2.svg', href: `/org/${orgName}/project/${projectSlug}/model-editor`, children: [
          { label: '模型管理', href: `/org/${orgName}/project/${projectSlug}/model-editor`, tabParam: 'view=schema' },
          { label: '数据管理', href: `/org/${orgName}/project/${projectSlug}/model-editor?view=data`, tabParam: 'view=data' },
        ]},
        { label: '数据库', icon: '/icons/icon-list.svg', href: `/org/${orgName}/project/${projectSlug}/databases` },
        { label: '枚举管理', icon: '/icons/icon-list.svg', href: `/org/${orgName}/project/${projectSlug}/enums` },
      ],
    },
    {
      header: '权限管理',
      items: [
        { label: '访问控制', icon: '/icons/icon-shield.svg', href: `/org/${orgName}/project/${projectSlug}/access-control`, children: [
          { label: '角色', href: `/org/${orgName}/project/${projectSlug}/access-control?tab=roles`, tabParam: 'tab=roles' },
          { label: '权限包', href: `/org/${orgName}/project/${projectSlug}/access-control?tab=bundles`, tabParam: 'tab=bundles' },
          { label: '权限点', href: `/org/${orgName}/project/${projectSlug}/access-control?tab=permissions`, tabParam: 'tab=permissions' },
        ]},
        { label: '用户授权', icon: '/icons/icon-key-round.svg', href: `/org/${orgName}/project/${projectSlug}/end-user-access` },
      ],
    },
    {
      header: '设置',
      items: [
        { label: '项目设置', icon: '/icons/icon-settings.svg', href: `/org/${orgName}/project/${projectSlug}/settings` },
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
                <Image src="/icons/icon-model-graphql.svg" alt="ModelCraft" width={14} height={14} />
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
                    onClick={() => router.push(item.href!)}
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
            variant="outline"
            size="sm"
            className="h-8 px-3 text-xs"
            onClick={() => {
              const s = useEndUserAuthStore.getState()
              if (s.accessToken && !s.isTokenExpired()) {
                router.push(`/end-user/${orgName}/dashboard`)
                return
              }
              void refreshEndUserAccessToken({ orgName }).then((token) => {
                router.push(token ? `/end-user/${orgName}/dashboard` : `/end-user/${orgName}/login`)
              })
            }}
          >
            切换到用户视图
          </Button>
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
                    const hasChildren = !sidebarCollapsed && item.children && item.children.length > 0
                    const expanded = expandedItems.has(item.href)

                    const toggleExpand = () => {
                      if (!hasChildren) return
                      setExpandedItems((prev) => {
                        const next = new Set(prev)
                        if (next.has(item.href)) next.delete(item.href)
                        else next.add(item.href)
                        return next
                      })
                    }

                    return (
                      <div key={item.href}>
                        {/* Main nav item */}
                        <a
                          href={item.href}
                          onClick={hasChildren ? (e) => { e.preventDefault(); toggleExpand() } : undefined}
                          className={cn(
                            'flex items-center gap-2.5 rounded-md py-2 text-[13px] font-medium transition-colors duration-150',
                            sidebarCollapsed ? 'justify-center px-2' : 'px-3',
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
                          {typeof item.icon === 'string' ? (
                            <Image
                              src={item.icon}
                              alt={item.label}
                              width={16}
                              height={16}
                              className={cn('size-4 flex-shrink-0', active ? 'opacity-100' : 'opacity-50')}
                            />
                          ) : (
                            <item.icon
                              className={cn('size-4 flex-shrink-0', active ? 'text-primary' : 'text-muted-foreground')}
                              strokeWidth={1.5}
                            />
                          )}
                          {!sidebarCollapsed && (
                            <>
                              <span className="flex-1">{item.label}</span>
                              {hasChildren && (
                                expanded
                                  ? <ChevronDown className="size-3.5 text-muted-foreground/60" />
                                  : <ChevronRight className="size-3.5 text-muted-foreground/60" />
                              )}
                            </>
                          )}
                        </a>

                        {/* Sub-items */}
                        {hasChildren && expanded && (
                          <div className="ml-3 mt-0.5 flex flex-col gap-0.5 border-l border-border pl-3">
                            {item.children!.map((sub) => {
                              const subActive = isSubItemActive(sub)
                              return (
                                <a
                                  key={sub.href}
                                  href={sub.href}
                                  className={cn(
                                    'rounded-md px-2 py-1.5 text-[12px] font-medium transition-colors duration-150',
                                    subActive
                                      ? 'bg-primary/[0.06] text-primary'
                                      : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                                  )}
                                >
                                  {sub.label}
                                </a>
                              )
                            })}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </nav>
              </div>
            ))}
          </div>

          {/* Sidebar footer */}
          <div className="flex-shrink-0 border-t border-border">
            {/* Collapse toggle */}
            <div className="flex h-11 items-center px-2">
              <button
                onClick={toggleSidebar}
                className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                title={sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'}
              >
                {sidebarCollapsed ? (
                  <Image src="/icons/icon-panel-left.svg" alt="展开侧边栏" width={16} height={16} />
                ) : (
                  <Image src="/icons/icon-panel-left.svg" alt="折叠侧边栏" width={16} height={16} className="-scale-x-100" />
                )}
              </button>
            </div>
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
