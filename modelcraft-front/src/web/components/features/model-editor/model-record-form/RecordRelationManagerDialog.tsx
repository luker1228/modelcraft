'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery } from '@apollo/client'
import type { RJSFSchema } from '@rjsf/utils'
import { createModelRuntimeClient, useProjectScopedClient } from '@bff/apollo/public'
import { buildCountQuery, buildFindManyQuery, buildUpdateMutation } from '@bff/cms/public'
import { GET_LOGICAL_FOREIGN_KEYS } from '@web/graphql'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
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
import { Check, ChevronsUpDown, Loader2, RefreshCw, Unlink } from 'lucide-react'
import { getXMC } from '@/types/xmc'
import type { LogicalForeignKey } from '@/types'

interface RecordRelationManagerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  jsonSchema: RJSFSchema | null
  orgName: string
  projectSlug: string
  modelId: string
  recordId: string | null
}

interface RelationRecord {
  id?: unknown
  _displayName?: unknown
}

interface OneToManyRelationField {
  name: string
  title: string
  relateFkId: string
  databaseName: string
  modelName: string
  format: string
}

const PAGE_SIZE = 10

function toReadableText(value: unknown): string {
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') {
    const record = value as Record<string, unknown>
    const nested = record._displayName ?? record.displayName ?? record.title ?? record.name ?? record.id
    if (nested !== undefined && nested !== value) {
      return toReadableText(nested)
    }
    try {
      return JSON.stringify(value)
    } catch {
      return '[object]'
    }
  }
  return String(value)
}

function toDisplayText(record: RelationRecord): string {
  const id = toReadableText(record.id)
  const label = toReadableText(record._displayName)
  if (label === '' && id !== '') return `空(${id})`
  if (label !== '' && id !== '') return `${label}(${id})`
  if (label !== '') return label
  return id
}

function extractOneToManyFields(schema: RJSFSchema | null): OneToManyRelationField[] {
  if (!schema?.properties) {
    return []
  }

  return Object.entries(schema.properties).flatMap(([fieldName, prop]) => {
    const fieldSchema = prop as Record<string, unknown>
    const xmc = getXMC(fieldSchema)
    const relation = xmc?.relation
    const widget = xmc?.widget
    const relateFkId = relation?.relateFkId
    const relationType = relation?.relationType
    const xmcFormat = xmc?.format ?? ''
    const databaseName = relation?.databaseName
    const modelName = relation?.modelName

    const isRelationField = xmcFormat === 'RELATION'
      || widget === 'relation-multi-readonly'
      || relationType === 'ONE_TO_MANY'
      || relationType === 'MANY_TO_ONE'

    if (fieldSchema.readOnly !== true || !isRelationField) {
      return []
    }

    if (typeof relateFkId !== 'string' || relateFkId === '') {
      return []
    }

    return [{
      name: fieldName,
      title: typeof fieldSchema.title === 'string' ? fieldSchema.title : fieldName,
      relateFkId,
      databaseName: typeof databaseName === 'string' ? databaseName : '',
      modelName: typeof modelName === 'string' ? modelName : '',
      format: xmcFormat,
    }]
  })
}

