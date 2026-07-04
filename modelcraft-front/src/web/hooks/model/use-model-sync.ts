import { useMutation } from '@apollo/client'
import {
  START_MODEL_SYNC_MUTATION,
  MODEL_SYNC_JOBS_QUERY,
} from '@/api-client/model'
import type {
  StartModelSyncMutation,
  StartModelSyncMutationVariables,
  ModelSyncJobsQuery,
  ModelSyncJobsQueryVariables,
  ModelSyncJob,
} from '@/generated/graphql'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { useQuery } from '@apollo/client'

export interface SyncTarget {
  databaseId: string
  tableNames?: string[]
}

export interface ModelSyncJobRef {
  databaseId: string
  jobId: string
}

export function useStartModelSync(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)
  const [mutate, { loading, error }] = useMutation<
    StartModelSyncMutation,
    StartModelSyncMutationVariables
  >(START_MODEL_SYNC_MUTATION, { client })

  const startSync = async (
    targets: SyncTarget[]
  ): Promise<{ batchId: string; jobs: ModelSyncJobRef[] } | null> => {
    const result = await mutate({ variables: { targets } })
    const payload = result.data?.startModelSync
    if (!payload) return null
    return {
      batchId: payload.batchId,
      jobs: payload.jobs.map((j) => ({ databaseId: j.databaseId, jobId: j.jobId })),
    }
  }

  return { startSync, loading, error }
}

export function useFetchModelSyncJobs(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)

  const fetchByBatch = async (batchId: string): Promise<ModelSyncJob[]> => {
    const result = await client.query<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>({
      query: MODEL_SYNC_JOBS_QUERY,
      variables: { batchId },
      fetchPolicy: 'network-only',
    })
    return result.data.modelSyncJobs ?? []
  }

  const fetchByJobIds = async (jobIds: string[]): Promise<ModelSyncJob[]> => {
    const result = await client.query<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>({
      query: MODEL_SYNC_JOBS_QUERY,
      variables: { jobIds },
      fetchPolicy: 'network-only',
    })
    return result.data.modelSyncJobs ?? []
  }

  return { fetchByBatch, fetchByJobIds }
}
