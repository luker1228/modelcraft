'use client'

import React, { useMemo } from 'react'
import { useQuery, useMutation, gql, ApolloClient } from '@apollo/client'
import { createForm } from '@formily/core'
import { createSchemaField } from '@formily/react'
import {
  Form,
  FormItem,
  Input,
  Select,
  DatePicker,
  Switch,
  NumberPicker,
  Submit,
} from '@formily/antd-v5'
import { useProjectScopedClient, createDevelopModelRuntimeClient } from '@api-client/apollo/public'
import {
  transformToFormilySchema,
  parseAndTransformSchema,
  type FormilySchema,
  type JSONSchema,
} from '@/shared/cms/schema-transformer'
import {
  buildFindUniqueQuery,
  buildCreateMutation,
  buildUpdateMutation,
  extractFieldsFromSchema,
  extractWritableFieldNamesFromSchema,
  sanitizeMutationInputData,
} from '@api-client/cms/public'

// GraphQL query to fetch Model JSON Schema from Design-Time API
const MODEL_JSON_SCHEMA_QUERY = gql`
  query ModelJsonSchema($projectId: ID!, $id: ID!) {
    modelJsonSchema(projectId: $projectId, id: $id) {
      modelId
      modelName
      schema
    }
  }
`

// Create SchemaField with Ant Design V5 components
const SchemaField = createSchemaField({
  components: {
    FormItem,
    Input,
    Select,
    DatePicker,
    Switch,
    NumberPicker,
  },
})

export interface FormRendererProps {
  orgName: string
  projectSlug?: string  // Changed from projectId to projectSlug?
  modelId: string
  modelName: string
  databaseName: string
  contentId?: string
  onSubmit?: (data: Record<string, unknown>) => void
  onCancel?: () => void
}

interface ModelJsonSchemaQueryData {
  modelJsonSchema?: {
    modelId: string
    modelName: string
    schema: string
  } | null
}

interface FindUniqueQueryData {
  findUnique?: {
    item?: Record<string, unknown>
  } | null
}

/**
 * FormRenderer Component
 * Renders a dynamic form based on Model JSON Schema
 *
 * Data Flow:
 * 1. Fetch JSON Schema from Design-Time API (modelJsonSchema query)
 * 2. Transform JSON Schema to Formily Schema
 * 3. Fetch existing content from Runtime API (if editing)
 * 4. Render form with Formily
 * 5. Submit to Runtime API (create or update mutation)
 */
