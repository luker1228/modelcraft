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
  contentClassName?: string
  className?: string
  renderContent: (tab: DataWorkspaceTab) => ReactNode
}

export function DataWorkspacePanel({
  tabs,
  activeTabId,
  onTabChange,
  onTabClose,
  emptyText = '从左侧选择模型以打开数据表',
  contentClassName = '',
  className = '',
  renderContent,
}: DataWorkspacePanelProps) {
  const activeTab = useMemo(
    () => tabs.find((tab) => tab.id === activeTabId) ?? null,
    [tabs, activeTabId]
  )

  return (
    <section className={`flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-[0_2px_4px_rgba(0,0,0,0.04),0_4px_8px_rgba(0,0,0,0.04),0_1px_1px_rgba(0,0,0,0.02)] ${className}`}>
      {/* Tab bar */}
      <div className="flex items-center gap-0 overflow-x-auto border-b border-border bg-card">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTabId
          const displayTitle = (tab.title || tab.name).trim()
          return (
            <div
              key={tab.id}
              className={`flex shrink-0 items-center gap-1.5 border-b-2 px-4 py-2.5 transition-colors ${
                isActive
                  ? 'border-b-primary bg-card text-foreground'
                  : 'border-b-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              <button
                onClick={() => onTabChange(tab.id)}
                className="flex items-center gap-1.5 text-left"
              >
                <span className={`font-mono text-[12px] ${isActive ? 'font-medium' : 'font-normal'}`}>
                  {displayTitle}
                </span>
                {tab.title && tab.title !== tab.name && (
                  <span className="text-[11px] text-muted-foreground">
                    ({tab.name})
                  </span>
                )}
              </button>
              <button
                onClick={() => onTabClose(tab.id)}
                className="flex size-4 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                aria-label={`关闭 ${displayTitle}`}
              >
                <X className="size-3" />
              </button>
            </div>
          )
        })}
      </div>

      {!activeTab ? (
        <div className="flex min-h-0 flex-1 items-center justify-center text-[13px] text-muted-foreground">
          {emptyText}
        </div>
      ) : (
        <div className={`flex min-h-0 flex-1 flex-col ${contentClassName}`}>{renderContent(activeTab)}</div>
      )}
    </section>
  )
}