export function RecordRelationManagerDialog({
  open,
  onOpenChange,
  jsonSchema,
  orgName,
  projectSlug,
  modelId,
  recordId,
}: RecordRelationManagerDialogProps) {
  const projectClient = useProjectScopedClient(projectSlug)
  const relationFields = useMemo(() => extractOneToManyFields(jsonSchema), [jsonSchema])

  const [selectedFieldName, setSelectedFieldName] = useState('')
  const [attachRecordId, setAttachRecordId] = useState('')
  const [attachRecordDisplay, setAttachRecordDisplay] = useState('')
  const [attachPickerOpen, setAttachPickerOpen] = useState(false)
  const [attachSearch, setAttachSearch] = useState('')
  const [debouncedAttachSearch, setDebouncedAttachSearch] = useState('')
  const [attachCandidates, setAttachCandidates] = useState<RelationRecord[]>([])
  const [attachCandidatesLoading, setAttachCandidatesLoading] = useState(false)
  const [attachCandidatesError, setAttachCandidatesError] = useState<string | null>(null)
  const [loadingRecords, setLoadingRecords] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [managerError, setManagerError] = useState<string | null>(null)
  const [relatedRecords, setRelatedRecords] = useState<RelationRecord[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [totalCount, setTotalCount] = useState(0)

  useEffect(() => {
    if (!open) {
      return
    }

    const fallback = relationFields[0]?.name ?? ''
    setSelectedFieldName((prev) => (prev && relationFields.some((f) => f.name === prev) ? prev : fallback))
    setAttachRecordId('')
    setAttachRecordDisplay('')
    setAttachPickerOpen(false)
    setAttachSearch('')
    setDebouncedAttachSearch('')
    setAttachCandidates([])
    setAttachCandidatesError(null)
    setManagerError(null)
    setCurrentPage(1)
  }, [open, relationFields])

  const selectedField = useMemo(
    () => relationFields.find((field) => field.name === selectedFieldName) ?? null,
    [relationFields, selectedFieldName]
  )

  const projectScopedContext = useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])

  const { data: fkData } = useQuery<{ logicalForeignKeys: LogicalForeignKey[] }>(
    GET_LOGICAL_FOREIGN_KEYS,
    {
      skip: !open || !recordId || !selectedField,
      variables: { modelId },
      context: projectScopedContext,
      client: projectClient,
      fetchPolicy: 'network-only',
    }
  )

  const logicalForeignKeys = useMemo(
    () => fkData?.logicalForeignKeys ?? [],
    [fkData?.logicalForeignKeys]
  )

  const currentFk = useMemo(() => {
    if (!selectedField) return null
    return logicalForeignKeys.find((fk) => fk.id === selectedField.relateFkId) ?? null
  }, [logicalForeignKeys, selectedField])

  const schemaDatabaseName = useMemo(() => {
    if (!jsonSchema) {
      return ''
    }
    const raw = (jsonSchema as Record<string, unknown>)['x-databaseName']
    return typeof raw === 'string' ? raw : ''
  }, [jsonSchema])

  const targetModelName = useMemo(() => {
    if (!selectedField) {
      return ''
    }
    if (selectedField.modelName !== '') {
      return selectedField.modelName
    }
    return currentFk?.refModelName ?? ''
  }, [selectedField, currentFk])

  const targetDatabaseName = useMemo(() => {
    if (!selectedField) {
      return ''
    }
    if (selectedField.databaseName !== '') {
      return selectedField.databaseName
    }
    return schemaDatabaseName
  }, [selectedField, schemaDatabaseName])

  const targetFkField = useMemo(() => {
    if (!currentFk) {
      return null
    }

    // REVERSE（HasMany）: current model one -> target model many，targetFields 对应目标模型外键列
    if (currentFk.direction === 'REVERSE') {
      if (currentFk.targetFields.length !== 1) {
        return null
      }
      return currentFk.targetFields[0]
    }

    // 兼容兜底：如果 direction 非 REVERSE，仍优先尝试 targetFields
    if (currentFk.targetFields.length !== 1) {
      return null
    }
    return currentFk.targetFields[0]
  }, [currentFk])

  const runtimeClient = useMemo(() => {
    if (!selectedField || !targetDatabaseName || !targetModelName) return null
    return createModelRuntimeClient(
      orgName,
      projectSlug,
      targetDatabaseName,
      targetModelName,
    )
  }, [selectedField, orgName, projectSlug, targetDatabaseName, targetModelName])

  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE))

  const refreshRelatedRecords = useCallback(async () => {
    if (!runtimeClient || !selectedField || !recordId || !targetFkField) {
      setRelatedRecords([])
      setTotalCount(0)
      return
    }

    setLoadingRecords(true)
    setManagerError(null)
    try {
      const where = { [targetFkField]: { equals: recordId } }
      const skip = (currentPage - 1) * PAGE_SIZE

      const [listResult, countResult] = await Promise.all([
        runtimeClient.query<{ findMany?: { items?: RelationRecord[] } }>({
          query: buildFindManyQuery(targetModelName, ['id', '_displayName']),
          variables: {
            where,
            take: PAGE_SIZE,
            skip,
          },
          fetchPolicy: 'network-only',
        }),
        runtimeClient.query<Record<string, unknown>>({
          query: buildCountQuery(targetModelName),
          variables: { where },
          fetchPolicy: 'network-only',
        }),
      ])

      const countValue = countResult.data?.count
      const nextTotal = typeof countValue === 'number'
        ? countValue
        : (listResult.data?.findMany?.items ?? []).length

      setRelatedRecords(listResult.data?.findMany?.items ?? [])
      setTotalCount(nextTotal)
    } catch (error) {
      const message = error instanceof Error ? error.message : '加载关联记录失败'
      setManagerError(message)
      setRelatedRecords([])
      setTotalCount(0)
    } finally {
      setLoadingRecords(false)
    }
  }, [runtimeClient, selectedField, recordId, targetFkField, currentPage, targetModelName])

  useEffect(() => {
    if (!open) {
      return
    }
    void refreshRelatedRecords()
  }, [open, refreshRelatedRecords])

  const hasSingleFK = Boolean(targetFkField)

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedAttachSearch(attachSearch)
    }, 300)
    return () => clearTimeout(timer)
  }, [attachSearch])

  const refreshAttachCandidates = useCallback(async () => {
    if (!open || !hasSingleFK || !runtimeClient || !targetModelName) {
      setAttachCandidates([])
      setAttachCandidatesError(null)
      return
    }

    setAttachCandidatesLoading(true)
    setAttachCandidatesError(null)
    try {
      const search = debouncedAttachSearch.trim()
      const variables: Record<string, unknown> = { take: 20 }
      if (search !== '') {
        variables.where = { _displayName: { contains: search } }
      }

      const result = await runtimeClient.query<{ findMany?: { items?: RelationRecord[] } }>({
        query: buildFindManyQuery(targetModelName, ['id', '_label']),
        variables,
        fetchPolicy: 'network-only',
      })
      setAttachCandidates(result.data?.findMany?.items ?? [])
    } catch (error) {
      const message = error instanceof Error ? error.message : '查询目标记录失败'
      setAttachCandidatesError(message)
      setAttachCandidates([])
    } finally {
      setAttachCandidatesLoading(false)
    }
  }, [open, hasSingleFK, runtimeClient, targetModelName, debouncedAttachSearch])

  useEffect(() => {
    void refreshAttachCandidates()
  }, [refreshAttachCandidates])

  const handleAttach = useCallback(async () => {
    if (!runtimeClient || !selectedField || !recordId || !targetFkField) {
      return
    }

    const targetId = attachRecordId.trim()
    if (targetId === '') {
      setManagerError('请输入要添加的目标记录 ID。')
      return
    }

    setUpdating(true)
    setManagerError(null)
    try {
      await runtimeClient.mutate({
        mutation: buildUpdateMutation(targetModelName),
        variables: {
          where: { id: targetId },
          data: { [targetFkField]: recordId },
        },
      })

      setAttachRecordId('')
      setAttachRecordDisplay('')
      setAttachSearch('')
      setDebouncedAttachSearch('')
      setCurrentPage(1)
      await refreshRelatedRecords()
    } catch (error) {
      const message = error instanceof Error ? error.message : '添加关联失败'
      setManagerError(message)
    } finally {
      setUpdating(false)
    }
  }, [runtimeClient, selectedField, recordId, targetFkField, attachRecordId, refreshRelatedRecords, targetModelName])

  const handleDetach = useCallback(async (targetId: string) => {
    if (!runtimeClient || !selectedField || !targetFkField) {
      return
    }

    setUpdating(true)
    setManagerError(null)
    try {
      await runtimeClient.mutate({
        mutation: buildUpdateMutation(targetModelName),
        variables: {
          where: { id: targetId },
          data: { [targetFkField]: null },
        },
      })

      await refreshRelatedRecords()
    } catch (error) {
      const message = error instanceof Error ? error.message : '移除关联失败'
      setManagerError(message)
    } finally {
      setUpdating(false)
    }
  }, [runtimeClient, selectedField, targetFkField, refreshRelatedRecords, targetModelName])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[760px]">
        <DialogHeader>
          <DialogTitle>关联关系管理</DialogTitle>
          <DialogDescription>
            记录 ID：<span className="font-mono">{recordId ?? '-'}</span>
          </DialogDescription>
        </DialogHeader>

        {relationFields.length === 0 ? (
          <p className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
            当前模型没有可管理的 RELATION 关系字段。
          </p>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <p className="text-xs text-muted-foreground">选择关系字段</p>
              <Select
                value={selectedFieldName}
                onValueChange={(value) => {
                  setSelectedFieldName(value)
                  setAttachRecordId('')
                  setAttachRecordDisplay('')
                  setAttachSearch('')
                  setDebouncedAttachSearch('')
                  setAttachCandidates([])
                  setAttachCandidatesError(null)
                  setCurrentPage(1)
                  setManagerError(null)
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择一对多关系字段" />
                </SelectTrigger>
                <SelectContent>
                  {relationFields.map((field) => (
                    <SelectItem key={field.name} value={field.name}>
                      {field.title} ({field.name})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {!hasSingleFK ? (
              <p className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
                当前关系为复合外键或缺少外键信息，暂不支持直接添加/移除。
              </p>
            ) : (
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground">
                  添加关联（搜索并选择目标记录）
                </p>
                <div className="flex items-center gap-2">
                  <Popover open={attachPickerOpen} onOpenChange={setAttachPickerOpen}>
                    <PopoverTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        role="combobox"
                        aria-expanded={attachPickerOpen}
                        disabled={updating}
                        className="h-9 flex-1 justify-between font-normal"
                      >
                        <span className="truncate text-left">
                          {attachRecordDisplay || '选择目标记录…'}
                        </span>
                        <ChevronsUpDown className="size-4 opacity-50" />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
                      <Command shouldFilter={false}>
                        <CommandInput
                          placeholder="搜索目标记录…"
                          value={attachSearch}
                          onValueChange={(value) => setAttachSearch(value)}
                        />
                        <CommandList>
                          {attachCandidatesLoading ? (
                            <div className="flex items-center justify-center py-6 text-sm text-muted-foreground">
                              <Loader2 className="mr-2 size-4 animate-spin" />
                              查询中...
                            </div>
                          ) : (
                            <>
                              {attachCandidatesError && (
                                <p className="px-3 py-2 text-xs text-destructive">{attachCandidatesError}</p>
                              )}
                              <CommandEmpty>未找到匹配记录</CommandEmpty>
                              <CommandGroup>
                                {attachCandidates.map((candidate, idx) => {
                                  const candidateId = toReadableText(candidate.id)
                                  if (candidateId === '') {
                                    return null
                                  }
                                  const displayText = toDisplayText(candidate)
                                  const selected = attachRecordId === candidateId
                                  return (
                                    <CommandItem
                                      key={`${candidateId}-${idx}`}
                                      value={candidateId}
                                      onSelect={() => {
                                        setAttachRecordId(candidateId)
                                        setAttachRecordDisplay(displayText)
                                        setAttachPickerOpen(false)
                                      }}
                                    >
                                      <Check className={`mr-2 size-4 ${selected ? 'opacity-100' : 'opacity-0'}`} />
                                      <span className="truncate">{displayText}</span>
                                    </CommandItem>
                                  )
                                })}
                              </CommandGroup>
                            </>
                          )}
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                  <Button onClick={handleAttach} disabled={updating}>
                    {updating ? '处理中...' : '添加'}
                  </Button>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-foreground">已关联列表</p>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  void refreshRelatedRecords()
                }}
                disabled={loadingRecords}
              >
                <RefreshCw className={`mr-1.5 size-3.5 ${loadingRecords ? 'animate-spin' : ''}`} />
                刷新
              </Button>
            </div>

            {managerError && (
              <p className="rounded-md bg-destructive/10 p-2 text-xs text-destructive">
                {managerError}
              </p>
            )}

            <div className="max-h-80 overflow-y-auto rounded-md border border-border">
              {loadingRecords ? (
                <div className="flex items-center justify-center py-10 text-muted-foreground">
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  加载中...
                </div>
              ) : relatedRecords.length === 0 ? (
                <div className="py-10 text-center text-sm text-muted-foreground">
                  暂无关联记录
                </div>
              ) : (
                <div className="divide-y divide-border">
                  {relatedRecords.map((record, idx) => {
                    const targetId = toReadableText(record.id)
                    return (
                      <div key={`${targetId}-${idx}`} className="flex items-center justify-between gap-3 p-3">
                        <div className="min-w-0">
                          <p className="truncate text-sm text-foreground">{toDisplayText(record)}</p>
                          <p className="truncate font-mono text-xs text-muted-foreground">{targetId || '-'}</p>
                        </div>
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          className="text-muted-foreground hover:text-destructive"
                          onClick={() => {
                            if (targetId !== '') {
                              void handleDetach(targetId)
                            }
                          }}
                          disabled={updating || !hasSingleFK || targetId === ''}
                        >
                          <Unlink className="mr-1.5 size-3.5" />
                          移除
                        </Button>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>

            <div className="flex items-center justify-between">
              <p className="text-xs text-muted-foreground">
                共 {totalCount} 条，当前第 {Math.min(currentPage, totalPages)} / {totalPages} 页
              </p>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={currentPage <= 1 || loadingRecords}
                  onClick={() => setCurrentPage((prev) => Math.max(1, prev - 1))}
                >
                  上一页
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={currentPage >= totalPages || loadingRecords}
                  onClick={() => setCurrentPage((prev) => Math.min(totalPages, prev + 1))}
                >
                  下一页
                </Button>
              </div>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
