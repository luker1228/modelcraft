'use client'

import { useRouter, useParams } from 'next/navigation'
import { ChevronUp, ChevronDown, RotateCcw, Check, ChevronRight } from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { useOnboarding, type OnboardingPendingAction } from './OnboardingContext'

export function OnboardingPanel({ orgName }: { orgName: string }) {
  const {
    groups,
    projectSlug: storedProjectSlug,
    hasProjects,
    completedCount,
    totalCount,
    isComplete,
    panelOpen,
    expandedGroupId,
    openPanel,
    closePanel,
    reset,
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
    const projectScopedIds = [
      'goto_model_editor', 'create_model',
      'goto_end_user_access', 'assign_role',
    ]
    return projectScopedIds.includes(stepId) && !projectSlug
  }

  if (isComplete) {
    if (!panelOpen) {
      return (
        <div
          className="fixed bottom-6 right-6 z-50 flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-white px-3 py-2.5 shadow-md transition-shadow hover:shadow-lg"
          onClick={openPanel}
          role="button"
          aria-label="展开教程完成面板"
        >
          <div className="h-8 w-0.5 flex-shrink-0 rounded-full bg-[#10b981]" />
          <div className="flex flex-col gap-1">
            <span className="text-[12px] font-semibold text-foreground">快速开始</span>
            <div className="h-1 w-32 overflow-hidden rounded-full bg-[#EBEEF2]">
              <div className="h-full w-full rounded-full bg-[#10b981]" />
            </div>
            <span className="text-[10px] font-medium text-[#10b981]">全部完成 🎉</span>
          </div>
          <ChevronUp className="ml-1 size-3.5 text-muted-foreground" />
        </div>
      )
    }
    return (
      <div className="fixed bottom-6 right-6 z-50 w-[260px] overflow-hidden rounded-xl border border-border bg-white shadow-lg">
        <div className="border-b border-border px-3.5 py-3">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          </div>
          <div className="h-1 overflow-hidden rounded-full bg-[#EBEEF2]">
            <div className="h-full w-full rounded-full bg-[#10b981] transition-all duration-300" />
          </div>
          <p className="mt-1 text-[10px] font-medium text-[#10b981]">{totalCount} / {totalCount} 步完成</p>
        </div>
        <div className="px-4 py-5 text-center">
          <div className="mb-2 text-2xl">🎉</div>
          <p className="text-[13px] font-semibold text-foreground">教程完成！</p>
          <p className="mt-1 text-[11px] text-muted-foreground">你已完成所有入门步骤</p>
        </div>
        <div className="border-t border-border">
          <button
            onClick={() => { reset(); openPanel() }}
            className="flex w-full items-center justify-center gap-1 py-1.5 text-[10px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
            <RotateCcw className="size-3" />
            重新开始
          </button>
          <button
            onClick={closePanel}
            className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
          >
            <ChevronDown className="size-3.5" />
          </button>
        </div>
      </div>
    )
  }

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
          // Groups that require a project are locked when no projects exist yet
          const groupNeedsProject = group.steps.some((s) => needsProject(s.id))
          const groupBlocked = groupNeedsProject && !hasProjects

          return (
            <div key={group.id}>
              {/* Group header — clickable to expand/collapse */}
              <button
                className={cn(
                  'flex w-full items-center gap-2 px-3.5 py-2 text-left transition-colors hover:bg-accent/50',
                  isCurrentGroup && group.status !== 'completed' && 'bg-primary/[0.03]',
                  groupBlocked && 'cursor-default opacity-50'
                )}
                onClick={() => { if (!groupBlocked) handleGroupClick(group.id) }}
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
                  {!projectSlug && group.steps.some((s) => needsProject(s.id)) && (
                    <div className="mb-1.5 mt-1 rounded-md border border-amber-200 bg-amber-50 px-2.5 py-2">
                      {!hasProjects ? (
                        <p className="text-[11px] font-medium text-amber-800">请先创建一个项目</p>
                      ) : (
                        <>
                          <p className="mb-1.5 text-[11px] font-medium text-amber-800">请先选择一个项目</p>
                          <button
                            className="flex w-full items-center justify-center gap-1.5 rounded-md bg-amber-500 px-2 py-1.5 text-[11px] font-semibold text-white transition-colors hover:bg-amber-600"
                            onClick={() => {
                              setPendingAction('highlight_first_project')
                              router.push(`/org/${orgName}/workspace`)
                            }}
                          >
                            选择项目 →
                          </button>
                        </>
                      )}
                    </div>
                  )}

                  {group.steps.map((step) => {
                    // ── Nav step ──────────────────────────────────────────
                    if (step.kind === 'nav') {
                      const route = step.route({ orgName, projectSlug })
                      // needsProject but has projects → redirect to workspace to select one
                      const needsProjectSelect = needsProject(step.id) && hasProjects
                      // needsProject and no projects at all → fully blocked
                      const navBlocked = needsProject(step.id) && !hasProjects
                      return (
                        <div key={step.id} className="py-1">
                          <button
                            className={cn(
                              'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors',
                              navBlocked ? 'cursor-default opacity-40' : 'hover:bg-primary/[0.04]'
                            )}
                            onClick={() => {
                              if (navBlocked) return
                              // goto_model_editor / goto_end_user_access: always guide the user
                              // to click a project card themselves — never jump directly into the project
                              const alwaysHighlight =
                                (step.id === 'goto_model_editor' || step.id === 'goto_end_user_access') &&
                                hasProjects
                              if (alwaysHighlight || needsProjectSelect) {
                                setPendingAction('highlight_first_project')
                                router.push(`/org/${orgName}/workspace`)
                                return
                              }
                              router.push(route)
                            }}
                          >
                            <div className="size-1.5 flex-shrink-0 rounded-full bg-border" />
                            <span className="flex-1 text-[11px] text-foreground">{step.label}</span>
                            {!navBlocked && <span className="text-[10px] text-muted-foreground">→</span>}
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
                        {isCurrent && !requiresProject ? (
                          // Current step with project available — amber highlight
                          <div className="rounded-md border border-amber-200 bg-amber-50 px-2.5 py-2">
                            <p className="mb-1.5 text-[10px] font-semibold text-amber-700">👆 当前步骤</p>
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

      {/* Footer */}
      <div className="border-t border-border">
        <button
          onClick={() => { reset(); openPanel() }}
          className="flex w-full items-center justify-center gap-1 py-1.5 text-[10px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          aria-label="重新开始教程"
        >
          <RotateCcw className="size-3" />
          重新开始
        </button>
        <button
          onClick={closePanel}
          className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
          aria-label="折叠面板"
        >
          <ChevronDown className="size-3.5" />
        </button>
      </div>
    </div>
  )
}
