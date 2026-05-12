export const ONBOARDING_KEY = 'mc_onboarding_v1'

export type OnboardingStepId =
  | 'create_project'
  | 'create_model'
  | 'add_field'
  | 'apply_preset'
  | 'create_role'
  | 'add_end_user'
  | 'assign_role'
  | 'end_user_login'

export interface OnboardingState {
  orgName: string
  projectSlug: string | null
  completedSteps: OnboardingStepId[]
  dismissed: boolean
  panelOpen: boolean
}

export function defaultOnboardingState(orgName: string): OnboardingState {
  return {
    orgName,
    projectSlug: null,
    completedSteps: [],
    dismissed: false,
    panelOpen: false,
  }
}

export function readOnboardingState(orgName: string): OnboardingState {
  try {
    const raw = localStorage.getItem(ONBOARDING_KEY)
    if (!raw) return defaultOnboardingState(orgName)
    const parsed = JSON.parse(raw) as OnboardingState
    if (parsed.orgName !== orgName) return defaultOnboardingState(orgName)
    return parsed
  } catch {
    return defaultOnboardingState(orgName)
  }
}

export function writeOnboardingState(state: OnboardingState): void {
  localStorage.setItem(ONBOARDING_KEY, JSON.stringify(state))
}
