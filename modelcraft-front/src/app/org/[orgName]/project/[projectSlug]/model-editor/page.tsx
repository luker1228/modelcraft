'use client'

import { Suspense } from 'react'
import { ModelEditorView } from './_components/ModelEditorView'
import { Loader2 } from 'lucide-react'

export default function ModelEditorPage() {
  return (
    <Suspense
      fallback={
        <div className="flex size-full items-center justify-center">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <ModelEditorView />
    </Suspense>
  )
}
