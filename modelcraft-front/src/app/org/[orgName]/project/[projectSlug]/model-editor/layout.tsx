'use client'

import React from 'react'

export default function ModelEditorLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="size-full overflow-hidden">
      {children}
    </div>
  )
}
