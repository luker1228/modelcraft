'use client'

import * as React from 'react'
import { cn } from '@/shared/utils'

interface EditorLayoutProps {
  children: React.ReactNode
  className?: string
}

interface EditorSidebarContainerProps {
  children: React.ReactNode
  className?: string
  width?: number | string
}

interface EditorContentProps {
  children: React.ReactNode
  className?: string
}

interface EditorEmptyStateProps {
  icon?: React.ReactNode
  title: string
  description?: string
  action?: React.ReactNode
  className?: string
}

/**
 * Editor Layout - Main container for editor-style interfaces
 * 
 * Usage:
 * <EditorLayout>
 *   <EditorSidebarContainer>
 *     <EditorSidebar ... />
 *   </EditorSidebarContainer>
 *   <EditorContent>
 *     {content}
 *   </EditorContent>
 * </EditorLayout>
 */
export function EditorLayout({ children, className }: EditorLayoutProps) {
  return (
    <div className={cn('flex h-full overflow-hidden bg-slate-50', className)}>
      {children}
    </div>
  )
}

/**
 * Container for the editor sidebar
 */
export function EditorSidebarContainer({
  children,
  className,
  width = 280,
}: EditorSidebarContainerProps) {
  return (
    <div
      className={cn('flex-shrink-0 h-full', className)}
      style={{ width: typeof width === 'number' ? `${width}px` : width }}
    >
      {children}
    </div>
  )
}

/**
 * Main content area of the editor
 */
export function EditorContent({ children, className }: EditorContentProps) {
  return (
    <div className={cn('flex-1 h-full overflow-hidden', className)}>
      {children}
    </div>
  )
}

/**
 * Empty state component for when no item is selected
 */
export function EditorEmptyState({
  icon,
  title,
  description,
  action,
  className,
}: EditorEmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center h-full text-center p-8',
        className
      )}
    >
      {icon && (
        <div className="mb-4 flex size-16 items-center justify-center rounded-2xl bg-slate-100">
          {icon}
        </div>
      )}
      <h3 className="mb-2 text-lg font-semibold text-foreground">{title}</h3>
      {description && (
        <p className="mb-6 max-w-sm text-sm text-muted-foreground">{description}</p>
      )}
      {action}
    </div>
  )
}
