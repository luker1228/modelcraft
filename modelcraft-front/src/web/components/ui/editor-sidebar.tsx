'use client'

import * as React from 'react'
import { useState, useMemo, useCallback, useRef, useEffect } from 'react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { ScrollArea } from '@web/components/ui/scroll-area'
import { cn } from '@/shared/utils'
import {
  Search,
  Plus,
  ChevronDown,
  Filter,
  MoreVertical,
  X,
} from 'lucide-react'

export interface EditorSidebarItem {
  id: string
  name: string
  title?: string
  icon?: React.ReactNode
  badge?: string | number
  group?: string
}

export interface EditorSidebarGroup {
  id: string
  label: string
  icon?: React.ReactNode
  defaultOpen?: boolean
}

interface EditorSidebarProps {
  title: string
  items: EditorSidebarItem[]
  groups?: EditorSidebarGroup[]
  selectedId?: string | null
  onSelect?: (item: EditorSidebarItem) => void
  onAdd?: () => void
  onItemAction?: (action: string, item: EditorSidebarItem) => void
  addButtonLabel?: string
  searchPlaceholder?: string
  emptyMessage?: string
  className?: string
  headerExtra?: React.ReactNode
  renderItem?: (item: EditorSidebarItem, isSelected: boolean) => React.ReactNode
}

/**
 * Enhanced Editor Sidebar Component
 *
 * A polished sidebar panel for editor-like interfaces with:
 * - Real-time search with clear button
 * - Keyboard navigation support (Arrow keys, Enter, Escape)
 * - Collapsible groups with smooth animations
 * - Item selection with visual feedback
 * - Optional action menu per item
 * - Performance optimizations (memoization, callbacks)
 * - Accessibility features (ARIA labels, focus management)
 */
