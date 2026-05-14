'use client'

import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { toast } from 'sonner'
import {
  createEndUserScopedClient,
  createEndUserModelRuntimeClient,
} from '@api-client/apollo/public'
import { getEndUserToken } from '@api-client/end-user/public'
import { ModelRecordForm } from '@web/components/features/model-editor/model-record-form/index'
import { ModelRecordTable } from '@web/components/shared/data-workspace/ModelRecordTable'
import type { ModelRecordTableFieldInfo, ModelRecordTableRow } from '@web/components/shared/data-workspace/ModelRecordTable'
import { getFieldProtocols } from '@web/components/features/model-editor/model-record-form/runtime/field-protocol'
import {
  buildFindManyQuery,
  buildFindUniqueQuery,
  buildDeleteMutation,
  buildCreateMutation,
  buildUpdateMutation,
  extractFieldsFromSchema,
  extractWritableFieldNamesFromSchema,
  sanitizeMutationInputData,
} from '@api-client/cms/public'
import type { FieldDefinition } from '@api-client/cms/public'
import { NOOP_MUTATION, NOOP_QUERY } from '@/api-client/noop'
import { GET_MODEL_RECORD_WORKSPACE_END_USER } from '@/api-client/model/graphql-docs.end-user'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import {
  Filter,
  List,
  Plus,
  Edit,
  Loader2,
  Search,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/shared/utils'
import { FilterPanel } from './FilterPanel'
import { getFilterCount } from './filter-utils'
import { getXMC } from '@/types/xmc'
import { RecordAccessAdapterProvider, type RecordAccessAdapter } from '@web/components/features/model-editor/model-record-form/access-adapter'

export interface EndUserRecordWorkspaceProps {
  modelId: string
  projectSlug: string
  orgName: string
  refreshToken?: number
}

interface GetModelQueryData {
  model?: {
    model?: {
      id: string
      name: string
      title?: string | null
      description?: string | null
      databaseName?: string | null
      createdVia?: 'NEW' | 'IMPORTED' | null
      jsonSchema?: string | null
      fields?: Array<{
        name: string
        isDeprecated?: boolean | null
      }> | null
    } | null
    error?: {
      message?: string | null
    } | null
  } | null
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function deriveStorageHintFromSchemaProp(prop: Record<string, unknown>): string | null {
  const xmc = getXMC(prop)
  if (typeof xmc?.storageHint === 'string' && xmc.storageHint.trim() !== '') {
    return xmc.storageHint
  }

  const schemaType = typeof prop.type === 'string' ? prop.type.toUpperCase() : ''
  if (schemaType === 'STRING') return 'TEXT'
  if (schemaType === 'INTEGER' || schemaType === 'NUMBER') return 'NUMBER'
  if (schemaType === 'BOOLEAN') return 'BOOLEAN'
  return null
}

function isEnumSchemaProp(prop: Record<string, unknown>): boolean {
  const xmc = getXMC(prop)
  return (
    Array.isArray(prop.enum) ||
    xmc?.widget === 'enum-select'
  )
}

function resolveEnumLabelFieldName(prop: Record<string, unknown>): string {
  const xmc = getXMC(prop)
  const configured = xmc?.enum?.labelFieldName?.trim()
  return configured ?? ''
}

/**
 * EndUserRecordWorkspace — 终端用户数据工作区。
 *
 * 能力范围（record CRUD）：
 * - record query / create / edit / delete
 *
 * 刻意**不包含**：
 * - 字段插入（InsertMenu 的"插入列"选项）
 * - 字段废弃/取消废弃/删除（字段生命周期维护）
 * - 关系维护对话框（RecordRelationManagerDialog）
 * - SQL 控制台入口
 *
 * 使用 end-user scoped client 作为所有数据端点，
 * 身份由当前 end-user token 决定，不依赖 workspaceMode。
 */
export default function EndUserRecordWorkspace({
  modelId,
  projectSlug,
  orgName,
  refreshToken = 0,
}: EndUserRecordWorkspaceProps) {
  const endUserContext = useMemo(
    () => ({ uri: `/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}` }),
    [orgName, projectSlug]
  )

  const endUserToken = getEndUserToken()

  const managementClient = useMemo(() => {
    if (!endUserToken) return null
    return createEndUserScopedClient(orgName, projectSlug, endUserToken)
  }, [orgName, projectSlug, endUserToken])

  const accessAdapter = useMemo<RecordAccessAdapter | null>(() => {
    if (!managementClient) return null
    const token = endUserToken
    return {
      managementClient,
      managementContext: endUserContext,
      createRuntimeClient: (databaseName: string, modelName: string) => {
        if (!token) {
          throw new Error('EndUserRecordWorkspace: end-user token is not available')
        }
        return createEndUserModelRuntimeClient(orgName, projectSlug, databaseName, modelName, token)
      },
    }
  }, [managementClient, endUserContext, orgName, projectSlug, endUserToken])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null)

  const [createDataOpen, setCreateDataOpen] = useState(false)
  const [createSaving, setCreateSaving] = useState(false)

  const [editDataOpen, setEditDataOpen] = useState(false)
  const [editItemId, setEditItemId] = useState<string | null>(null)
  const [editFormData, setEditFormData] = useState<Record<string, unknown>>({})
  const [editSaving, setEditSaving] = useState(false)
  const [editLoading, setEditLoading] = useState(false)

  const [searchKeyword, setSearchKeyword] = useState('')

  // --- Filter state ---
  const [filterOpen, setFilterOpen] = useState(false)
  // Draft: changes on every keystroke, does NOT trigger a query
  const [whereJsonDraft, setWhereJsonDraft] = useState<string>('')
  // Committed: only updated on "应用筛选", drives the actual GraphQL where clause
  const [whereJsonCommitted, setWhereJsonCommitted] = useState<string | null>(null)

  const whereInput = useMemo(() => {
    if (!whereJsonCommitted?.trim()) return undefined
    try {
      return JSON.parse(whereJsonCommitted) as Record<string, unknown>
    } catch {
      return undefined
    }
  }, [whereJsonCommitted])

  function handleApplyFilter() {
    const trimmed = whereJsonDraft.trim()
    setWhereJsonCommitted(trimmed || null)
  }

  function handleClearFilter() {
    // Atomically clear both draft and committed state so the filter is removed immediately.
    // We cannot call setWhereJsonDraft('') then handleApplyFilter() because React would
    // batch the updates and handleApplyFilter would still see the old draft value.
    setWhereJsonDraft('')
    setWhereJsonCommitted(null)
  }

  const filterCount = getFilterCount(whereJsonCommitted)
  const hasActiveFilter = filterCount !== null
  // --- End filter state ---

  const { data: modelData, loading: modelLoading, refetch: refetchModel } = useQuery<GetModelQueryData, { id: string }>(
    GET_MODEL_RECORD_WORKSPACE_END_USER,
    {
      client: managementClient!,
      skip: !managementClient,
      variables: { id: modelId },
      context: endUserContext,
    }
  )

  const model = modelData?.model?.model
  const modelError = modelData?.model?.error
  const modelName = model?.name
  const isManagedReadOnlyModel = model?.createdVia === 'IMPORTED'

  const canCreateRecord = !isManagedReadOnlyModel
  const canEditRecord = !isManagedReadOnlyModel
  const canDeleteRecord = !isManagedReadOnlyModel

  const showManagedReadonlyToast = useCallback(() => {
    toast.warning('托管模型仅支持查看，不支持写入')
  }, [])

  const runtimeClient = useMemo(() => {
    if (!endUserToken) return null
    const dbName = model?.databaseName
    const mName = model?.name
    if (!dbName || !mName) return null
    return createEndUserModelRuntimeClient(orgName, projectSlug, dbName, mName, endUserToken)
  }, [orgName, projectSlug, endUserToken, model?.databaseName, model?.name])

  const jsonSchema = useMemo<Record<string, unknown> | null>(() => {
    if (!model?.jsonSchema) return null
    try {
      return JSON.parse(model.jsonSchema) as Record<string, unknown>
    } catch (error) {
      console.error('Failed to parse schema:', error)
      return null
    }
  }, [model?.jsonSchema])

  const runtimeFields = useMemo(() => {
    if (jsonSchema?.properties) {
      const props = jsonSchema.properties as Record<string, unknown>
      const schemaFieldDefs: FieldDefinition[] = Object.entries(props).flatMap(([name, rawProp]) => {
        if (!isRecord(rawProp)) return []
        const prop = rawProp
        const labelFieldName = resolveEnumLabelFieldName(prop)
        const hasLabelFieldInSchema = Object.prototype.hasOwnProperty.call(props, labelFieldName)
        const schemaType = typeof prop.type === 'string' ? prop.type.toUpperCase() : undefined
        const format = typeof prop.format === 'string' ? prop.format.toUpperCase() : undefined
        const baseDef: FieldDefinition = {
          name,
          type: schemaType ?? 'string',
          format,
          schemaType,
          storageHint: deriveStorageHintFromSchemaProp(prop) ?? undefined,
        }

        if (!isEnumSchemaProp(prop)) {
          return [baseDef]
        }

        if (!labelFieldName || labelFieldName === name || hasLabelFieldInSchema) {
          return [baseDef]
        }

        return [{ ...baseDef }, { name: labelFieldName, type: 'string', schemaType: 'STRING', storageHint: 'TEXT' }]
      })

      if (schemaFieldDefs.length > 0) {
        if (!schemaFieldDefs.some((field) => field.name === 'id')) {
          schemaFieldDefs.unshift({ name: 'id', type: 'string', schemaType: 'STRING' })
        }
        return schemaFieldDefs
      }
    }

    if (!jsonSchema) return ['id']
    return extractFieldsFromSchema(jsonSchema as { properties?: Record<string, unknown> })
  }, [jsonSchema])

  const writableFieldNames = useMemo(
    () => extractWritableFieldNamesFromSchema(jsonSchema as { properties?: Record<string, unknown> } | null | undefined),
    [jsonSchema]
  )

  const tableFieldInfos = useMemo<ModelRecordTableFieldInfo[]>(() => {
    if (jsonSchema?.properties) {
      const props = jsonSchema.properties as Record<string, unknown>
      return Object.entries(props).flatMap(([name, rawProp]) => {
        if (!isRecord(rawProp)) return []
        const prop = rawProp
        const labelFieldName = resolveEnumLabelFieldName(prop)
        const hasLabelFieldInSchema = Object.prototype.hasOwnProperty.call(props, labelFieldName)
        const xmc = getXMC(prop)
        const baseInfo: ModelRecordTableFieldInfo = {
          name,
          title: typeof prop.title === 'string' ? prop.title : null,
          isPrimary: xmc?.isPrimary === true,
          isDeprecated: false,
          storageHint: deriveStorageHintFromSchemaProp(prop),
          schemaType: typeof prop.type === 'string' ? prop.type.toUpperCase() : null,
        }

        if (!isEnumSchemaProp(prop)) {
          return [baseInfo]
        }

        if (!labelFieldName || labelFieldName === name || hasLabelFieldInSchema) {
          return [baseInfo]
        }

        return [
          baseInfo,
          {
            name: labelFieldName,
            title: null,
            isPrimary: false,
            isDeprecated: false,
            storageHint: 'TEXT',
            schemaType: 'STRING',
          },
        ]
      })
    }

    return []
  }, [jsonSchema])

  const displayFields = useMemo(() => {
    if (tableFieldInfos.length > 0) {
      return tableFieldInfos.map((field) => field.name)
    }
    return runtimeFields.map((f) => typeof f === 'string' ? f : f.name)
  }, [tableFieldInfos, runtimeFields])

  const propByName = useMemo(() => {
    if (!jsonSchema) return {}
    const protocols = getFieldProtocols(jsonSchema as import('@rjsf/utils').RJSFSchema)
    return Object.fromEntries(protocols.map(({ name, prop }) => [name, prop]))
  }, [jsonSchema])

  const getFieldInfo = useCallback(
    (fieldName: string): ModelRecordTableFieldInfo | null => {
      return tableFieldInfos.find((field) => field.name === fieldName) ?? null
    },
    [tableFieldInfos]
  )

  const getFieldTypeDisplay = useCallback((fieldInfo: ModelRecordTableFieldInfo | null) => {
    if (!fieldInfo) return ''
    return fieldInfo.storageHint || fieldInfo.schemaType || ''
  }, [])

  const findManyQuery = useMemo(() => {
    if (!modelName) return null
    return buildFindManyQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  const {
    data: contentData,
    loading: contentLoading,
    refetch,
  } = useQuery<Record<string, unknown>>(findManyQuery || NOOP_QUERY, {
    client: runtimeClient!,
    skip: !findManyQuery || !runtimeClient,
    variables: {
      take: 50,
      skip: 0,
      where: whereInput,
    },
  })

  const deleteMutation = useMemo(() => {
    if (!modelName) return null
    return buildDeleteMutation(modelName)
  }, [modelName])

  const [deleteContent] = useMutation(deleteMutation || NOOP_MUTATION, {
    client: runtimeClient!,
    onCompleted: () => {
      refetch()
      setDeleteDialogOpen(false)
      setDeleteItemId(null)
    },
  })

  const createMutation = useMemo(() => {
    if (!modelName) return null
    return buildCreateMutation(modelName)
  }, [modelName])

  const [createContent] = useMutation(createMutation || NOOP_MUTATION, {
    client: runtimeClient!,
    onCompleted: () => {
      refetch()
      setCreateDataOpen(false)
      setCreateSaving(false)
    },
    onError: (error) => {
      console.error('Failed to create content:', error)
      toast.error('创建数据失败: ' + error.message)
      setCreateSaving(false)
    },
  })

  const findUniqueQuery = useMemo(() => {
    if (!modelName) return null
    return buildFindUniqueQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  const updateMutation = useMemo(() => {
    if (!modelName) return null
    return buildUpdateMutation(modelName)
  }, [modelName])

  const [updateContent] = useMutation(updateMutation || NOOP_MUTATION, {
    client: runtimeClient!,
    onCompleted: () => {
      refetch()
      setEditDataOpen(false)
      setEditFormData({})
      setEditItemId(null)
      setEditSaving(false)
    },
    onError: (error) => {
      console.error('Failed to update content:', error)
      toast.error('更新数据失败: ' + error.message)
      setEditSaving(false)
    },
  })

  const contentList: ModelRecordTableRow[] = useMemo(
    () => (Array.isArray((contentData as Record<string, unknown> | undefined)?.findMany && ((contentData as Record<string, unknown>).findMany as Record<string, unknown>)?.items)
      ? ((contentData as Record<string, unknown>).findMany as Record<string, unknown>).items as ModelRecordTableRow[]
      : []),
    [contentData]
  )

  useEffect(() => {
    if (!refreshToken) return
    void refetchModel()
    void refetch()
  }, [refreshToken, refetchModel, refetch])

  const handleDelete = async () => {
    if (!deleteItemId) return
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }

    try {
      await deleteContent({
        variables: {
          where: { id: deleteItemId },
        },
      })
    } catch (error) {
      console.error('Failed to delete content:', error)
      toast.error('删除失败')
    }
  }

  const handleEdit = async (id: string) => {
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }
    if (!runtimeClient || !findUniqueQuery) return

    setEditItemId(id)
    setEditLoading(true)
    setEditDataOpen(true)

    try {
      const { data } = await runtimeClient.query<Record<string, unknown>>({
        query: findUniqueQuery,
        variables: { where: { id } },
        fetchPolicy: 'network-only',
      })

      const item = (data as Record<string, unknown> | undefined)?.findUnique && ((data as Record<string, unknown>).findUnique as Record<string, unknown>)?.item
      if (isRecord(item)) {
        const schemaDrivenData = writableFieldNames.reduce<Record<string, unknown>>((formData, fieldName) => {
          formData[fieldName] = item[fieldName] ?? ''
          return formData
        }, {})
        setEditFormData(schemaDrivenData)
      }
    } catch (error) {
      console.error('Failed to fetch content:', error)
      toast.error('获取数据失败')
      setEditDataOpen(false)
    } finally {
      setEditLoading(false)
    }
  }

  const handleCreate = () => {
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }
    setCreateDataOpen(true)
  }

  const confirmDelete = (id: string) => {
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }
    setDeleteItemId(id)
    setDeleteDialogOpen(true)
  }

  const filteredContentList = useMemo(() => {
    const keyword = searchKeyword.trim().toLowerCase()
    if (!keyword) {
      return contentList
    }

    return contentList.filter((row) => {
      return Object.values(row).some((value) => {
        if (value == null) return false
        if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
          return String(value).toLowerCase().includes(keyword)
        }
        return false
      })
    })
  }, [contentList, searchKeyword])

  if (modelLoading || !managementClient) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!model) {
    const errorMessage = modelError?.message || '模型未找到'
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <p className="mb-2 text-destructive">{errorMessage}</p>
        </div>
      </div>
    )
  }

  return (
    <RecordAccessAdapterProvider value={accessAdapter}>
      <div className="flex h-full flex-col">
        {isManagedReadOnlyModel && (
          <div className="border-b border-border bg-card px-4 py-2">
            <Alert variant="warning" className="py-2">
              <AlertDescription className="text-xs">
                当前为托管模型，数据为只读模式，新增/编辑/删除操作已禁用。
              </AlertDescription>
            </Alert>
          </div>
        )}

        {/* 顶部搜索栏 */}
        <div className="flex items-center justify-between border-b border-border bg-card px-4 py-3">
          <div className="flex w-full max-w-xl items-center gap-2">
            <Search className="size-4 text-muted-foreground" />
            <Input
              value={searchKeyword}
              onChange={(event) => setSearchKeyword(event.target.value)}
              placeholder={`搜索 ${model.title || model.name} 的记录...`}
              className="h-8"
            />
          </div>
          <div className="text-xs text-muted-foreground">
            {filteredContentList.length} / {contentList.length} 条
          </div>
        </div>

        {/* 工具栏 */}
        <div className="flex h-10 items-center justify-between gap-2 overflow-x-auto border-b border-border bg-card p-1.5">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setFilterOpen((open) => !open)}
                className={cn(
                  'h-[26px] px-2.5 text-xs font-normal',
                  filterOpen
                    ? 'border border-primary text-primary ring-2 ring-primary/20'
                    : hasActiveFilter
                      ? 'border border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <Filter className="mr-1.5 size-3.5" />
                <span>筛选</span>
                {filterCount !== null && (
                  <span className="ml-1.5 flex size-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
                    {filterCount}
                  </span>
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-[26px] border-transparent px-2.5 text-xs font-normal text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <List className="mr-1.5 size-3.5" />
                <span>排序</span>
              </Button>
            </div>

            <div className="h-5 w-px bg-border" />

            <Button
              size="sm"
              className="h-[26px] border-0 bg-primary px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-primary/90"
              onClick={handleCreate}
              disabled={!canCreateRecord}
            >
              <Plus className="mr-1.5 size-3.5" />
              <span>添加数据</span>
            </Button>
          </div>

          <div className="flex items-center gap-2">
            {!contentLoading && (
              <span className="text-xs text-muted-foreground">
                {filteredContentList.length} 条记录
              </span>
            )}
            <Button
              variant="outline"
              size="sm"
              className="h-[26px] w-7 p-0"
              onClick={() => refetch()}
              disabled={contentLoading}
              title="刷新数据"
            >
              <RefreshCw className={`size-3.5 ${contentLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* 筛选面板（工具栏下方内联展开） */}
        {filterOpen && (
          <FilterPanel
            fields={runtimeFields}
            whereJsonDraft={whereJsonDraft}
            onWhereJsonDraftChange={setWhereJsonDraft}
            onApply={handleApplyFilter}
            onClear={handleClearFilter}
          />
        )}

        {/* 数据表格 */}
        <ModelRecordTable
          contentLoading={contentLoading}
          contentList={filteredContentList}
          displayFields={displayFields}
          getFieldInfo={getFieldInfo}
          getFieldTypeDisplay={getFieldTypeDisplay}
          propByName={propByName}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={confirmDelete}
          canManageFieldLifecycle={false}
          canCreateRecord={canCreateRecord}
          canEditRecord={canEditRecord}
          canDeleteRecord={canDeleteRecord}
        />

        {/* 新增数据侧边栏 */}
        <Sheet open={createDataOpen} onOpenChange={setCreateDataOpen}>
          <SheetContent side="right" className="w-[450px] overflow-y-auto sm:max-w-[500px]">
            <SheetHeader>
              <SheetTitle className="flex items-center gap-2">
                <Plus className="size-5 text-primary" />
                添加数据
              </SheetTitle>
              <SheetDescription>
                向 <span className="font-mono text-primary">{model.name}</span> 添加一条新记录
              </SheetDescription>
            </SheetHeader>

            {jsonSchema && (
              <ModelRecordForm
                jsonSchema={jsonSchema as import('@rjsf/utils').RJSFSchema}
                onSubmit={async (data) => {
                  if (isManagedReadOnlyModel) {
                    showManagedReadonlyToast()
                    throw new Error('托管模型仅支持查看')
                  }
                  setCreateSaving(true)
                  try {
                    const sanitizedData = sanitizeMutationInputData(data, writableFieldNames)
                    await createContent({ variables: { data: sanitizedData } })
                    setCreateDataOpen(false)
                  } catch (error) {
                    throw error
                  } finally {
                    setCreateSaving(false)
                  }
                }}
                onCancel={() => setCreateDataOpen(false)}
                isSubmitting={createSaving}
                orgName={orgName}
                projectSlug={projectSlug}
                clusterName=""
                databaseName={model.databaseName ?? ''}
                modelId={modelId}
                workspaceMode="end_user"
              />
            )}
          </SheetContent>
        </Sheet>

        {/* 编辑数据侧边栏 */}
        <Sheet open={editDataOpen} onOpenChange={setEditDataOpen}>
          <SheetContent side="right" className="w-[450px] overflow-y-auto sm:max-w-[500px]">
            <SheetHeader>
              <SheetTitle className="flex items-center gap-2">
                <Edit className="size-5 text-primary" />
                编辑数据
              </SheetTitle>
              <SheetDescription>
                编辑 <span className="font-mono text-primary">{model.name}</span> 中的记录
                {editItemId && (
                  <span className="mt-1 block text-xs text-muted-foreground">ID: {editItemId}</span>
                )}
              </SheetDescription>
            </SheetHeader>

            {editLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="size-8 animate-spin text-muted-foreground" />
              </div>
            ) : jsonSchema && (
              <ModelRecordForm
                jsonSchema={jsonSchema as import('@rjsf/utils').RJSFSchema}
                initialData={editFormData}
                recordId={editItemId ?? undefined}
                onSubmit={async (data) => {
                  if (isManagedReadOnlyModel) {
                    showManagedReadonlyToast()
                    throw new Error('托管模型仅支持查看')
                  }
                  setEditSaving(true)
                  try {
                    const sanitizedData = sanitizeMutationInputData(data, writableFieldNames)
                    await updateContent({ variables: { where: { id: editItemId }, data: sanitizedData } })
                    setEditDataOpen(false)
                  } catch (error) {
                    throw error
                  } finally {
                    setEditSaving(false)
                  }
                }}
                onCancel={() => setEditDataOpen(false)}
                isSubmitting={editSaving}
                orgName={orgName}
                projectSlug={projectSlug}
                clusterName=""
                databaseName={model.databaseName ?? ''}
                modelId={modelId}
                workspaceMode="end_user"
              />
            )}
          </SheetContent>
        </Sheet>

        {/* 删除确认对话框 */}
        <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>确认删除</DialogTitle>
              <DialogDescription>确定要删除这条数据吗？此操作不可撤销。</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
                取消
              </Button>
              <Button variant="destructive" onClick={handleDelete} disabled={isManagedReadOnlyModel}>
                删除
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </RecordAccessAdapterProvider>
  )
}
