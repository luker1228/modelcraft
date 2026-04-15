import type { RJSFSchema } from '@rjsf/utils'
import { getXMC } from '@/types/xmc'

type SchemaProperty = Record<string, unknown>

/**
 * Determines whether a field should appear in create/edit forms.
 * Protocol: fields with readOnly: true are excluded from forms.
 */
export function shouldShowInForm(prop: SchemaProperty): boolean {
  return prop.readOnly !== true
}

/**
 * Format a relation value according to the `id + _label` protocol.
 *
 * Display format: `_label(id)`
 * If `_label` is empty string, display: `空(id)`
 */
function formatRelationDisplay(rel: Record<string, unknown>): string {
  const id = String(rel.id ?? '')
  const label = rel._label
  const labelStr = typeof label === 'string' ? label : ''

  if (!id) return ''

  if (labelStr === '') {
    return `空(${id})`
  }
  return `${labelStr}(${id})`
}

/**
 * Renders a table cell value for a given field's schema property.
 *
 * Protocol:
 *   - RELATION fields (type=object, x-mc.relateFkId or x-mc.belongsToFkId present)
 *     → display format: `_label(id)`, or `空(id)` if _label is empty
 *   - All other values → convert to string, truncated to 100 chars
 *
 * Note: both RelateFKID and BelongsToFKID produce FormatRelation fields in the
 * backend, which are always type=object in JSON Schema with value { id, _label, ... }.
 */
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''

  // RELATION fields: schema explicitly marks them (x-mc.relateFkId / x-mc.belongsToFkId)
  const xmc = getXMC(prop)
  if (prop.type === 'object' && (xmc?.relateFkId ?? xmc?.belongsToFkId)) {
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      return formatRelationDisplay(rel)
    }
    return ''
  }

  // Generic object fallback: any object value with id and _label fields is treated as a relation
  if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
    const rel = value as Record<string, unknown>
    if (rel.id !== undefined && '_label' in rel) {
      return formatRelationDisplay(rel)
    }
    // For other objects, render as JSON
    try {
      return JSON.stringify(value).slice(0, 100)
    } catch {
      return '[object]'
    }
  }

  return String(value).slice(0, 100)
}

/**
 * Returns an ordered list of field protocol entries from the JSON Schema,
 * sorted by x-displayOrder using locale-aware string comparison.
 */
export function getFieldProtocols(
  schema: RJSFSchema
): Array<{ name: string; prop: SchemaProperty }> {
  if (!schema.properties) return []

  return Object.entries(schema.properties)
    .map(([name, prop]) => ({ name, prop: prop as SchemaProperty }))
    .sort((a, b) => {
      const oa = String(getXMC(a.prop)?.displayOrder ?? '')
      const ob = String(getXMC(b.prop)?.displayOrder ?? '')
      return oa.localeCompare(ob, undefined, { numeric: true })
    })
}