export function EditorSidebar({
  title,
  items,
  groups,
  selectedId,
  onSelect,
  onAdd,
  onItemAction,
  addButtonLabel = '新建',
  searchPlaceholder = '搜索...',
  emptyMessage = '暂无数据',
  className,
  headerExtra,
  renderItem,
}: EditorSidebarProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(() => {
    const initial = new Set<string>()
    groups?.forEach(g => {
      if (g.defaultOpen !== false) initial.add(g.id)
    })
    return initial
  })

  const searchInputRef = useRef<HTMLInputElement>(null)
  const [focusedIndex, setFocusedIndex] = useState<number>(-1)

  // Filter items based on search query
  const filteredItems = useMemo(() => {
    if (!searchQuery.trim()) return items
    const query = searchQuery.toLowerCase()
    return items.filter(
      item =>
        item.name.toLowerCase().includes(query) ||
        item.title?.toLowerCase().includes(query) ||
        item.badge?.toString().toLowerCase().includes(query)
    )
  }, [items, searchQuery])

  // Group items
  const groupedItems = useMemo(() => {
    if (!groups || groups.length === 0) {
      return { ungrouped: filteredItems }
    }

    const result: Record<string, EditorSidebarItem[]> = {}
    groups.forEach(g => {
      result[g.id] = []
    })
    result['ungrouped'] = []

    filteredItems.forEach(item => {
      const groupId = item.group || 'ungrouped'
      if (result[groupId]) {
        result[groupId].push(item)
      } else {
        result['ungrouped'].push(item)
      }
    })

    return result
  }, [filteredItems, groups])

  // Toggle group expansion
  const toggleGroup = useCallback((groupId: string) => {
    setExpandedGroups(prev => {
      const next = new Set(prev)
      if (next.has(groupId)) {
        next.delete(groupId)
      } else {
        next.add(groupId)
      }
      return next
    })
  }, [])

  // Clear search
  const clearSearch = useCallback(() => {
    setSearchQuery('')
    searchInputRef.current?.focus()
  }, [])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target !== searchInputRef.current) return

      if (e.key === 'Escape') {
        clearSearch()
        searchInputRef.current?.blur()
      } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        e.preventDefault()
        const direction = e.key === 'ArrowDown' ? 1 : -1
        setFocusedIndex(prev => {
          const newIndex = prev + direction
          return Math.max(-1, Math.min(newIndex, filteredItems.length - 1))
        })
      } else if (e.key === 'Enter' && focusedIndex >= 0) {
        e.preventDefault()
        const item = filteredItems[focusedIndex]
        if (item && onSelect) {
          onSelect(item)
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [clearSearch, filteredItems, focusedIndex, onSelect])

  // Focus management for keyboard navigation
  useEffect(() => {
    if (focusedIndex >= 0 && focusedIndex < filteredItems.length) {
      const item = filteredItems[focusedIndex]
      const element = document.getElementById(`sidebar-item-${item.id}`)
      element?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }
  }, [focusedIndex, filteredItems])

  const renderDefaultItem = useCallback((item: EditorSidebarItem, isSelected: boolean, index?: number) => {
    const isFocused = index !== undefined && index === focusedIndex

    return (
      <div
        id={`sidebar-item-${item.id}`}
        className={cn(
          'group relative flex items-center gap-2 px-3 py-2.5 rounded-lg cursor-pointer transition-all duration-200',
          isSelected
            ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/25 scale-[1.02]'
            : isFocused
            ? 'bg-blue-50 text-foreground ring-2 ring-blue-300'
            : 'text-muted-foreground hover:bg-primary/10 hover:text-foreground hover:scale-[1.01]'
        )}
        onClick={() => onSelect?.(item)}
        role="listitem"
        tabIndex={isSelected ? 0 : -1}
        aria-label={item.title || item.name}
        aria-current={isSelected ? 'true' : undefined}
      >
        {/* Selection indicator */}
        {isSelected && (
          <div className="absolute left-0 top-1/2 h-8 w-1 -translate-y-1/2 rounded-r-full bg-white opacity-80 shadow-sm" />
        )}

        {item.icon && (
          <span className={cn(
            "flex-shrink-0 transition-all duration-200",
            isSelected ? "text-white scale-110" : "text-blue-500 group-hover:text-blue-600 group-hover:scale-110"
          )}>
            {item.icon}
          </span>
        )}
        <span className={cn(
          "flex-1 truncate text-sm font-medium transition-all",
          isSelected ? "font-mono tracking-tight" : "font-sans"
        )}>
          {item.title || item.name}
        </span>
        {item.badge !== undefined && (
          <span className={cn(
            "flex-shrink-0 text-xs px-2 py-0.5 rounded-md font-mono font-semibold",
            isSelected
              ? "bg-white/20 text-white"
              : "text-muted-foreground bg-slate-100"
          )}>
            {item.badge}
          </span>
        )}
        {onItemAction && (
          <button
            className={cn(
              "flex-shrink-0 opacity-0 group-hover:opacity-100 p-1.5 rounded-md transition-all",
              isSelected ? "hover:bg-white/20" : "hover:bg-selected"
            )}
            onClick={(e) => {
              e.stopPropagation()
              onItemAction('more', item)
            }}
            aria-label="更多操作"
          >
            <MoreVertical className="size-3.5" />
          </button>
        )}
      </div>
    )
  }, [focusedIndex, onSelect, onItemAction])

  const renderGroupSection = useCallback((groupId: string, groupItems: EditorSidebarItem[], startIndex: number) => {
    const group = groups?.find(g => g.id === groupId)
    if (!group) return null

    const isExpanded = expandedGroups.has(groupId)
    const isEmpty = groupItems.length === 0

    return (
      <div key={groupId} className="mb-4">
        <button
          className="group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-semibold uppercase tracking-widest text-foreground transition-all hover:bg-blue-50/50 hover:text-blue-600"
          onClick={() => toggleGroup(groupId)}
          aria-expanded={isExpanded}
          aria-controls={`group-${groupId}`}
        >
          <ChevronDown className={cn(
            "w-3.5 h-3.5 transition-transform duration-200 text-blue-500",
            !isExpanded && "-rotate-90"
          )} />
          {group.icon && <span className="text-blue-500 transition-transform group-hover:scale-110">{group.icon}</span>}
          <span className="flex-1 text-left">{group.label}</span>
          {!isEmpty && (
            <span className="rounded-full bg-blue-100 px-2 py-0.5 font-mono text-xs font-semibold tabular-nums text-blue-600">
              {groupItems.length}
            </span>
          )}
        </button>

        {isExpanded && !isEmpty && (
          <div
            id={`group-${groupId}`}
            className="animate-slideDown mt-2 space-y-1"
            role="group"
            aria-label={group.label}
          >
            {groupItems.map((item, index) => (
              <div
                key={item.id}
                style={{
                  animation: `fadeInUp 0.3s ease-out ${index * 0.04}s backwards`
                }}
              >
                {renderItem
                  ? renderItem(item, selectedId === item.id)
                  : renderDefaultItem(item, selectedId === item.id, startIndex + index)}
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }, [groups, expandedGroups, toggleGroup, selectedId, renderItem, renderDefaultItem])

  return (
    <div
      className={cn(
        'flex flex-col h-full bg-white border-r border-slate-200',
        className
      )}
    >
      {/* Header */}
      <div className="flex h-14 flex-shrink-0 items-center border-b border-slate-200 px-4">
        <h2 className="text-sm font-semibold text-foreground">{title}</h2>
      </div>

      {/* Actions */}
      <div className="flex-shrink-0 space-y-3 border-b border-slate-200 p-4">
        {headerExtra}

        {/* Add Button */}
        {onAdd && (
          <Button
            variant="outline"
            className="w-full justify-start gap-2 border-2 border-dashed border-blue-300 font-semibold text-blue-600 transition-all hover:border-blue-400 hover:bg-blue-50 hover:text-blue-700"
            onClick={onAdd}
          >
            <div className="flex size-5 items-center justify-center rounded-md bg-blue-100">
              <Plus className="size-3.5" />
            </div>
            {addButtonLabel}
          </Button>
        )}

        {/* Search */}
        <div className="group relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground transition-colors group-focus-within:text-blue-500" />
          <Input
            ref={searchInputRef}
            type="text"
            placeholder={searchPlaceholder}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-10 border-slate-200 bg-slate-50 pl-9 pr-20 text-sm transition-all focus:border-blue-300 focus:bg-white focus:ring-2 focus:ring-blue-100"
            aria-label="搜索项目"
          />
          {searchQuery && (
            <button
              onClick={clearSearch}
              className="absolute right-9 top-1/2 z-10 -translate-y-1/2 rounded p-1 transition-colors hover:bg-selected"
              aria-label="清除搜索"
            >
              <X className="size-3.5 text-muted-foreground" />
            </button>
          )}
          <button
            className="group absolute right-2 top-1/2 z-10 -translate-y-1/2 rounded-md p-1.5 transition-colors hover:bg-blue-100"
            aria-label="筛选选项"
          >
            <Filter className="size-3.5 text-muted-foreground transition-colors group-hover:text-blue-600" />
          </button>
        </div>
      </div>

      {/* Items List */}
      <ScrollArea className="flex-1">
        <div className="p-2" role="list" aria-label="项目列表">
          {filteredItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center px-4 py-12">
              <div className="mb-3 flex size-12 items-center justify-center rounded-full bg-slate-100">
                <Search className="size-6 text-muted-foreground" />
              </div>
              <p className="text-center text-sm text-muted-foreground">
                {searchQuery ? `未找到包含 "${searchQuery}" 的项目` : emptyMessage}
              </p>
              {searchQuery && (
                <button
                  onClick={clearSearch}
                  className="mt-2 text-xs text-blue-600 transition-colors hover:text-blue-700 hover:underline"
                >
                  清除搜索
                </button>
              )}
            </div>
          ) : groups && groups.length > 0 ? (
            // Render grouped items
            <>
              {(() => {
                let currentIndex = 0
                return groups.map(group => {
                  const groupItems = groupedItems[group.id] || []
                  if (groupItems.length === 0) return null
                  const startIndex = currentIndex
                  currentIndex += groupItems.length
                  return renderGroupSection(group.id, groupItems, startIndex)
                })
              })()}
              {/* Render ungrouped items */}
              {groupedItems['ungrouped']?.length > 0 && (
                <div className="space-y-1" role="group" aria-label="未分组项目">
                  {groupedItems['ungrouped'].map((item, index) => (
                    <div key={item.id}>
                      {renderItem
                        ? renderItem(item, selectedId === item.id)
                        : renderDefaultItem(item, selectedId === item.id, index)}
                    </div>
                  ))}
                </div>
              )}
            </>
          ) : (
            // Render flat list
            <div className="space-y-1" role="group" aria-label="所有项目">
              {filteredItems.map((item, index) => (
                <div key={item.id}>
                  {renderItem
                    ? renderItem(item, selectedId === item.id)
                    : renderDefaultItem(item, selectedId === item.id, index)}
                </div>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
