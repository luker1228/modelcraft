'use client'

import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import {
  type OnboardingState,
  type OnboardingStepId,
  defaultOnboardingState,
  readOnboardingState,
  writeOnboardingState,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS, type OnboardingSubStep, type OnboardingGroup } from './steps'

// ── Derived types ──────────────────────────────────────────────────────────

export interface OnboardingSubStepWithStatus extends OnboardingSubStep {
  status: 'completed' | 'current' | 'locked'
}

export type OnboardingGroupStatus = 'completed' | 'current' | 'locked'

export interface OnboardingGroupWithStatus extends Omit<OnboardingGroup, 'steps'> {
  status: OnboardingGroupStatus
  steps: OnboardingSubStepWithStatus[]
  /** index of the current sub-step within this group (0-based), or -1 if not current group */
  currentSubStepIndex: number
}

// ── Context value ──────────────────────────────────────────────────────────

interface OnboardingContextValue {
  groups: OnboardingGroupWithStatus[]
  currentGroup: OnboardingGroupWithStatus | null
  currentStep: OnboardingSubStepWithStatus | null
  projectSlug: string | null
  completedCount: number
  totalCount: number
  isComplete: boolean
  panelOpen: boolean
  dismissed: boolean
  markStep: (id: OnboardingStepId, projectSlug?: string) => void
  openPanel: () => void
  closePanel: () => void
  dismiss: () => void
  reset: () => void
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

  // Hydrate from localStorage on mount (client only)
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

  // ── Derive grouped status ──────────────────────────────────────────────

  const completedSet = new Set(state.completedSteps)
  let foundCurrentGroup = false

  const groups: OnboardingGroupWithStatus[] = ONBOARDING_GROUPS.map((group) => {
    const stepsWithStatus: OnboardingSubStepWithStatus[] = group.steps.map((step) => {
      const completed = completedSet.has(step.id)
      return { ...step, status: completed ? 'completed' : 'locked' } satisfies OnboardingSubStepWithStatus
    })

    const allCompleted = stepsWithStatus.every((s) => completedSet.has(s.id))

    let groupStatus: OnboardingGroupStatus
    if (allCompleted) {
      groupStatus = 'completed'
    } else if (!foundCurrentGroup) {
      groupStatus = 'current'
      foundCurrentGroup = true
    } else {
      groupStatus = 'locked'
    }

    // Assign sub-step statuses within current group
    let foundCurrentSubStep = false
    const resolvedSteps: OnboardingSubStepWithStatus[] = stepsWithStatus.map((step) => {
      if (groupStatus === 'completed') return { ...step, status: 'completed' }
      if (groupStatus === 'locked') return { ...step, status: 'locked' }
      // current group: assign current/locked to sub-steps
      if (completedSet.has(step.id)) return { ...step, status: 'completed' }
      if (!foundCurrentSubStep) {
        foundCurrentSubStep = true
        return { ...step, status: 'current' }
      }
      return { ...step, status: 'locked' }
    })

    const currentSubStepIndex = resolvedSteps.findIndex((s) => s.status === 'current')

    return {
      id: group.id,
      label: group.label,
      status: groupStatus,
      steps: resolvedSteps,
      currentSubStepIndex,
    }
  })

  const currentGroup = groups.find((g) => g.status === 'current') ?? null
  const currentStep = currentGroup?.steps.find((s) => s.status === 'current') ?? null
  const completedCount = state.completedSteps.length
  const totalCount = ONBOARDING_STEPS.length
  const isComplete = completedCount === totalCount

  return (
    <OnboardingContext.Provider
      value={{
        groups,
        currentGroup,
        currentStep,
        projectSlug: state.projectSlug,
        completedCount,
        totalCount,
        isComplete,
        panelOpen: state.panelOpen,
        dismissed: state.dismissed,
        markStep,
        openPanel,
        closePanel,
        dismiss,
        reset,
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
