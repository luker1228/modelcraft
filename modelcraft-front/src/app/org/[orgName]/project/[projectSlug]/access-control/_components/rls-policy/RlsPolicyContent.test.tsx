// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import * as React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_MODEL, GET_MODELS_BY_DATABASE } from '@/api-client/model'
import { REGISTERED_DATABASES } from '@/api-client/cluster'
import { RlsPolicyContent } from './RlsPolicyContent'

vi.mock('@apollo/client', async () => {
  const actual = await vi.importActual<typeof import('@apollo/client')>('@apollo/client')
  return {
    ...actual,
    useQuery: vi.fn(),
  }
})

vi.mock('@api-client/apollo/develop-client', () => ({
  useProjectScopedClient: vi.fn(() => 'project-client'),
}))

vi.mock('../../_hooks/rls-policy/useRlsPolicyList', () => ({
  useRlsPolicyList: vi.fn(() => ({
    policies: [],
    loading: false,
    error: undefined,
  })),
}))

vi.mock('../../_hooks/rls-policy/useRlsPolicyManage', () => ({
  useRlsPolicyManage: vi.fn(() => ({
    upsertPolicy: vi.fn(),
    deletePolicy: vi.fn(),
    validateRlsExpression: vi.fn(),
    upserting: false,
    deleting: false,
    validating: false,
  })),
}))

vi.mock('./PolicyEditorDialog', () => ({
  PolicyEditorDialog: ({
    modelFields,
    authVariables,
  }: {
    modelFields?: Array<{ name: string }>
    authVariables?: Array<{ name: string }>
  }) => (
    <div>
      <div data-testid="model-fields">{JSON.stringify(modelFields ?? [])}</div>
      <div data-testid="auth-variables">{JSON.stringify(authVariables ?? [])}</div>
    </div>
  ),
}))

vi.mock('@web/components/ui/button', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button {...props}>{children}</button>
  ),
}))

vi.mock('@web/components/ui/select', () => {
  const SelectContext = React.createContext<{
    onValueChange?: (value: string) => void
  } | null>(null)

  return {
    Select: ({
      children,
      onValueChange,
    }: {
      children: React.ReactNode
      onValueChange?: (value: string) => void
    }) => (
      <SelectContext.Provider value={{ onValueChange }}>{children}</SelectContext.Provider>
    ),
    SelectTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    SelectValue: ({ placeholder }: { placeholder?: string }) => <span>{placeholder}</span>,
    SelectContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    SelectItem: ({ children, value }: { children: React.ReactNode; value: string }) => {
      const context = React.useContext(SelectContext)
      return (
        <button type="button" onClick={() => context?.onValueChange?.(value)}>
          {children}
        </button>
      )
    },
  }
})

