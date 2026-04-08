'use client'

import { useErrorStore } from '@web/stores/error'
import { GraphQLErrorDialog } from '@web/components/common/GraphQLErrorDialog'

export function ErrorProvider({ children }: { children: React.ReactNode }) {
  const { 
    isErrorDialogOpen, 
    currentErrors, 
    currentContext, 
    hideErrorDialog 
  } = useErrorStore()

  return (
    <>
      {children}
      <GraphQLErrorDialog
        open={isErrorDialogOpen}
        onOpenChange={hideErrorDialog}
        errors={currentErrors}
        context={currentContext}
      />
    </>
  )
}