import { gql } from '@apollo/client'

export const GET_MODEL_RECORD_WORKSPACE_END_USER = gql`
  query GetModelRecordWorkspaceEndUser($id: ID!) {
    model(id: $id) {
      model {
        id
        name
        title
        description
        databaseName
        createdVia
        jsonSchema
        fields {
          name
          isDeprecated
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`
