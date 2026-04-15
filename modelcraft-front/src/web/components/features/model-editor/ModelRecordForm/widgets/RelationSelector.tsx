'use client'

import React, { useState, useCallback, useEffect, useRef } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { gql } from '@apollo/client'
import { createModelRuntimeClient } from '@bff/apollo/public'
import { buildFindManyQuery } from '@bff/cms/public'
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

// ──────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────

interface XRelation {
  databaseName: string
  modelName: string
}

interface RemoteRecord {
  id: string
  __label?: string | null
}

interface FindManyQueryData {
  findMany?: {
    items: RemoteRecord[]
  } | null
}

interface FormContext {
  orgName: string
  projectSlug: string
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

/**
 * Format a remote record for display.
 * Protocol: `__label(id)`, or `空(id)` when __label is empty string.
 */
function formatRecordDisplay(r: RemoteRecord): string {
  const labelStr = typeof r.__label === 'string' ? r.__label : ''
  if (labelStr === '') {
    return `空(${r.id})`
  }
  return `${labelStr}(${r.id})`
}

/**
 * Format just the trigger label when a value is selected.
 * Shows `__label (id)` — human-readable label first, id as context.
 * Falls back to bare id if no __label.
 */
function formatTriggerLabel(r: RemoteRecord): string {
  const labelStr = typeof r.__label === 'string' ? r.__label : ''
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
}: UseRelationSearchOptions): UseRelationSearchResult {
  const [records, setRecords] = useState<RemoteRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  // Stable Apollo client for this relation — recreated only when connection params change
  const clientRef = useRef<ReturnType<typeof createModelRuntimeClient> | null>(null)
  const clientKey = `${orgName}:${projectSlug}:${databaseName}:${modelName}`
  const clientKeyRef = useRef<string>('')

  if (clientKeyRef.current !== clientKey && orgName && projectSlug && databaseName && modelName) {
    clientKeyRef.current = clientKey
    clientRef.current = createModelRuntimeClient(orgName, projectSlug, databaseName, modelName)
  }

  // Debounce search input 300ms
  const [debouncedSearch, setDebouncedSearch] = useState(search)
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(timer)
  }, [search])

  useEffect(() => {
    if (!enabled || !clientRef.current) return

    let cancelled = false
    const client = clientRef.current

    const query = buildFindManyQuery(modelName, ['id', '__label'])

    // Build where filter only when there is a search term
    const variables: Record<string, unknown> = { take: 50 }
    if (debouncedSearch.trim()) {
      variables.where = { __label: { contains: debouncedSearch.trim() } }
    }

    setLoading(true)
    setError(false)

    client
      .query<FindManyQueryData>({ query, variables, fetchPolicy: 'network-only' })
      .then(({ data }) => {
        if (!cancelled) {
          setRecords(data?.findMany?.items ?? [])
          setLoading(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setError(true)
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [enabled, debouncedSearch, modelName])

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
}

function useCurrentRecord({
  orgName,
  projectSlug,
  databaseName,
  modelName,
  currentId,
}: UseCurrentRecordOptions): RemoteRecord | null {
  const [record, setRecord] = useState<RemoteRecord | null>(null)
  const fetchedIdRef = useRef<string>('')

  const clientRef = useRef<ReturnType<typeof createModelRuntimeClient> | null>(null)
  const clientKey = `${orgName}:${projectSlug}:${databaseName}:${modelName}`
  const clientKeyRef = useRef<string>('')

  if (clientKeyRef.current !== clientKey && orgName && projectSlug && databaseName && modelName) {
    clientKeyRef.current = clientKey
    clientRef.current = createModelRuntimeClient(orgName, projectSlug, databaseName, modelName)
  }

  useEffect(() => {
    if (!currentId || fetchedIdRef.current === currentId || !clientRef.current) {
      if (!currentId) setRecord(null)
      return
    }

    const client = clientRef.current
    fetchedIdRef.current = currentId

    const query = gql`
      query FindById($where: ${modelName}UniqueWhereInput!) {
        findUnique(where: $where) {
          item {
            id
            __label
          }
        }
      }
    `

    let cancelled = false
    client
      .query<{ findUnique?: { item?: RemoteRecord | null } | null }>({
        query,
        variables: { where: { id: currentId } },
        fetchPolicy: 'cache-first',
      })
      .then(({ data }) => {
        if (!cancelled) {
          setRecord(data?.findUnique?.item ?? { id: currentId })
        }
      })
      .catch(() => {
        if (!cancelled) {
          // Fallback to bare id when lookup fails
          setRecord({ id: currentId })
        }
      })

    return () => {
      cancelled = true
    }
  }, [currentId, modelName])

  return record
}

// ──────────────────────────────────────────────
// Component
// ──────────────────────────────────────────────

/**
 * RelationSelector widget for RJSF.
 *
 * Activated when a JSON Schema property contains an `x-relation` extension:
 * ```json
 * {
 *   "type": "string",
 *   "x-belongsToFkId": "fk-123",
 *   "x-relation": {
 *     "databaseName": "users_db",
 *     "modelName": "User"
 *   }
 * }
 * ```
 *
 * Features:
 * - Popover + Command combobox with 300ms debounced server-side search
 * - Each item shows `__label` (primary) + id (secondary, muted)
 * - Falls back to bare id when __label is empty
 * - Supports nullable fields (clear button shown when field is not required)
 * - Shows `__label (id)` on the trigger for already-selected values
 */
export function RelationSelector(props: WidgetProps) {
  const { value, onChange, disabled, required, schema, formContext } = extractProps(props)

  const ctx = (formContext ?? {}) as unknown as FormContext
  const { orgName, projectSlug } = ctx

  const xRelation = schema['x-relation'] as XRelation | undefined
  const databaseName = xRelation?.databaseName ?? ''
  const modelName = xRelation?.modelName ?? ''

  const currentId: string = typeof value === 'string' && value !== '' ? value : ''

  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')

  // Resolve __label for the currently selected id
  const currentRecord = useCurrentRecord({
    orgName,
    projectSlug,
    databaseName,
    modelName,
    currentId,
  })

  const enabled = open && !!orgName && !!projectSlug && !!databaseName && !!modelName

  const { records, loading, error } = useRelationSearch({
    orgName,
    projectSlug,
    databaseName,
    modelName,
    search,
    enabled,
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

  // Guard: missing x-relation metadata
  if (!xRelation || !databaseName || !modelName) {
    return (
      <span className="text-sm text-destructive">
        缺少 x-relation 元数据，无法渲染关联选择器
      </span>
    )
  }

  const triggerLabel = currentId
    ? currentRecord
      ? formatTriggerLabel(currentRecord)
      : currentId // show bare id while loading
    : null

  return (
    <Popover open={open} onOpenChange={setOpen}>
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
              <span
                role="button"
                aria-label="清空选择"
                onClick={handleClear}
                className="rounded p-0.5 hover:bg-muted"
              >
                <X className="size-3.5 text-muted-foreground" />
              </span>
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
                      {typeof record.__label === 'string' && record.__label !== '' ? (
                        <>
                          <span className="text-foreground">{record.__label}</span>
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
