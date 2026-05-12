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
  ALL_TRACKED_STEPS,
  type OnboardingTrackedStep,
  type OnboardingNavStep,
  type OnboardingGroup,
} from './steps'

// ── Derived types ──────────────────────────────────────────────────────────

export type OnboardingStepStatus = 'completed' | 'current' | 'locked'

export type OnboardingTrackedStepWithStatus = OnboardingTrackedStep & {
  status: OnboardingStepStatus
}

export type OnboardingSubStepWithStatus =
  | (OnboardingNavStep & { status: 'nav' })
  | OnboardingTrackedStepWithStatus

export interface OnboardingGroupWithStatus extends Omit<OnboardingGroup, 'steps'> {
  /** completed = all required tracked sub-steps done; todo = in progress or not started */
  status: 'completed' | 'todo'
  steps: OnboardingSubStepWithStatus[]
}

// ── Context value ──────────────────────────────────────────────────────────

export type OnboardingPendingAction =
  | 'create_project'
  | 'select_database'
  | 'create_model'
  | 'insert_column'
  | 'insert_data'
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
  expandedGroupId: string | null
  markStep: (id: OnboardingStepId, projectSlug?: string) => void
  openPanel: () => void
  closePanel: () => void
  dismiss: () => void
  reset: () => void
  setPendingAction: (action: OnboardingPendingAction) => void
  setExpandedGroupId: (id: string | null) => void
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
  const [expandedGroupId, setExpandedGroupId] = useState<string | null>(null)

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

  // Find the first non-completed, non-optional tracked step ID (global current)
  const globalCurrentId = ALL_TRACKED_STEPS.find(
    (s) => !s.optional && !completedSet.has(s.id)
  )?.id ?? null

  let markedCurrent = false
  const groups: OnboardingGroupWithStatus[] = ONBOARDING_GROUPS.map((group) => {
    const isOptionalGroup = group.optional === true

    const steps: OnboardingSubStepWithStatus[] = group.steps.map((step) => {
      if (step.kind === 'nav') return { ...step, status: 'nav' as const }
      if (completedSet.has(step.id)) return { ...step, status: 'completed' } satisfies OnboardingTrackedStepWithStatus
      // For optional groups: all incomplete steps are 'locked' (greyed, still clickable)
      if (isOptionalGroup) return { ...step, status: 'locked' } satisfies OnboardingTrackedStepWithStatus
      // For required groups: first non-completed = current
      if (step.id === globalCurrentId && !markedCurrent) {
        markedCurrent = true
        return { ...step, status: 'current' } satisfies OnboardingTrackedStepWithStatus
      }
      return { ...step, status: 'locked' } satisfies OnboardingTrackedStepWithStatus
    })

    const trackedSteps = steps.filter((s): s is OnboardingTrackedStepWithStatus => s.kind === 'tracked')
    // Optional group: completed when all its steps are done (but doesn't affect overall completion)
    const requiredTracked = trackedSteps.filter((s) => !s.optional)
    const allDone = isOptionalGroup
      ? trackedSteps.every((s) => s.status === 'completed')
      : requiredTracked.length > 0 && requiredTracked.every((s) => s.status === 'completed')

    return {
      id: group.id,
      label: group.label,
      status: allDone ? 'completed' : 'todo',
      steps,
    }
  })

  // Only required (non-optional) steps count toward completion
  const completedCount = ONBOARDING_STEPS.filter((s) => completedSet.has(s.id)).length
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
        expandedGroupId,
        markStep,
        openPanel,
        closePanel,
        dismiss,
        reset,
        setPendingAction,
        setExpandedGroupId,
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
