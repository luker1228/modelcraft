'use client'

// TODO: developers section is temporarily hidden.
// Direct access redirects to dashboard.

import { useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'

export default function DevelopersPage() {
  const router = useRouter()
  const params = useParams()
  const orgName = params.orgName as string

  useEffect(() => {
    router.replace(`/org/${orgName}/dashboard`)
  }, [router, orgName])

  return null
}
