'use client'

import { useCallback, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import type { UseProfilePageStateReturn } from './types'

interface UseProfilePageStateOptions {
  mode: 'view' | 'edit'
}

export function useProfilePageState({ mode }: UseProfilePageStateOptions): UseProfilePageStateReturn {
  const params = useParams()
  const router = useRouter()
  const [saving, setSaving] = useState(false)

  const orgName = typeof params.orgName === 'string' ? params.orgName : ''

  const goToEdit = useCallback(() => {
    if (!orgName) {
      return
    }

    router.push(`/org/${orgName}/profile/edit`)
  }, [orgName, router])

  const goToOverview = useCallback(() => {
    if (!orgName) {
      return
    }

    router.push(`/org/${orgName}/profile`)
  }, [orgName, router])

  return {
    mode,
    saving,
    setSaving,
    goToEdit,
    goToOverview,
  }
}
