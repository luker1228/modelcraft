/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-argument */
import { useState, useCallback } from 'react'
import { useMutation } from '@apollo/client'

import { useProjectScopedClient } from '@api-client/apollo/public'
import { CREATE_END_USER_PERMISSION, GET_END_USER_PERMISSIONS } from '@/api-client/rbac'
import type { EndUserPermissionAction, EndUserRowScope, ColumnPolicy } from '@/types'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type WizardStep = 'model-action' | 'row-scope' | 'column-policy'

export interface WizardState {
  step: WizardStep
  modelId: string
  modelDisplayName: string
  action: EndUserPermissionAction | null
  displayName: string
  description: string
  rowScope: EndUserRowScope
  columnPolicy: ColumnPolicy
  /** Whether the selected model has an owner (EndUserRef) field */
  hasOwnerField: boolean
  /** Whether the selected model has a dept_id field */
  hasDeptIdField: boolean
}

const INITIAL_STATE: WizardState = {
  step: 'model-action',
  modelId: '',
  modelDisplayName: '',
  action: null,
  displayName: '',
  description: '',
  rowScope: 'ALL',
  columnPolicy: { defaultMode: 'VISIBLE', rules: [] },
  hasOwnerField: true,
  hasDeptIdField: false,
}

const STEP_ORDER: WizardStep[] = ['model-action', 'row-scope', 'column-policy']

// ---------------------------------------------------------------------------
// Mock field presence derivation (Wave 1)
// In Wave 2 this will be replaced by a real GraphQL query.
// ---------------------------------------------------------------------------

function deriveModelFieldPresence(modelId: string): {
  hasOwnerField: boolean
  hasDeptIdField: boolean
} {
  return {
    // Any non-empty modelId is treated as having an owner field for now
    hasOwnerField: modelId !== '',
    hasDeptIdField: false,
  }
}

// ---------------------------------------------------------------------------
// Return type
// ---------------------------------------------------------------------------

export interface UseCreatePermissionWizardReturn {
  state: WizardState
  updateField: <K extends keyof WizardState>(key: K, value: WizardState[K]) => void
  goNext: () => void
  goBack: () => void
  reset: () => void
  submit: () => Promise<void>
  submitting: boolean
  /** Non-null when the last submit call produced an error */
  submitError: string | null
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useCreatePermissionWizard(
  orgName: string,
  projectSlug: string,
  /** Called on successful creation. Default: navigate to permissions list. */
  onSuccess?: () => void,
): UseCreatePermissionWizardReturn {
  const client = useProjectScopedClient(projectSlug)

  const [state, setState] = useState<WizardState>(INITIAL_STATE)
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const [createPermissionMutation] = useMutation(CREATE_END_USER_PERMISSION, {
    client,
    refetchQueries: [GET_END_USER_PERMISSIONS],
  })

  // ---- field update ----

  const updateField = useCallback(
    <K extends keyof WizardState>(key: K, value: WizardState[K]) => {
      setState((prev) => ({ ...prev, [key]: value }))
    },
    [],
  )

  // ---- step navigation ----

  const goNext = useCallback(() => {
    setState((prev) => {
      const currentIndex = STEP_ORDER.indexOf(prev.step)
      const nextStep = STEP_ORDER[currentIndex + 1]
      if (!nextStep) return prev

      // Entering row-scope: resolve mock field presence for the selected model
      if (nextStep === 'row-scope') {
        const { hasOwnerField, hasDeptIdField } = deriveModelFieldPresence(prev.modelId)
        return { ...prev, step: nextStep, hasOwnerField, hasDeptIdField }
      }

      return { ...prev, step: nextStep }
    })
  }, [])

  const goBack = useCallback(() => {
    setState((prev) => {
      const currentIndex = STEP_ORDER.indexOf(prev.step)
      const prevStep = STEP_ORDER[currentIndex - 1]
      if (!prevStep) return prev
      return { ...prev, step: prevStep }
    })
  }, [])

  const reset = useCallback(() => {
    setState(INITIAL_STATE)
    setSubmitError(null)
  }, [])

  // ---- submit ----

  const submit = useCallback(async () => {
    if (!state.action) return

    setSubmitting(true)
    setSubmitError(null)

    try {
      const result = await createPermissionMutation({
        variables: {
          projectSlug,
          input: {
            modelId: state.modelId,
            action: state.action,
            rowScope: state.rowScope,
            displayName: state.displayName || undefined,
            description: state.description || undefined,
            columnPolicy: {
              defaultMode: state.columnPolicy.defaultMode,
              rules: state.columnPolicy.rules.map((r) => ({
                fieldName: r.fieldName,
                mode: r.mode,
                maskPattern: r.maskPattern ?? null,
              })),
            },
          },
        },
      })

      const payload = result.data?.createEndUserPermission
      if (payload?.error) {
        setSubmitError(payload.error.message ?? '创建权限点失败，请重试')
        return
      }

      // Success — call the provided callback or navigate back to the permissions list
      if (onSuccess) {
        onSuccess()
      } else {
        window.location.href = `/org/${orgName}/project/${projectSlug}/access-control?tab=permissions`
      }
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : '创建权限点时发生错误，请重试')
    } finally {
      setSubmitting(false)
    }
  }, [createPermissionMutation, orgName, projectSlug, onSuccess, state])

  return { state, updateField, goNext, goBack, reset, submit, submitting, submitError }
}
