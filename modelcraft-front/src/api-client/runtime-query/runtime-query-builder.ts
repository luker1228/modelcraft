import { gql, DocumentNode } from '@apollo/client'
import * as gqlBuilder from 'gql-query-builder'

/**
 * Runtime GraphQL Query Builder
 * Uses gql-query-builder library to generate dynamic GraphQL queries
 */

/**
 * Convert a model name to a valid GraphQL type name prefix by prepending "T".
 *
 * Background: GraphQL type names must start with a letter or underscore.
 * ModelCraft supports importing external database tables (createdVia: IMPORTED),
 * whose names may start with digits (e.g. "123orders"). Self-created models are
 * blocked from digit-leading names at creation time, but imported tables are not.
 * Using a raw digit-leading name like "123ordersWhereInput" causes gql() to throw
 * "Invalid number, expected digit but got: s".
 *
 * Fix: prepend "T" to every model name so all generated type names are always valid.
 * This must match the backend's gqlTypeName() helper in graphql_type_name.go.
 *   "User"      -> "TUser"      -> "TUserWhereInput"
 *   "123orders" -> "T123orders" -> "T123ordersWhereInput"
 */
function gqlTypeName(modelName: string): string {
  return 'T' + modelName
}

export interface FieldDefinition {
  name: string
  type: string
  format?: string
  schemaType?: string
  storageHint?: string
}

/**
 * Convert fields array to field names
 * @deprecated Use buildFieldSelections() instead
 */
function getFieldNames(fields: string[] | FieldDefinition[]): string[] {
  if (fields.length === 0) {
    return ['id']
  }

  return fields.map((field) =>
    typeof field === 'string' ? field : field.name
  )
}

/**
 * Check if a field is a RELATION type that requires sub-selections
 */
function isRelationField(field: FieldDefinition): boolean {
  return field.format === 'RELATION'
}

/**
 * Build field selections for gql-query-builder.
 * Scalar fields remain as strings: 'fieldName'
 * RELATION fields become objects with sub-selections: { fieldName: ['id', '_displayName'] }
 *
 * Protocol: relation fields only request `id` and `_displayName` (computed display name).
 */
export function buildFieldSelections(
  fields: string[] | FieldDefinition[] | (string | FieldDefinition)[]
): (string | Record<string, string[]>)[] {
  if (fields.length === 0) {
    return ['id']
  }

  return fields.map((field) => {
    if (typeof field === 'string') {
      return field
    }
    if (isRelationField(field)) {
      return { [field.name]: ['id', '_displayName'] }
    }
    return field.name
  })
}

/**
 * Build findMany query for a model
 * Response format: { timeCost, reqId, items: [...] }
 * Note: Runtime API uses generic 'findMany' operation name
 */
export function buildFindManyQuery(
  modelName: string,
  fields: string[] | FieldDefinition[] | (string | FieldDefinition)[]
): DocumentNode {
  const fieldSelection = buildFieldSelections(fields)

  const { query } = gqlBuilder.query({
    operation: 'findMany',
    variables: {
      where: { type: `${gqlTypeName(modelName)}WhereInput`, required: false },
      orderBy: { type: `[${gqlTypeName(modelName)}OrderByInput!]`, required: false },
      skip: { type: 'Int', required: false },
      take: { type: 'Int', required: false },
    },
    fields: [
      'timeCost',
      'reqId',
      {
        items: fieldSelection,
      },
    ],
  })

  return gql(query)
}

/**
 * Build findUnique query for a model
 * Response format: { reqId, timeCost, item: {...} }
 * Query format: findUnique(where: UniqueWhereInput!): FindUniqueResponse
 */
export function buildFindUniqueQuery(
  modelName: string,
  fields: string[] | FieldDefinition[]
): DocumentNode {
  const fieldSelection = buildFieldSelections(fields)

  const { query } = gqlBuilder.query({
    operation: 'findUnique',
    variables: {
      where: { type: `${gqlTypeName(modelName)}UniqueWhereInput`, required: true },
    },
    fields: [
      'reqId',
      'timeCost',
      {
        item: fieldSelection,
      },
    ],
  })

  return gql(query)
}

/**
 * Build findFirst query for a model
 * Query format: findFirst(where, orderBy, skip, take): Model
 */
