'use client'

import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import {
  type OnboardingState,
  type OnboardingStepId,
  defaultOnboardingState,
  readOnboardingState,
  writeOnboardingState,
} from './storage'
import {
  ONBOARDING_GROUPS,
  ONBOARDING_STEPS,
  type OnboardingTrackedStep,
  type OnboardingNavStep,
  type OnboardingGroup,
} from './steps'

// ── Derived types ──────────────────────────────────────────────────────────

export type OnboardingStepStatus = 'completed' | 'todo'

export type OnboardingTrackedStepWithStatus = OnboardingTrackedStep & {
  status: OnboardingStepStatus
}

export type OnboardingSubStepWithStatus =
  | (OnboardingNavStep & { status: 'nav' })
  | OnboardingTrackedStepWithStatus

export interface OnboardingGroupWithStatus extends Omit<OnboardingGroup, 'steps'> {
  /** completed = all tracked sub-steps done */
  status: OnboardingStepStatus
  steps: OnboardingSubStepWithStatus[]
}

// ── Context value ──────────────────────────────────────────────────────────

export type OnboardingPendingAction =
  | 'create_project'
  | 'create_model'
  | 'create_permission'
  | 'create_bundle'
  | 'create_role'
  | 'add_end_user'
  | 'assign_role'
  | null

interface OnboardingContextValue {
  groups: OnboardingGroupWithStatus[]
  projectSlug: string | null
  completedCount: number
  totalCount: number
  isComplete: boolean
  panelOpen: boolean
  dismissed: boolean
  pendingAction: OnboardingPendingAction
  markStep: (id: OnboardingStepId, projectSlug?: string) => void
  openPanel: () => void
  closePanel: () => void
  dismiss: () => void
  reset: () => void
  setPendingAction: (action: OnboardingPendingAction) => void
}

const OnboardingContext = createContext<OnboardingContextValue | null>(null)

// ── Provider ───────────────────────────────────────────────────────────────

export function OnboardingProvider({
  orgName,
  children,
}: {
  orgName: string
  children: React.ReactNode
}) {
  const [state, setState] = useState<OnboardingState>(() =>
    defaultOnboardingState(orgName)
  )
  const [pendingAction, setPendingAction] = useState<OnboardingPendingAction>(null)

  useEffect(() => {
    setState(readOnboardingState(orgName))
  }, [orgName])

  const markStep = useCallback((id: OnboardingStepId, projectSlug?: string) => {
    setState((prev) => {
      if (prev.completedSteps.includes(id)) return prev
      const next: OnboardingState = {
        ...prev,
        completedSteps: [...prev.completedSteps, id],
        projectSlug:
          id === 'create_project' && projectSlug ? projectSlug : prev.projectSlug,
      }
      writeOnboardingState(next)
      return next
    })
  }, [])

  const openPanel = useCallback(() => {
    setState((prev) => {
      const next = { ...prev, panelOpen: true }
      writeOnboardingState(next)
      return next
    })
  }, [])

  const closePanel = useCallback(() => {
    setState((prev) => {
      const next = { ...prev, panelOpen: false }
      writeOnboardingState(next)
      return next
    })
  }, [])

  const dismiss = useCallback(() => {
    setState((prev) => {
      const next = { ...prev, dismissed: true, panelOpen: false }
      writeOnboardingState(next)
      return next
    })
  }, [])

  const reset = useCallback(() => {
    const next = defaultOnboardingState(orgName)
    setState(next)
    writeOnboardingState(next)
  }, [orgName])

  // ── Derive group/step status ───────────────────────────────────────────────

  const completedSet = new Set(state.completedSteps)

  const groups: OnboardingGroupWithStatus[] = ONBOARDING_GROUPS.map((group) => {
    const steps: OnboardingSubStepWithStatus[] = group.steps.map((step) => {
      if (step.kind === 'nav') {
        return { ...step, status: 'nav' as const }
      }
      return {
        ...step,
        status: completedSet.has(step.id) ? 'completed' : 'todo',
      } satisfies OnboardingTrackedStepWithStatus
    })

    // Group is completed only when all tracked sub-steps are done
    const trackedSteps = steps.filter((s): s is OnboardingTrackedStepWithStatus => s.kind === 'tracked')
    const allDone = trackedSteps.length > 0 && trackedSteps.every((s) => s.status === 'completed')

    return {
      id: group.id,
      label: group.label,
      status: allDone ? 'completed' : 'todo',
      steps,
    }
  })

  const completedCount = state.completedSteps.length
  const totalCount = ONBOARDING_STEPS.length
  const isComplete = completedCount === totalCount

  return (
    <OnboardingContext.Provider
      value={{
        groups,
        projectSlug: state.projectSlug,
        completedCount,
        totalCount,
        isComplete,
        panelOpen: state.panelOpen,
        dismissed: state.dismissed,
        pendingAction,
        markStep,
        openPanel,
        closePanel,
        dismiss,
        reset,
        setPendingAction,
      }}
    >
      {children}
    </OnboardingContext.Provider>
  )
}

// ── Hook ───────────────────────────────────────────────────────────────────

export function useOnboarding(): OnboardingContextValue {
  const ctx = useContext(OnboardingContext)
  if (!ctx) {
    throw new Error('useOnboarding must be used within OnboardingProvider')
  }
  return ctx
}
