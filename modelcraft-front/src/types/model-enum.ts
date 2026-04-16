import type { ValidationConfigInput } from '@/generated/graphql'

export type ModelEnumErrorType =
  | 'InvalidInput'
  | 'FieldFormatImmutable'
  | 'Unknown'

export type ModelEnumErrorCode =
  | 'FIELD_FORMAT_IMMUTABLE'
  | 'UNKNOWN'

export interface ModelEnumDomainError {
  type: ModelEnumErrorType
  code?: ModelEnumErrorCode
  message: string
  suggestion?: string
}

export interface EnumSourceOption {
  fieldName: string
  title: string
  enumName: string
  occupied: boolean
}

export interface EnumRelationOption {
  id: string
  sourceFieldName: string
  enumName: string
  labelFieldName: string
}

export interface CreateEnumFieldFormValues {
  name: string
  title: string
  description?: string
  relateEnumName: string
}

export interface UpdateFieldMetaFormValues {
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}
