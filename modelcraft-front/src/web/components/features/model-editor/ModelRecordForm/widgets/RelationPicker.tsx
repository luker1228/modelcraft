'use client'

import React, { useState } from 'react'
import type { WidgetProps, RJSFSchema } from '@rjsf/utils'
import { useQuery, gql } from '@apollo/client'
import { createModelRuntimeClient } from '@bff/apollo/public'
import { buildFindManyQuery } from '@bff/cms/public'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { RefreshCw } from 'lucide-react'

interface FormContext {
  orgName: string
  projectSlug: string
  clusterName: string
  databaseName: string
  modelId: string
  logicalForeignKeys: Array<{
    id: string
    refModelId: string
    refModelName: string
    sourceFields: string[]
    targetFields: string[]
  }>
}

interface RemoteRecord {
  id: string
  __label?: string
  [key: string]: unknown
}

interface FindManyQueryData {
  findMany?: {
    items: RemoteRecord[]
  } | null
}

/**
 * Extract typed fields from WidgetProps to avoid `any` index signature pollution.
 * WidgetProps extends GenericObjectType which has `[name: string]: any`,
 * so direct destructuring triggers @typescript-eslint/no-unsafe-* rules.
 */
function extractWidgetFields(props: WidgetProps) {
  return {
    value: props.value as unknown,
    onChange: props.onChange as (value: unknown) => void,
    disabled: props.disabled as boolean,
    uiSchema: props.uiSchema as Record<string, unknown> | undefined,
    formContext: props.formContext as Record<string, unknown> | undefined,
  }
}

/**
 * RelationPicker widget for RJSF.
 *
 * Renders a searchable <Select> populated with records from the referenced model.
 * Requires `formContext` to carry connection info and logicalForeignKeys.
 *
 * ui:options expected:
 *   relateFkId: string — ID of the LogicalForeignKey defining this relation
 */
export function RelationPicker(props: WidgetProps) {
  const { value, onChange, disabled, uiSchema, formContext } = extractWidgetFields(props)

  const ctx = (formContext ?? {}) as unknown as FormContext
  const { orgName, projectSlug, databaseName, logicalForeignKeys } = ctx

  const uiOptions = (uiSchema?.['ui:options'] ?? {}) as { relateFkId?: string }
  const relateFkId = uiOptions.relateFkId ?? ''

  const fk = logicalForeignKeys?.find((f) => f.id === relateFkId)
  const refModelName = fk?.refModelName ?? ''

  const [search, setSearch] = useState('')

  // Only create the client when we have enough info
  const client = React.useMemo(() => {
    if (!orgName || !projectSlug || !databaseName || !refModelName) return null
    return createModelRuntimeClient(orgName, projectSlug, databaseName, refModelName)
  }, [orgName, projectSlug, databaseName, refModelName])

  const findManyQuery = React.useMemo(() => {
    if (!refModelName) return null
    return buildFindManyQuery(refModelName, ['id', '__label'])
  }, [refModelName])

  // Provide a no-op fallback query/client so hooks are always called unconditionally.
  // The `skip` flag prevents actual execution when client or query is not ready.
  const safeQuery = findManyQuery ?? gql`query Noop { __typename }`
  const safeClient = client ?? undefined

  const {
    data,
    loading,
    error,
    refetch,
  } = useQuery<FindManyQueryData>(safeQuery, {
    client: safeClient,
    skip: !client || !findManyQuery,
    variables: { take: 200, skip: 0 },
  })

  // Guard: no FK / no ref model
  if (!fk || !refModelName) {
    return (
      <span className="text-sm text-destructive">无法找到关联模型</span>
    )
  }

  if (loading) {
    return <span className="text-sm text-muted-foreground">加载中...</span>
  }

  if (error) {
    return (
      <div className="flex items-center gap-2">
        <span className="text-sm text-destructive">加载关联数据失败</span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          className="h-7 px-2"
        >
          <RefreshCw className="mr-1 size-3" />
          重试
        </Button>
      </div>
    )
  }

  const records: RemoteRecord[] = data?.findMany?.items ?? []

  if (records.length === 0) {
    return <span className="text-sm text-muted-foreground">暂无数据</span>
  }

  /**
   * Format display label according to id + __label protocol.
   * Format: `__label(id)`, or `空(id)` if __label is empty.
   */
  const formatLabel = (r: RemoteRecord): string => {
    const labelStr = typeof r.__label === 'string' ? r.__label : ''
    if (labelStr === '') {
      return `空(${r.id})`
    }
    return `${labelStr}(${r.id})`
  }

  const filtered = search
    ? records.filter((r) => {
        const displayLabel = formatLabel(r)
        return displayLabel.toLowerCase().includes(search.toLowerCase())
      })
    : records

  const currentValue: string = typeof value === 'string' ? value : ''

  return (
    <Select
      value={currentValue}
      onValueChange={(val) => onChange(val)}
      disabled={disabled}
    >
      <SelectTrigger className="w-full">
        <SelectValue placeholder="选择关联记录…" />
      </SelectTrigger>
      <SelectContent>
        <div className="px-2 py-1">
          <Input
            placeholder="搜索…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-7 text-sm"
            // Prevent the select from closing on keydown
            onKeyDown={(e) => e.stopPropagation()}
          />
        </div>
        {filtered.map((record) => (
          <SelectItem key={record.id} value={record.id}>
            {formatLabel(record)}
          </SelectItem>
        ))}
        {filtered.length === 0 && (
          <div className="px-2 py-1.5 text-sm text-muted-foreground">无匹配记录</div>
        )}
      </SelectContent>
    </Select>
  )
}
