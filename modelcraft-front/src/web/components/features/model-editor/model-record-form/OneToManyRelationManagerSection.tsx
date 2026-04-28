'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery } from '@apollo/client'
import type { RJSFSchema } from '@rjsf/utils'
import { createModelRuntimeClient, useProjectScopedClient, useProjectScopedContext } from '@api-client/apollo/public'
import { buildFindManyQuery, buildUpdateMutation } from '@api-client/cms/public'
import { GET_LOGICAL_FOREIGN_KEYS } from '@/api-client/model'
import { Badge } from '@web/components/ui/badge'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Loader2, Link2, RefreshCw, Unlink } from 'lucide-react'
import { getXMC } from '@/types/xmc'
import type { LogicalForeignKey } from '@/types'

interface OneToManyRelationManagerSectionProps {
  jsonSchema: RJSFSchema
  initialData?: Record<string, unknown>
  recordId?: string
  orgName: string
  projectSlug: string
  modelId: string
}

interface RelationRecord {
  id?: string | null
  _displayName?: string | null
}

interface OneToManyRelationField {
  name: string
  title: string
  relateFkId: string
  databaseName: string
  modelName: string
  relationType: 'ONE_TO_MANY' | 'MANY_TO_ONE' | ''
}

function toDisplayText(record: RelationRecord): string {
  const id = typeof record.id === 'string' ? record.id : ''
  const label = typeof record._displayName === 'string' ? record._displayName : ''
  if (label === '' && id !== '') return `空(${id})`
  if (label !== '' && id !== '') return `${label}(${id})`
  if (label !== '') return label
  return id
}

function toRelationRecords(rawValue: unknown): RelationRecord[] {
  if (!Array.isArray(rawValue)) {
    return []
  }

  return rawValue.filter((item): item is RelationRecord => typeof item === 'object' && item !== null)
}

function extractOneToManyFields(schema: RJSFSchema): OneToManyRelationField[] {
  if (!schema.properties) {
    return []
  }

  return Object.entries(schema.properties).flatMap(([fieldName, prop]) => {
    const fieldSchema = prop as Record<string, unknown>
    const xmc = getXMC(fieldSchema)
    const relation = xmc?.relation
    const widget = xmc?.widget
    const relateFkId = relation?.relateFkId
    const relationType = relation?.relationType
    const xmcFormat = xmc?.format
    const databaseName = relation?.databaseName
    const modelName = relation?.modelName

    const isOneToMany =
      relationType === 'ONE_TO_MANY'
      || widget === 'relation-multi-readonly'
      || (xmcFormat === 'RELATION' && relationType !== 'MANY_TO_ONE')

    if (fieldSchema.readOnly !== true || !isOneToMany) {
      return []
    }

    if (
      typeof relateFkId !== 'string'
      || typeof databaseName !== 'string'
      || typeof modelName !== 'string'
      || databaseName === ''
      || modelName === ''
    ) {
      return []
    }

    const title = typeof fieldSchema.title === 'string' ? fieldSchema.title : fieldName
    return [{
      name: fieldName,
      title,
      relateFkId,
      databaseName,
      modelName,
      relationType: relationType ?? '',
    }]
  })
}

