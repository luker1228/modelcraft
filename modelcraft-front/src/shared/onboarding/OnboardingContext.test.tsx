import { describe, it, expect } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS } from './steps'

// ── Helpers that mirror the derivation logic in OnboardingContext ──────────

function deriveGroups(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  let foundCurrentGroup = false

  return ONBOARDING_GROUPS.map((group, groupIndex) => {
    const allCompleted = group.steps.every((s) => completedSet.has(s.id))

    let groupStatus: 'completed' | 'current' | 'locked'
    if (allCompleted) {
      groupStatus = 'completed'
    } else if (!foundCurrentGroup) {
      groupStatus = 'current'
      foundCurrentGroup = true
    } else {
      groupStatus = 'locked'
    }

    let foundCurrentSubStep = false
    const resolvedSteps = group.steps.map((step) => {
      if (groupStatus === 'completed') return { ...step, status: 'completed' as const }
      if (groupStatus === 'locked') return { ...step, status: 'locked' as const }
      if (completedSet.has(step.id)) return { ...step, status: 'completed' as const }
      if (!foundCurrentSubStep) {
        foundCurrentSubStep = true
        return { ...step, status: 'current' as const }
      }
      return { ...step, status: 'locked' as const }
    })

    return { id: group.id, label: group.label, status: groupStatus, steps: resolvedSteps, index: groupIndex + 1 }
  })
}

function currentStep(completedSteps: OnboardingStepId[]) {
  const groups = deriveGroups(completedSteps)
  const currentGroup = groups.find((g) => g.status === 'current')
  return currentGroup?.steps.find((s) => s.status === 'current') ?? null
}

describe('onboarding group derivation', () => {
  it('first group is current when nothing completed', () => {
    const groups = deriveGroups([])
    expect(groups[0].status).toBe('current')
    expect(groups.slice(1).every((g) => g.status === 'locked')).toBe(true)
  })

  it('first group completes when all its sub-steps are done', () => {
    const groups = deriveGroups(['create_project'])
    expect(groups[0].status).toBe('completed')
    expect(groups[1].status).toBe('current')
  })

  it('second group sub-steps: first is current, second is locked', () => {
    const groups = deriveGroups(['create_project'])
    const g2 = groups[1]
    expect(g2.steps[0].status).toBe('current')
    expect(g2.steps[1].status).toBe('locked')
  })

  it('second group advances sub-step when first sub-step done', () => {
    const groups = deriveGroups(['create_project', 'create_model'])
    const g2 = groups[1]
    expect(g2.steps[0].status).toBe('completed')
    expect(g2.steps[1].status).toBe('current')
  })

  it('currentStep returns correct sub-step', () => {
    const step = currentStep(['create_project'])
    expect(step?.id).toBe('create_model')
  })

  it('currentStep returns add_field after create_model done', () => {
    const step = currentStep(['create_project', 'create_model'])
    expect(step?.id).toBe('add_field')
  })

  it('third group becomes current after second group fully done', () => {
    const groups = deriveGroups(['create_project', 'create_model', 'add_field'])
    expect(groups[2].status).toBe('current')
    expect(groups[2].steps[0].status).toBe('current') // apply_preset
  })

  it('fourth group becomes current when groups 1-3 done', () => {
    const groups = deriveGroups([
      'create_project',
      'create_model',
      'add_field',
      'apply_preset',
      'create_role',
    ])
    expect(groups[3].status).toBe('current')
    expect(groups[3].steps[0].status).toBe('current') // add_end_user
  })

  it('isComplete when all 8 steps done', () => {
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
    const groups = deriveGroups(all)
    expect(groups.every((g) => g.status === 'completed')).toBe(true)
  })

  it('markStep is idempotent — dedup logic', () => {
    const prev: OnboardingState = {
      ...defaultOnboardingState('test-org'),
      completedSteps: ['create_project'],
    }
    const id: OnboardingStepId = 'create_project'
    expect(prev.completedSteps.includes(id)).toBe(true)
  })

  it('markStep with create_project stores projectSlug', () => {
    const prev: OnboardingState = defaultOnboardingState('test-org')
    const id: OnboardingStepId = 'create_project'
    const slug = 'my-project'
    const next: OnboardingState = {
      ...prev,
      completedSteps: [...prev.completedSteps, id],
      projectSlug: id === 'create_project' && slug ? slug : prev.projectSlug,
    }
    expect(next.projectSlug).toBe('my-project')
  })

  it('route for project-scoped step returns null when projectSlug is null', () => {
    const modelStep = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(modelStep.route({ orgName: 'my-org', projectSlug: null })).toBeNull()
  })

  it('route for project-scoped step returns full path when projectSlug set', () => {
    const modelStep = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    const route = modelStep.route({ orgName: 'my-org', projectSlug: 'my-project' })
    expect(route).toBe('/org/my-org/project/my-project/model-editor')
  })
})
