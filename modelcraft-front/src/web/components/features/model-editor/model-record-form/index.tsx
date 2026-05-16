'use client'

import React, { useMemo, useRef } from 'react'
import Form, { IChangeEvent } from '@rjsf/core'
import validator from '@rjsf/validator-ajv8'
import type { RJSFSchema, UiSchema } from '@rjsf/utils'
import { buildUiSchema } from './build-ui-schema'
import { filterJsonSchemaForForm } from './filter-json-schema-for-form'
import { EnumSelect, EnumSchemaSelect, RelationSelector, EndUserSelectorWidget } from './widgets'
import { OneToManyRelationManagerSection } from './OneToManyRelationManagerSection'
import { FieldTemplate, BaseInputTemplate, ObjectFieldTemplate } from './templates'
import { Button } from '@web/components/ui/button'
import { toast } from 'sonner'
import { useRecordAccessAdapter } from './access-adapter'

interface ModelRecordFormProps {
  jsonSchema: RJSFSchema
  initialData?: Record<string, unknown>
  /** Prefill values for new record creation (merged on top of defaults). Used by AI actions. */
  initialValues?: Record<string, unknown>
  onSubmit: (data: Record<string, unknown>) => Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string
  recordId?: string
  /** Controls which Apollo client EndUserSelectorWidget uses to fetch users.
   *  - 'design'   → Tenant client via /graphql/org/ (default)
   *  - 'end_user' → End-user client via /graphql/end-user/org/
   */
  workspaceMode?: 'design' | 'end_user'
}

const customWidgets = {
  EnumSelect,
  EnumSchemaSelect,
  RelationSelector,
  EndUserSelectorWidget,
}

const customTemplates = {
  FieldTemplate,
  BaseInputTemplate,
  ObjectFieldTemplate,
}

// Alias the Form ref type so we don't have to repeat the long generic signature
type RJSFFormRef = Form<Record<string, unknown>>

export function ModelRecordForm({
  jsonSchema,
  initialData,
  initialValues,
  onSubmit,
  onCancel,
  isSubmitting = false,
  orgName,
  projectSlug,
  clusterName,
  databaseName,
  modelId,
  recordId,
  workspaceMode = 'design',
}: ModelRecordFormProps) {
  console.log('[ModelRecordForm] MOUNTED', { orgName, projectSlug, workspaceMode, modelId })
  const formRef = useRef<RJSFFormRef>(null)
  const adapter = useRecordAccessAdapter()

  // Filter readOnly fields (primary keys, RELATION fields) out of the form schema
  const editableSchema = useMemo(
    () => filterJsonSchemaForForm(jsonSchema),
    [jsonSchema]
  )

  // Build uiSchema from x-mc.widget in JSON Schema, then enforce ui:order
  const uiSchema = useMemo<UiSchema>(() => {
    const widgetUiSchema = buildUiSchema(editableSchema)
    const orderedFieldNames = editableSchema.properties
      ? Object.keys(editableSchema.properties)
      : []

    if (orderedFieldNames.length === 0) {
      return widgetUiSchema
    }

    return {
      ...widgetUiSchema,
      'ui:order': orderedFieldNames,
    }
  }, [editableSchema])

  // Form context for widgets — 注入 createRuntimeClient 使 RelationSelector 通过 access adapter 访问数据
  const formContext = useMemo(() => {
    const ctx = {
      orgName,
      projectSlug,
      clusterName,
      databaseName,
      modelId,
      workspaceMode,
      createRuntimeClient: adapter.createRuntimeClient,
    }
    console.log('[ModelRecordForm] formContext built', {
      orgName,
      projectSlug,
      workspaceMode,
      modelId,
    })
    return ctx
  }, [orgName, projectSlug, clusterName, databaseName, modelId, workspaceMode, adapter.createRuntimeClient])

  // Handle form submission
  const handleSubmit = async (data: IChangeEvent<Record<string, unknown>>) => {
    if (!data.formData) return

    try {
      await onSubmit(data.formData)
    } catch (error) {
      const message = error instanceof Error ? error.message : '保存失败'
      toast.error(message)
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-auto p-4">
        <Form
          ref={formRef}
          schema={editableSchema}
          uiSchema={uiSchema}
          formData={initialValues && Object.keys(initialValues).length > 0 ? { ...initialValues, ...initialData } : initialData}
          validator={validator}
          widgets={customWidgets as never}
          templates={customTemplates as never}
          formContext={formContext as never}
          onSubmit={handleSubmit}
          disabled={isSubmitting}
          liveValidate={false}
          showErrorList={false}
        >
          {/* Empty children to suppress default submit button */}
          <></>
        </Form>

        <OneToManyRelationManagerSection
          jsonSchema={jsonSchema}
          initialData={initialData}
          recordId={recordId}
          orgName={orgName}
          projectSlug={projectSlug}
          modelId={modelId}
        />
      </div>
      <div className="flex justify-end gap-2 border-t p-4">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          取消
        </Button>
        <Button
          type="submit"
          disabled={isSubmitting}
          onClick={() => formRef.current?.submit()}
        >
          {isSubmitting ? '保存中...' : '保存'}
        </Button>
      </div>
    </div>
  )
}

export default ModelRecordForm
