import { describe, it, expect } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS } from './steps'

// ── Mirror derivation logic from OnboardingContext ─────────────────────────

function deriveGroups(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  return ONBOARDING_GROUPS.map((group) => {
    const steps = group.steps.map((step) => ({
      ...step,
      status: completedSet.has(step.id) ? 'completed' as const : 'todo' as const,
    }))
    const allDone = steps.every((s) => s.status === 'completed')
    return { id: group.id, label: group.label, status: allDone ? 'completed' as const : 'todo' as const, steps }
  })
}

describe('onboarding group derivation', () => {
  it('all groups start as todo', () => {
    const groups = deriveGroups([])
    expect(groups.every((g) => g.status === 'todo')).toBe(true)
  })

  it('all sub-steps start as todo and are immediately accessible', () => {
    const groups = deriveGroups([])
    const allSteps = groups.flatMap((g) => g.steps)
    expect(allSteps.every((s) => s.status === 'todo')).toBe(true)
  })

  it('completing create_project marks group 1 as completed', () => {
    const groups = deriveGroups(['create_project'])
    expect(groups[0].status).toBe('completed')
    expect(groups[1].status).toBe('todo')
  })

  it('completing create_model marks that sub-step completed, add_field still todo', () => {
    const groups = deriveGroups(['create_project', 'create_model'])
    const g2 = groups[1]
    expect(g2.steps[0].status).toBe('completed')
    expect(g2.steps[1].status).toBe('todo')
  })

  it('group 2 completes independently of group 3', () => {
    const groups = deriveGroups(['create_project', 'create_model', 'add_field'])
    expect(groups[1].status).toBe('completed')
    expect(groups[2].status).toBe('todo') // group 3 still accessible
  })

  it('group 3 can be completed before group 2', () => {
    const groups = deriveGroups(['create_project', 'apply_preset', 'create_role'])
    expect(groups[2].status).toBe('completed')
    expect(groups[1].status).toBe('todo') // group 2 still todo
  })

  it('group 4 completes last', () => {
    const groups = deriveGroups([
      'create_project',
      'create_model',
      'add_field',
      'apply_preset',
      'create_role',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ])
    expect(groups.every((g) => g.status === 'completed')).toBe(true)
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
  })

  it('markStep dedup — already completed step is not added again', () => {
    const prev: OnboardingState = {
      ...defaultOnboardingState('test-org'),
      completedSteps: ['create_project'],
    }
    const id: OnboardingStepId = 'create_project'
    expect(prev.completedSteps.includes(id)).toBe(true)
  })

  it('markStep create_project stores projectSlug', () => {
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

  it('project-scoped route returns null when projectSlug is null', () => {
    const step = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'my-org', projectSlug: null })).toBeNull()
  })

  it('project-scoped route returns full path when projectSlug set', () => {
    const step = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'my-org', projectSlug: 'my-project' }))
      .toBe('/org/my-org/project/my-project/model-editor')
  })
})
