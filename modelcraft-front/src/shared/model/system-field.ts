export const SYSTEM_LABEL_SUFFIX_REGEX = /(_label|_labels)$/

export interface FieldLike {
  name: string
  format?: string | null
}

export function hasSystemLabelSuffix(fieldName: string): boolean {
  return SYSTEM_LABEL_SUFFIX_REGEX.test(fieldName.trim())
}

export function getEnumDisplayFieldName(field: FieldLike): string | null {
  if (!field.name) {
    return null
  }

  if (field.format === 'ENUM') {
    return `${field.name}_label`
  }

  if (field.format === 'ENUM_ARRAY') {
    return `${field.name}_labels`
  }

  return null
}

export function findEnumSourceFieldForSystemLabel(
  fieldName: string,
  fields: FieldLike[],
): FieldLike | null {
  if (fieldName.endsWith('_labels')) {
    const sourceName = fieldName.slice(0, -'_labels'.length)
    return fields.find((field) => field.name === sourceName && field.format === 'ENUM_ARRAY') ?? null
  }

  if (fieldName.endsWith('_label')) {
    const sourceName = fieldName.slice(0, -'_label'.length)
    return fields.find((field) => field.name === sourceName && field.format === 'ENUM') ?? null
  }

  return null
}

export function isSystemGeneratedLabelField(field: FieldLike, fields: FieldLike[]): boolean {
  if (field.format === 'ENUM_LABEL') {
    return true
  }

  return findEnumSourceFieldForSystemLabel(field.name, fields) !== null
}

