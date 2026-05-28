import { gql } from '@apollo/client'

// ── Fragment ────────────────────────────────────────────────────────────────────

export const MODEL_DATABASE_FRAGMENT = gql`
  fragment ModelDatabaseFields on ModelDatabase {
    id
    name
    title
    description
    mode
    createdAt
    updatedAt
  }
`

// ── Queries ─────────────────────────────────────────────────────────────────────

export const LIST_MODEL_DATABASES = gql`
  ${MODEL_DATABASE_FRAGMENT}
  query ListModelDatabases {
    modelDatabases {
      ...ModelDatabaseFields
    }
  }
`

export const LIST_CLUSTER_RAW_DATABASES = gql`
  query ListClusterRawDatabases {
    clusterRawDatabases {
      name
      isRegistered
    }
  }
`

// ── Mutations ───────────────────────────────────────────────────────────────────

export const REGISTER_MODEL_DATABASE = gql`
  ${MODEL_DATABASE_FRAGMENT}
  mutation RegisterModelDatabase($input: RegisterModelDatabaseInput!) {
    registerModelDatabase(input: $input) {
      ... on ModelDatabase {
        ...ModelDatabaseFields
      }
      ... on InvalidInput {
        message
      }
      ... on ResourceNotFound {
        message
        resourceType
      }
    }
  }
`

export const UPDATE_MODEL_DATABASE = gql`
  ${MODEL_DATABASE_FRAGMENT}
  mutation UpdateModelDatabase($id: ID!, $input: UpdateModelDatabaseInput!) {
    updateModelDatabase(id: $id, input: $input) {
      ...ModelDatabaseFields
    }
  }
`

export const UNREGISTER_MODEL_DATABASE = gql`
  mutation UnregisterModelDatabase($id: ID!) {
    unregisterModelDatabase(id: $id)
  }
`
