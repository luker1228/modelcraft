import { useMutation, useQuery } from '@apollo/client'
import {
  LIST_MODEL_DATABASES,
  LIST_CLUSTER_RAW_DATABASES,
  REGISTER_MODEL_DATABASE,
  BATCH_REGISTER_MODEL_DATABASES,
  UPDATE_MODEL_DATABASE,
  UNREGISTER_MODEL_DATABASE,
  START_MODEL_DATABASE_SYNC,
  GET_MODEL_DATABASE_SYNC_JOB,
} from '@/api-client/project'
import { useProjectScopedClient } from '@api-client/apollo/public'

// ── Types ────────────────────────────────────────────────────────────────────────

export type DatabaseMode = 'SELF_HOSTED' | 'MANAGED'

export interface ModelDatabase {
  id: string
  name: string
  title: string
  description: string
  mode: DatabaseMode
  createdAt: string
  updatedAt: string
}

export type ModelDatabaseSyncJobStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCEEDED'
  | 'PARTIAL_SUCCESS'
  | 'FAILED'

export interface ModelDatabaseSyncFailedTable {
  tableName: string
  message: string
}

export interface ModelDatabaseSyncJob {
  id: string
  databaseId: string
  status: ModelDatabaseSyncJobStatus
  totalTables: number
  processedTables: number
  createdModels: number
  syncedModels: number
  failedCount: number
  failedTables: ModelDatabaseSyncFailedTable[]
  startedAt?: string | null
  finishedAt?: string | null
  createdAt: string
  updatedAt: string
}

export interface RawDatabase {
  name: string
  isRegistered: boolean
}

export interface RegisterModelDatabaseInput {
  name: string
  title: string
  description?: string
  mode: DatabaseMode
}

export interface UpdateModelDatabaseInput {
  title?: string
  description?: string
  mode?: DatabaseMode
}

// ── Internal response types ──────────────────────────────────────────────────────

interface InvalidInputError {
  __typename: 'InvalidInput'
  message: string
}

interface ResourceNotFoundError {
  __typename: 'ResourceNotFound'
  message: string
  resourceType: string
}

interface ModelDatabaseResult {
  __typename: 'ModelDatabase'
  id: string
  name: string
  title: string
  description: string
  mode: DatabaseMode
  createdAt: string
  updatedAt: string
}

type RegisterModelDatabaseResult = ModelDatabaseResult | InvalidInputError | ResourceNotFoundError

interface RegisterModelDatabaseData {
  registerModelDatabase: RegisterModelDatabaseResult
}

interface UpdateModelDatabaseData {
  updateModelDatabase: ModelDatabase
}

interface StartModelDatabaseSyncData {
  startModelDatabaseSync: {
    job: ModelDatabaseSyncJob
  }
}

interface GetModelDatabaseSyncJobData {
  modelDatabaseSyncJob: ModelDatabaseSyncJob | null
}

// ── Hooks ────────────────────────────────────────────────────────────────────────

export function useModelDatabases(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const { data, loading, error, refetch } = useQuery<{ modelDatabases: ModelDatabase[] }>(
    LIST_MODEL_DATABASES,
    { client, skip: !projectSlug }
  )
  return {
    databases: data?.modelDatabases ?? [],
    loading,
    error,
    refetch,
  }
}

export function useClusterRawDatabases(
  projectSlug: string | null | undefined,
  skip?: boolean
) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const { data, loading, error, refetch } = useQuery<{ clusterRawDatabases: RawDatabase[] }>(
    LIST_CLUSTER_RAW_DATABASES,
    { client, skip: !projectSlug || skip }
  )
  return {
    rawDatabases: data?.clusterRawDatabases ?? [],
    loading,
    error,
    refetch,
  }
}

export function useRegisterModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading, error }] = useMutation<RegisterModelDatabaseData>(
    REGISTER_MODEL_DATABASE,
    {
      client,
      refetchQueries: [LIST_MODEL_DATABASES],
    }
  )

  const register = async (input: RegisterModelDatabaseInput) => {
    const result = await mutate({ variables: { input } })
    const data = result.data?.registerModelDatabase
    if (data?.__typename === 'InvalidInput' || data?.__typename === 'ResourceNotFound') {
      throw new Error(data.message)
    }
    return data as ModelDatabase
  }

  return { register, loading, error }
}

export function useUpdateModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading, error }] = useMutation<UpdateModelDatabaseData>(
    UPDATE_MODEL_DATABASE,
    {
      client,
      refetchQueries: [LIST_MODEL_DATABASES],
    }
  )

  const update = async (id: string, input: UpdateModelDatabaseInput) => {
    const result = await mutate({ variables: { id, input } })
    return result.data?.updateModelDatabase
  }

  return { update, loading, error }
}

export function useUnregisterModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading }] = useMutation(UNREGISTER_MODEL_DATABASE, {
    client,
    refetchQueries: [LIST_MODEL_DATABASES],
  })

  const unregister = async (id: string) => {
    await mutate({ variables: { id } })
  }

  return { unregister, loading }
}

export function useStartModelDatabaseSync(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading, error }] = useMutation<StartModelDatabaseSyncData>(
    START_MODEL_DATABASE_SYNC,
    { client }
  )

  const startSync = async (databaseId: string) => {
    const result = await mutate({ variables: { databaseId } })
    return result.data?.startModelDatabaseSync.job ?? null
  }

  return { startSync, loading, error }
}

export function useFetchModelDatabaseSyncJob(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)

  const fetchJob = async (jobId: string) => {
    const result = await client.query<GetModelDatabaseSyncJobData>({
      query: GET_MODEL_DATABASE_SYNC_JOB,
      variables: { jobId },
      fetchPolicy: 'network-only',
    })
    return result.data.modelDatabaseSyncJob
  }

  return { fetchJob }
}

export interface BatchRegisterError {
  name: string
  message: string
}

export interface BatchRegisterResult {
  succeeded: ModelDatabase[]
  failed: BatchRegisterError[]
}

interface BatchRegisterModelDatabaseData {
  batchRegisterModelDatabases: BatchRegisterResult
}

export function useBatchRegisterModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading }] = useMutation<BatchRegisterModelDatabaseData>(
    BATCH_REGISTER_MODEL_DATABASES,
    {
      client,
      refetchQueries: [LIST_MODEL_DATABASES],
    }
  )

  const batchRegister = async (inputs: RegisterModelDatabaseInput[]): Promise<BatchRegisterResult> => {
    const result = await mutate({ variables: { input: { databases: inputs } } })
    return result.data?.batchRegisterModelDatabases ?? { succeeded: [], failed: [] }
  }

  return { batchRegister, loading }
}