export function FormRenderer({
  orgName,
  projectSlug,
  modelId,
  modelName,
  databaseName,
  contentId,
  onSubmit,
  onCancel,
}: FormRendererProps) {
  const projectClient = useProjectScopedClient(projectSlug)

  // Create Runtime Client with correct endpoint: /graphql/org/:orgName/project/:projectSlug/db/:databaseName/model/:modelName
  const runtimeClient = useMemo(() => {
    return createDevelopModelRuntimeClient(orgName, projectSlug || 'default', databaseName, modelName)
  }, [orgName, projectSlug, databaseName, modelName]) as ApolloClient<object>

  // 1. Fetch JSON Schema from Design-Time API
  const { data: schemaData, loading: schemaLoading, error: schemaError } = useQuery<ModelJsonSchemaQueryData>(
    MODEL_JSON_SCHEMA_QUERY,
    {
      client: projectClient,
      variables: { projectId: projectSlug, id: modelId },  // Backend still uses projectId parameter name
    }
  )

  // Extract schema and parse
  const jsonSchema = useMemo<Record<string, unknown> | null>(() => {
    if (!schemaData?.modelJsonSchema?.schema) return null
    try {
      return JSON.parse(schemaData.modelJsonSchema.schema) as Record<string, unknown>
    } catch (error) {
      console.error('Failed to parse JSON Schema:', error)
      return null
    }
  }, [schemaData])

  // 2. Transform to Formily Schema
  const formilySchema: FormilySchema | null = useMemo(() => {
    if (!jsonSchema) return null
    return transformToFormilySchema(jsonSchema as unknown as JSONSchema)
  }, [jsonSchema])

  // Extract fields for Runtime queries
  const fields = useMemo(() => {
    if (!jsonSchema) return ['id']
    return extractFieldsFromSchema(jsonSchema as { properties?: Record<string, unknown> })
  }, [jsonSchema])

  const writableFieldNames = useMemo(
    () => extractWritableFieldNamesFromSchema(jsonSchema as { properties?: Record<string, unknown> } | null | undefined),
    [jsonSchema]
  )

  // Build Runtime GraphQL queries dynamically
  const findUniqueQuery = useMemo(
    () => buildFindUniqueQuery(modelName, fields),
    [modelName, fields]
  )

  const createMutation = useMemo(
    () => buildCreateMutation(modelName),
    [modelName]
  )

  const updateMutation = useMemo(
    () => buildUpdateMutation(modelName),
    [modelName]
  )

  // 3. Fetch existing content from Runtime API (if editing)
  // Response format: { findUnique: { reqId, timeCost, item: {...} } }
  const {
    data: contentData,
    loading: contentLoading,
    error: contentError,
  } = useQuery<FindUniqueQueryData>(findUniqueQuery, {
    client: runtimeClient,
    variables: { where: { id: contentId } },
    skip: !contentId,
  })

  // Runtime mutations
  const [createContent, { loading: createLoading }] = useMutation(
    createMutation,
    {
      client: runtimeClient,
    }
  )

  const [updateContent, { loading: updateLoading }] = useMutation(
    updateMutation,
    {
      client: runtimeClient,
    }
  )

  // 4. Create Formily form instance
  const form = useMemo(() => {
    // Extract item from findUnique response: { findUnique: { reqId, timeCost, item: {...} } }
    const initialValues: Record<string, unknown> = contentId
      ? (contentData?.findUnique?.item ?? {})
      : {}

    return createForm({
      initialValues,
      validateFirst: true,
    })
  }, [contentId, contentData])

  // Update form values when content data changes
  React.useEffect(() => {
    if (contentId && contentData) {
      // Extract item from findUnique response
      const values = contentData?.findUnique?.item as Record<string, unknown> | undefined
      if (values) {
        form.setInitialValues(values)
        form.reset()
      }
    }
  }, [contentData, contentId, form])

  // 5. Handle form submission
  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      const sanitizedValues = sanitizeMutationInputData(values, writableFieldNames)

      if (contentId) {
        // Update existing content
        await updateContent({
          variables: {
            where: { id: contentId },
            data: sanitizedValues,
          },
        })
      } else {
        // Create new content
        await createContent({
          variables: {
            data: sanitizedValues,
          },
        })
      }

      onSubmit?.(sanitizedValues)
    } catch (error) {
      console.error('Failed to submit form:', error)
      // Error handling can be improved with toast notifications
    }
  }

  // Loading states
  if (schemaLoading) {
    return <div className="p-4">Loading schema...</div>
  }

  if (contentId && contentLoading) {
    return <div className="p-4">Loading content...</div>
  }

  // Error states
  if (schemaError) {
    return (
      <div className="p-4 text-red-500">
        Failed to load schema: {schemaError.message}
      </div>
    )
  }

  if (contentId && contentError) {
    return (
      <div className="p-4 text-red-500">
        Failed to load content: {contentError.message}
      </div>
    )
  }

  if (!formilySchema) {
    return <div className="p-4">Invalid schema format</div>
  }

  // Render form
  return (
    <div className="mx-auto max-w-4xl p-6">
      <Form
        form={form}
        onAutoSubmit={handleSubmit}
        layout="vertical"
      >
        {/* Formily SchemaField expects its internal Schema type which is structurally
            incompatible with our FormilySchema. Bridging via unknown is necessary at
            the boundary between our typed schema transformer and Formily's runtime. */}
        <SchemaField schema={formilySchema as unknown as never} />

        <div className="mt-6 flex gap-4">
          <Submit
            loading={createLoading || updateLoading}
            size="large"
          >
            {contentId ? 'Update' : 'Create'}
          </Submit>

          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="rounded border px-4 py-2 hover:bg-gray-100"
            >
              Cancel
            </button>
          )}
        </div>
      </Form>
    </div>
  )
}
