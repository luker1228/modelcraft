import { useMutation, useQuery } from '@apollo/client'
import {
  MODEL_SYNC_JOBS_QUERY,
  SYNC_MODELS_FROM_DB_MUTATION,
} from '@/api-client/model'
import type {
  SyncModelsFromDbMutation,
  SyncModelsFromDbMutationVariables,
  ModelSyncJobsQuery,
  ModelSyncJobsQueryVariables,
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
  return useQuery<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>(MODEL_SYNC_JOBS_QUERY, {
    client,
    variables: { jobIds: jobId ? [jobId] : [] },
    skip: !jobId,
    pollInterval: 2000,
  })
}
