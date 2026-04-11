import type { ProfileDomainError, UpdateMyProfileFormValues, UserProfileView } from '@/types'

export interface ProfilePageState {
  mode: 'view' | 'edit'
  saving: boolean
}

export interface ProfilePageData {
  profile: UserProfileView | null
  loading: boolean
  error: ProfileDomainError | null
}

export interface UseProfilePageStateReturn extends ProfilePageState {
  setSaving: (saving: boolean) => void
  goToEdit: () => void
  goToOverview: () => void
}

export interface UseProfilePageDataReturn extends ProfilePageData {
  refetch: () => Promise<void>
}

export interface UseProfileEditFormReturn {
  initialValues: UpdateMyProfileFormValues
  submit: (values: UpdateMyProfileFormValues) => Promise<void>
  reset: () => void
  saving: boolean
  error: ProfileDomainError | null
}

export interface UseProfileEditFormOptions {
  profile: UserProfileView | null
  refetchProfile: () => Promise<void>
  onSuccess?: () => void
  onSavingChange?: (saving: boolean) => void
}
