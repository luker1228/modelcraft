'use client'

import { ReactNode, useMemo } from 'react'
import { X } from 'lucide-react'

export interface DataWorkspaceTab {
  id: string
  name: string
  title?: string | null
}

interface DataWorkspacePanelProps {
  tabs: DataWorkspaceTab[]
  activeTabId: string
  onTabChange: (tabId: string) => void
  onTabClose: (tabId: string) => void
  emptyText?: string
  maxContentHeightClassName?: string
  className?: string
  renderContent: (tab: DataWorkspaceTab) => ReactNode
}

export function DataWorkspacePanel({
  tabs,
  activeTabId,
  onTabChange,
  onTabClose,
  emptyText = '从左侧选择模型以打开数据表',
  maxContentHeightClassName = 'min-h-[620px]',
  className = '',
  renderContent,
}: DataWorkspacePanelProps) {
  const activeTab = useMemo(
    () => tabs.find((tab) => tab.id === activeTabId) ?? null,
    [tabs, activeTabId]
  )

  return (
    <section className={`min-w-0 flex-1 rounded-xl border border-border bg-background shadow-sm ${className}`}>
      <div className="flex items-center gap-2 overflow-x-auto border-b border-border p-3">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTabId
          const displayTitle = (tab.title || tab.name).trim()
          return (
            <div
              key={tab.id}
              className={`flex shrink-0 items-center gap-2 rounded-lg border px-3 py-2.5 transition-all ${
                isActive
                  ? 'border-blue-200 bg-blue-50 text-blue-800 shadow-sm'
                  : 'border-border bg-muted/20 text-muted-foreground hover:bg-muted/50'
              }`}
            >
              <button onClick={() => onTabChange(tab.id)} className="text-left">
                <span className={`${isActive ? 'text-sm font-semibold' : 'text-sm font-medium'}`}>
                  {displayTitle} ({tab.name})
                </span>
              </button>
              <button
                onClick={() => onTabClose(tab.id)}
                className="rounded p-0.5 hover:bg-black/5"
                aria-label={`关闭 ${displayTitle}`}
              >
                <X className="size-3.5" />
              </button>
            </div>
          )
        })}
      </div>

      {!activeTab ? (
        <div className="flex min-h-[420px] items-center justify-center text-sm text-muted-foreground">
          {emptyText}
        </div>
      ) : (
        <div className={maxContentHeightClassName}>{renderContent(activeTab)}</div>
      )}
    </section>
  )
}
