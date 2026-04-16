import type { UiSchema, RJSFSchema } from '@rjsf/utils'
import { getXMC, type XMCWidget } from '@/types/xmc'

/**
 * Widget name mapping: x-mc.widget value → RJSF widget string.
 */
const WIDGET_MAP: Record<XMCWidget, string> = {
  'enum-select': 'EnumSchemaSelect',
  'date': 'date',
  'datetime-local': 'datetime-local',
  'time': 'time',
  'textarea': 'textarea',
  'relation-selector': 'RelationSelector',
  'relation-multi-readonly': 'RelationMultiReadonly',
}

/**
 * Build RJSF uiSchema directly from the (filtered) JSON Schema.
 *
 * Reads `x-mc.widget` on each property and maps it to the appropriate
 * RJSF widget string via WIDGET_MAP.
 *
 * Connection context (orgName, projectSlug, etc.) is passed at runtime via
 * RJSF formContext and read directly inside each widget, so it is not needed
 * here when building the static uiSchema.
 */
export function buildUiSchema(jsonSchema: RJSFSchema): UiSchema {
  const uiSchema: UiSchema = {}

  if (!jsonSchema.properties) return uiSchema

  for (const [fieldName, prop] of Object.entries(jsonSchema.properties)) {
    const xmc = getXMC(prop as Record<string, unknown>)
    const widget = xmc?.widget

    if (widget && WIDGET_MAP[widget]) {
      uiSchema[fieldName] = { 'ui:widget': WIDGET_MAP[widget] }
    }
  }

  return uiSchema
}
