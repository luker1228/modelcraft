import type { ValidationConfigInput } from '@/generated/graphql'
import type { EnumSourceOption, ModelEnumDomainError } from '@/types'

export interface ModelEnumContextQuery {
  orgName: string
  projectSlug: string
  modelId: string
}

export interface CreateEnumFieldCommand {
  orgName: string
  projectSlug: string
  modelId: string
  name: string
  title: string
  description?: string
  relateEnumName: string
}

export interface UpdateFieldMetaCommand {
  orgName: string
  projectSlug: string
  modelId: string
  fieldName: string
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}

export interface ModelEnumContextResult {
  enumSources: EnumSourceOption[]
  relations: EnumRelationOption[]
  error: ModelEnumDomainError | null
}

export interface ModelEnumActionResult {
  success: boolean
  error: ModelEnumDomainError | null
}
