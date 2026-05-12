'use client'

import { useRouter } from 'next/navigation'
import { ChevronUp, ChevronDown, X, Check, ArrowRight } from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { useOnboarding } from './OnboardingContext'

export function OnboardingPanel({ orgName }: { orgName: string }) {
  const {
    groups,
    currentStep,
    projectSlug,
    completedCount,
    totalCount,
    isComplete,
    panelOpen,
    openPanel,
    closePanel,
    dismiss,
    markStep,
  } = useOnboarding()

  const router = useRouter()

  if (isComplete) return null

  const progressPct = (completedCount / totalCount) * 100

  const handleCta = () => {
    if (!currentStep) return
    const route = currentStep.route({ orgName, projectSlug })
    if (route) {
      router.push(route)
      closePanel()
    }
  }

  // ── Collapsed state ──────────────────────────────────────────────────────
  if (!panelOpen) {
    const currentGroup = groups.find((g) => g.status === 'current')
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
            {currentGroup ? currentGroup.label : '全部完成'} · {completedCount}/{totalCount} 步
          </span>
        </div>
        <ChevronUp className="ml-1 size-3.5 text-muted-foreground" />
      </div>
    )
  }

  // ── Expanded state ───────────────────────────────────────────────────────
  return (
    <div className="fixed bottom-6 right-6 z-50 w-[248px] overflow-hidden rounded-xl border border-border bg-white shadow-lg">

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

      {/* Groups list */}
      <div className="py-1.5">
        {groups.map((group, groupIndex) => (
          <div key={group.id}>
            {/* Group row */}
            <div
              className={cn(
                'flex items-center gap-2 px-3.5 py-2',
                group.status === 'current' &&
                  'border-l-[3px] border-primary bg-primary/[0.06] pl-[11px]'
              )}
            >
              {/* Group indicator */}
              {group.status === 'completed' ? (
                <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/30 bg-[#10b981]/10">
                  <Check className="size-3 text-[#10b981]" strokeWidth={2.5} />
                </div>
              ) : group.status === 'current' ? (
                <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border-[1.5px] border-primary bg-primary/10">
                  <span className="text-[9px] font-semibold text-primary">{groupIndex + 1}</span>
                </div>
              ) : (
                <div className="flex size-5 flex-shrink-0 items-center justify-center rounded-full border border-border">
                  <span className="text-[9px] text-muted-foreground">{groupIndex + 1}</span>
                </div>
              )}

              {/* Group label */}
              <span
                className={cn(
                  'flex-1 text-[12px]',
                  group.status === 'completed' && 'text-muted-foreground line-through',
                  group.status === 'current' && 'font-semibold text-foreground',
                  group.status === 'locked' && 'text-muted-foreground'
                )}
              >
                {group.label}
              </span>

              {/* Sub-step progress for current group */}
              {group.status === 'current' && group.steps.length > 1 && (
                <span className="text-[10px] text-muted-foreground">
                  {group.steps.filter((s) => s.status === 'completed').length}/{group.steps.length}
                </span>
              )}
            </div>

            {/* Sub-steps — only shown for current group */}
            {group.status === 'current' && group.steps.length > 1 && (
              <div className="mb-1 ml-[26px] border-l border-border pl-3">
                {group.steps.map((step) => (
                  <div
                    key={step.id}
                    className={cn(
                      'flex items-center gap-2 py-1.5 pr-3.5',
                      step.status === 'current' && 'text-primary'
                    )}
                  >
                    {step.status === 'completed' ? (
                      <div className="flex size-3.5 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/40 bg-[#10b981]/10">
                        <Check className="size-2 text-[#10b981]" strokeWidth={3} />
                      </div>
                    ) : step.status === 'current' ? (
                      <div className="size-1.5 flex-shrink-0 rounded-full bg-primary" />
                    ) : (
                      <div className="size-1.5 flex-shrink-0 rounded-full bg-border" />
                    )}
                    <span
                      className={cn(
                        'flex-1 text-[11px]',
                        step.status === 'completed' && 'text-muted-foreground line-through',
                        step.status === 'current' && 'font-medium text-primary',
                        step.status === 'locked' && 'text-muted-foreground'
                      )}
                    >
                      {step.label}
                    </span>
                    {step.status === 'current' && (
                      <ArrowRight className="size-3 flex-shrink-0 text-primary" />
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}

        {/* Step 8 manual confirm (end_user_login is current) */}
        {currentStep?.id === 'end_user_login' && (
          <div className="mx-3.5 mt-1 rounded-md border border-border bg-[#F6F8FA] px-3 py-2">
            <p className="mb-1.5 text-[10px] text-muted-foreground">终端用户登录地址：</p>
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
        )}
      </div>

      {/* CTA footer */}
      {currentStep && currentStep.id !== 'end_user_login' && (
        <div className="border-t border-border px-3.5 py-2.5">
          <Button size="sm" className="h-8 w-full text-[11px]" onClick={handleCta}>
            前往：{currentStep.label} →
          </Button>
        </div>
      )}

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
