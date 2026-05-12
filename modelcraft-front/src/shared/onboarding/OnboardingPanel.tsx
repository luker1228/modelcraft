'use client'

import { useRouter, useParams } from 'next/navigation'
import { ChevronUp, ChevronDown, X, Check } from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { useOnboarding } from './OnboardingContext'

export function OnboardingPanel({ orgName }: { orgName: string }) {
  const {
    groups,
    projectSlug: storedProjectSlug,
    completedCount,
    totalCount,
    isComplete,
    panelOpen,
    openPanel,
    closePanel,
    dismiss,
    markStep,
    setPendingAction,
  } = useOnboarding()

  const router = useRouter()
  const params = useParams()

  const urlProjectSlug = (params.projectSlug as string | undefined) ?? null
  const projectSlug = urlProjectSlug ?? storedProjectSlug

  if (isComplete) return null

  const progressPct = (completedCount / totalCount) * 100

  // ── Collapsed ──────────────────────────────────────────────────────────────
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

  // ── Expanded ──────────────────────────────────────────────────────────────
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
      <div className="max-h-[520px] overflow-y-auto py-2">
        {groups.map((group, groupIndex) => {
          const trackedDone = group.steps
            .filter((s) => s.kind === 'tracked')
            .filter((s) => s.kind === 'tracked' && s.status === 'completed').length
          const trackedTotal = group.steps.filter((s) => s.kind === 'tracked').length

          return (
            <div key={group.id} className="mb-1">
              {/* Group header */}
              <div className="flex items-center gap-2 px-3.5 py-1.5">
                {group.status === 'completed' ? (
                  <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/30 bg-[#10b981]/10">
                    <Check className="size-3 text-[#10b981]" strokeWidth={2.5} />
                  </div>
                ) : (
                  <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-border bg-[#F6F8FA]">
                    <span className="text-[9px] font-medium text-muted-foreground">{groupIndex + 1}</span>
                  </div>
                )}
                <span className={cn(
                  'text-[12px] font-semibold',
                  group.status === 'completed' ? 'text-muted-foreground' : 'text-foreground'
                )}>
                  {group.label}
                </span>
                {group.status !== 'completed' && trackedTotal > 0 && (
                  <span className="ml-auto text-[10px] text-muted-foreground">
                    {trackedDone}/{trackedTotal}
                  </span>
                )}
              </div>

              {/* Sub-steps */}
              <div className="ml-[26px] border-l border-border pb-1 pl-3 pr-3.5">
                {group.steps.map((step) => {

                  // ── Nav step — pure navigation, always clickable ──────────
                  if (step.kind === 'nav') {
                    const route = step.route({ orgName, projectSlug })
                    return (
                      <div key={step.id} className="py-1">
                        <button
                          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-primary/[0.04]"
                          onClick={() => router.push(route)}
                        >
                          <div className="size-1.5 flex-shrink-0 rounded-full bg-border" />
                          <span className="flex-1 text-[11px] text-foreground">{step.label}</span>
                          <span className="text-[10px] text-muted-foreground">→</span>
                        </button>
                      </div>
                    )
                  }

                  // ── Tracked step ─────────────────────────────────────────
                  const isDone = step.status === 'completed'
                  const route = step.route({ orgName, projectSlug })

                  // Manual confirm (end_user_login)
                  if (step.type === 'manual' && !isDone) {
                    return (
                      <div key={step.id} className="py-1">
                        <div className="rounded-md border border-border bg-[#F6F8FA] px-2.5 py-2">
                          <div className="mb-1 flex items-center gap-1.5">
                            <div className="size-1.5 flex-shrink-0 rounded-full bg-primary/40" />
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
                            onClick={() => markStep('end_user_login')}
                          >
                            已完成 ✓
                          </Button>
                        </div>
                      </div>
                    )
                  }

                  // Action step — navigate + trigger page dialog
                  if (step.type === 'action' && !isDone) {
                    return (
                      <div key={step.id} className="py-1">
                        <button
                          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-primary/[0.06]"
                          onClick={() => {
                            setPendingAction(step.id as 'create_project' | 'create_model' | 'add_field')
                            if (route) router.push(route)
                          }}
                        >
                          <div className="size-1.5 flex-shrink-0 rounded-full bg-primary/40" />
                          <span className="flex-1 text-[11px] font-medium text-primary">{step.label}</span>
                          <span className="text-[10px] text-primary">↗</span>
                        </button>
                      </div>
                    )
                  }

                  // Completed tracked step (any type)
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
                })}
              </div>
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
