"use client"

import * as React from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@web/components/ui/dialog"
import { Button } from "@web/components/ui/button"
import { OrganizationNameInput } from "@web/components/organization-name-input"
import { generateRandomOrgName } from "@/shared/organization-name-generator"
import { validateOrgName } from "@/shared/organization-name-validator"
import { Alert, AlertDescription } from "@web/components/ui/alert"
import { Loader2 } from "lucide-react"

export interface OrganizationOnboardingDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentOrgName?: string
  onComplete?: () => void
}

export function OrganizationOnboardingDialog({
  open,
  onOpenChange,
  currentOrgName,
  onComplete,
}: OrganizationOnboardingDialogProps) {
  const [orgName, setOrgName] = React.useState(() => generateRandomOrgName())
  const [mode, setMode] = React.useState<'default' | 'custom'>('default')
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | undefined>()

  const validation = React.useMemo(() => {
    if (!orgName) return { valid: false }
    return validateOrgName(orgName)
  }, [orgName])

  const handleSave = async () => {
    if (!validation.valid) {
      setError(validation.error || 'Invalid organization name')
      return
    }

    setLoading(true)
    setError(undefined)

    try {
      const response = await fetch('/api/organization/update', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          oldName: currentOrgName,
          newName: orgName,
        }),
      })

      if (!response.ok) {
        const data = await response.json() as Record<string, unknown>
        throw new Error(typeof data.error === 'string' ? data.error : 'Failed to update organization name')
      }

      // Success! Close dialog and optionally trigger re-login
      onComplete?.()
      onOpenChange(false)
      
      // Optional: Show success message and redirect to re-login
      // For now, we'll just close the dialog
    } catch (err) {
      console.error('Organization update failed:', err)
      setError(err instanceof Error ? err.message : 'Failed to update organization name')
    } finally {
      setLoading(false)
    }
  }

  const handleSkip = () => {
    // User chooses to skip customization
    onOpenChange(false)
    onComplete?.()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Welcome to ModelCraft! 🎉</DialogTitle>
          <DialogDescription>
            Your organization was created with a temporary identifier. Let's customize it to make it memorable and easy to share.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {currentOrgName && (
            <Alert variant="info">
              <AlertDescription className="text-sm">
                <strong>Current name:</strong> {currentOrgName}
                <br />
                <span className="text-muted-foreground">
                  This will be replaced with your custom name.
                </span>
              </AlertDescription>
            </Alert>
          )}

          <OrganizationNameInput
            value={orgName}
            onChange={setOrgName}
            mode={mode}
            onModeChange={setMode}
            disabled={loading}
          />

          {error && (
            <Alert variant="destructive">
              <AlertDescription className="text-sm">{error}</AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            onClick={handleSkip}
            disabled={loading}
          >
            Skip for Now
          </Button>
          <Button
            type="button"
            onClick={handleSave}
            disabled={!validation.valid || loading}
          >
            {loading && <Loader2 className="mr-2 size-4 animate-spin" />}
            Save and Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
