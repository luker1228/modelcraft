import { gql } from '@apollo/client'

// ── Fragment ────────────────────────────────────────────────────────────────────

export const MODEL_DATABASE_FRAGMENT = gql`
  fragment ModelDatabaseFields on ModelDatabase {
    id
    name
    title
    description
    mode
    latestSyncJobId
    createdAt
    updatedAt
  }
`

export const MODEL_DATABASE_SYNC_JOB_FRAGMENT = gql`
  fragment ModelDatabaseSyncJobFields on ModelDatabaseSyncJob {
    id
    databaseId
    status
    totalTables
    processedTables
    createdModels
    syncedModels
    failedCount
    failedTables {
      tableName
      message
    }
    startedAt
    finishedAt
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

export const GET_MODEL_DATABASE_SYNC_JOB = gql`
  ${MODEL_DATABASE_SYNC_JOB_FRAGMENT}
  query GetModelDatabaseSyncJob($jobId: ID!) {
    modelDatabaseSyncJob(jobId: $jobId) {
      ...ModelDatabaseSyncJobFields
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

export const BATCH_REGISTER_MODEL_DATABASES = gql`
  ${MODEL_DATABASE_FRAGMENT}
  mutation BatchRegisterModelDatabases($input: BatchRegisterModelDatabaseInput!) {
    batchRegisterModelDatabases(input: $input) {
      succeeded {
        ...ModelDatabaseFields
      }
      failed {
        name
        message
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

export const START_MODEL_DATABASE_SYNC = gql`
  ${MODEL_DATABASE_SYNC_JOB_FRAGMENT}
  mutation StartModelDatabaseSync($databaseId: ID!) {
    startModelDatabaseSync(databaseId: $databaseId) {
      job {
        ...ModelDatabaseSyncJobFields
      }
    }
  }
`
