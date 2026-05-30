import { describe, expect, it } from 'vitest'
import type { ModelDatabase } from '@web/hooks/model-database/use-model-databases'
import {
  buildDisplayedDatabases,
  createOptimisticPendingSyncJob,
} from './database-list-state'

const db = (id: string, name: string): ModelDatabase => ({
  id,
  name,
  title: name,
  description: '',
  mode: 'SELF_HOSTED',
  createdAt: '2026-05-30T00:00:00Z',
  updatedAt: '2026-05-30T00:00:00Z',
})

describe('buildDisplayedDatabases', () => {
  it('puts optimistic databases into the table before refetch completes', () => {
    const displayed = buildDisplayedDatabases([db('fetched-1', 'users')], [db('optimistic-1', 'orders')])
    expect(displayed.map((item) => item.id)).toEqual(['optimistic-1', 'fetched-1'])
  })

  it('deduplicates databases by id when fetched data catches up', () => {
    const displayed = buildDisplayedDatabases([db('shared-1', 'orders')], [db('shared-1', 'orders')])
    expect(displayed).toHaveLength(1)
  })
})

describe('createOptimisticPendingSyncJob', () => {
  it('creates a pending sync job placeholder for a newly registered database', () => {
    const job = createOptimisticPendingSyncJob('db-1')
    expect(job.databaseId).toBe('db-1')
    expect(job.status).toBe('PENDING')
    expect(job.failedTables).toEqual([])
  })
})
