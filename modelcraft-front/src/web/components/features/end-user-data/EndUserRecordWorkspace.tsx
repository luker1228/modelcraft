'use client'

import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { toast } from 'sonner'
import {
  createEndUserModelRuntimeClient,
  useEndUserProjectScopedClient,
  useEndUserModelRuntimeClient,
} from '@api-client/apollo/end-user-client'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { ModelRecordForm } from '@web/components/features/model-editor/model-record-form/index'
import { ModelRecordTable } from '@web/components/shared/data-workspace/ModelRecordTable'
import type { ModelRecordTableFieldInfo } from '@web/components/shared/data-workspace/ModelRecordTable'
import { getFieldProtocols } from '@web/components/features/model-editor/model-record-form/runtime/field-protocol'
import {
  buildFindUniqueQuery,
  buildDeleteMutation,
  buildCreateMutation,
  buildUpdateMutation,
  extractWritableFieldNamesFromSchema,
  sanitizeMutationInputData,
} from '@api-client/cms/public'
import type { FieldDefinition } from '@api-client/cms/public'
import { NOOP_MUTATION } from '@/api-client/noop'
import { GET_MODEL_RECORD_WORKSPACE_END_USER } from '@/api-client/model/graphql-docs.end-user'
import { Button } from '@web/components/ui/button'
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
  Plus,
  Edit,
  BookOpen,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/shared/utils'
import { FilterBar } from './FilterPanel'
import { SortPopover, buildOrderBy, type SortState } from './SortPopover'
import { getXMC } from '@/types/xmc'
import { RecordAccessAdapterProvider, type RecordAccessAdapter } from '@web/components/features/model-editor/model-record-form/access-adapter'
import { useCopilotKitAvailable, FilterCopilotActions } from './FilterCopilotActions'
import { getRecordPageCountText } from '@web/components/shared/data-workspace/recordPageCount'
import { useRuntimeListByPage } from '@web/components/shared/data-workspace/useRuntimeListByPage'
import { ModelApiDocsDialog } from './ModelApiDocsDialog'
import type { ModelApiDocContext } from './model-api-docs'


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

const PAGE_SIZE = 20

