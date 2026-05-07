'use client'

import React, { useState, useCallback, useEffect, useMemo } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import type { ApolloClient } from '@apollo/client'
import { createModelRuntimeClient } from '@api-client/apollo/public'
import { buildFindManyQuery, buildFindUniqueQuery } from '@api-client/cms/public'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@web/components/ui/command'
import { Button } from '@web/components/ui/button'
import { ChevronsUpDown, X, Check, Loader2 } from 'lucide-react'
import { cn } from '@/shared/utils'
import type { XMC } from '@/types/xmc'
import { useWidgetRouteContext } from '../_hooks/useWidgetRouteContext'

// ──────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────

interface RemoteRecord {
  id: string
  _displayName?: string | null
}

interface FindManyQueryData {
  findMany?: {
    items: RemoteRecord[]
  } | null
}

/**
 * FormContext 注入的 createRuntimeClient 函数类型。
 * develop workspace 注入 createModelRuntimeClient；
 * runtime workspace 注入 createEndUserScopedClient（忽略 databaseName/modelName）。
 */
interface FormContext {
  orgName?: string
  projectSlug?: string
  /** 可选：通过 access adapter 注入的 runtime client 工厂，优先于内建路由推导。 */
  createRuntimeClient?: (databaseName: string, modelName: string) => ApolloClient<object>
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

/**
 * Format a remote record for display.
 * Protocol: `_displayName(id)`, or `空(id)` when _displayName is empty string.
 */
function formatRecordDisplay(r: RemoteRecord): string {
  const labelStr = typeof r._displayName === 'string' ? r._displayName : ''
  if (labelStr === '') {
    return `空(${r.id})`
  }
  return `${labelStr}(${r.id})`
}

/**
 * Format just the trigger label when a value is selected.
 * Shows `_displayName (id)` — human-readable label first, id as context.
 * Falls back to bare id if no _displayName.
 */
function formatTriggerLabel(r: RemoteRecord): string {
  const labelStr = typeof r._displayName === 'string' ? r._displayName : ''
  if (labelStr === '') return r.id
  return `${labelStr} (${r.id})`
}

/** Safely extract WidgetProps fields to avoid unsafe-* ESLint violations */
function extractProps(props: WidgetProps) {
  return {
    id: props.id as string,
    value: props.value as unknown,
    onChange: props.onChange as (v: unknown) => void,
    disabled: props.disabled as boolean,
    required: props.required as boolean,
    schema: props.schema as Record<string, unknown>,
    formContext: props.formContext as Record<string, unknown> | undefined,
  }
}

// ──────────────────────────────────────────────
// Hook: debounced relation search
// ──────────────────────────────────────────────

interface UseRelationSearchOptions {
  orgName: string
  projectSlug: string
  databaseName: string
  modelName: string
  search: string
  enabled: boolean
  triggerNonce: number
  createRuntimeClientOverride?: (databaseName: string, modelName: string) => ApolloClient<object>
}

interface UseRelationSearchResult {
  records: RemoteRecord[]
  loading: boolean
  error: boolean
}

function useRelationSearch({
  orgName,
  projectSlug,
  databaseName,
  modelName,
  search,
  enabled,
  triggerNonce,
  createRuntimeClientOverride,
}: UseRelationSearchOptions): UseRelationSearchResult {
  const [records, setRecords] = useState<RemoteRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  const client = useMemo(() => {
    if (!orgName || !projectSlug || !databaseName || !modelName) {
      return null
    }
    // 优先使用注入的 createRuntimeClient（来自 access adapter）
    if (createRuntimeClientOverride) {
      return createRuntimeClientOverride(databaseName, modelName)
    }
    // 降级：使用 develop-specific model runtime client
    return createModelRuntimeClient(orgName, projectSlug, databaseName, modelName)
  }, [orgName, projectSlug, databaseName, modelName, createRuntimeClientOverride])

  // Debounce search input 300ms
  const [debouncedSearch, setDebouncedSearch] = useState(search)
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(timer)
  }, [search])

  useEffect(() => {
    if (!enabled || !client) return

    let cancelled = false

    // Build where filter only when there is a search term
    const variables: Record<string, unknown> = { take: 50 }
    if (debouncedSearch.trim()) {
      variables.where = { _displayName: { contains: debouncedSearch.trim() } }
    }

    setLoading(true)
    setError(false)

    ;(async () => {
      try {
        const withLabel = await client.query<FindManyQueryData>({
          query: buildFindManyQuery(modelName, ['id', '_displayName']),
          variables,
          fetchPolicy: 'network-only',
        })

        const runtimeError = (withLabel as unknown as { error?: unknown; errors?: unknown }).error
          ?? (withLabel as unknown as { errors?: unknown }).errors
        if (runtimeError) {
          if (!cancelled) {
            setError(true)
            setLoading(false)
          }
          return
        }

        if (!cancelled) {
          setRecords(withLabel.data?.findMany?.items ?? [])
          setLoading(false)
        }
      } catch {
        if (!cancelled) {
          setError(true)
          setLoading(false)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [enabled, client, debouncedSearch, modelName, triggerNonce])

  return { records, loading, error }
}

// ──────────────────────────────────────────────
// Hook: resolve display label for current value
// ──────────────────────────────────────────────

interface UseCurrentRecordOptions {
  orgName: string
  projectSlug: string
  databaseName: string
  modelName: string
  currentId: string
  createRuntimeClientOverride?: (databaseName: string, modelName: string) => ApolloClient<object>
}

function useCurrentRecord({
  orgName,
  projectSlug,
  databaseName,
  modelName,
  currentId,
  createRuntimeClientOverride,
}: UseCurrentRecordOptions): RemoteRecord | null {
  const [record, setRecord] = useState<RemoteRecord | null>(null)
  const client = useMemo(() => {
    if (!orgName || !projectSlug || !databaseName || !modelName) {
      return null
    }
    if (createRuntimeClientOverride) {
      return createRuntimeClientOverride(databaseName, modelName)
    }
    return createModelRuntimeClient(orgName, projectSlug, databaseName, modelName)
  }, [orgName, projectSlug, databaseName, modelName, createRuntimeClientOverride])

  useEffect(() => {
    if (!currentId || !client) {
      setRecord(null)
      return
    }

    let cancelled = false
    ;(async () => {
      try {
        const withLabel = await client.query<{ findUnique?: { item?: RemoteRecord | null } | null }>({
          query: buildFindUniqueQuery(modelName, ['id', '_displayName']),
          variables: { where: { id: currentId } },
          fetchPolicy: 'cache-first',
        })

        const runtimeError = (withLabel as unknown as { error?: unknown; errors?: unknown }).error
          ?? (withLabel as unknown as { errors?: unknown }).errors
        if (runtimeError) {
          if (!cancelled) {
            setRecord({ id: currentId })
          }
          return
        }

        if (!cancelled) {
          setRecord(withLabel.data?.findUnique?.item ?? { id: currentId })
        }
      } catch {
        if (!cancelled) {
          setRecord({ id: currentId })
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [client, currentId, modelName])

  return record
}

// ──────────────────────────────────────────────
// Component
// ──────────────────────────────────────────────

/**
 * RelationSelector widget for RJSF.
 *
 * Activated when a JSON Schema property contains an `x-mc.widget: "relation-selector"`
 * extension along with `x-mc.relation`:
 * ```json
 * {
 *   "type": "string",
 *   "x-mc": {
 *     "widget": "relation-selector",
 *     "belongsToFkId": "fk-123",
 *     "relation": {
 *       "databaseName": "users_db",
 *       "modelName": "User"
 *     }
 *   }
 * }
 * ```
 *
 * Features:
 * - Popover + Command combobox with 300ms debounced server-side search
 * - Each item shows `_displayName` (primary) + id (secondary, muted)
 * - Falls back to bare id when _displayName is empty
 * - Supports nullable fields (clear button shown when field is not required)
 * - Shows `_displayName (id)` on the trigger for already-selected values
 *
 * 数据访问边界：
 * - 若 formContext.createRuntimeClient 存在，通过注入的工厂访问（access adapter 模式）
 * - 否则降级为 createModelRuntimeClient（兼容旧调用方）
 */
export function RelationSelector(props: WidgetProps) {
  const { value, onChange, disabled, required, schema, formContext } = extractProps(props)

  const ctx = (formContext ?? {}) as unknown as FormContext
  const { orgName, projectSlug } = useWidgetRouteContext(ctx)
  const createRuntimeClientOverride = ctx.createRuntimeClient

  const xmc = schema['x-mc'] as XMC | undefined
  const xRelation = xmc?.relation
  const databaseName = xRelation?.databaseName ?? ''
  const modelName = xRelation?.modelName ?? ''

  const currentId: string = typeof value === 'string' && value !== '' ? value : ''

  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [openTriggerNonce, setOpenTriggerNonce] = useState(0)

  // Resolve _displayName for the currently selected id
  const currentRecord = useCurrentRecord({
    orgName,
    projectSlug,
    databaseName,
    modelName,
    currentId,
    createRuntimeClientOverride,
  })

  const enabled = open && !!orgName && !!projectSlug && !!databaseName && !!modelName

  const { records, loading, error } = useRelationSearch({
    orgName,
    projectSlug,
    databaseName,
    modelName,
    search,
    enabled,
    triggerNonce: openTriggerNonce,
    createRuntimeClientOverride,
  })

  const handleSelect = useCallback(
    (id: string) => {
      onChange(id)
      setOpen(false)
      setSearch('')
    },
    [onChange]
  )

  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation()
      onChange(null)
    },
    [onChange]
  )

  // Guard: missing x-mc.relation metadata
  if (!xRelation || !databaseName || !modelName) {
    return (
      <span className="text-sm text-destructive">
        缺少 x-mc.relation 元数据，无法渲染关联选择器
      </span>
    )
  }

  const triggerLabel = currentId
    ? currentRecord
      ? formatTriggerLabel(currentRecord)
      : currentId // show bare id while loading
    : null

  return (
    <Popover
      open={open}
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen)
        if (nextOpen) {
          setOpenTriggerNonce((n) => n + 1)
        }
      }}
    >
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            'h-10 w-full justify-between font-normal',
            !triggerLabel && 'text-muted-foreground'
          )}
        >
          <span className="flex-1 truncate text-left">
            {triggerLabel ?? '选择关联记录…'}
          </span>
          <span className="ml-2 flex shrink-0 items-center gap-1">
            {/* Clear button — only for non-required fields */}
            {!required && currentId && (
              <button
                type="button"
                aria-label="清空选择"
                onClick={handleClear}
                className="rounded p-0.5 hover:bg-muted"
              >
                <X className="size-3.5 text-muted-foreground" />
              </button>
            )}
            <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
          </span>
        </Button>
      </PopoverTrigger>

      <PopoverContent
        className="w-[var(--radix-popover-trigger-width)] p-0"
        align="start"
        sideOffset={4}
      >
        <Command shouldFilter={false}>
          <CommandInput
            placeholder="搜索记录…"
            value={search}
            onValueChange={setSearch}
          />
          <CommandList>
            {loading && (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="size-4 animate-spin text-muted-foreground" />
              </div>
            )}

            {!loading && error && (
              <div className="py-4 text-center text-sm text-destructive">
                加载关联数据失败
              </div>
            )}

            {!loading && !error && records.length === 0 && (
              <CommandEmpty>暂无可选记录</CommandEmpty>
            )}

            {!loading && !error && records.length > 0 && (
              <CommandGroup>
                {records.map((record) => (
                  <CommandItem
                    key={record.id}
                    value={record.id}
                    onSelect={handleSelect}
                    className="flex items-center gap-2"
                  >
                    <Check
                      className={cn(
                        'size-4 shrink-0',
                        currentId === record.id ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                    <span className="flex-1 truncate">
                      {typeof record._displayName === 'string' && record._displayName !== '' ? (
                        <>
                          <span className="text-foreground">{record._displayName}</span>
                          <span className="ml-1.5 text-xs text-muted-foreground">
                            {record.id}
                          </span>
                        </>
                      ) : (
                        <span className="text-muted-foreground">{record.id}</span>
                      )}
                    </span>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
