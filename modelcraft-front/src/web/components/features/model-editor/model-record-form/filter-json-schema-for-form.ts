import type { RJSFSchema } from '@rjsf/utils'
import { getXMC } from '@/types/xmc'

/**
 * Filters out readOnly fields from a JSON Schema, returning a schema that
 * contains only user-editable fields. Also removes filtered field names from
 * the `required` array.
 *
 * Protocol:
 *   readOnly: true  →  field is excluded from create/edit forms
 *   (Primary key fields and RELATION fields are marked readOnly by the backend)
 */
export function filterJsonSchemaForForm(schema: RJSFSchema): RJSFSchema {
  if (!schema.properties) return schema

  const editableEntries = Object.entries(schema.properties)
    .filter(([, prop]) => !(prop as Record<string, unknown>).readOnly)
    .map(([key, prop], index) => ({
      key,
      prop,
      index,
      displayOrder: getXMC(prop as Record<string, unknown>)?.displayOrder,
    }))

  editableEntries.sort((a, b) => {
    const aOrder = typeof a.displayOrder === 'string' ? a.displayOrder : null
    const bOrder = typeof b.displayOrder === 'string' ? b.displayOrder : null

    if (aOrder === null && bOrder === null) {
      return a.index - b.index
    }

    if (aOrder === null) {
      return 1
    }

    if (bOrder === null) {
      return -1
    }

    const orderComparison = aOrder.localeCompare(bOrder, undefined, {
      numeric: true,
    })

    if (orderComparison !== 0) {
      return orderComparison
    }

    return a.index - b.index
  })

  const sortedEditableEntries = editableEntries.map(({ key, prop }) => [
    key,
    prop,
  ] as const)

  const editableKeys = new Set(sortedEditableEntries.map(([key]) => key))

  return {
    ...schema,
    properties: Object.fromEntries(sortedEditableEntries),
    required: Array.isArray(schema.required)
      ? schema.required.filter((key) => editableKeys.has(key as string))
      : schema.required,
  }
}
