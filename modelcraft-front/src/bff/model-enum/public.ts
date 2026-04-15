import type { ApolloClient, ApolloError } from '@apollo/client'
import {
  mockCreateEnumField,
  mockCreateEnumLabelField,
  mockCreateFieldEnumRelation,
  mockListFieldEnumRelations,
  mockUpdateFieldMeta,
} from './mock-client'
import { GET_FIELD_ENUM_RELATIONS, GET_MODEL_ENUM_SOURCE_FIELDS } from './graphql-docs'
import type { GetFieldEnumRelationsQuery, GetModelQuery } from '@/generated/graphql'
import type { EnumRelationOption, EnumSourceOption, ModelEnumDomainError } from '@/types'
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

export async function queryModelEnumContext(
  query: ModelEnumContextQuery,
  client: ApolloClient<object>,
): Promise<ModelEnumContextResult> {
  try {
    const [modelResult, relationsResult] = await Promise.all([
      client.query<GetModelQuery>({
        query: GET_MODEL_ENUM_SOURCE_FIELDS,
        variables: { id: query.modelId, withActualSchema: false },
        fetchPolicy: 'network-only',
      }),
      client.query<GetFieldEnumRelationsQuery>({
        query: GET_FIELD_ENUM_RELATIONS,
        variables: { modelID: query.modelId },
        fetchPolicy: 'network-only',
      }),
    ])

    // 检查 GraphQL errors（errorPolicy: 'all' 不会 throw，但 errors 会填充到结果中）
    const gqlErrors = [...(modelResult.errors ?? []), ...(relationsResult.errors ?? [])]
    if (gqlErrors.length > 0) {
      const error: ModelEnumDomainError = {
        type: 'Unknown',
        code: 'UNKNOWN',
        message: gqlErrors.map((e) => e.message).join('；'),
      }
      return { enumSources: [], relations: [], error }
    }

    const relations: EnumRelationOption[] = (relationsResult.data?.fieldEnumRelations ?? []).map((r) => ({
      id: r.id,
      sourceFieldName: r.sourceFieldName,
      enumName: r.enumName,
      labelFieldName: r.labelFieldName,
    }))
    const occupiedSourceNames = new Set(relations.map((relation) => relation.sourceFieldName))

    const enumSources: EnumSourceOption[] = (modelResult.data?.model.model?.fields ?? [])
      .filter((field) => field.format === 'ENUM' && Boolean(field.enum?.name))
      .map((field) => ({
        fieldName: field.name,
        title: field.title,
        enumName: field.enum?.name ?? '',
        occupied: occupiedSourceNames.has(field.name),
      }))

    return { enumSources, relations, error: null }
  } catch (err) {
    // 处理网络错误、CORS 错误等 ApolloError，避免向调用方 throw
    const apolloError = err as ApolloError
    const message = apolloError?.message ?? '网络请求失败，请检查网络连接后重试。'
    const error: ModelEnumDomainError = {
      type: 'Unknown',
      code: 'UNKNOWN',
      message,
    }
    return { enumSources: [], relations: [], error }
  }
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
