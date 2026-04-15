import type { FieldDefinition } from '@bff/cms/public'
import type { ModelRecordTableFieldInfo } from './ModelRecordTable'

export interface ModelField {
  name: string
  title?: string | null
  format?: string | null
  schemaType?: string | null
  storageHint?: string | null
  isPrimary?: boolean | null
  isDeprecated?: boolean | null
  type?: string | null
}

function toOptionalString(value: string | null | undefined): string | undefined {
  return typeof value === 'string' ? value : undefined
}

export function mapModelFieldsToRuntimeFields(modelFields: readonly ModelField[]): FieldDefinition[] {
  return modelFields.map((field) => ({
    name: field.name,
    type: field.schemaType ?? field.type ?? 'string',
    format: toOptionalString(field.format),
    schemaType: toOptionalString(field.schemaType),
    storageHint: toOptionalString(field.storageHint),
  }))
}

export function mapModelFieldsToTableFieldInfos(
  modelFields: readonly ModelField[]
): ModelRecordTableFieldInfo[] {
  return modelFields.map((field) => ({
    name: field.name,
    title: typeof field.title === 'string' ? field.title : null,
    isPrimary: field.isPrimary === true,
    isDeprecated: field.isDeprecated === true,
    storageHint: typeof field.storageHint === 'string' ? field.storageHint : null,
    schemaType: typeof field.schemaType === 'string' ? field.schemaType : null,
  }))
}

export function buildEditFormData(
  modelFields: readonly ModelField[],
  item: Readonly<Record<string, unknown>>
): Record<string, unknown> {
  return modelFields.reduce<Record<string, unknown>>((formData, field) => {
    if (field.isPrimary) {
      return formData
    }

    formData[field.name] = item[field.name] ?? ''
    return formData
  }, {})
}
