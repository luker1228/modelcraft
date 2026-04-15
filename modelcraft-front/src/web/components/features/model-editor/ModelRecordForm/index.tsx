'use client'

import React, { useMemo, useRef } from 'react'
import { useQuery } from '@apollo/client'
import Form, { IChangeEvent } from '@rjsf/core'
import validator from '@rjsf/validator-ajv8'
import type { RJSFSchema, UiSchema } from '@rjsf/utils'
import { useProjectScopedClient } from '@bff/apollo/public'
import { GET_LOGICAL_FOREIGN_KEYS, GET_MODELS } from '@web/graphql'
import type { Field } from '@/types/index'
import { buildUiSchema, buildRelationUiSchema } from './buildUiSchema'
import { filterJsonSchemaForForm } from './filterJsonSchemaForForm'
import { EnumSelect, EnumSchemaSelect, RelationPicker, RelationSelector } from './widgets'
import { FieldTemplate, BaseInputTemplate, ObjectFieldTemplate } from './templates'
import { Button } from '@web/components/ui/button'
import { Skeleton } from '@web/components/ui/skeleton'
import { toast } from 'sonner'

interface ModelRecordFormProps {
  fields: Field[]
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
}

interface LogicalForeignKey {
  id: string
  pairId?: string
  direction?: string
  modelId: string
  refModelId: string
  sourceFields: string[]
  targetFields: string[]
}

interface EnrichedFK extends LogicalForeignKey {
  refModelName: string
}

interface ModelNode {
  id: string
  name: string
}

interface GetModelsQueryData {
  models?: {
    edges?: Array<{
      node?: {
        id: string
        name: string
      } | null
    } | null>
  } | null
}

interface GetLogicalForeignKeysQueryData {
  logicalForeignKeys?: LogicalForeignKey[] | null
}

const customWidgets = {
  EnumSelect,
  EnumSchemaSelect,
  RelationPicker,
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
  fields,
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
}: ModelRecordFormProps) {
  const formRef = useRef<RJSFFormRef>(null)

  const projectClient = useProjectScopedClient(projectSlug)

  const projectScopedContext = useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])

  // Query logical foreign keys
  const {
    data: fkData,
    loading: fkLoading,
    error: fkError,
  } = useQuery<GetLogicalForeignKeysQueryData>(GET_LOGICAL_FOREIGN_KEYS, {
    client: projectClient,
    context: projectScopedContext,
    variables: { projectSlug, modelId },
    skip: !projectSlug || !modelId,
  })

  // Query all models to resolve refModelId → refModelName
  const {
    data: modelsData,
    loading: modelsLoading,
  } = useQuery<GetModelsQueryData>(GET_MODELS, {
    client: projectClient,
    context: projectScopedContext,
    variables: { input: { databaseName } },
    skip: !databaseName,
  })

  const loading = fkLoading || modelsLoading

  // Build models lookup map
  const modelsMap = useMemo<Map<string, ModelNode>>(() => {
    const map = new Map<string, ModelNode>()
    const edges = modelsData?.models?.edges ?? []
    for (const edge of edges) {
      if (!edge?.node?.id) {
        continue
      }

      map.set(edge.node.id, { id: edge.node.id, name: edge.node.name })
    }
    return map
  }, [modelsData])

  // Enrich FK data with refModelName
  const logicalForeignKeys = useMemo<EnrichedFK[]>(() => {
    const fks: LogicalForeignKey[] = fkData?.logicalForeignKeys ?? []
    return fks.map((fk) => ({
      ...fk,
      refModelName: modelsMap.get(fk.refModelId)?.name ?? fk.refModelId,
    }))
  }, [fkData, modelsMap])

  // Filter readOnly fields (primary keys, RELATION fields) out of the form schema
  const editableSchema = useMemo(
    () => filterJsonSchemaForForm(jsonSchema),
    [jsonSchema]
  )

  // Build uiSchema and enforce a stable render order from filtered schema properties.
  // Priority: field-type widgets (ENUM, DATE, …) → x-relation widgets → ui:order
  const uiSchema = useMemo<UiSchema>(() => {
    const fieldUiSchema = buildUiSchema(fields)
    const relationUiSchema = buildRelationUiSchema(editableSchema)
    const orderedFieldNames = editableSchema.properties
      ? Object.keys(editableSchema.properties)
      : []

    const merged = { ...fieldUiSchema, ...relationUiSchema }

    if (orderedFieldNames.length === 0) {
      return merged
    }

    return {
      ...merged,
      'ui:order': orderedFieldNames,
    }
  }, [fields, editableSchema])

  // Form context for widgets
  const formContext = useMemo(() => ({
    orgName,
    projectSlug,
    clusterName,
    databaseName,
    modelId,
    logicalForeignKeys,
  }), [orgName, projectSlug, clusterName, databaseName, modelId, logicalForeignKeys])

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

  // Loading state
  if (loading) {
    return (
      <div className="space-y-4 p-4">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    )
  }

  // FK query error - show form anyway, RelationPicker will show error for affected fields
  if (fkError) {
    // Intentionally silent: FK load failure is non-fatal.
    // RelationPicker widgets will surface per-field errors when rendered.
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
