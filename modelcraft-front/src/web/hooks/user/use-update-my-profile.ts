import { useCallback, useState } from 'react'
import { useApolloClient, useMutation } from '@apollo/client'
import { useOrganization } from '@web/hooks/organization/use-organization'
import { UPDATE_MY_PROFILE } from '@/api-client/profile'
import { MY_USER_PROFILE } from '@/api-client/profile'
import { useOrgScopedContext } from '@api-client/apollo/public'
import type {
  ProfileDomainError,
  UpdateMyProfileFormValues,
  UserProfileStatus,
  UserProfileView,
} from '@/types/profile'

interface MutationPayloadError {
  __typename: string
  message: string
  suggestion?: string | null
}

interface UpdateMyProfileMutationData {
  updateMyProfile?: {
    profile?: {
      id: string
      userId: string
      nickname: string
      avatarUrl?: string | null
      bio?: string | null
      createdAt: string
      updatedAt: string
    } | null
    error?: MutationPayloadError | null
  } | null
}

interface UpdateMyProfileMutationVariables {
  input: UpdateMyProfileFormValues
}

interface QueryProfilePayloadError {
  __typename: string
  message: string
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
  error?: QueryProfilePayloadError | null
}

interface MyUserProfileQueryData {
  myUserProfile?: MyUserProfilePayload | null
}

export interface UseUpdateMyProfileReturn {
  mutate: (input: UpdateMyProfileFormValues) => Promise<UserProfileView | null>
  loading: boolean
  error: ProfileDomainError | null
  clearError: () => void
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

function mapMutationError(error: MutationPayloadError): ProfileDomainError {
  if (error.__typename === 'ProfileNotFound') {
    return {
      type: 'ProfileNotFound',
      message: error.message,
    }
  }

  if (error.__typename === 'InvalidInput') {
    return {
      type: 'InvalidInput',
      message: error.message,
      suggestion: error.suggestion ?? undefined,
    }
  }

  return {
    type: 'Unknown',
    message: error.message,
  }
}

function mapRefetchPayloadError(error?: QueryProfilePayloadError | null): ProfileDomainError | null {
  if (!error) {
    return null
  }

  if (error.__typename === 'ProfileNotFound') {
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

export function useUpdateMyProfile(): UseUpdateMyProfileReturn {
  const apolloClient = useApolloClient()
  const { orgName } = useOrganization()
  const [domainError, setDomainError] = useState<ProfileDomainError | null>(null)

  const orgScopedContext = useOrgScopedContext(orgName)

  const [updateMyProfileMutation, { loading, error: mutationError }] = useMutation<
    UpdateMyProfileMutationData,
    UpdateMyProfileMutationVariables
  >(UPDATE_MY_PROFILE, {
    context: orgScopedContext,
  })

  const clearError = useCallback(() => {
    setDomainError(null)
  }, [])

  const mutate = useCallback(
    async (input: UpdateMyProfileFormValues) => {
      if (!orgName) {
        setDomainError({
          type: 'Unknown',
          message: '未检测到组织上下文，无法更新资料。',
        })
        return null
      }

      setDomainError(null)

      const mutationResult = await updateMyProfileMutation({
        variables: {
          input,
        },
      })

      const payload = mutationResult.data?.updateMyProfile

      if (!payload) {
        setDomainError({
          type: 'Unknown',
          message: '更新资料失败，请稍后重试。',
        })
        return null
      }

      if (payload.error) {
        setDomainError(mapMutationError(payload.error))
        return null
      }

      const refreshedResult = await apolloClient.query<MyUserProfileQueryData>({
        query: MY_USER_PROFILE,
        context: orgScopedContext,
        fetchPolicy: 'network-only',
      })

      const refreshedPayload = refreshedResult.data?.myUserProfile
      const refetchPayloadError = mapRefetchPayloadError(refreshedPayload?.error)

      if (refetchPayloadError) {
        setDomainError(refetchPayloadError)
        return null
      }

      const refreshedProfile = mapQueryPayloadToProfileView(refreshedPayload)

      if (!refreshedProfile) {
        setDomainError({
          type: 'Unknown',
          message: '资料已更新，但读取最新资料失败。',
        })
        return null
      }

      return refreshedProfile
    },
    [apolloClient, orgName, orgScopedContext, updateMyProfileMutation]
  )

  const mergedError = useMemo(() => {
    if (domainError) {
      return domainError
    }

    if (mutationError) {
      return {
        type: 'Unknown',
        message: mutationError.message,
      } satisfies ProfileDomainError
    }

    return null
  }, [domainError, mutationError])

  return {
    mutate,
    loading,
    error: mergedError,
    clearError,
  }
}
