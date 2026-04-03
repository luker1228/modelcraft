import { memo } from 'react'

interface LoadingScreenProps {
  message?: string
}

/**
 * Full-screen loading indicator
 * 
 * Displays a centered loading spinner with optional message
 * Used during authentication checks and page transitions
 */
export const LoadingScreen = memo(({ 
  message = 'Loading...' 
}: LoadingScreenProps) => {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="text-center">
        <div className="mx-auto size-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
        <p className="mt-3 text-sm text-muted-foreground">{message}</p>
      </div>
    </div>
  )
})

LoadingScreen.displayName = 'LoadingScreen'
