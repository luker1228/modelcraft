'use client'

import { ReactNode, useCallback, useMemo, useState, useEffect } from 'react'
import Image from 'next/image'
import { useRouter, useParams, usePathname } from 'next/navigation'
import { AppSidebarNav, type NavSection } from '@web/components/features/layout/AppSidebarNav'
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
import { useAppStore } from '@web/stores/app'
import { getToken, getUserInfoFromToken, removeToken } from '@api-client/auth/public'
import {
  Search,
  RefreshCw,
  HelpCircle,
  ChevronRight,
  Check,
  KeyRound,
} from 'lucide-react'
import { cn } from '@/shared/utils'
import { buildAppLayoutBreadcrumbs } from './app-layout-breadcrumbs'
import { buildModelEditorPath } from '@shared/routes/model-editor-path'

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
  const [contentRefreshKey, setContentRefreshKey] = useState(0)
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
  const selectedDatabase = useAppStore((state) => state.selectedDatabase)

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

  const handleRefreshContent = useCallback(() => {
    setContentRefreshKey((prev) => prev + 1)
    router.refresh()
  }, [router])

  // ── Navigation structure ──────────────────────────────────────────────────

  const workspaceNavSections: NavSection[] = [
    {
      header: '工作区',
      items: [
        { label: '项目', icon: '/icons/icon-folder-open.svg', href: `/org/${orgName}/dashboard` },
      ],
    },
    {
      header: '设置',
      items: [
        { label: '组织设置', icon: '/icons/icon-settings.svg', href: `/org/${orgName}/settings/general` },
        { label: 'API Token', icon: KeyRound, href: `/org/${orgName}/api-tokens` },
      ],
    },
  ]

  const projectNavSections: NavSection[] = [
    {
      header: '数据建模',
      items: [
        { label: '数据模型', icon: '/icons/icon-table2.svg', href: `/org/${orgName}/project/${projectSlug}/model-editor`, children: [
          { label: '模型管理', href: `/org/${orgName}/project/${projectSlug}/model-editor`, tabParam: 'view=schema' },
          {
            label: '数据管理',
            href: buildModelEditorPath(orgName, projectSlug, {
              view: 'data',
              databaseName: selectedDatabase,
            }),
            tabParam: 'view=data',
          },
        ]},
        { label: '数据库', icon: '/icons/icon-list.svg', href: `/org/${orgName}/project/${projectSlug}/databases` },
        { label: '枚举管理', icon: '/icons/icon-list.svg', href: `/org/${orgName}/project/${projectSlug}/enums` },
      ],
    },
    {
      header: '权限管理',
      items: [
        { label: '访问控制', icon: '/icons/icon-shield.svg', href: `/org/${orgName}/project/${projectSlug}/access-control` },
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

        {/* Right actions */}
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="size-8 p-0 text-muted-foreground hover:bg-accent hover:text-foreground"
            title="刷新"
            onClick={handleRefreshContent}
          >
            <RefreshCw className="size-4" strokeWidth={1.5} />
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
        <AppSidebarNav
          navSections={navSections}
          collapsed={sidebarCollapsed}
          onToggleCollapse={toggleSidebar}
          showCollapseToggle
        />

        {/* Content area */}
        <main className="flex-1 overflow-hidden bg-card">
          <div key={contentRefreshKey} className="h-full overflow-y-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
