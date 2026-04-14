import { gql } from '@apollo/client'

// 创建字段枚举关联
export const CREATE_FIELD_ENUM_RELATION = gql`
  mutation CreateFieldEnumRelation($input: CreateFieldEnumRelationInput!) {
    createFieldEnumRelation(input: $input) {
      relation {
        id
        sourceFieldName
        labelFieldName
        enumName
      }
      error {
        __typename
        ... on InvalidInput {
          message
          suggestion
        }
        ... on FieldEnumSourceConflict {
          message
          code
          suggestion
        }
      }
    }
  }
`
