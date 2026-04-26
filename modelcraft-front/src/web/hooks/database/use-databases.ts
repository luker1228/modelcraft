import { useMemo } from 'react'
import { useQuery } from '@apollo/client'
import { LIST_DATABASES } from '@/api-client/cluster'
import { useProjectScopedClient } from '@api-client/apollo/public'
import type { ApolloError } from '@apollo/client'

interface Database {
  name: string
}

interface PageInfo {
  hasNextPage: boolean
  hasPreviousPage: boolean
  startCursor: string | null
  endCursor: string | null
}

interface UseDatabasesOptions {
  initialLimit?: number
  search?: string
}

// -- GraphQL response types ---------------------------------------------------

interface DatabaseEdge {
  node: Database
  cursor: string
}

interface ListDatabasesConnection {
  edges: DatabaseEdge[]
  pageInfo: PageInfo
  totalCount: number
}

interface ListDatabasesData {
  listDatabases: ListDatabasesConnection
}

interface UseDatabasesReturn {
  databases: Database[]
  pageInfo: PageInfo
  totalCount: number
  loading: boolean
  error: ApolloError | undefined
  fetchMore: (limit?: number) => Promise<unknown>
  refetch: () => void
}

export function useDatabases(
  projectSlug: string | null | undefined,
  options?: UseDatabasesOptions
): UseDatabasesReturn {
  // Use project-scoped client -- listDatabases lives on the project endpoint
  // /graphql/org/{orgName}/project/{projectSlug}/
  const client = useProjectScopedClient(projectSlug ?? undefined)

  const { data, loading, error, fetchMore, refetch } = useQuery<ListDatabasesData>(
    LIST_DATABASES,
    {
      client,
      variables: {
        input: {
          limit: options?.initialLimit || 50,
          offset: 0,
          search: options?.search,
        },
      },
      skip: !projectSlug,
    }
  )

  const databases: Database[] = useMemo(
    () => data?.listDatabases?.edges?.map((edge) => edge.node) || [],
    [data?.listDatabases?.edges]
  )

  const pageInfo: PageInfo = useMemo(
    () =>
      data?.listDatabases?.pageInfo || {
        hasNextPage: false,
        hasPreviousPage: false,
        startCursor: null,
        endCursor: null,
      },
    [data?.listDatabases?.pageInfo]
  )

  const totalCount: number = useMemo(
    () => data?.listDatabases?.totalCount || 0,
    [data?.listDatabases?.totalCount]
  )

  const handleFetchMore = (limit: number = 50): Promise<unknown> => {
    if (!pageInfo.hasNextPage || !projectSlug) return Promise.resolve()

    return fetchMore({
      variables: {
        input: {
          limit,
          offset: databases.length,
          search: options?.search,
        },
      },
      updateQuery: (prev, { fetchMoreResult }) => {
        if (!fetchMoreResult) return prev
        return {
          listDatabases: {
            ...fetchMoreResult.listDatabases,
            edges: [
              ...prev.listDatabases.edges,
              ...fetchMoreResult.listDatabases.edges,
            ],
          },
        }
      },
    })
  }

  return {
    databases,
    pageInfo,
    totalCount,
    loading,
    error,
    fetchMore: handleFetchMore,
    refetch,
  }
}