export function OneToManyRelationManagerSection({
  jsonSchema,
  initialData,
  recordId,
  orgName,
  projectSlug,
  modelId,
}: OneToManyRelationManagerSectionProps) {
  const fields = useMemo(() => extractOneToManyFields(jsonSchema), [jsonSchema])
  const projectClient = useProjectScopedClient(projectSlug)

  const projectScopedContext = useProjectScopedContext(orgName, projectSlug)

  const { data: fkData } = useQuery<{ logicalForeignKeys: LogicalForeignKey[] }>(
    GET_LOGICAL_FOREIGN_KEYS,
    {
      skip: !recordId || fields.length === 0,
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
  const [managingField, setManagingField] = useState<OneToManyRelationField | null>(null)
  const [loadingRecords, setLoadingRecords] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [managerError, setManagerError] = useState<string | null>(null)
  const [attachRecordId, setAttachRecordId] = useState('')
  const [relatedRecords, setRelatedRecords] = useState<RelationRecord[]>([])

  const currentFk = useMemo(() => {
    if (!managingField) return null
    return logicalForeignKeys.find((fk) => fk.id === managingField.relateFkId) ?? null
  }, [logicalForeignKeys, managingField])

  const targetFkField = useMemo(() => {
    if (!currentFk || currentFk.targetFields.length !== 1) {
      return null
    }
    return currentFk.targetFields[0]
  }, [currentFk])

  const runtimeClient = useMemo(() => {
    if (!managingField) return null
    return createModelRuntimeClient(
      orgName,
      projectSlug,
      managingField.databaseName,
      managingField.modelName,
    )
  }, [managingField, orgName, projectSlug])

  const refreshRelatedRecords = useCallback(async () => {
    if (!runtimeClient || !managingField || !recordId) {
      return
    }

    if (!targetFkField) {
      setManagerError('当前关系为复合外键或缺少外键信息，暂不支持在此直接维护。')
      setRelatedRecords([])
      return
    }

    setLoadingRecords(true)
    setManagerError(null)
    try {
      const result = await runtimeClient.query<{ findMany?: { items?: RelationRecord[] } }>({
        query: buildFindManyQuery(managingField.modelName, ['id', '_displayName', targetFkField]),
        variables: {
          where: { [targetFkField]: { equals: recordId } },
          take: 100,
        },
        fetchPolicy: 'network-only',
      })

      setRelatedRecords(result.data?.findMany?.items ?? [])
    } catch (error) {
      const message = error instanceof Error ? error.message : '加载关联记录失败'
      setManagerError(message)
      setRelatedRecords([])
    } finally {
      setLoadingRecords(false)
    }
  }, [runtimeClient, managingField, recordId, targetFkField])

  const handleOpenManager = useCallback((field: OneToManyRelationField) => {
    setManagingField(field)
    setAttachRecordId('')
    setManagerError(null)
    setRelatedRecords([])
  }, [])

  const closeManager = useCallback(() => {
    setManagingField(null)
    setAttachRecordId('')
    setManagerError(null)
    setRelatedRecords([])
  }, [])

  const handleAttach = useCallback(async () => {
    if (!runtimeClient || !managingField || !recordId || !targetFkField) {
      return
    }

    const targetId = attachRecordId.trim()
    if (targetId === '') {
      setManagerError('请输入要绑定的目标记录 ID。')
      return
    }

    setUpdating(true)
    setManagerError(null)
    try {
      await runtimeClient.mutate({
        mutation: buildUpdateMutation(managingField.modelName),
        variables: {
          where: { id: targetId },
          data: { [targetFkField]: recordId },
        },
      })

      setAttachRecordId('')
      await refreshRelatedRecords()
    } catch (error) {
      const message = error instanceof Error ? error.message : '绑定关联失败'
      setManagerError(message)
    } finally {
      setUpdating(false)
    }
  }, [runtimeClient, managingField, recordId, targetFkField, attachRecordId, refreshRelatedRecords])

  const handleDetach = useCallback(async (targetId: string) => {
    if (!runtimeClient || !managingField || !targetFkField) {
      return
    }

    setUpdating(true)
    setManagerError(null)
    try {
      await runtimeClient.mutate({
        mutation: buildUpdateMutation(managingField.modelName),
        variables: {
          where: { id: targetId },
          data: { [targetFkField]: null },
        },
      })

      await refreshRelatedRecords()
    } catch (error) {
      const message = error instanceof Error ? error.message : '解绑关联失败'
      setManagerError(message)
    } finally {
      setUpdating(false)
    }
  }, [runtimeClient, managingField, targetFkField, refreshRelatedRecords])

  useEffect(() => {
    if (!managingField) {
      return
    }
    void refreshRelatedRecords()
  }, [managingField, refreshRelatedRecords])

  if (!recordId || fields.length === 0) {
    return null
  }

  return (
    <>
      <div className="mt-4 rounded-md border border-border bg-muted/20 p-3">
        <div className="mb-3 flex items-center gap-2">
          <Link2 className="size-4 text-primary" />
          <h3 className="text-sm font-semibold text-foreground">关联关系管理（一对多）</h3>
        </div>

        <div className="space-y-3">
          {fields.map((field) => {
            const values = toRelationRecords(initialData?.[field.name])
            return (
              <div key={field.name} className="rounded-md border border-border/70 bg-background p-2.5">
                <div className="mb-2 flex items-center justify-between gap-2">
                  <div>
                    <p className="text-sm font-medium text-foreground">{field.title}</p>
                    <p className="font-mono text-xs text-muted-foreground">{field.name}</p>
                  </div>
                  <Button size="sm" variant="outline" onClick={() => handleOpenManager(field)}>
                    关联关系管理
                  </Button>
                </div>

                {values.length === 0 ? (
                  <p className="text-xs text-muted-foreground">暂无关联记录（只读展示）</p>
                ) : (
                  <div className="flex flex-wrap gap-1.5">
                    {values.map((record, idx) => (
                      <Badge key={`${record.id ?? 'unknown'}-${idx}`} variant="secondary" className="font-mono text-[11px]">
                        {toDisplayText(record)}
                      </Badge>
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      <Dialog
        open={Boolean(managingField)}
        onOpenChange={(open) => {
          if (!open) closeManager()
        }}
      >
        <DialogContent className="sm:max-w-[680px]">
          <DialogHeader>
            <DialogTitle>关联关系管理</DialogTitle>
            <DialogDescription>
              {managingField
                ? `字段 ${managingField.title}（${managingField.name}）`
                : '管理当前记录的一对多关联'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {!targetFkField ? (
              <p className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
                当前关系为复合外键或缺少外键信息，暂不支持在此直接维护。
              </p>
            ) : (
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground">
                  目标模型 <span className="font-mono">{managingField?.modelName}</span>，通过字段
                  <span className="mx-1 font-mono">{targetFkField}</span>关联当前记录。
                </p>
                <div className="flex items-center gap-2">
                  <Input
                    value={attachRecordId}
                    onChange={(event) => setAttachRecordId(event.target.value)}
                    placeholder="输入目标记录 ID 后绑定"
                    className="font-mono"
                    disabled={updating}
                  />
                  <Button onClick={handleAttach} disabled={updating}>
                    {updating ? '处理中...' : '绑定'}
                  </Button>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-foreground">当前已关联记录</p>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  void refreshRelatedRecords()
                }}
                disabled={loadingRecords || !targetFkField}
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

            <div className="max-h-72 overflow-y-auto rounded-md border border-border">
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
                  {relatedRecords.map((record) => {
                    const recordIdValue = typeof record.id === 'string' ? record.id : ''
                    return (
                      <div key={recordIdValue} className="flex items-center justify-between gap-3 p-3">
                        <div className="min-w-0">
                          <p className="truncate text-sm text-foreground">{toDisplayText(record)}</p>
                          <p className="truncate font-mono text-xs text-muted-foreground">{recordIdValue}</p>
                        </div>
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          className="text-muted-foreground hover:text-destructive"
                          onClick={() => {
                            if (recordIdValue !== '') {
                              void handleDetach(recordIdValue)
                            }
                          }}
                          disabled={updating || recordIdValue === ''}
                        >
                          <Unlink className="mr-1.5 size-3.5" />
                          解绑
                        </Button>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={closeManager}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
