export type ProfileLoadStatus = 'idle' | 'loading' | 'success' | 'error'

export type UserProfileStatus = 'REGISTERED' | 'ACTIVE' | 'SUSPENDED'

export interface UserProfileView {
  userId: string
  phone: string
  userName: string
  status: UserProfileStatus
  profileId: string
  nickname: string
  avatarUrl?: string
  bio?: string
  createdAt: string
  updatedAt: string
}

export interface UpdateMyProfileFormValues {
  nickname?: string
  avatarUrl?: string
  bio?: string
}

export type ProfileDomainErrorType = 'ProfileNotFound' | 'InvalidProfileInput' | 'Unknown'

export interface ProfileDomainError {
  type: ProfileDomainErrorType
  message: string
  suggestion?: string
}
