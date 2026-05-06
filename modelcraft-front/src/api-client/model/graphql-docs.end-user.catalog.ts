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
        ... on ResourceNotFound {
          message
        }
      }
    }
  }
`

export const MODEL_CATALOG_END_USER = gql`
  query ModelCatalogEndUser($input: ModelQueryInput!) {
    models(input: $input) {
      edges {
        node {
          id
          name
          title
          databaseName
        }
      }
      totalCount
    }
  }
`