function resolveListByPageOrderBy(sort: SortState): Record<string, string>[] {
  return buildOrderBy(sort) ?? [{ id: 'desc' }]
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

interface NormalizedSchemaField {
  name: string
  prop: Record<string, unknown>
  labelFieldName: string
  hasLabelFieldInSchema: boolean
}

function normalizeSchemaFields(jsonSchema: Record<string, unknown> | null): NormalizedSchemaField[] {
  if (!jsonSchema?.properties) return []

  const props = jsonSchema.properties as Record<string, unknown>
  const propEntries = Object.entries(props) as Array<[string, unknown]>

  return propEntries.flatMap(([name, rawProp]) => {
    if (typeof rawProp !== 'object' || rawProp === null || Array.isArray(rawProp)) return []

    const prop = rawProp as Record<string, unknown>
    const labelFieldName = resolveEnumLabelFieldName(prop)

    return [{
      name,
      prop,
      labelFieldName,
      hasLabelFieldInSchema: Object.prototype.hasOwnProperty.call(props, labelFieldName),
    }]
  })
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

  const hasEndUserToken = useEndUserAuthStore((s) => !!s.accessToken)

  const managementClientFromHook = useEndUserProjectScopedClient(projectSlug)
  const managementClient = hasEndUserToken ? managementClientFromHook : null

  const accessAdapter = useMemo<RecordAccessAdapter | null>(() => {
    if (!managementClient) return null
    return {
      managementClient,
      managementContext: endUserContext,
      createRuntimeClient: (databaseName: string, modelName: string) => {
        return createEndUserModelRuntimeClient(orgName, projectSlug, databaseName, modelName)
      },
    }
  }, [managementClient, endUserContext, orgName, projectSlug])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null)
  const [apiDocsOpen, setApiDocsOpen] = useState(false)

  const [createDataOpen, setCreateDataOpen] = useState(false)
  const [createSaving, setCreateSaving] = useState(false)

  const [editDataOpen, setEditDataOpen] = useState(false)
  const [editItemId, setEditItemId] = useState<string | null>(null)
  const [editFormData, setEditFormData] = useState<Record<string, unknown>>({})
  const [editSaving, setEditSaving] = useState(false)
  const [editLoading, setEditLoading] = useState(false)

  const [searchKeyword, setSearchKeyword] = useState('')

  // --- Filter state ---
  // Draft state is now owned by FilterBar (StructuredFilterTab rows).
  // Only committed where JSON drives the actual GraphQL query.
  const [whereJsonCommitted, setWhereJsonCommitted] = useState<string | null>(null)

  const whereInput = useMemo(() => {
    if (!whereJsonCommitted?.trim()) return undefined
    try {
      return JSON.parse(whereJsonCommitted) as Record<string, unknown>
    } catch {
      return undefined
    }
  }, [whereJsonCommitted])

  function handleApplyFilter(whereJson: string) {
    setWhereJsonCommitted(whereJson)
  }

  function handleClearFilter() {
    setWhereJsonCommitted(null)
  }

  // --- Sort state ---
  const [sortState, setSortState] = useState<SortState>({
    field: '',
    direction: 'asc',
    stableSort: true,
  })
  // --- End filter/sort state ---
  // ----- CopilotKit Frontend Actions -----
  // Guard: only mount the actions component when CopilotKit context is available.
  // EndUserRecordWorkspace can render outside CopilotWrapper (e.g. /end-user/ routes),
  // and useCopilotAction throws when called outside CopilotKit context.
  //
  // NOTE: useContext(CopilotContext) !== null is NOT a reliable guard because both
  // @copilotkit/react-core's CopilotContext and @copilotkitnext/react's CopilotKitContext
  // have non-null default values. We must inspect copilotkit instance directly.
  const hasCopilot = useCopilotKitAvailable()

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
  const apiDocContext = useMemo<ModelApiDocContext | null>(() => {
    if (!model?.databaseName || !model?.name) {
      return null
    }

    return {
      orgName,
      projectSlug,
      databaseName: model.databaseName,
      modelName: model.name,
    }
  }, [orgName, projectSlug, model?.databaseName, model?.name])

  const canCreateRecord = !isManagedReadOnlyModel
  const canEditRecord = !isManagedReadOnlyModel
  const canDeleteRecord = !isManagedReadOnlyModel

  const showManagedReadonlyToast = useCallback(() => {
    toast.warning('托管模型仅支持查看，不支持写入')
  }, [])

  const runtimeClientFromHook = useEndUserModelRuntimeClient(
    projectSlug,
    model?.databaseName,
    model?.name
  )
  const runtimeClient = hasEndUserToken ? runtimeClientFromHook : null

  const jsonSchema = useMemo<Record<string, unknown> | null>(() => {
    if (!model?.jsonSchema) return null
    try {
      return JSON.parse(model.jsonSchema) as Record<string, unknown>
    } catch (error) {
      console.error('Failed to parse schema:', error)
      return null
    }
  }, [model?.jsonSchema])

  const normalizedSchemaFields = useMemo(() => normalizeSchemaFields(jsonSchema), [jsonSchema])

  const runtimeFields = useMemo(() => {
    if (normalizedSchemaFields.length > 0) {
      const schemaFieldDefs: FieldDefinition[] = normalizedSchemaFields.flatMap(({ name, prop, labelFieldName, hasLabelFieldInSchema }) => {
        const rawType = prop['type']
        const rawFormat = prop['format']
        const schemaType = typeof rawType === 'string' ? rawType.toUpperCase() : undefined
        const format = typeof rawFormat === 'string' ? rawFormat.toUpperCase() : undefined
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

    if (!jsonSchema) return [{ name: 'id', type: 'string', schemaType: 'STRING' }] as FieldDefinition[]
    return [] as FieldDefinition[]
  }, [jsonSchema, normalizedSchemaFields])

  const writableFieldNames = useMemo(
    () => extractWritableFieldNamesFromSchema(jsonSchema as { properties?: Record<string, unknown> } | null | undefined),
    [jsonSchema]
  )

  const tableFieldInfos = useMemo<ModelRecordTableFieldInfo[]>(() => {
    if (normalizedSchemaFields.length > 0) {
      return normalizedSchemaFields.flatMap(({ name, prop, labelFieldName, hasLabelFieldInSchema }) => {
        const xmc = getXMC(prop)
        const rawTitle = prop['title']
        const rawType = prop['type']
        const baseInfo: ModelRecordTableFieldInfo = {
          name,
          title: typeof rawTitle === 'string' ? rawTitle : null,
          isPrimary: xmc?.isPrimary === true,
          isDeprecated: false,
          storageHint: deriveStorageHintFromSchemaProp(prop),
          schemaType: typeof rawType === 'string' ? rawType.toUpperCase() : null,
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
  }, [normalizedSchemaFields])

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

  const listByPageOrderBy = useMemo(() => resolveListByPageOrderBy(sortState), [sortState])

  const {
    currentPage,
    setCurrentPage,
    contentLoading,
    contentList,
    totalCount,
    totalPages,
    hasNextPage,
    refetch,
  } = useRuntimeListByPage({
    modelName,
    runtimeFields,
    runtimeClient,
    whereInput,
    orderBy: listByPageOrderBy,
    pageSize: PAGE_SIZE,
    resetDeps: [modelId, whereJsonCommitted, sortState.field, sortState.direction, sortState.stableSort],
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

    setEditFormData({})
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
      if (typeof item === 'object' && item !== null && !Array.isArray(item)) {
        const schemaDrivenData: Record<string, unknown> = {}
        for (const fieldName of writableFieldNames) {
          schemaDrivenData[fieldName] = item[fieldName] ?? ''
        }
        setEditFormData(schemaDrivenData)
      } else {
        setEditFormData({})
      }
    } catch (error) {
      console.error('Failed to fetch content:', error)
      setEditFormData({})
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

  const pageCountText = useMemo(
    () =>
      getRecordPageCountText({
        pageCount: contentList.length,
        filteredCount: filteredContentList.length,
        searchKeyword,
      }),
    [contentList.length, filteredContentList.length, searchKeyword]
  )

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
      <div className="flex h-full min-h-0 flex-col">
        {isManagedReadOnlyModel && (
          <div className="border-b border-border bg-card px-4 py-2">
            <Alert variant="warning" className="py-2">
              <AlertDescription className="text-xs">
                当前为托管模型，数据为只读模式，新增/编辑/删除操作已禁用。
              </AlertDescription>
            </Alert>
          </div>
        )}

        {/* Filter bar — chip-style inline filters + search + count (Supabase style) */}
        <FilterBar
          fields={runtimeFields}
          onApply={handleApplyFilter}
          onClear={handleClearFilter}
          searchValue={searchKeyword}
          onSearchChange={setSearchKeyword}
          searchPlaceholder={`搜索 ${model.title || model.name}...`}
          summaryText={pageCountText}
        />

        {/* 工具栏 */}
        <div className="flex h-10 items-center justify-between gap-2 overflow-x-auto border-b border-border bg-card p-1.5">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <SortPopover
                fields={runtimeFields.filter((f): f is FieldDefinition => typeof f !== 'string')}
                sort={sortState}
                onSortChange={setSortState}
              />
            </div>

            <div className="h-5 w-px bg-border" />

            <Button
              size="sm"
              className={cn(
                'h-[26px] border-0 px-2.5 text-xs font-normal transition-colors duration-200',
                canCreateRecord
                  ? 'bg-primary text-white hover:bg-primary/90'
                  : 'cursor-not-allowed bg-muted text-muted-foreground'
              )}
              onClick={handleCreate}
              disabled={!canCreateRecord}
            >
              <Plus className="mr-1.5 size-3.5" />
              <span>添加数据</span>
            </Button>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-[26px] px-2.5 text-xs"
              onClick={() => setApiDocsOpen(true)}
              disabled={apiDocContext == null}
            >
              <BookOpen className="mr-1.5 size-3.5" />
              <span>API 文档</span>
            </Button>
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
          pagination={{
            currentPage,
            pageSize: PAGE_SIZE,
            totalCount,
            totalPages,
            hasNextPage,
            onPreviousPage: () => setCurrentPage((prev) => Math.max(1, prev - 1)),
            onNextPage: () => setCurrentPage((prev) => Math.min(totalPages, prev + 1)),
            onGoToPage: (page) => setCurrentPage(Math.min(totalPages, Math.max(1, page))),
          }}
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

        <ModelApiDocsDialog
          open={apiDocsOpen}
          onOpenChange={setApiDocsOpen}
          context={apiDocContext}
        />
      </div>
      {/* Mount CopilotKit actions only when context exists — avoids null.subscribe crash */}
      {hasCopilot && (
        <FilterCopilotActions
          onSetFilter={(json) => { setWhereJsonCommitted(json) }}
          onClearFilter={handleClearFilter}
        />
      )}
    </RecordAccessAdapterProvider>
  )
}
