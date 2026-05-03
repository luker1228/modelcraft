import { gql } from '@apollo/client'

export const MODEL_DATABASE_CATALOG_END_USER = gql`
  query ModelDatabaseCatalogEndUser($input: ModelDatabaseCatalogInput) {
    modelDatabaseCatalog(input: $input) {
      data {
        databases {
          name
        }
        totalCount
        page
        pageSize
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ProjectNotFound {
          message
        }
        ... on Unauthorized {
          message
        }
      }
    }
  }
`

export const MODEL_CATALOG_END_USER = gql`
  query ModelCatalogEndUser($input: ModelCatalogInput!) {
    modelCatalog(input: $input) {
      data {
        models {
          id
          name
          title
          databaseName
        }
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ProjectNotFound {
          message
        }
        ... on Unauthorized {
          message
        }
      }
    }
  }
`
