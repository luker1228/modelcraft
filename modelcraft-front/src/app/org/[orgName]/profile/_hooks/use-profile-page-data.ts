'use client'

import { useMemo } from 'react'
import { useMyUserProfile } from '@web/hooks/user/use-my-user-profile'
import type { UseProfilePageDataReturn } from './types'

export function useProfilePageData(): UseProfilePageDataReturn {
  const { data, loading, error, refetch } = useMyUserProfile()

  return useMemo(
    () => ({
      profile: data,
      loading,
      error,
      refetch,
    }),
    [data, loading, error, refetch]
  )
}
