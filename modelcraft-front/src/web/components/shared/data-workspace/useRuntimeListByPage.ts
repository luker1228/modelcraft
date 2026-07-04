'use client'

import { useEffect, useMemo, useState } from 'react'
import { useQuery, type ApolloClient } from '@apollo/client'
import { buildListByPageQuery } from '@api-client/cms/public'
import type { FieldDefinition } from '@api-client/cms/public'
import type { ModelRecordTableRow } from './ModelRecordTable'
import { NOOP_QUERY } from '@/api-client/noop'

interface ListByPageResult {
  items?: ModelRecordTableRow[]
  total?: number | null
  pageIndex?: number | null
  pageSize?: number | null
}

interface UseRuntimeListByPageOptions {
  modelName?: string | null
  runtimeFields: string[] | FieldDefinition[] | (string | FieldDefinition)[]
  runtimeClient: ApolloClient<object> | null
  whereInput?: Record<string, unknown>
  orderBy?: RuntimeOrderBy
  pageSize?: number
  resetDeps?: ReadonlyArray<unknown>
}

export type RuntimeSortDirection = 'asc' | 'desc'
export type RuntimeOrderBy = Record<string, RuntimeSortDirection>[]

export interface RuntimeSort {
  field: string
  direction: RuntimeSortDirection
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function getDefaultListByPageOrderBy(
  runtimeFields: string[] | FieldDefinition[] | (string | FieldDefinition)[]
): RuntimeOrderBy | undefined {
  const fieldNames = runtimeFields
    .map((field) => (typeof field === 'string' ? field : field.name))
    .filter((fieldName) => fieldName.length > 0)
  const sortField = fieldNames.includes('id') ? 'id' : fieldNames[0]
  return sortField ? [{ [sortField]: sortField === 'id' ? 'desc' : 'asc' }] : undefined
}

export function buildListByPageOrderBy(
  primarySort: RuntimeSort | null | undefined,
  stableSortEnabled: boolean,
  sortableFieldNames: readonly string[]
): RuntimeOrderBy | undefined {
  const orderBy: RuntimeOrderBy = primarySort
    ? [{ [primarySort.field]: primarySort.direction }]
    : []

  if (
    stableSortEnabled &&
    sortableFieldNames.includes('id') &&
    primarySort?.field !== 'id'
  ) {
    orderBy.push({ id: 'desc' })
  }

  return orderBy.length > 0 ? orderBy : undefined
}

export function useRuntimeListByPage({
  modelName,
  runtimeFields,
  runtimeClient,
  whereInput,
  orderBy,
  pageSize = 20,
  resetDeps = [],
}: UseRuntimeListByPageOptions) {
  const [currentPage, setCurrentPage] = useState(1)

  const listByPageQuery = useMemo(() => {
    if (!modelName) return null
    return buildListByPageQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  const effectiveOrderBy = useMemo(() => {
    return orderBy && orderBy.length > 0 ? orderBy : getDefaultListByPageOrderBy(runtimeFields)
  }, [orderBy, runtimeFields])

  const {
    data: contentData,
    loading: contentLoading,
    refetch,
  } = useQuery<Record<string, unknown>>(listByPageQuery || NOOP_QUERY, {
    client: runtimeClient ?? undefined,
    skip: !listByPageQuery || !runtimeClient,
    fetchPolicy: 'network-only',
    variables: {
      where: whereInput,
      pageIndex: currentPage,
      pageSize,
      orderBy: effectiveOrderBy,
    },
  })

  const listByPageData = useMemo<ListByPageResult | null>(() => {
    const payload = (contentData as Record<string, unknown> | undefined)?.listByPage
    return isRecord(payload) ? (payload as ListByPageResult) : null
  }, [contentData])

  const contentList: ModelRecordTableRow[] = useMemo(
    () => (Array.isArray(listByPageData?.items) ? listByPageData.items : []),
    [listByPageData]
  )

  const totalCount = typeof listByPageData?.total === 'number' ? listByPageData.total : 0
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize))
  const hasNextPage = currentPage < totalPages

  useEffect(() => {
    setCurrentPage(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, resetDeps)

  return {
    currentPage,
    setCurrentPage,
    contentLoading,
    contentList,
    totalCount,
    totalPages,
    hasNextPage,
    refetch,
  }
}
