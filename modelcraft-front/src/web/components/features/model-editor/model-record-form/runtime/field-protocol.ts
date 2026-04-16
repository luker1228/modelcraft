import type { RJSFSchema } from '@rjsf/utils'
import { getXMC } from '@/types/xmc'

type SchemaProperty = Record<string, unknown>

function toReadableText(value: unknown): string {
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') {
    const record = value as Record<string, unknown>
    const nested = record._displayName ?? record.displayName ?? record.title ?? record.name ?? record.id
    if (nested !== undefined && nested !== value) {
      return toReadableText(nested)
    }
    try {
      return JSON.stringify(value)
    } catch {
      return '[object]'
    }
  }
  return String(value)
}

/**
 * Determines whether a field should appear in create/edit forms.
 * Protocol: fields with readOnly: true are excluded from forms.
 */
export function shouldShowInForm(prop: SchemaProperty): boolean {
  return prop.readOnly !== true
}

/**
 * Format a relation value according to the `id + _displayName` protocol.
 *
 * Display format: `_displayName(id)`
 * If `_displayName` is empty string, display: `空(id)`
 */
function formatRelationDisplay(rel: Record<string, unknown>): string {
  const id = toReadableText(rel.id)
  const labelStr = toReadableText(rel._displayName)

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
 *   - RELATION fields (type=object, x-mc.relation.relateFkId or x-mc.relation.belongsToFkId present)
 *     → display format: `_displayName(id)`, or `空(id)` if _displayName is empty
 *   - All other values → convert to string, truncated to 100 chars
 *
 * Note: both relateFkId and belongsToFkId produce FormatRelation fields in the
 * backend, which are always type=object in JSON Schema with value { id, _displayName, ... }.
 */
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''

  // RELATION fields: schema explicitly marks them (x-mc.relation.relateFkId / belongsToFkId)
  const xmc = getXMC(prop)
  if ((xmc?.relation?.relateFkId ?? xmc?.relation?.belongsToFkId)) {
    // one-to-many / runtime array payload
    if (Array.isArray(value)) {
      const items = value
        .filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)
        .map((item) => {
          if (item.id !== undefined) {
            return formatRelationDisplay(item)
          }
          return toReadableText(item)
        })
        .filter((text) => text !== '')

      return items.join(', ').slice(0, 100)
    }

    // many-to-one / runtime object payload
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      return formatRelationDisplay(rel)
    }
    return ''
  }

  // Generic array fallback
  if (Array.isArray(value)) {
    return value.map((item) => toReadableText(item)).join(', ').slice(0, 100)
  }

  // Generic object fallback: any object value with id and _displayName fields is treated as a relation
  if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
    const rel = value as Record<string, unknown>
    if (rel.id !== undefined && '_displayName' in rel) {
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
