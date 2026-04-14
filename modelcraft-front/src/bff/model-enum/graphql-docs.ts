/**
 * GraphQL documents used internally by the model-enum BFF module.
 * These are duplicated here (rather than imported from @web/graphql) to respect
 * the architectural rule that BFF must not depend on the Web layer.
 */
import { gql } from '@apollo/client'

export const GET_ENUMS = gql`
  query GetEnums {
    enums {
      id
      projectSlug
      name
      displayName
      description
      isMultiSelect
      options {
        code
        label
        order
        description
      }
      createdAt
      updatedAt
    }
  }
`

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