vi.mock('@web/components/ui/table', () => ({
  Table: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TableHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TableBody: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TableRow: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TableHead: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TableCell: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

vi.mock('@web/components/ui/skeleton', () => ({
  Skeleton: () => <div>loading</div>,
}))

vi.mock('@web/components/ui/alert-dialog', () => ({
  AlertDialog: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogTitle: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogDescription: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogAction: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AlertDialogCancel: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

describe('RlsPolicyContent', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.mocked(useQuery).mockReset()
    vi.mocked(useProjectScopedClient).mockClear()
  })

  type MockQueryOptions = {
    variables?: {
      input?: {
        databaseName?: string | null
      }
      id?: string | null
    }
  }

  it('uses the org-scoped project query and loads row completions from model details', () => {
    vi.mocked(useQuery).mockImplementation((query, options) => {
      if (query === REGISTERED_DATABASES) {
        return {
          data: {
            registeredDatabases: {
              data: {
                databases: [{ name: 'main' }],
              },
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODELS_BY_DATABASE) {
        return {
          data: {
            models: {
              items: [{ id: 'model-1', name: 'users', title: 'Users' }],
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODEL) {
        return {
          data: options?.variables?.id === 'model-1'
            ? {
                model: {
                  model: {
                    id: 'model-1',
                    name: 'users',
                    fields: [{ name: 'owner_id', title: 'Owner' }],
                  },
                },
              }
            : undefined,
          loading: false,
        } as never
      }

      return { data: undefined, loading: false } as never
    })

    render(<RlsPolicyContent orgName="acme" projectSlug="demo" />)

    fireEvent.click(screen.getByRole('button', { name: 'main' }))
    fireEvent.click(screen.getAllByRole('button', { name: 'users' })[0])

    expect(useProjectScopedClient).toHaveBeenCalledWith('demo')
    expect(screen.getByTestId('model-fields')).toHaveTextContent('owner_id')
    expect(screen.getByTestId('auth-variables')).toHaveTextContent('userid')
    expect(screen.getByTestId('auth-variables')).toHaveTextContent('username')
  })

  it('restores the last selected database and model for the current project', () => {
    window.localStorage.setItem(
      'modelcraft:access-control:rls-policy:acme:demo',
      JSON.stringify({ databaseName: 'demo_ecommerce', modelId: 'model-1' }),
    )

    vi.mocked(useQuery).mockImplementation((query, options) => {
      if (query === REGISTERED_DATABASES) {
        return {
          data: {
            registeredDatabases: {
              data: {
                databases: [{ name: 'demo_ecommerce' }],
              },
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODELS_BY_DATABASE) {
        return {
          data: {
            models: {
              items: [{ id: 'model-1', name: 'users', title: 'Users' }],
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODEL) {
        return {
          data: options?.variables?.id === 'model-1'
            ? {
                model: {
                  model: {
                    id: 'model-1',
                    name: 'users',
                    fields: [{ name: 'owner_id', title: 'Owner' }],
                  },
                },
              }
            : undefined,
          loading: false,
        } as never
      }

      return { data: undefined, loading: false } as never
    })

    render(<RlsPolicyContent orgName="acme" projectSlug="demo" />)

    expect(useProjectScopedClient).toHaveBeenCalledWith('demo')
    expect(vi.mocked(useQuery)).toHaveBeenCalledWith(
      GET_MODELS_BY_DATABASE,
      expect.objectContaining({
        variables: { input: { databaseName: 'demo_ecommerce' } },
      }),
    )
    expect(vi.mocked(useQuery)).toHaveBeenCalledWith(
      GET_MODEL,
      expect.objectContaining({
        variables: { id: 'model-1' },
      }),
    )
  })

  it('clears the selected model when switching to another database', () => {
    vi.mocked(useQuery).mockImplementation((query: unknown, options?: MockQueryOptions) => {
      if (query === REGISTERED_DATABASES) {
        return {
          data: {
            registeredDatabases: {
              data: {
                databases: [{ name: 'demo_ecommerce' }, { name: 'analytics' }],
              },
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODELS_BY_DATABASE) {
        const databaseName = options?.variables?.input?.databaseName
        return {
          data: {
            models: {
              items:
                databaseName === 'demo_ecommerce'
                  ? [{ id: 'model-1', name: 'users', title: 'Users' }]
                  : [{ id: 'model-2', name: 'events', title: 'Events' }],
            },
          },
          loading: false,
        } as never
      }

      if (query === GET_MODEL) {
        return {
          data: {
            model: {
              model: {
                id: options?.variables?.id,
                name: 'users',
                fields: [{ name: 'owner_id', title: 'Owner' }],
              },
            },
          },
          loading: false,
        } as never
      }

      return { data: undefined, loading: false } as never
    })

    render(<RlsPolicyContent orgName="acme" projectSlug="demo" />)

    fireEvent.click(screen.getAllByRole('button', { name: 'demo_ecommerce' })[0])
    fireEvent.click(screen.getAllByRole('button', { name: 'users' })[0])

    fireEvent.click(screen.getAllByRole('button', { name: 'demo_ecommerce' })[0])
    fireEvent.click(screen.getAllByRole('button', { name: 'analytics' })[0])

    expect(window.localStorage.getItem('modelcraft:access-control:rls-policy:acme:demo')).toContain(
      '"databaseName":"analytics"',
    )
    expect(window.localStorage.getItem('modelcraft:access-control:rls-policy:acme:demo')).toContain(
      '"modelId":null',
    )
  })
})
