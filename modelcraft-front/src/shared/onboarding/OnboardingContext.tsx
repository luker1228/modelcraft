'use client'

import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import {
  type OnboardingState,
  type OnboardingStepId,
  defaultOnboardingState,
  readOnboardingState,
  writeOnboardingState,
} from './storage'
import { ONBOARDING_STEPS, type OnboardingStepDef } from './steps'

export interface OnboardingStep extends OnboardingStepDef {
  status: 'completed' | 'current' | 'locked'
  index: number // 1-based
}

interface OnboardingContextValue {
  steps: OnboardingStep[]
  currentStep: OnboardingStep | null
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

  const persist = useCallback(
    (next: OnboardingState) => {
      setState(next)
      writeOnboardingState(next)
    },
    []
  )

  const markStep = useCallback(
    (id: OnboardingStepId, projectSlug?: string) => {
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
    },
    []
  )

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
    persist(next)
  }, [orgName, persist])

  // Derive step statuses
  const completedSet = new Set(state.completedSteps)
  let foundCurrent = false
  const steps: OnboardingStep[] = ONBOARDING_STEPS.map((def, i) => {
    const completed = completedSet.has(def.id)
    let status: OnboardingStep['status']
    if (completed) {
      status = 'completed'
    } else if (!foundCurrent) {
      status = 'current'
      foundCurrent = true
    } else {
      status = 'locked'
    }
    return { ...def, status, index: i + 1 }
  })

  const currentStep = steps.find((s) => s.status === 'current') ?? null
  const completedCount = state.completedSteps.length
  const totalCount = ONBOARDING_STEPS.length
  const isComplete = completedCount === totalCount

  return (
    <OnboardingContext.Provider
      value={{
        steps,
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

export function useOnboarding(): OnboardingContextValue {
  const ctx = useContext(OnboardingContext)
  if (!ctx) {
    throw new Error('useOnboarding must be used within OnboardingProvider')
  }
  return ctx
}
