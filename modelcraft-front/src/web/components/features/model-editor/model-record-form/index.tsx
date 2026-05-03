'use client'

import React, { useMemo, useRef } from 'react'
import Form, { IChangeEvent } from '@rjsf/core'
import validator from '@rjsf/validator-ajv8'
import type { RJSFSchema, UiSchema } from '@rjsf/utils'
import { buildUiSchema } from './build-ui-schema'
import { filterJsonSchemaForForm } from './filter-json-schema-for-form'
import { EnumSelect, EnumSchemaSelect, RelationSelector } from './widgets'
import { OneToManyRelationManagerSection } from './OneToManyRelationManagerSection'
import { FieldTemplate, BaseInputTemplate, ObjectFieldTemplate } from './templates'
import { Button } from '@web/components/ui/button'
import { toast } from 'sonner'

interface ModelRecordFormProps {
  jsonSchema: RJSFSchema
  initialData?: Record<string, unknown>
  onSubmit: (data: Record<string, unknown>) => Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string
  recordId?: string
  workspaceMode?: 'develop' | 'end_user'
}

const customWidgets = {
  EnumSelect,
  EnumSchemaSelect,
  RelationSelector,
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
  onSubmit,
  onCancel,
  isSubmitting = false,
  orgName,
  projectSlug,
  clusterName,
  databaseName,
  modelId,
  recordId,
  workspaceMode = 'develop',
}: ModelRecordFormProps) {
  const formRef = useRef<RJSFFormRef>(null)

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

  // Form context for widgets
  const formContext = useMemo(() => ({
    orgName,
    projectSlug,
    clusterName,
    databaseName,
    modelId,
  }), [orgName, projectSlug, clusterName, databaseName, modelId])

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
          formData={initialData}
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
          workspaceMode={workspaceMode}
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
