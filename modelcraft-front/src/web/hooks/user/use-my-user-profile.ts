import { useCallback, useMemo } from 'react'
import { useQuery } from '@apollo/client'
import { useOrganization } from '@web/hooks/organization/use-organization'
import { MY_USER_PROFILE } from '@/api-client/profile'
import { useOrgScopedContext } from '@api-client/apollo/context'
import type { ProfileDomainError, UserProfileStatus, UserProfileView } from '@/types/profile'

interface ProfilePayloadError {
  __typename: string
  message: string
  resourceType?: string
}

interface QueryUserProfile {
  id: string
  phone: string
  userName: string
  status: string
  createdAt: string
  updatedAt: string
  profile?: {
    id: string
    userId: string
    nickname: string
    avatarUrl?: string | null
    bio?: string | null
    createdAt: string
    updatedAt: string
  } | null
}

interface MyUserProfilePayload {
  user?: QueryUserProfile | null
  error?: ProfilePayloadError | null
}

interface MyUserProfileQueryData {
  myUserProfile?: MyUserProfilePayload | null
}

export interface UseMyUserProfileReturn {
  data: UserProfileView | null
  loading: boolean
  error: ProfileDomainError | null
  refetch: () => Promise<void>
}

function normalizeUserStatus(status: string): UserProfileStatus {
  if (status === 'ACTIVE' || status === 'SUSPENDED' || status === 'REGISTERED') {
    return status
  }

  return 'REGISTERED'
}

function mapQueryPayloadToProfileView(payload?: MyUserProfilePayload | null): UserProfileView | null {
  const user = payload?.user
  const profile = user?.profile

  if (!user || !profile) {
    return null
  }

  return {
    userId: user.id,
    phone: user.phone,
    userName: user.userName,
    status: normalizeUserStatus(user.status),
    profileId: profile.id,
    nickname: profile.nickname,
    avatarUrl: profile.avatarUrl ?? undefined,
    bio: profile.bio ?? undefined,
    createdAt: profile.createdAt,
    updatedAt: profile.updatedAt,
  }
}

function mapPayloadError(error?: ProfilePayloadError | null): ProfileDomainError | null {
  if (!error) {
    return null
  }

  if (error.__typename === 'ResourceNotFound') {
    return {
      type: 'ProfileNotFound',
      message: error.message,
    }
  }

  return {
    type: 'Unknown',
    message: error.message,
  }
}

export function useMyUserProfile(): UseMyUserProfileReturn {
  const { orgName } = useOrganization()

  const orgScopedContext = useOrgScopedContext(orgName ?? undefined)

  const { data, loading, error, refetch: apolloRefetch } = useQuery<MyUserProfileQueryData>(MY_USER_PROFILE, {
    skip: !orgName,
    context: orgScopedContext,
  })

  const mappedResult = useMemo<{ data: UserProfileView | null; error: ProfileDomainError | null }>(() => {
    if (error) {
      return {
        data: null,
        error: {
          type: 'Unknown',
          message: error.message,
        },
      }
    }

    const payload: MyUserProfilePayload | null | undefined = data?.myUserProfile
    const mappedPayloadError: ProfileDomainError | null = mapPayloadError(payload?.error)

    if (mappedPayloadError) {
      return {
        data: null,
        error: mappedPayloadError,
      }
    }

    const profile: UserProfileView | null = mapQueryPayloadToProfileView(payload)

    if (!profile && payload?.user) {
      return {
        data: null,
        error: {
          type: 'ProfileNotFound',
          message: '当前用户暂无个人资料，请先完成资料创建。',
        },
      }
    }

    return {
      data: profile,
      error: null,
    }
  }, [data, error])

  const refetch = useCallback(async () => {
    if (!orgName) {
      return
    }

    await apolloRefetch()
  }, [apolloRefetch, orgName])

  return {
    data: mappedResult.data,
    loading,
    error: mappedResult.error,
    refetch,
  }
}
