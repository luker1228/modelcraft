import {
  mockCreateEnumField,
  mockCreateEnumLabelField,
  mockCreateFieldEnumRelation,
  mockListFieldEnumRelations,
  mockQueryModelEnumContext,
  mockUpdateFieldMeta,
} from './mock-client'
import type {
  CreateEnumFieldCommand,
  CreateEnumLabelFieldCommand,
  CreateFieldEnumRelationCommand,
  FieldEnumRelationListResult,
  ModelEnumActionResult,
  ModelEnumContextQuery,
  ModelEnumContextResult,
  UpdateFieldMetaCommand,
} from './types'

export function queryModelEnumContext(query: ModelEnumContextQuery): Promise<ModelEnumContextResult> {
  return mockQueryModelEnumContext(query)
}

export function createEnumField(command: CreateEnumFieldCommand): Promise<ModelEnumActionResult> {
  return mockCreateEnumField(command)
}

export function createEnumLabelField(command: CreateEnumLabelFieldCommand): Promise<ModelEnumActionResult> {
  return mockCreateEnumLabelField(command)
}

export function updateFieldMeta(command: UpdateFieldMetaCommand): Promise<ModelEnumActionResult> {
  return mockUpdateFieldMeta(command)
}

export function listFieldEnumRelations(
  query: ModelEnumContextQuery,
): Promise<FieldEnumRelationListResult> {
  return mockListFieldEnumRelations(query)
}

export function createFieldEnumRelation(
  command: CreateFieldEnumRelationCommand,
): Promise<ModelEnumActionResult> {
  return mockCreateFieldEnumRelation(command)
}
