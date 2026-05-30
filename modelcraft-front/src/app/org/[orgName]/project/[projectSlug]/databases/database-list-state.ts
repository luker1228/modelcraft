import type { ModelDatabase, ModelDatabaseSyncJob } from '@web/hooks/model-database/use-model-databases'

export function buildDisplayedDatabases(
  fetchedDatabases: ModelDatabase[],
  optimisticDatabases: ModelDatabase[]
): ModelDatabase[] {
  const byId = new Map<string, ModelDatabase>()

  optimisticDatabases.forEach((database) => {
    byId.set(database.id, database)
  })

  fetchedDatabases.forEach((database) => {
    byId.set(database.id, database)
  })

  return [...byId.values()]
}

export function createOptimisticPendingSyncJob(databaseId: string): ModelDatabaseSyncJob {
  const now = new Date().toISOString()
  return {
    id: `pending-${databaseId}`,
    databaseId,
    status: 'PENDING',
    totalTables: 0,
    processedTables: 0,
    createdModels: 0,
    syncedModels: 0,
    failedCount: 0,
    failedTables: [],
    startedAt: now,
    finishedAt: null,
    createdAt: now,
    updatedAt: now,
  }
}
