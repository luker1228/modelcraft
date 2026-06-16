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
import { useProjectScopedClient } from '@api-client/apollo/develop-client'

export function useSyncModelsFromDB(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)
  return useMutation<SyncModelsFromDbMutation, SyncModelsFromDbMutationVariables>(
    SYNC_MODELS_FROM_DB_MUTATION,
    { client }
  )
}

export function useModelSyncJob(
  jobId: string | null,
  projectSlug: string | null | undefined
) {
  const client = useProjectScopedClient(projectSlug)
  return useQuery<ModelSyncJobQuery, ModelSyncJobQueryVariables>(MODEL_SYNC_JOB_QUERY, {
    client,
    variables: { jobId: jobId! },
    skip: !jobId,
    pollInterval: 2000,
  })
}
