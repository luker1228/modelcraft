import type { UiSchema, RJSFSchema } from '@rjsf/utils'
import { getXMC, type XMCWidget } from '@/types/xmc'

/**
 * Widget name mapping for non-end-user-ref widgets.
 */
const WIDGET_MAP: Partial<Record<XMCWidget, string>> = {
  'enum-select': 'EnumSelect',
  'date': 'date',
  'datetime-local': 'datetime',
  'time': 'time',
  'textarea': 'textarea',
  'relation-selector': 'RelationSelector',
  'relation-multi-readonly': 'RelationMultiReadonly',
}

/**
 * Build RJSF uiSchema directly from the (filtered) JSON Schema.
 *
 * Reads `x-mc.widget` on each property and maps it to the appropriate
 * RJSF widget string.
 *
 * For `end-user-ref` fields: always renders EndUserSelectorWidget.
 * The widget picks the correct token (end-user or admin) by availability.
 */
export function buildUiSchema(jsonSchema: RJSFSchema): UiSchema {
  const uiSchema: UiSchema = {}

  if (!jsonSchema.properties) return uiSchema

  for (const [fieldName, prop] of Object.entries(jsonSchema.properties)) {
    const xmc = getXMC(prop as Record<string, unknown>)
    const widget = xmc?.widget

    if (!widget) continue

    if (widget === 'end-user-ref') {
      uiSchema[fieldName] = { 'ui:widget': 'EndUserSelectorWidget' }
      continue
    }

    if (WIDGET_MAP[widget]) {
      uiSchema[fieldName] = { 'ui:widget': WIDGET_MAP[widget] }
    }
  }

  return uiSchema
}
