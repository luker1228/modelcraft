import React from 'react'

interface EndUserProjectLayoutProps {
  children: React.ReactNode
}

/**
 * End-user 专用项目布局。
 * 与开发者 web 完全分离，不复用 AppLayout。
 */
export default function EndUserProjectLayout({ children }: EndUserProjectLayoutProps) {
  return <>{children}</>
}

