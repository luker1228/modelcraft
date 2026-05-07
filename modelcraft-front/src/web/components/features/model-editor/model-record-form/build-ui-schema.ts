import type { UiSchema, RJSFSchema } from '@rjsf/utils'
import { getXMC, type XMCWidget } from '@/types/xmc'

export type WorkspaceMode = 'design' | 'end_user'

/**
 * Widget name mapping for non-end-user-ref widgets.
 */
const WIDGET_MAP: Partial<Record<XMCWidget, string>> = {
  'enum-select': 'EnumSelect',
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
 * RJSF widget string.
 *
 * For `end-user-ref` fields:
 * - `end_user` mode: hidden (auto-injected by backend from JWT)
 * - `design` mode: EndUserSelectorWidget (admin selects an EndUser)
 */
export function buildUiSchema(
  jsonSchema: RJSFSchema,
  workspaceMode: WorkspaceMode = 'design',
): UiSchema {
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
