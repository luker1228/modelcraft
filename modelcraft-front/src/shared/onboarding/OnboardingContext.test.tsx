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
    (s) => !s.optional && !completedSet.has(s.id)
  )?.id ?? null

  let markedCurrent = false
  return ONBOARDING_GROUPS.map((group) => {
    const isOptionalGroup = group.optional === true
    const steps = group.steps.map((step) => {
      if (step.kind === 'nav') return { ...step, status: 'nav' as const }
      if (completedSet.has(step.id)) return { ...step, status: 'completed' as const }
      if (isOptionalGroup) return { ...step, status: 'locked' as const }
      if (step.id === globalCurrentId && !markedCurrent) {
        markedCurrent = true
        return { ...step, status: 'current' as const }
      }
      return { ...step, status: 'locked' as const }
    })
    const tracked = steps.filter((s) => s.kind === 'tracked')
    const allDone = tracked.length > 0 && tracked.every((s) => s.kind === 'tracked' && s.status === 'completed')
    return { id: group.id, label: group.label, optional: group.optional, status: allDone ? 'completed' as const : 'todo' as const, steps }
  })
}

describe('onboarding group derivation', () => {
  it('first required step is current at start', () => {
    const groups = deriveGroups([])
    const g1 = groups[0]
    const tracked = g1.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current')
  })

  it('nav steps are always nav status', () => {
    const groups = deriveGroups([])
    const navStep = groups[0].steps.find((s) => s.kind === 'nav')
    expect(navStep?.status).toBe('nav')
  })

  it('optional group steps are all locked regardless of current pointer', () => {
    const groups = deriveGroups([])
    const optGroup = groups.find((g) => g.optional)
    expect(optGroup).toBeDefined()
    const tracked = optGroup!.steps.filter((s) => s.kind === 'tracked')
    expect(tracked.every((s) => s.status === 'locked')).toBe(true)
  })

  it('completing create_project advances current to select_database', () => {
    const groups = deriveGroups(['create_project'])
    const g2 = groups[1]
    const tracked = g2.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // select_database
  })

  it('group 2 completes when all its tracked steps done', () => {
    const groups = deriveGroups(['create_project', 'select_database', 'create_model', 'insert_column', 'insert_data'])
    expect(groups[1].status).toBe('completed')
  })

  it('optional group does not block group 4 from having a current step', () => {
    const groups = deriveGroups(['create_project', 'select_database', 'create_model', 'insert_column', 'insert_data'])
    const g4 = groups[3]
    const tracked = g4.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // add_end_user
  })

  it('optional steps in opt group are still completeable independently', () => {
    const groups = deriveGroups(['create_permission'])
    const optGroup = groups.find((g) => g.optional)!
    const perm = optGroup.steps.find((s) => s.kind === 'tracked' && s.id === 'create_permission')
    expect(perm?.status).toBe('completed')
  })

  it('required completion count excludes optional steps', () => {
    const requiredIds: OnboardingStepId[] = [
      'create_project', 'select_database', 'create_model', 'insert_column', 'insert_data',
      'add_end_user', 'assign_role', 'end_user_login',
    ]
    expect(requiredIds.length).toBe(ONBOARDING_STEPS.length)
  })

  it('markStep dedup', () => {
    const prev: OnboardingState = { ...defaultOnboardingState('test'), completedSteps: ['create_project'] }
    expect(prev.completedSteps.includes('create_project')).toBe(true)
  })

  it('create_project stores projectSlug', () => {
    const prev = defaultOnboardingState('test')
    const id: OnboardingStepId = 'create_project'
    const slug = 'my-project'
    const next: OnboardingState = {
      ...prev,
      completedSteps: [...prev.completedSteps, id],
      projectSlug: id === 'create_project' && slug ? slug : prev.projectSlug,
    }
    expect(next.projectSlug).toBe('my-project')
  })

  it('create_model route resolves to model-editor when projectSlug set', () => {
    const step = ALL_TRACKED_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'org', projectSlug: 'proj' }))
      .toBe('/org/org/project/proj/model-editor')
  })
})
