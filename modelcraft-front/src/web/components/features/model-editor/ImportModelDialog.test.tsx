// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { ImportModelDialog } from './ImportModelDialog'

const mockUseQuery = vi.fn()
const mockUseModelSyncJob = vi.fn()
const mockStartSync = vi.fn()
const mockToastError = vi.fn()
const mockToastSuccess = vi.fn()

vi.mock('@apollo/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@apollo/client')>()
  return {
    ...actual,
    useQuery: (...args: unknown[]) => mockUseQuery(...args),
  }
})

vi.mock('@web/hooks/model/use-sync-models-from-db', () => ({
  useModelSyncJob: (...args: unknown[]) => mockUseModelSyncJob(...args),
}))

vi.mock('@web/hooks/model/use-model-sync', () => ({
  useStartModelSync: () => ({
    startSync: mockStartSync,
    loading: false,
  }),
}))

vi.mock('@api-client/apollo/develop-client', () => ({
  useProjectScopedClient: () => ({}),
}))

vi.mock('sonner', () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
    success: (...args: unknown[]) => mockToastSuccess(...args),
  },
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('ImportModelDialog', () => {
  it('does not keep showing importing only because a jobId exists', async () => {
    mockUseQuery.mockReturnValue({
      data: {
        listTables: {
          items: [{ name: 'users' }],
          totalCount: 1,
        },
      },
      loading: false,
    })

    mockStartSync.mockResolvedValue({
      batchId: 'batch-1',
      jobs: [{ databaseId: 'db-1', jobId: 'job-1' }],
    })

    mockUseModelSyncJob.mockImplementation((jobId: string | null) => {
      if (!jobId) {
        return { data: undefined, loading: false }
      }
      return {
        data: {
          modelSyncJobs: [{ status: 'UNKNOWN_STATUS' }],
        },
        loading: false,
      }
    })

    render(
      <ImportModelDialog
        open
        onOpenChange={vi.fn()}
        projectSlug="demo-project"
        databaseName="demo_ecommerce"
        databaseId="db-1"
        onSuccess={vi.fn()}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: 'users' }))
    fireEvent.click(screen.getByRole('button', { name: '导入' }))

    await waitFor(() => {
      expect(mockUseModelSyncJob).toHaveBeenLastCalledWith('job-1', 'demo-project')
    })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '导入' })).toBeEnabled()
    })
  })
})
