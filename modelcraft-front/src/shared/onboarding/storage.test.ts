import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  readOnboardingState,
  writeOnboardingState,
  ONBOARDING_KEY,
  defaultOnboardingState,
} from './storage'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

vi.stubGlobal('localStorage', localStorageMock)

describe('onboarding storage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns default state when localStorage is empty', () => {
    const state = readOnboardingState('my-org')
    expect(state).toEqual(defaultOnboardingState('my-org'))
  })

  it('default state has panelOpen: true', () => {
    const state = defaultOnboardingState('my-org')
    expect(state.panelOpen).toBe(true)
  })

  it('default state has no dismissed field', () => {
    const state = defaultOnboardingState('my-org')
    expect('dismissed' in state).toBe(false)
  })

  it('round-trips state to localStorage', () => {
    const state = {
      orgName: 'my-org',
      projectSlug: 'my-project',
      completedSteps: ['create_project' as const],
      panelOpen: true,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('my-org')).toEqual(state)
  })

  it('returns default state when orgName does not match', () => {
    const state = {
      orgName: 'org-a',
      projectSlug: null,
      completedSteps: [],
      panelOpen: false,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('org-b')).toEqual(defaultOnboardingState('org-b'))
  })

  it('handles corrupt localStorage gracefully', () => {
    localStorage.setItem(ONBOARDING_KEY, 'not-json')
    expect(readOnboardingState('my-org')).toEqual(defaultOnboardingState('my-org'))
  })

  it('uses v2 key to avoid collision with old data', () => {
    expect(ONBOARDING_KEY).toBe('mc_onboarding_v2')
  })
})
