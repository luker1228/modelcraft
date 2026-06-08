// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import React from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import EndUserRecordWorkspace from './EndUserRecordWorkspace'

type MockModel = {
  id: string
  name?: string | null
  title?: string | null
  description?: string | null
  databaseName?: string | null
  createdVia?: 'NEW' | 'IMPORTED' | null
  jsonSchema?: string | null
  fields?: Array<{
    name: string
    isDeprecated?: boolean | null
  }> | null
}

const mockUseQuery = vi.fn()
const mockUseMutation = vi.fn()
const mockModelApiDocsDialog = vi.fn()

vi.mock('@apollo/client', () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
  useMutation: (...args: unknown[]) => mockUseMutation(...args),
}))

vi.mock('sonner', () => ({
  toast: {
    warning: vi.fn(),
    error: vi.fn(),
  },
}))

vi.mock('@api-client/apollo/end-user-client', () => ({
  createEndUserScopedClient: vi.fn(() => ({ kind: 'management-client' })),
  createEndUserModelRuntimeClient: vi.fn(() => ({ kind: 'runtime-client' })),
  useEndUserProjectScopedClient: vi.fn(() => ({ kind: 'management-client' })),
  useEndUserModelRuntimeClient: vi.fn(() => ({ kind: 'runtime-client' })),
}))

vi.mock('@shared/stores/end-user-auth-store', () => ({
  useEndUserAuthStore: vi.fn((selector: (state: { accessToken: string | null }) => unknown) =>
    selector({ accessToken: 'test-token' })
  ),
}))

vi.mock('@web/components/features/model-editor/model-record-form/index', () => ({
  ModelRecordForm: () => <div data-testid="model-record-form" />,
}))

vi.mock('@web/components/shared/data-workspace/ModelRecordTable', () => ({
  ModelRecordTable: () => <div data-testid="model-record-table" />,
}))

vi.mock('@web/components/features/model-editor/model-record-form/runtime/field-protocol', () => ({
  getFieldProtocols: vi.fn(() => []),
}))

vi.mock(
  '@api-client/cms/public',
  () => ({
    buildFindUniqueQuery: vi.fn(() => ({ kind: 'find-unique-query' })),
    buildDeleteMutation: vi.fn(() => ({ kind: 'delete-mutation' })),
    buildCreateMutation: vi.fn(() => ({ kind: 'create-mutation' })),
    buildUpdateMutation: vi.fn(() => ({ kind: 'update-mutation' })),
    extractWritableFieldNamesFromSchema: vi.fn(() => []),
    sanitizeMutationInputData: vi.fn((data: unknown) => data),
  }),
  { virtual: true }
)

vi.mock('@/api-client/noop', () => ({
  NOOP_MUTATION: {},
}))

vi.mock('@/api-client/model/graphql-docs.end-user', () => ({
  GET_MODEL_RECORD_WORKSPACE_END_USER: {},
}))

vi.mock('./FilterPanel', () => ({
  FilterBar: () => <div data-testid="filter-bar" />,
}))

vi.mock('./SortPopover', () => ({
  SortPopover: () => <div data-testid="sort-popover" />,
  buildOrderBy: vi.fn(() => [{ id: 'desc' }]),
}))

vi.mock('@/types/xmc', () => ({
  getXMC: vi.fn(() => null),
}))

vi.mock('@web/components/features/model-editor/model-record-form/access-adapter', () => ({
  RecordAccessAdapterProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

vi.mock('./FilterCopilotActions', () => ({
  useCopilotKitAvailable: vi.fn(() => false),
  FilterCopilotActions: () => <div data-testid="filter-copilot-actions" />,
}))

vi.mock('@web/components/shared/data-workspace/recordPageCount', () => ({
  getRecordPageCountText: vi.fn(() => 'count summary'),
}))

vi.mock('@web/components/shared/data-workspace/useRuntimeListByPage', () => ({
  useRuntimeListByPage: vi.fn(() => ({
    currentPage: 1,
    setCurrentPage: vi.fn(),
    contentLoading: false,
    contentList: [{ id: 'record-1' }],
    totalCount: 1,
    totalPages: 1,
    hasNextPage: false,
    refetch: vi.fn(),
  })),
}))

vi.mock('./ModelApiDocsDialog', () => ({
  ModelApiDocsDialog: (props: {
    open: boolean
    onOpenChange: (open: boolean) => void
    context: unknown
  }) => {
    mockModelApiDocsDialog(props)
    return (
      <div
        data-testid="model-api-docs-dialog"
        data-open={String(props.open)}
        data-context={JSON.stringify(props.context)}
      />
    )
  },
}))

const baseProps = {
  modelId: 'model-1',
  orgName: 'acme',
  projectSlug: 'alpha-project',
}

function buildModel(overrides: Partial<MockModel> = {}): MockModel {
  return {
    id: 'model-1',
    name: 'User',
    title: 'User',
    description: null,
    databaseName: 'users_db',
    createdVia: 'NEW',
    jsonSchema: null,
    fields: [],
    ...overrides,
  }
}

function mockLoadedModel(model: MockModel) {
  mockUseQuery.mockReturnValue({
    data: {
      model: {
        model,
        error: null,
      },
    },
    loading: false,
    refetch: vi.fn(),
  })
}

describe('EndUserRecordWorkspace API docs integration', () => {
  beforeEach(() => {
    mockUseMutation.mockReturnValue([vi.fn()])
    mockModelApiDocsDialog.mockClear()
  })

  afterEach(() => {
    cleanup()
    vi.clearAllMocks()
  })

  it('keeps the add data button and opens API docs with the derived model context', async () => {
    mockLoadedModel(buildModel())

    render(<EndUserRecordWorkspace {...baseProps} />)

    expect(screen.getByRole('button', { name: '添加数据' })).toBeInTheDocument()

    const apiDocsButton = screen.getByRole('button', { name: 'API 文档' })
    expect(apiDocsButton).toBeEnabled()

    fireEvent.click(apiDocsButton)

    await waitFor(() => {
      expect(mockModelApiDocsDialog).toHaveBeenLastCalledWith(
        expect.objectContaining({
          open: true,
          context: {
            orgName: 'acme',
            projectSlug: 'alpha-project',
            databaseName: 'users_db',
            modelName: 'User',
          },
        })
      )
    })
  })

  it('disables API docs when the loaded model is missing runtime context', () => {
    mockLoadedModel(buildModel({ databaseName: null }))

    render(<EndUserRecordWorkspace {...baseProps} />)

    expect(screen.getByRole('button', { name: '添加数据' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'API 文档' })).toBeDisabled()
    expect(mockModelApiDocsDialog).toHaveBeenLastCalledWith(
      expect.objectContaining({
        open: false,
        context: null,
      })
    )
  })
})
