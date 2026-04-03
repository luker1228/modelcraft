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
 * Renders a table cell value for a given field's schema property.
 *
 * Protocol:
 *   - RELATION fields (type=object, x-relateFkId or x-belongsToFkId present)
 *     → display the `name` attribute of the relation object, or fall back to `id`
 *   - All other values → convert to string, truncated to 100 chars
 *
 * Note: both RelateFKID and BelongsToFKID produce FormatRelation fields in the
 * backend, which are always type=object in JSON Schema with value { id, name, ... }.
 */
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''

  if (prop.type === 'object' && (prop['x-relateFkId'] || prop['x-belongsToFkId'])) {
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      return String(rel.name ?? rel.id ?? '')
    }
    return ''
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
