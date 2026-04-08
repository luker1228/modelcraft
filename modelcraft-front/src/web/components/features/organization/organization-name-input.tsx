"use client"

import * as React from "react"
import { Copy, Check, RefreshCw } from "lucide-react"

import { cn } from "@/shared/utils"
import { Button } from "@web/components/ui/button"
import { Input } from "@web/components/ui/input"
import { Label } from "@web/components/ui/label"
import { Alert, AlertDescription } from "@web/components/ui/alert"
import {
  generateRandomOrgName,
  appendRandomSuffix,
  generateRandomSuffix,
} from "@/shared/organization-name-generator"
import {
  validateCustomBase,
  validateOrgName,
} from "@/shared/organization-name-validator"

export type OrganizationNameMode = 'default' | 'custom'

export interface OrganizationNameInputProps {
  value: string
  onChange: (name: string) => void
  mode: OrganizationNameMode
  onModeChange: (mode: OrganizationNameMode) => void
  className?: string
  disabled?: boolean
}

export function OrganizationNameInput({
  value,
  onChange,
  mode,
  onModeChange,
  className,
  disabled = false,
}: OrganizationNameInputProps) {
  const [customBase, setCustomBase] = React.useState('')
  const [copied, setCopied] = React.useState(false)
  const [validationError, setValidationError] = React.useState<string | undefined>()

  // Extract base and suffix from value if in custom mode
  React.useEffect(() => {
    if (mode === 'custom' && value.includes('_')) {
      const parts = value.split('_')
      if (parts.length === 2) {
        setCustomBase(parts[0])
      }
    }
  }, [mode, value])

  const handleCustomize = () => {
    onModeChange('custom')
    setCustomBase('')
    setValidationError(undefined)
  }

  const handleUseDefault = () => {
    onModeChange('default')
    const newName = generateRandomOrgName()
    onChange(newName)
    setCustomBase('')
    setValidationError(undefined)
  }

  const handleRegenerateDefault = () => {
    const newName = generateRandomOrgName()
    onChange(newName)
  }

  const handleCustomBaseChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newBase = e.target.value.toLowerCase()
    setCustomBase(newBase)

    if (!newBase) {
      onChange('')
      setValidationError(undefined)
      return
    }

    // Validate base
    const validation = validateCustomBase(newBase)
    if (!validation.valid) {
      setValidationError(validation.error)
      onChange('')
      return
    }

    // Generate full name with suffix
    const fullName = appendRandomSuffix(newBase)
    onChange(fullName)
    setValidationError(undefined)
  }

  const handleRegenerateSuffix = () => {
    if (customBase) {
      const newSuffix = generateRandomSuffix()
      const fullName = `${customBase}_${newSuffix}`
      onChange(fullName)
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  // Validate final name
  const finalValidation = React.useMemo(() => {
    if (!value) return { valid: false }
    return validateOrgName(value)
  }, [value])

  const previewSuffix = value.includes('_') ? value.split('_')[1] : null

  return (
    <div className={cn("space-y-4", className)}>
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="org-name">
            Organization Name
            <span className="ml-1 text-xs text-muted-foreground">
              (Cannot be changed later)
            </span>
          </Label>
          {mode === 'default' ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleCustomize}
              disabled={disabled}
            >
              Customize
            </Button>
          ) : (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleUseDefault}
              disabled={disabled}
            >
              Use Default
            </Button>
          )}
        </div>

        {mode === 'default' ? (
          <div className="space-y-2">
            <div className="flex gap-2">
              <Input
                id="org-name"
                value={value}
                readOnly
                disabled={disabled}
                className="font-mono"
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={handleRegenerateDefault}
                disabled={disabled}
                title="Generate new name"
              >
                <RefreshCw className="size-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={handleCopy}
                disabled={disabled || !value}
                title="Copy to clipboard"
              >
                {copied ? (
                  <Check className="size-4" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              A random 12-character organization identifier
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            <div className="flex gap-2">
              <div className="flex-1">
                <Input
                  id="org-name"
                  value={customBase}
                  onChange={handleCustomBaseChange}
                  placeholder="mycompany"
                  disabled={disabled}
                  maxLength={12}
                  className={cn(
                    "font-mono",
                    validationError && "border-destructive focus-visible:ring-destructive"
                  )}
                />
              </div>
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={handleRegenerateSuffix}
                disabled={disabled || !customBase || !!validationError}
                title="Regenerate suffix"
              >
                <RefreshCw className="size-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={handleCopy}
                disabled={disabled || !value}
                title="Copy to clipboard"
              >
                {copied ? (
                  <Check className="size-4" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>

            {customBase && previewSuffix && !validationError && (
              <div className="rounded-md border bg-muted/50 px-3 py-2">
                <p className="mb-1 text-xs text-muted-foreground">Preview:</p>
                <p className="font-mono text-sm">
                  {customBase}
                  <span className="text-muted-foreground">_</span>
                  {previewSuffix}
                </p>
              </div>
            )}

            {validationError && (
              <Alert variant="destructive">
                <AlertDescription className="text-xs">
                  {validationError}
                </AlertDescription>
              </Alert>
            )}

            <p className="text-xs text-muted-foreground">
              Enter 1-12 characters (lowercase letters, digits, hyphens). A 6-character random suffix will be added automatically.
            </p>
          </div>
        )}
      </div>

      {!finalValidation.valid && finalValidation.error && mode === 'custom' && !validationError && (
        <Alert variant="warning">
          <AlertDescription className="text-xs">
            {finalValidation.error}
          </AlertDescription>
        </Alert>
      )}

      {value && finalValidation.valid && (
        <Alert variant="info">
          <AlertDescription className="text-xs">
            <strong>Important:</strong> Your organization name "{value}" cannot be changed after registration. Make sure you're happy with it before proceeding.
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
