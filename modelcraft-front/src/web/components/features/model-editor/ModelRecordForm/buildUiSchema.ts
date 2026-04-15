import type { UiSchema, RJSFSchema } from '@rjsf/utils'
import type { Field } from '@/types/index'

/**
 * Build RJSF uiSchema from design-time Field definitions.
 *
 * Mapping rules (in priority order):
 *  - format ENUM/ENUM_ARRAY → custom widget "EnumSchemaSelect"
 *                             reads enum codes from jsonSchema.enum (set by backend)
 *  - format DATE       → ui:widget "date"
 *  - format DATETIME   → ui:widget "datetime-local"
 *  - format TIME       → ui:widget "time"
 *  - storageHint TEXT  → ui:widget "textarea"
 *
 * Note: isPrimary and RELATION fields are filtered out before reaching RJSF
 * via filterJsonSchemaForForm(), so no ui:widget mappings are needed for them.
 *
 * Connection context (orgName, projectSlug, etc.) is passed at runtime via
 * RJSF formContext and read directly inside each widget, so it is not needed
 * here when building the static uiSchema.
 */
export function buildUiSchema(fields: Field[]): UiSchema {
  const uiSchema: UiSchema = {}

  for (const field of fields) {
    if (field.format === 'ENUM' || field.format === 'ENUM_ARRAY') {
      uiSchema[field.name] = { 'ui:widget': 'EnumSchemaSelect' }
      continue
    }

    if (field.format === 'DATE') {
      uiSchema[field.name] = { 'ui:widget': 'date' }
      continue
    }

    if (field.format === 'DATETIME') {
      uiSchema[field.name] = { 'ui:widget': 'datetime-local' }
      continue
    }

    if (field.format === 'TIME') {
      uiSchema[field.name] = { 'ui:widget': 'time' }
      continue
    }

    if (field.storageHint === 'TEXT') {
      uiSchema[field.name] = { 'ui:widget': 'textarea' }
      continue
    }
  }

  return uiSchema
}

/**
 * Scan a (filtered) JSON Schema and return widget overrides for any property
 * that carries an `x-relation` extension.
 *
 * `x-relation` is injected by the backend at runtime into FK scalar fields:
 * ```json
 * "user_id": {
 *   "type": "string",
 *   "x-belongsToFkId": "fk-123",
 *   "x-relation": { "databaseName": "users_db", "modelName": "User" }
 * }
 * ```
 *
 * Such fields should be rendered as a `RelationSelector` combobox instead of a
 * plain text input.  The widget reads `x-relation` directly from `props.schema`
 * at render time, so no extra ui:options are required here.
 */
export function buildRelationUiSchema(jsonSchema: RJSFSchema): UiSchema {
  const uiSchema: UiSchema = {}

  if (!jsonSchema.properties) return uiSchema

  for (const [fieldName, prop] of Object.entries(jsonSchema.properties)) {
    const typedProp = prop as Record<string, unknown>
    if (typedProp['x-relation']) {
      uiSchema[fieldName] = { 'ui:widget': 'RelationSelector' }
    }
  }

  return uiSchema
}
