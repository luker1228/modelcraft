export const ONBOARDING_KEY = 'mc_onboarding_v2'

export type OnboardingStepId =
  | 'confirm_project'
  | 'confirm_model'
  | 'confirm_end_user'
  | 'confirm_permission'
  | 'end_user_login'

export interface OnboardingState {
  orgName: string
  projectSlug: string | null
  completedSteps: OnboardingStepId[]
  panelOpen: boolean
  dismissed: boolean
}

export function defaultOnboardingState(orgName: string): OnboardingState {
  return {
    orgName,
    projectSlug: null,
    completedSteps: [],
    panelOpen: true,
    dismissed: false,
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
