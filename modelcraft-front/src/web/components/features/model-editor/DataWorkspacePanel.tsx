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
      <div className="flex items-center gap-1 overflow-x-auto border-b border-border px-3 py-2">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTabId
          return (
            <div
              key={tab.id}
              className={`flex shrink-0 items-center gap-1 rounded-md border px-2 py-1 text-xs ${
                isActive
                  ? 'border-blue-200 bg-blue-50 text-blue-700'
                  : 'border-border bg-muted/30 text-muted-foreground'
              }`}
            >
              <button onClick={() => onTabChange(tab.id)}>
                {(tab.title || tab.name).trim()} ({tab.name})
              </button>
              <button
                onClick={() => onTabClose(tab.id)}
                className="rounded p-0.5 hover:bg-black/5"
                aria-label={`关闭 ${(tab.title || tab.name).trim()}`}
              >
                <X className="size-3" />
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
