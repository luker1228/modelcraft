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
import { useProjectStore } from '@web/stores/project'
import { getToken, getUserInfoFromToken, removeToken } from '@bff/auth/public'
import {
  Sparkles,
  RefreshCw,
  Search,
  Bell,
  HelpCircle,
  FolderOpen,
  Users,
  Settings,
  Check,
  PanelLeftClose,
  PanelLeft,
  LayoutDashboard,
  Server,
  Table2,
  List,
  LogIn,
  Shield,
} from 'lucide-react'
import { cn } from '@/shared/utils'

interface AppLayoutProps {
  children: ReactNode
  /** Whether to show project-specific navigation in sidebar */
  showProjectNav?: boolean
  /** Current page title for breadcrumb */
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
  icon: React.ComponentType<{ className?: string; strokeWidth?: number }>
  href: string
}

/**
 * AppLayout - Global Application Layout
 *
 * Full-width topbar layout:
 * - Topbar spans full width: left org selector + center breadcrumb + right actions
 * - Main area: sidebar + content
 *
 * Design specs:
 * - Topbar height: 56px
 * - Sidebar width: 200px (expanded) / 64px (collapsed)
 * - Logo background: #2563eb
 * - Selected nav: #dadee5
 */
export function AppLayout({
  children,
  showProjectNav = false,
  pageTitle,
}: AppLayoutProps) {
  const router = useRouter()
  const params = useParams()
  const pathname = usePathname()

  const token = getToken()
  const userInfo = getUserInfoFromToken(token || '')
  const [storedUserName, setStoredUserName] = useState('')
  const displayName = userInfo?.name || userInfo?.userName || storedUserName || userInfo?.phone || 'User'

  // Sidebar collapse state
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  // State for dropdowns
  const [memberships, setMemberships] = useState<MembershipInfo[]>([])
  const [orgSearchQuery, setOrgSearchQuery] = useState('')

  // Get memberships from store
  const storedMemberships = useOrganizationStore((state) => state.memberships)
  const loadMembershipsStore = useOrganizationStore((state) => state.loadMemberships)

  // Current context
  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  useEffect(() => {
    setStoredUserName(localStorage.getItem('defaultUserName') || '')
  }, [])

  // Fetch memberships once from store
  useEffect(() => {
    if (storedMemberships && storedMemberships.length > 0) {
      setMemberships(storedMemberships)
      return
    }

    const token = getToken()
    if (!token) return

    loadMembershipsStore(token, false).then((data) => {
      setMemberships(data)
    }).catch((error) => {
      console.error('[AppLayout] Failed to fetch memberships:', error)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Get projects from store (needed for currentProject)
  const { projects: storeProjects } = useProjectStore()

  // Filter orgs by search query
  const filteredOrgs = useMemo(() => {
    if (!orgSearchQuery) return memberships
    return memberships.filter(
      (m) =>
        m.displayName.toLowerCase().includes(orgSearchQuery.toLowerCase()) ||
        m.orgName.toLowerCase().includes(orgSearchQuery.toLowerCase())
    )
  }, [memberships, orgSearchQuery])

  // Current org display
  const currentOrg = useMemo(() => {
    return memberships.find((m) => m.orgName === orgName)
  }, [memberships, orgName])

  // Current project display
  const currentProject = useMemo(() => {
    return storeProjects.find((p) => p.slug === projectSlug)
  }, [storeProjects, projectSlug])

  // Handlers
  const handleLogout = useCallback(() => {
    removeToken()
    localStorage.removeItem('defaultUserName')
    localStorage.removeItem('defaultOrgName')
    useOrganizationStore.getState().clearOrganization()
    router.push('/login')
  }, [router])

  const handleRefresh = useCallback(() => {
    window.location.reload()
  }, [])

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

  // Check if nav item is active
  const isNavActive = useCallback(
    (href: string) => {
      return pathname === href || pathname?.startsWith(href + '/')
    },
    [pathname]
  )

  // Workspace navigation items
  const workspaceNavItems: NavItem[] = [
    { label: '项目', icon: FolderOpen, href: `/org/${orgName}/workspace` },
    { label: '团队', icon: Users, href: `/org/${orgName}/team` },
    { label: '组织设置', icon: Settings, href: `/org/${orgName}/settings` },
  ]

  const projectNavItems: NavItem[] = [
    { label: '数据模型', icon: Table2, href: `/org/${orgName}/project/${projectSlug}/model-editor` },
    { label: '项目设置', icon: Settings, href: `/org/${orgName}/project/${projectSlug}/settings` },
    { label: '枚举管理', icon: List, href: `/org/${orgName}/project/${projectSlug}/enums` },
    { label: '认证与授权', icon: Shield, href: `/org/${orgName}/project/${projectSlug}/rls-settings` },
  ]

  const authNavItems: NavItem[] = [
    { label: 'RLS 设置', icon: Shield, href: `/org/${orgName}/project/${projectSlug}/rls-settings` },
    { label: '登录配置', icon: LogIn, href: `/org/${orgName}/project/${projectSlug}/login-settings` },
    { label: '用户管理', icon: Users, href: `/org/${orgName}/project/${projectSlug}/end-users` },
  ]

  const navItems = showProjectNav ? projectNavItems : workspaceNavItems

  const authBasePath = `/org/${orgName}/project/${projectSlug}`
  const isAuthSection = pathname === `${authBasePath}/rls-settings`
    || pathname?.startsWith(`${authBasePath}/rls-settings/`)
    || pathname === `${authBasePath}/login-settings`
    || pathname?.startsWith(`${authBasePath}/login-settings/`)
    || pathname === `${authBasePath}/end-users`
    || pathname?.startsWith(`${authBasePath}/end-users/`)

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* ===== Full-Width Topbar ===== */}
      <header className="flex h-14 flex-shrink-0 items-center border-b border-gray-200 bg-white px-4">
        {/* Left Section - Organization Selector */}
        <div className="w-auto flex-shrink-0 transition-all duration-300">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="flex items-center justify-center rounded-lg p-2 transition-colors hover:bg-[#fafafa]">
                <div
                  className="flex size-8 flex-shrink-0 items-center justify-center rounded-lg"
                  style={{ background: '#2563eb' }}
                >
                  <Sparkles className="size-4 text-white" strokeWidth={1.5} />
                </div>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-64" align="start">
              <div className="p-2">
                <div className="relative">
                  <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
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
                    <span>{org.displayName}</span>
                    {org.orgName === orgName && (
                      <Check className="size-4 text-blue-600" />
                    )}
                  </DropdownMenuItem>
                ))}
                {filteredOrgs.length === 0 && (
                  <div className="px-3 py-2 text-sm text-muted-foreground">未找到组织</div>
                )}
              </div>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Center Section - Breadcrumb */}
        <div className="flex flex-1 items-center pl-4">
          <nav className="flex items-center text-sm">
            <span className="font-medium text-foreground">
              {currentOrg?.orgName || orgName || 'Organization'}
            </span>
            <span className="mx-2 text-muted-foreground/50">/</span>
            <span className="text-foreground">
              {currentProject?.slug || projectSlug || pageTitle || '所有项目'}
            </span>
            {showProjectNav && (() => {
              const activeNavItem = projectNavItems.find((item) =>
                pathname === item.href || pathname?.startsWith(item.href + '/')
              )
              if (!activeNavItem) return null
              const activeAuthItem = isAuthSection
                ? authNavItems.find((item) => pathname === item.href || pathname?.startsWith(item.href + '/'))
                : null
              return (
                <>
                  <span className="mx-2 text-muted-foreground/50">/</span>
                  <span className="text-foreground">{activeNavItem.label}</span>
                  {activeAuthItem && (
                    <>
                      <span className="mx-2 text-muted-foreground/50">/</span>
                      <span className="text-foreground">{activeAuthItem.label}</span>
                    </>
                  )}
                </>
              )
            })()}
          </nav>
        </div>

        {/* Right Section - Actions */}
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="size-8 p-0 text-muted-foreground hover:bg-[#fafafa] hover:text-foreground"
            title="搜索"
          >
            <Search className="size-4" strokeWidth={1.5} />
          </Button>

          <Button
            variant="ghost"
            size="sm"
            className="relative size-8 p-0 text-muted-foreground hover:bg-[#fafafa] hover:text-foreground"
            title="通知"
          >
            <Bell className="size-4" strokeWidth={1.5} />
            <span className="absolute right-1.5 top-1.5 size-1.5 rounded-full bg-red-500" />
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefresh}
            className="size-8 p-0 text-muted-foreground hover:bg-[#fafafa] hover:text-foreground"
            title="刷新"
          >
            <RefreshCw className="size-4" strokeWidth={1.5} />
          </Button>

          <Button
            variant="ghost"
            size="sm"
            className="size-8 p-0 text-muted-foreground hover:bg-[#fafafa] hover:text-foreground"
            title="帮助"
          >
            <HelpCircle className="size-4" strokeWidth={1.5} />
          </Button>

          <div className="mx-2 h-5 w-px bg-gray-200" />

          <UserMenu
            userName={displayName}
            userEmail={userInfo?.phone}
            onLogout={handleLogout}
          />
        </div>
      </header>

      {/* ===== Main Area (Sidebar + Content) ===== */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside
          className={cn(
            "bg-white border-r border-gray-200 flex flex-col overflow-hidden flex-shrink-0 transition-all duration-300",
            sidebarCollapsed ? "w-16" : "w-[240px]"
          )}
        >
          {/* Sidebar Content */}
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-3">
            <nav>
              {navItems.map((item) => (
                <a
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "mb-0.5 flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-all",
                    isNavActive(item.href)
                      ? "bg-blue-100 text-blue-600"
                      : "text-muted-foreground hover:bg-gray-100 hover:text-foreground"
                  )}
                >
                  <item.icon className="size-4 flex-shrink-0" strokeWidth={1.5} />
                  {!sidebarCollapsed && <span>{item.label}</span>}
                </a>
              ))}
            </nav>
          </div>

          {/* Sidebar Footer - Toggle */}
          <div className="flex h-12 flex-shrink-0 items-center border-t border-gray-200 px-2">
            <button
              onClick={toggleSidebar}
              className="flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-[#fafafa] hover:text-foreground"
            >
              {sidebarCollapsed ? (
                <PanelLeft className="size-4" strokeWidth={1.5} />
              ) : (
                <PanelLeftClose className="size-4" strokeWidth={1.5} />
              )}
            </button>
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1 overflow-hidden bg-[#fafafa]">
          {showProjectNav && isAuthSection ? (
            <div className="flex h-full">
              <aside className="w-[200px] flex-shrink-0 overflow-y-auto border-r border-gray-200 bg-white">
                <div className="p-3">
                  <p className="px-2 text-sm font-semibold text-foreground">
                    认证与授权
                  </p>
                  <div className="my-2 border-t border-gray-200" />
                  <nav className="flex flex-col gap-0.5">
                    {authNavItems.map((item) => (
                      <a
                        key={item.href}
                        href={item.href}
                        className={cn(
                          'flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150',
                          isNavActive(item.href)
                            ? 'bg-blue-100 text-blue-600'
                            : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                        )}
                      >
                        <item.icon className="size-4 flex-shrink-0" strokeWidth={1.5} />
                        <span>{item.label}</span>
                      </a>
                    ))}
                  </nav>
                </div>
              </aside>

              <div className="flex-1 overflow-y-auto">
                {children}
              </div>
            </div>
          ) : (
            <div className="h-full overflow-y-auto">
              {children}
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