export function buildFindFirstQuery(
  modelName: string,
  fields: string[] | FieldDefinition[]
): DocumentNode {
  const fieldSelection = buildFieldSelections(fields)

  const { query } = gqlBuilder.query({
    operation: 'findFirst',
    variables: {
      where: { type: `${gqlTypeName(modelName)}WhereInput`, required: false },
      orderBy: { type: `[${gqlTypeName(modelName)}OrderByInput!]`, required: false },
      skip: { type: 'Int', required: false },
      take: { type: 'Int', required: false },
    },
    fields: fieldSelection,
  })

  return gql(query)
}

/**
 * Build count query for a model
 * Query format: count(where): Int!
 */
export function buildCountQuery(modelName: string): DocumentNode {
  const { query } = gqlBuilder.query({
    operation: 'count',
    variables: {
      where: { type: `${gqlTypeName(modelName)}WhereInput`, required: false },
    },
  })

  return gql(query)
}

/**
 * Helper to extract field names from JSON Schema
 */
export function extractFieldsFromSchema(
  schema: { properties?: Record<string, unknown> } | null | undefined
): string[] {
  if (!schema?.properties) {
    return ['id']
  }

  const fields = Object.keys(schema.properties)

  // Always include 'id' if not present
  if (!fields.includes('id')) {
    fields.unshift('id')
  }

  return fields
}

/**
 * Build complete query operations object for a model
 */
export interface ModelQueryOperations {
  findUnique: DocumentNode
  findFirst: DocumentNode
  findMany: DocumentNode
  count: DocumentNode
}

export function buildModelQueryOperations(
  modelName: string,
  fields: string[] | FieldDefinition[]
): ModelQueryOperations {
  return {
    findUnique: buildFindUniqueQuery(modelName, fields),
    findFirst: buildFindFirstQuery(modelName, fields),
    findMany: buildFindManyQuery(modelName, fields),
    count: buildCountQuery(modelName),
  }
}

/**
 * Helper to capitalize model name (for Pascal case conversion)
 */
export function capitalizeModelName(name: string): string {
  return name.charAt(0).toUpperCase() + name.slice(1)
}

/**
 * Extract writable field names from JSON Schema properties.
 * Fields with `readOnly: true` are excluded.
 */
export function extractWritableFieldNamesFromSchema(
  schema: { properties?: Record<string, unknown> } | null | undefined
): string[] {
  if (!schema?.properties) {
    return []
  }

  return Object.entries(schema.properties)
    .filter(([, prop]) => (prop as Record<string, unknown>).readOnly !== true)
    .map(([name]) => name)
}

/**
 * Filter mutation input data by allowed field names.
 *
 * - Removes unknown fields that are not in current schema input
 * - Preserves explicit `null` values (used for clearing fields)
 * - Drops only `undefined` values
 */
export function sanitizeMutationInputData(
  data: Record<string, unknown> | null | undefined,
  allowedFieldNames: readonly string[]
): Record<string, unknown> {
  if (!data || typeof data !== 'object' || allowedFieldNames.length === 0) {
    return {}
  }

  const allowed = new Set(allowedFieldNames)

  return Object.fromEntries(
    Object.entries(data).filter(
      ([key, value]) => allowed.has(key) && value !== undefined
    )
  )
}

/**
 * Build create mutation for a model
 * Mutation format: create(data: CreateInput!): { id }
 * Note: Create mutation always returns only 'id' field
 */
export function buildCreateMutation(modelName: string): DocumentNode {
  const { query } = gqlBuilder.mutation({
    operation: 'create',
    variables: {
      data: { type: `${gqlTypeName(modelName)}CreateInput`, required: true },
    },
    fields: ['id'],
  })

  return gql(query)
}

/**
 * Build update mutation for a model
 * Mutation format: update(where: UniqueWhereInput!, data: UpdateInput!): { success }
 * Note: Update mutation always returns only 'success' field
 */
export function buildUpdateMutation(modelName: string): DocumentNode {
  const { query } = gqlBuilder.mutation({
    operation: 'update',
    variables: {
      where: { type: `${gqlTypeName(modelName)}UniqueWhereInput`, required: true },
      data: { type: `${gqlTypeName(modelName)}UpdateInput`, required: true },
    },
    fields: ['success'],
  })

  return gql(query)
}

/**
 * Build delete mutation for a model
 * Mutation format: delete(where: UniqueWhereInput!): { success }
 * Note: Delete mutation always returns only 'success' field
 */
export function buildDeleteMutation(modelName: string): DocumentNode {
  const { query } = gqlBuilder.mutation({
    operation: 'delete',
    variables: {
      where: { type: `${gqlTypeName(modelName)}UniqueWhereInput`, required: true },
    },
    fields: ['success'],
  })

  return gql(query)
}
