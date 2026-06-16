import { useEffect, useRef, useCallback } from 'react'
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
  ModelSyncJobStatus,
} from '@/generated/graphql'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'

export function useSyncModelsFromDB(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)
  return useMutation<SyncModelsFromDbMutation, SyncModelsFromDbMutationVariables>(
    SYNC_MODELS_FROM_DB_MUTATION,
    { client }
  )
}

const POLLING_STATUSES: ModelSyncJobStatus[] = ['PENDING', 'RUNNING']

export function useModelSyncJob(
  jobId: string | null,
  projectSlug: string | null | undefined
) {
  const client = useProjectScopedClient(projectSlug)
  const skip = !jobId

  const result = useQuery<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>(MODEL_SYNC_JOBS_QUERY, {
    client,
    variables: { jobIds: jobId ? [jobId] : [] },
    skip,
    // No pollInterval — polling managed imperatively below
  })

  const job = result.data?.modelSyncJobs?.[0]
  const isPolling = job ? POLLING_STATUSES.includes(job.status) : !skip

  // Stable callback so the effect doesn't re-run on every render.
  const startPolling = useCallback(() => result.startPolling(2000), [result])
  const stopPollingFn = useCallback(() => result.stopPolling(), [result])

  const wasPolling = useRef(false)

  useEffect(() => {
    if (isPolling && !wasPolling.current) {
      startPolling()
      wasPolling.current = true
    } else if (!isPolling) {
      stopPollingFn()
      wasPolling.current = false
    }
  }, [isPolling, startPolling, stopPollingFn])

  // Reset when jobId changes
  useEffect(() => {
    wasPolling.current = false
  }, [jobId])

  return result
}
