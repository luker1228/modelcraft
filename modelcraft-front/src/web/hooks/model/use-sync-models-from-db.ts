import { useMutation, useQuery } from '@apollo/client'
import {
  MODEL_SYNC_JOB_QUERY,
  SYNC_MODELS_FROM_DB_MUTATION,
} from '@/api-client/model'
import type {
  SyncModelsFromDbMutation,
  SyncModelsFromDbMutationVariables,
  ModelSyncJobQuery,
  ModelSyncJobQueryVariables,
} from '@/generated/graphql'

export function useSyncModelsFromDB() {
  return useMutation<SyncModelsFromDbMutation, SyncModelsFromDbMutationVariables>(
    SYNC_MODELS_FROM_DB_MUTATION
  )
}

export function useModelSyncJob(jobId: string | null) {
  return useQuery<ModelSyncJobQuery, ModelSyncJobQueryVariables>(MODEL_SYNC_JOB_QUERY, {
    variables: { jobId: jobId! },
    skip: !jobId,
    pollInterval: 2000,
  })
}
