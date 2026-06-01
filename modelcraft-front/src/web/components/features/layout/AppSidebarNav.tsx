'use client'

import { useState, useEffect, useCallback } from 'react'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import { ChevronRight, ChevronDown, type LucideIcon } from 'lucide-react'
import { cn } from '@/shared/utils'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface NavSubItem {
  label: string
  href: string
  /** URL search param to match for active state, e.g. "tab=bundles" */
  tabParam?: string
}

export interface NavItem {
  label: string
  /** LucideIcon component or path string to a custom SVG (e.g. "/icons/icon-foo.svg") */
  icon: LucideIcon | string
  href: string
  /** Optional sub-items rendered as expandable children when sidebar is expanded */
  children?: NavSubItem[]
}

export interface NavSection {
  header?: string
  items: NavItem[]
}

interface AppSidebarNavProps {
  navSections: NavSection[]
  collapsed: boolean
  onToggleCollapse: () => void
  /** Whether to render the collapse toggle button at the sidebar bottom footer (Admin=true, EndUser=false) */
  showCollapseToggle?: boolean
  /** Width of sidebar when collapsed (default: "w-16" = 64px) */
  collapsedWidth?: string
  className?: string
}

// ── Component ─────────────────────────────────────────────────────────────────

/**
 * AppSidebarNav — Shared sidebar navigation component.
 *
 * Used by both Admin (AppLayout) and EndUser (EndUserAppLayout) to ensure
 * consistent nav item styling, active state, and collapse behaviour.
 */
export function AppSidebarNav({
  navSections,
  collapsed,
  onToggleCollapse,
  showCollapseToggle = false,
  collapsedWidth = 'w-16',
  className,
}: AppSidebarNavProps) {
  const pathname = usePathname()
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set())

  const isNavActive = useCallback(
    (href: string) => {
      if (pathname === href || pathname?.startsWith(href + '/')) return true
      return false
    },
    [pathname]
  )

  const isSubItemActive = useCallback(
    (sub: NavSubItem) => {
      const pathMatch = pathname === sub.href.split('?')[0]
      if (!pathMatch) return false
      if (!sub.tabParam) return true
      const paramKey = sub.tabParam.split('=')[0]
      const paramVal = sub.tabParam.split('=')[1]
      if (typeof window === 'undefined') return false
      const currentVal = new URLSearchParams(window.location.search).get(paramKey ?? '')
      if (paramVal === 'roles') return !currentVal || currentVal === 'roles'
      if (paramVal === 'schema') return !currentVal || currentVal === 'schema'
      return currentVal === paramVal
    },
    [pathname]
  )

  // Auto-expand items whose children contain the active route
  useEffect(() => {
    const toExpand = new Set<string>()
    for (const section of navSections) {
      for (const item of section.items) {
        if (item.children && isNavActive(item.href)) {
          toExpand.add(item.href)
        }
      }
    }
    if (toExpand.size > 0) {
      setExpandedItems((prev) => new Set([...prev, ...toExpand]))
    }
  }, [pathname, navSections, isNavActive])

  return (
    <aside
      className={cn(
        'flex flex-shrink-0 flex-col overflow-hidden border-r border-border bg-card transition-all duration-200',
        collapsed ? collapsedWidth : 'w-[240px]',
        className
      )}
    >
      {/* Nav sections */}
      <div className="flex-1 overflow-y-auto overflow-x-hidden p-2 pt-3">
        {navSections.map((section, si) => (
          <div key={si} className={cn(si > 0 && 'mt-4')}>
            {/* Section header */}
            {section.header && !collapsed && (
              <div className="mb-1 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70">
                {section.header}
              </div>
            )}
            {/* Section divider when collapsed */}
            {si > 0 && collapsed && (
              <div className="mb-2 border-t border-border" />
            )}
            <nav className="flex flex-col gap-0.5">
              {section.items.map((item) => {
                const active = isNavActive(item.href)
                const hasChildren = !collapsed && item.children && item.children.length > 0
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
                        collapsed ? 'justify-center px-2' : 'px-3',
                        !collapsed && 'border-l-[3px]',
                        active
                          ? [
                              'bg-primary/[0.08] text-primary',
                              !collapsed && 'border-l-primary',
                            ]
                          : [
                              'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
                              !collapsed && 'border-l-transparent',
                            ]
                      )}
                      title={collapsed ? item.label : undefined}
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
                      {!collapsed && (
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

      {/* Sidebar footer — collapse toggle (Admin only) */}
      {showCollapseToggle && (
        <div className="flex-shrink-0 border-t border-border">
          <div className="flex h-11 items-center px-2">
            <button
              type="button"
              onClick={onToggleCollapse}
              className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              title={collapsed ? '展开侧边栏' : '折叠侧边栏'}
            >
              {collapsed ? (
                <Image src="/icons/icon-panel-left.svg" alt="展开侧边栏" width={16} height={16} />
              ) : (
                <Image src="/icons/icon-panel-left.svg" alt="折叠侧边栏" width={16} height={16} className="-scale-x-100" />
              )}
            </button>
          </div>
        </div>
      )}
    </aside>
  )
}
