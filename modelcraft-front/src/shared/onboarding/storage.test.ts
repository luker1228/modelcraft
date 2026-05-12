import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  readOnboardingState,
  writeOnboardingState,
  ONBOARDING_KEY,
  defaultOnboardingState,
} from './storage'

// Mock localStorage in Node environment
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

  it('round-trips state to localStorage', () => {
    const state = {
      orgName: 'my-org',
      projectSlug: 'my-project',
      completedSteps: ['create_project' as const],
      dismissed: false,
      panelOpen: true,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('my-org')).toEqual(state)
  })

  it('returns default state when orgName does not match stored state', () => {
    const state = {
      orgName: 'org-a',
      projectSlug: null,
      completedSteps: [],
      dismissed: false,
      panelOpen: false,
    }
    writeOnboardingState(state)
    const result = readOnboardingState('org-b')
    expect(result).toEqual(defaultOnboardingState('org-b'))
  })

  it('handles corrupt localStorage gracefully', () => {
    localStorage.setItem(ONBOARDING_KEY, 'not-json')
    const result = readOnboardingState('my-org')
    expect(result).toEqual(defaultOnboardingState('my-org'))
  })
})
