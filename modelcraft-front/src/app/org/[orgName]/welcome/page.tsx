'use client'

import { useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'

/**
 * 重定向页面：/org/[orgName]/welcome -> /org/[orgName]/dashboard
 */
export default function WelcomeRedirect() {
  const router = useRouter()
  const params = useParams()
  const orgName = params.orgName as string

  useEffect(() => {
    if (orgName) {
      router.replace(`/org/${orgName}/dashboard`)
    }
  }, [orgName, router])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <div className="mx-auto size-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
        <p className="mt-3 text-sm text-muted-foreground">重定向中...</p>
      </div>
    </div>
  )
}
