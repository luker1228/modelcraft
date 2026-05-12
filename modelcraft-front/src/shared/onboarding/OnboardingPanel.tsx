'use client'

import { useRouter, useParams } from 'next/navigation'
import { ChevronUp, ChevronDown, X, Check, ChevronRight } from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { useOnboarding, type OnboardingPendingAction } from './OnboardingContext'

export function OnboardingPanel({ orgName }: { orgName: string }) {
  const {
    groups,
    projectSlug: storedProjectSlug,
    completedCount,
    totalCount,
    isComplete,
    panelOpen,
    expandedGroupId,
    openPanel,
    closePanel,
    dismiss,
    markStep,
    setPendingAction,
    setExpandedGroupId,
  } = useOnboarding()

  const router = useRouter()
  const params = useParams()

  const urlProjectSlug = (params.projectSlug as string | undefined) ?? null
  const projectSlug = urlProjectSlug ?? storedProjectSlug

  /** True if this step requires a project but none is available */
  const needsProject = (stepId: string): boolean => {
    const projectScopedIds = ['select_database', 'create_model', 'insert_column', 'insert_data', 'create_permission', 'create_bundle', 'create_role']
    return projectScopedIds.includes(stepId) && !projectSlug
  }

  if (isComplete) return null

  const progressPct = (completedCount / totalCount) * 100

  // The group containing the global current step
  const currentGroupId = groups.find((g) =>
    g.steps.some((s) => s.kind === 'tracked' && s.status === 'current')
  )?.id ?? null

  // Determine which group is shown expanded:
  // - If user explicitly picked one (expandedGroupId), use that
  // - Otherwise default to currentGroupId
  const activeGroupId = expandedGroupId ?? currentGroupId

  const handleGroupClick = (groupId: string) => {
    setExpandedGroupId(activeGroupId === groupId ? null : groupId)
  }

  // ── Collapsed pill ─────────────────────────────────────────────────────────
  if (!panelOpen) {
    return (
      <div
        className="fixed bottom-6 right-6 z-50 flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-white px-3 py-2.5 shadow-md transition-shadow hover:shadow-lg"
        onClick={openPanel}
        role="button"
        aria-label="展开快速开始面板"
      >
        <div className="h-8 w-0.5 flex-shrink-0 rounded-full bg-primary" />
        <div className="flex flex-col gap-1">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          <div className="h-1 w-32 overflow-hidden rounded-full bg-[#EBEEF2]">
            <div
              className="h-full rounded-full bg-primary transition-all duration-300"
              style={{ width: `${progressPct}%` }}
            />
          </div>
          <span className="text-[10px] font-medium text-muted-foreground">
            {completedCount} / {totalCount} 步完成
          </span>
        </div>
        <ChevronUp className="ml-1 size-3.5 text-muted-foreground" />
      </div>
    )
  }

  // ── Expanded panel ─────────────────────────────────────────────────────────
  return (
    <div className="fixed bottom-6 right-6 z-50 w-[260px] overflow-hidden rounded-xl border border-border bg-white shadow-lg">

      {/* Header */}
      <div className="border-b border-border px-3.5 py-3">
        <div className="mb-2 flex items-center justify-between">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          <button
            onClick={dismiss}
            className="flex size-5 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            aria-label="关闭并不再显示"
          >
            <X className="size-3" />
          </button>
        </div>
        <div className="h-1 overflow-hidden rounded-full bg-[#EBEEF2]">
          <div
            className="h-full rounded-full bg-primary transition-all duration-300"
            style={{ width: `${progressPct}%` }}
          />
        </div>
        <p className="mt-1 text-[10px] font-medium text-muted-foreground">
          {completedCount} / {totalCount} 步完成
        </p>
      </div>

      {/* Groups */}
      <div className="max-h-[520px] overflow-y-auto py-1.5">
        {groups.map((group, groupIndex) => {
          const isExpanded = activeGroupId === group.id
          const isCurrentGroup = group.id === currentGroupId
          const trackedDone = group.steps.filter((s) => s.kind === 'tracked' && s.status === 'completed').length
          const trackedTotal = group.steps.filter((s) => s.kind === 'tracked').length

          return (
            <div key={group.id}>
              {/* Group header — clickable to expand/collapse */}
              <button
                className={cn(
                  'flex w-full items-center gap-2 px-3.5 py-2 text-left transition-colors hover:bg-accent/50',
                  isCurrentGroup && group.status !== 'completed' && 'bg-primary/[0.03]'
                )}
                onClick={() => handleGroupClick(group.id)}
              >
                {group.status === 'completed' ? (
                  <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/30 bg-[#10b981]/10">
                    <Check className="size-3 text-[#10b981]" strokeWidth={2.5} />
                  </div>
                ) : isCurrentGroup ? (
                  <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border-[1.5px] border-primary bg-primary/10">
                    <span className="text-[9px] font-semibold text-primary">{groupIndex + 1}</span>
                  </div>
                ) : (
                  <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-border bg-[#F6F8FA]">
                    <span className="text-[9px] font-medium text-muted-foreground">{groupIndex + 1}</span>
                  </div>
                )}

                <span className={cn(
                  'flex-1 text-[12px] font-semibold',
                  group.status === 'completed' ? 'text-muted-foreground' : isCurrentGroup ? 'text-foreground' : 'text-muted-foreground'
                )}>
                  {group.label}
                </span>

                {group.status !== 'completed' && trackedTotal > 0 && (
                  <span className="text-[10px] text-muted-foreground">
                    {trackedDone}/{trackedTotal}
                  </span>
                )}

                <ChevronRight className={cn(
                  'size-3.5 flex-shrink-0 text-muted-foreground transition-transform duration-200',
                  isExpanded && 'rotate-90'
                )} />
              </button>

              {/* Sub-steps — shown only when expanded */}
              {isExpanded && (
                <div className="ml-[26px] border-l border-border pb-2 pl-3 pr-3.5">
                  {/* Project-required banner */}
                  {!projectSlug && group.steps.some((s) => s.kind === 'tracked' && needsProject(s.id)) && (
                    <div className="mb-1.5 mt-1 flex items-center gap-1.5 rounded-md border border-border bg-[#F6F8FA] px-2.5 py-1.5">
                      <span className="text-[11px] text-muted-foreground">请先进入一个项目再操作此步骤</span>
                      <button
                        className="ml-auto shrink-0 text-[11px] font-medium text-primary hover:underline"
                        onClick={() => router.push(`/org/${orgName}/workspace`)}
                      >
                        前往 →
                      </button>
                    </div>
                  )}

                  {group.steps.map((step) => {
                    // ── Nav step ──────────────────────────────────────────
                    if (step.kind === 'nav') {
                      const route = step.route({ orgName, projectSlug })
                      // Nav steps pointing to a project path require projectSlug
                      const navNeedsProject = route.includes('/project/') && !projectSlug
                      return (
                        <div key={step.id} className="py-1">
                          <button
                            className={cn(
                              'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors',
                              navNeedsProject ? 'cursor-default opacity-40' : 'hover:bg-primary/[0.04]'
                            )}
                            onClick={() => { if (!navNeedsProject) router.push(route) }}
                          >
                            <div className="size-1.5 flex-shrink-0 rounded-full bg-border" />
                            <span className="flex-1 text-[11px] text-foreground">{step.label}</span>
                            {!navNeedsProject && <span className="text-[10px] text-muted-foreground">→</span>}
                          </button>
                        </div>
                      )
                    }

                    // ── Tracked step ─────────────────────────────────────
                    const isCurrent = step.status === 'current'
                    const isDone = step.status === 'completed'
                    const isLocked = step.status === 'locked'
                    const route = step.route({ orgName, projectSlug })

                    // Manual confirm (end_user_login)
                    if (step.type === 'manual') {
                      if (isDone) {
                        return (
                          <div key={step.id} className="py-1">
                            <div className="flex items-center gap-2 px-2 py-1.5 opacity-60">
                              <div className="flex size-3.5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/40 bg-[#10b981]/10">
                                <Check className="size-2 text-[#10b981]" strokeWidth={3} />
                              </div>
                              <span className="flex-1 text-[11px] text-muted-foreground line-through">{step.label}</span>
                            </div>
                          </div>
                        )
                      }
                      return (
                        <div key={step.id} className={cn('py-1', isLocked && 'opacity-40')}>
                          <div className={cn(
                            'rounded-md border px-2.5 py-2',
                            isCurrent ? 'border-amber-200 bg-amber-50' : 'border-border bg-[#F6F8FA]'
                          )}>
                            {isCurrent && (
                              <p className="mb-1.5 text-[10px] font-semibold text-amber-700">👆 当前步骤</p>
                            )}
                            <div className="mb-1 flex items-center gap-1.5">
                              <div className={cn('size-1.5 flex-shrink-0 rounded-full', isCurrent ? 'bg-amber-500' : 'bg-muted-foreground/40')} />
                              <span className="text-[11px] font-medium text-foreground">{step.label}</span>
                            </div>
                            <p className="mb-1 text-[10px] text-muted-foreground">终端用户登录地址：</p>
                            <code className="block break-all font-mono text-[10px] text-foreground">
                              /end-user/org/{orgName}/login
                            </code>
                            <Button
                              size="sm"
                              variant="outline"
                              className="mt-2 h-7 w-full text-[11px]"
                              disabled={isLocked}
                              onClick={() => markStep('end_user_login')}
                            >
                              已完成 ✓
                            </Button>
                          </div>
                        </div>
                      )
                    }

                    // Action step
                    const requiresProject = needsProject(step.id)
                    return (
                      <div key={step.id} className={cn('py-1', isLocked && 'opacity-40')}>
                        {isCurrent ? (
                          // Current step — amber highlight
                          <div className="rounded-md border border-amber-200 bg-amber-50 px-2.5 py-2">
                            <p className="mb-1.5 text-[10px] font-semibold text-amber-700">👆 当前步骤</p>
                            {requiresProject ? (
                              <div className="flex items-center gap-1.5">
                                <div className="size-1.5 flex-shrink-0 rounded-full bg-amber-400" />
                                <span className="text-[11px] text-amber-800">{step.label}</span>
                                <span className="ml-auto text-[10px] text-amber-600">请先进入项目 →</span>
                              </div>
                            ) : (
                              <button
                                className="flex w-full items-center gap-2 text-left"
                                onClick={() => {
                                  setPendingAction(step.id as OnboardingPendingAction)
                                  if (route) router.push(route)
                                }}
                              >
                                <div className="size-1.5 flex-shrink-0 rounded-full bg-amber-500" />
                                <span className="flex-1 text-[11px] font-semibold text-amber-900">{step.label}</span>
                                <span className="text-[10px] text-amber-600">↗</span>
                              </button>
                            )}
                          </div>
                        ) : isDone ? (
                          // Completed
                          <div className="flex items-center gap-2 px-2 py-1.5 opacity-60">
                            <div className="flex size-3.5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/40 bg-[#10b981]/10">
                              <Check className="size-2 text-[#10b981]" strokeWidth={3} />
                            </div>
                            <span className="flex-1 text-[11px] text-muted-foreground line-through">{step.label}</span>
                          </div>
                        ) : (
                          // Locked — dimmed, still navigable (soft guidance)
                          <button
                            className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-primary/[0.04]"
                            onClick={() => {
                              if (requiresProject) return
                              setPendingAction(step.id as OnboardingPendingAction)
                              if (route) router.push(route)
                            }}
                          >
                            <div className="size-1.5 flex-shrink-0 rounded-full bg-muted-foreground/30" />
                            <span className="flex-1 text-[11px] text-muted-foreground">{step.label}</span>
                            {requiresProject
                              ? <span className="text-[10px] text-muted-foreground/40">需进入项目</span>
                              : <span className="text-[10px] text-muted-foreground/50">→</span>
                            }
                          </button>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Collapse chevron */}
      <button
        onClick={closePanel}
        className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
        aria-label="折叠面板"
      >
        <ChevronDown className="size-3.5" />
      </button>
    </div>
  )
}
