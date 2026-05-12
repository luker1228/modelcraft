'use client'

import { useEffect, useState } from 'react'

export function MSWProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false)

  useEffect(() => {
    // 双重保护：生产环境强制不启用，开发环境下还需 NEXT_PUBLIC_API_MOCKING=enabled
    if (
      process.env.NODE_ENV === 'production' ||
      process.env.NEXT_PUBLIC_API_MOCKING !== 'enabled'
    ) {
      setReady(true)
      return
    }

    import('@/mocks/browser').then(({ worker }) => {
      worker.start({ onUnhandledRequest: 'bypass' }).then(() => {
        setReady(true)
      })
    })
  }, [])

  if (!ready) return null

  return <>{children}</>
}
