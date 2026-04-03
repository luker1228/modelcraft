import { gql } from '@apollo/client'

export const GET_CLUSTER = gql`
  query GetCluster($projectSlug: String!) {
    databaseCluster(projectSlug: $projectSlug) {
      cluster {
        id
        title
        description
        status
        connectionInfo {
          host
          port
          username
          password
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ClusterNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

export const LIST_DATABASES = gql`
  query ListDatabases($input: ListDatabasesInput!) {
    listDatabases(input: $input) {
      edges {
        node {
          name
        }
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`
