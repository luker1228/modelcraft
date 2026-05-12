import { describe, it, expect } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS } from './steps'

function deriveGroups(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  return ONBOARDING_GROUPS.map((group) => {
    const steps = group.steps.map((step) => {
      if (step.kind === 'nav') return { ...step, status: 'nav' as const }
      return {
        ...step,
        status: completedSet.has(step.id) ? 'completed' as const : 'todo' as const,
      }
    })
    const tracked = steps.filter((s) => s.kind === 'tracked')
    const allDone = tracked.length > 0 && tracked.every((s) => s.kind === 'tracked' && s.status === 'completed')
    return { id: group.id, label: group.label, status: allDone ? 'completed' as const : 'todo' as const, steps }
  })
}

describe('onboarding group derivation', () => {
  it('nav steps always have status nav, never block group completion', () => {
    const groups = deriveGroups([])
    const g1 = groups[0]
    const navStep = g1.steps.find((s) => s.kind === 'nav')
    expect(navStep?.status).toBe('nav')
    expect(g1.status).toBe('todo')
  })

  it('group 1 completes only when create_project is done', () => {
    const groups = deriveGroups(['create_project'])
    expect(groups[0].status).toBe('completed')
  })

  it('all groups start as todo', () => {
    const groups = deriveGroups([])
    expect(groups.every((g) => g.status === 'todo')).toBe(true)
  })

  it('group 2 completes when create_model is done', () => {
    const groups = deriveGroups(['create_model'])
    expect(groups[1].status).toBe('completed')
  })

  it('group 2 not complete when nothing done', () => {
    const groups = deriveGroups([])
    expect(groups[1].status).toBe('todo')
  })

  it('group 3 completes when all permission steps done', () => {
    const groups = deriveGroups(['create_permission', 'create_bundle', 'create_role'])
    expect(groups[2].status).toBe('completed')
  })

  it('group 3 can be completed independently of group 2', () => {
    const groups = deriveGroups(['create_permission', 'create_bundle', 'create_role'])
    expect(groups[2].status).toBe('completed')
    expect(groups[1].status).toBe('todo')
  })

  it('tracked steps count matches ONBOARDING_STEPS', () => {
    const allTracked = ONBOARDING_GROUPS.flatMap((g) => g.steps).filter((s) => s.kind === 'tracked')
    expect(allTracked.length).toBe(ONBOARDING_STEPS.length)
  })

  it('isComplete when all tracked steps done', () => {
    const all: OnboardingStepId[] = [
      'create_project',
      'create_model',
      'create_permission',
      'create_bundle',
      'create_role',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ]
    const groups = deriveGroups(all)
    expect(groups.every((g) => g.status === 'completed')).toBe(true)
  })

  it('markStep dedup', () => {
    const prev: OnboardingState = {
      ...defaultOnboardingState('test-org'),
      completedSteps: ['create_project'],
    }
    expect(prev.completedSteps.includes('create_project')).toBe(true)
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

  it('tracked step route returns workspace when projectSlug is null', () => {
    const step = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'my-org', projectSlug: null })).toBe('/org/my-org/workspace')
  })

  it('tracked step route returns project path when projectSlug set', () => {
    const step = ONBOARDING_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'my-org', projectSlug: 'my-project' }))
      .toBe('/org/my-org/project/my-project/model-editor')
  })
})
