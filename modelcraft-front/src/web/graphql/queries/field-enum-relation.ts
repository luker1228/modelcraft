import { gql } from '@apollo/client'

// 查询字段枚举关联关系
export const GET_FIELD_ENUM_RELATIONS = gql`
  query GetFieldEnumRelations($modelID: ID!) {
    fieldEnumRelations(modelID: $modelID) {
      id
      modelId
      sourceFieldName
      labelFieldName
      enumName
      createdAt
      updatedAt
    }
  }
`
