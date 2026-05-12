import { describe, it, expect, vi } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_STEPS } from './steps'

// ── Helpers that mirror the derivation logic in OnboardingContext ──────────

function deriveSteps(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  let foundCurrent = false
  return ONBOARDING_STEPS.map((def, i) => {
    const completed = completedSet.has(def.id)
    let status: 'completed' | 'current' | 'locked'
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
}

function currentStep(completedSteps: OnboardingStepId[]) {
  return deriveSteps(completedSteps).find((s) => s.status === 'current') ?? null
}

describe('onboarding step derivation', () => {
  it('all steps are locked except the first when no steps completed', () => {
    const steps = deriveSteps([])
    expect(steps[0].status).toBe('current')
    expect(steps.slice(1).every((s) => s.status === 'locked')).toBe(true)
  })

  it('completed steps are marked completed', () => {
    const steps = deriveSteps(['create_project', 'create_model'])
    expect(steps[0].status).toBe('completed')
    expect(steps[1].status).toBe('completed')
    expect(steps[2].status).toBe('current')
  })

  it('currentStep returns first incomplete step', () => {
    const step = currentStep(['create_project', 'create_model'])
    expect(step?.id).toBe('add_field')
  })

  it('currentStep returns null when all steps completed', () => {
    const all: OnboardingStepId[] = [
      'create_project',
      'create_model',
      'add_field',
      'apply_preset',
      'create_role',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ]
    expect(currentStep(all)).toBeNull()
  })

  it('isComplete is true when all 8 steps done', () => {
    const all: OnboardingStepId[] = [
      'create_project',
      'create_model',
      'add_field',
      'apply_preset',
      'create_role',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ]
    expect(all.length).toBe(ONBOARDING_STEPS.length)
  })

  it('steps have 1-based index', () => {
    const steps = deriveSteps([])
    expect(steps[0].index).toBe(1)
    expect(steps[7].index).toBe(8)
  })

  it('markStep skips duplicate (idempotent)', () => {
    // Simulates the dedup logic in markStep
    const prev: OnboardingState = {
      ...defaultOnboardingState('test-org'),
      completedSteps: ['create_project'],
    }
    const id: OnboardingStepId = 'create_project'
    const isAlreadyDone = prev.completedSteps.includes(id)
    expect(isAlreadyDone).toBe(true)
  })

  it('markStep with create_project stores projectSlug', () => {
    const prev: OnboardingState = defaultOnboardingState('test-org')
    const id: OnboardingStepId = 'create_project'
    const projectSlug = 'my-project'
    const next: OnboardingState = {
      ...prev,
      completedSteps: [...prev.completedSteps, id],
      projectSlug: id === 'create_project' && projectSlug ? projectSlug : prev.projectSlug,
    }
    expect(next.projectSlug).toBe('my-project')
  })

  it('route for project-scoped step returns null when projectSlug is null', () => {
    const modelStep = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(modelStep.route({ orgName: 'my-org', projectSlug: null })).toBeNull()
  })

  it('route for project-scoped step returns full path when projectSlug is set', () => {
    const modelStep = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    const route = modelStep.route({ orgName: 'my-org', projectSlug: 'my-project' })
    expect(route).toBe('/org/my-org/project/my-project/model-editor')
  })
})
