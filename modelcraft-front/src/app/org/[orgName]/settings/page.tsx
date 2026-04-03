'use client'

import { useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'

export default function SettingsPage() {
  const router = useRouter()
  const params = useParams()
  const orgName = params.orgName as string

  useEffect(() => {
    router.replace(`/org/${orgName}/settings/members`)
  }, [router, orgName])

  return null
}
