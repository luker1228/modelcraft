import type { ModelDatabase, ModelDatabaseSyncJob } from '@web/hooks/model-database/use-model-databases'

type RegisteredDatabaseSummary = Pick<ModelDatabase, 'id' | 'name'>
type StartSyncFn = (databaseId: string) => Promise<ModelDatabaseSyncJob | null>
type UpsertJobFn = (job: ModelDatabaseSyncJob) => void
type StartPollingFn = (databaseId: string, jobId: string) => void

export async function startSyncForRegisteredDatabases(
  databases: RegisteredDatabaseSummary[],
  startSync: StartSyncFn,
  upsertJob: UpsertJobFn,
  startPolling: StartPollingFn
): Promise<number> {
  let started = 0

  for (const database of databases) {
    const job = await startSync(database.id)
    if (!job) {
      continue
    }

    upsertJob(job)
    startPolling(database.id, job.id)
    started += 1
  }

  return started
}
