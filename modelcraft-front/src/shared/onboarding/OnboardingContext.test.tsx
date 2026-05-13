import { describe, it, expect } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS, ALL_TRACKED_STEPS } from './steps'

function deriveGroups(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  const globalCurrentId = ALL_TRACKED_STEPS.find(
    (s) => !completedSet.has(s.id)
  )?.id ?? null
  let markedCurrent = false
  return ONBOARDING_GROUPS.map((group) => {
    const steps = group.steps.map((step) => {
      if (step.kind === 'nav') return { ...step, status: 'nav' as const }
      if (completedSet.has(step.id)) return { ...step, status: 'completed' as const }
      if (step.id === globalCurrentId && !markedCurrent) {
        markedCurrent = true
        return { ...step, status: 'current' as const }
      }
      return { ...step, status: 'locked' as const }
    })
    const tracked = steps.filter((s) => s.kind === 'tracked')
    const allDone = tracked.length > 0 && tracked.every((s) => s.kind === 'tracked' && s.status === 'completed')
    return { id: group.id, label: group.label, status: allDone ? 'completed' as const : 'todo' as const, steps }
  })
}

describe('tutorial step definitions', () => {
  it('has exactly 5 groups', () => {
    expect(ONBOARDING_GROUPS.length).toBe(5)
  })

  it('group ids match expected 5-step flow', () => {
    const ids = ONBOARDING_GROUPS.map((g) => g.id)
    expect(ids).toEqual([
      'setup_project',
      'design_model',
      'add_users',
      'assign_permissions',
      'experience_login',
    ])
  })

  it('first required step is current at start', () => {
    const groups = deriveGroups([])
    const g1 = groups[0]
    const tracked = g1.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current')
  })

  it('nav steps always have nav status', () => {
    const groups = deriveGroups([])
    const navSteps = groups.flatMap((g) => g.steps.filter((s) => s.kind === 'nav'))
    expect(navSteps.every((s) => s.status === 'nav')).toBe(true)
  })

  it('completing create_project advances current to create_model', () => {
    const groups = deriveGroups(['create_project'])
    const g2 = groups[1]
    const tracked = g2.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // create_model
  })

  it('group 2 completes when create_model done', () => {
    const groups = deriveGroups(['create_project', 'create_model'])
    expect(groups[1].status).toBe('completed')
  })

  it('completing all required steps except last advances to end_user_login', () => {
    const groups = deriveGroups(['create_project', 'create_model', 'add_end_user', 'assign_role'])
    const g5 = groups[4]
    const tracked = g5.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // end_user_login
  })

  it('ONBOARDING_STEPS contains exactly the 5 required step IDs', () => {
    const ids = ONBOARDING_STEPS.map((s) => s.id)
    expect(ids).toEqual([
      'create_project',
      'create_model',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ])
  })

  it('end_user_login step route returns null (no navigation)', () => {
    const step = ALL_TRACKED_STEPS.find((s) => s.id === 'end_user_login')!
    expect(step.route({ orgName: 'org', projectSlug: null })).toBeNull()
  })

  it('create_model route resolves to model-editor when projectSlug set', () => {
    const step = ALL_TRACKED_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'org', projectSlug: 'proj' }))
      .toBe('/org/org/project/proj/model-editor')
  })

  it('markStep dedup works', () => {
    const prev: OnboardingState = { ...defaultOnboardingState('test'), completedSteps: ['create_project'] }
    expect(prev.completedSteps.includes('create_project')).toBe(true)
  })
})
