'use client'

import { useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'

export default function DevelopersPage() {
  const router = useRouter()
  const params = useParams()
  const orgName = params.orgName as string

  useEffect(() => {
    router.replace(`/org/${orgName}/developers/members`)
  }, [router, orgName])

  return null
}
