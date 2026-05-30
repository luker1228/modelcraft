import { describe, expect, it, vi } from 'vitest'
import { startSyncForRegisteredDatabases } from './batch-register-sync'

describe('startSyncForRegisteredDatabases', () => {
  it('starts sync and polling for each successfully registered database', async () => {
    const startSync = vi
      .fn()
      .mockResolvedValueOnce({ id: 'job-1', databaseId: 'db-1', status: 'PENDING' })
      .mockResolvedValueOnce({ id: 'job-2', databaseId: 'db-2', status: 'RUNNING' })

    const upsertJob = vi.fn()
    const startPolling = vi.fn()

    const started = await startSyncForRegisteredDatabases(
      [
        { id: 'db-1', name: 'orders' },
        { id: 'db-2', name: 'billing' },
      ],
      startSync,
      upsertJob,
      startPolling
    )

    expect(started).toBe(2)
    expect(startSync).toHaveBeenNthCalledWith(1, 'db-1')
    expect(startSync).toHaveBeenNthCalledWith(2, 'db-2')
    expect(upsertJob).toHaveBeenCalledTimes(2)
    expect(startPolling).toHaveBeenNthCalledWith(1, 'db-1', 'job-1')
    expect(startPolling).toHaveBeenNthCalledWith(2, 'db-2', 'job-2')
  })

  it('skips polling when sync job creation returns null', async () => {
    const startSync = vi.fn().mockResolvedValue(null)
    const upsertJob = vi.fn()
    const startPolling = vi.fn()

    const started = await startSyncForRegisteredDatabases(
      [{ id: 'db-1', name: 'orders' }],
      startSync,
      upsertJob,
      startPolling
    )

    expect(started).toBe(0)
    expect(upsertJob).not.toHaveBeenCalled()
    expect(startPolling).not.toHaveBeenCalled()
  })
})
