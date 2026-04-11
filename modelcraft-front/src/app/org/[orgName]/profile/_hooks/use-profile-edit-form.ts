'use client'

import { useCallback, useEffect, useMemo } from 'react'
import { useUpdateMyProfile } from '@web/hooks/user/use-update-my-profile'
import type { UseProfileEditFormOptions, UseProfileEditFormReturn } from './types'

export function useProfileEditForm({
  profile,
  refetchProfile,
  onSuccess,
  onSavingChange,
}: UseProfileEditFormOptions): UseProfileEditFormReturn {
  const { mutate, loading, error, clearError } = useUpdateMyProfile()

  const initialValues = useMemo(
    () => ({
      nickname: profile?.nickname ?? '',
      avatarUrl: profile?.avatarUrl ?? '',
      bio: profile?.bio ?? '',
    }),
    [profile]
  )

  const submit = useCallback(
    async (values: { nickname?: string; avatarUrl?: string; bio?: string }) => {
      const updatedProfile = await mutate(values)

      if (!updatedProfile) {
        return
      }

      await refetchProfile()
      onSuccess?.()
    },
    [mutate, onSuccess, refetchProfile]
  )

  const reset = useCallback(() => {
    clearError()
  }, [clearError])

  useEffect(() => {
    onSavingChange?.(loading)

    return () => {
      onSavingChange?.(false)
    }
  }, [loading, onSavingChange])

  return {
    initialValues,
    submit,
    reset,
    saving: loading,
    error,
  }
}
