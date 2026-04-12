import type { RJSFSchema } from '@rjsf/utils'

type SchemaProperty = Record<string, unknown>

/**
 * Determines whether a field should appear in create/edit forms.
 * Protocol: fields with readOnly: true are excluded from forms.
 */
export function shouldShowInForm(prop: SchemaProperty): boolean {
  return prop.readOnly !== true
}

/**
 * Format a relation value according to the `id + __label` protocol.
 *
 * Display format: `__label(id)`
 * If `__label` is empty string, display: `空(id)`
 */
function formatRelationDisplay(rel: Record<string, unknown>): string {
  const id = String(rel.id ?? '')
  const label = rel.__label
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
 *   - RELATION fields (type=object, x-relateFkId or x-belongsToFkId present)
 *     → display format: `__label(id)`, or `空(id)` if __label is empty
 *   - All other values → convert to string, truncated to 100 chars
 *
 * Note: both RelateFKID and BelongsToFKID produce FormatRelation fields in the
 * backend, which are always type=object in JSON Schema with value { id, __label, ... }.
 */
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''

  // RELATION fields: schema explicitly marks them (x-relateFkId / x-belongsToFkId)
  if (prop.type === 'object' && (prop['x-relateFkId'] || prop['x-belongsToFkId'])) {
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      return formatRelationDisplay(rel)
    }
    return ''
  }

  // Generic object fallback: any object value with id and __label fields is treated as a relation
  if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
    const rel = value as Record<string, unknown>
    if (rel.id !== undefined && '__label' in rel) {
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
      const oa = String(a.prop['x-displayOrder'] ?? '')
      const ob = String(b.prop['x-displayOrder'] ?? '')
      return oa.localeCompare(ob)
    })
}
