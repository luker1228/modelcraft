'use client'

import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import { toast } from 'sonner'
import {
  useProjectScopedClient,
  createModelRuntimeClient,
  useProjectScopedContext,
} from '@api-client/apollo/public'
import { ModelRecordForm } from './index'
import { ModelRecordInsertMenu } from './ModelRecordInsertMenu'
import { ModelRecordTable } from './ModelRecordTable'
import { RecordRelationManagerDialog } from './RecordRelationManagerDialog'
import type { ModelRecordTableFieldInfo, ModelRecordTableRow } from './ModelRecordTable'
import { getFieldProtocols } from './runtime/field-protocol'
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
import { DEPRECATE_FIELD, REMOVE_FIELD, UNDEPRECATE_FIELD } from '@/api-client/model'
import { GET_MODEL_RECORD_WORKSPACE } from '@/api-client/model'
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
  Loader2,
  TerminalSquare,
  RefreshCw,
} from 'lucide-react'
import { getXMC } from '@/types/xmc'
import { RecordAccessAdapterProvider, type RecordAccessAdapter } from './access-adapter'
import { useWorkspaceAIRef } from '@web/contexts/workspace-ai-ref-context'
import { FilterBar } from '@web/components/features/end-user-data/FilterPanel'
import { getRecordPageCountText } from '@web/components/shared/data-workspace/recordPageCount'

export interface DevelopRecordWorkspaceAIRef {
  openCreate: (prefill: Record<string, unknown>) => void
  openEdit: (recordId: string, patch: Record<string, unknown>) => Promise<void>
  setHighlight: (ids: string[], reasons: Record<string, string>) => void
}

function toQueryValue(value: string | null | undefined): string {
  return (value ?? '').trim()
}

export interface DevelopRecordWorkspaceProps {
  modelId: string
  projectSlug: string
  orgName: string
  refreshToken?: number
  quickNav?: React.ReactNode
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
 * DevelopRecordWorkspace — 开发者模式数据工作区。
 *
 * 能力范围：
 * - record query / create / edit / delete
 * - 字段插入（InsertMenu）
 * - 字段废弃 / 取消废弃 / 删除（字段生命周期维护）
 * - 关系维护（RecordRelationManagerDialog）
 *
 * 使用 developer project-scoped client 作为管理端点，
 * 使用 createModelRuntimeClient 作为数据端点。
 * 不依赖 workspaceMode，不接受 workspaceMode prop。
 */
export default function DevelopRecordWorkspace({
  modelId,
  projectSlug,
  orgName,
  refreshToken = 0,
  quickNav,
}: DevelopRecordWorkspaceProps) {
  const projectClient = useProjectScopedClient(projectSlug)
  const projectScopedContext = useProjectScopedContext(orgName, projectSlug)

  // 构建 develop workspace 的 RecordAccessAdapter
  // orgName/projectSlug 在此组件中必须存在，断言收窄类型
  const accessAdapter = useMemo<RecordAccessAdapter>(() => ({
    managementClient: projectClient,
    managementContext: projectScopedContext ?? { uri: `/api/bff/graphql/org/${orgName}/project/${projectSlug}/` },
    createRuntimeClient: (databaseName: string, modelName: string) =>
      createModelRuntimeClient(orgName, projectSlug, databaseName, modelName),
  }), [projectClient, projectScopedContext, orgName, projectSlug])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null)
  const [removeFieldDialogOpen, setRemoveFieldDialogOpen] = useState(false)
  const [removeFieldTarget, setRemoveFieldTarget] = useState<ModelRecordTableFieldInfo | null>(null)
  const [relationManagerOpen, setRelationManagerOpen] = useState(false)
  const [relationRecordId, setRelationRecordId] = useState<string | null>(null)

  // 新增数据状态
  const [createDataOpen, setCreateDataOpen] = useState(false)
  const [createSaving, setCreateSaving] = useState(false)

  // 编辑数据状态
  const [editDataOpen, setEditDataOpen] = useState(false)
  const [editItemId, setEditItemId] = useState<string | null>(null)
  const [editFormData, setEditFormData] = useState<Record<string, unknown>>({})
  const [editSaving, setEditSaving] = useState(false)
  const [editLoading, setEditLoading] = useState(false)

  // 搜索关键词
  const [searchKeyword, setSearchKeyword] = useState('')

  // 结构化筛选
  const [whereFilter, setWhereFilter] = useState<Record<string, unknown> | null>(null)

  // AI 控制状态
  const [createPrefill, setCreatePrefill] = useState<Record<string, unknown>>({})
  const [highlightedIds, setHighlightedIds] = useState<string[]>([])
  const [highlightReasons, setHighlightReasons] = useState<Record<string, string>>({})

  // 拉取模型 schema（develop 用 project client）
  const { data: modelData, loading: modelLoading, refetch: refetchModel } = useQuery<GetModelQueryData, { id: string }>(
    GET_MODEL_RECORD_WORKSPACE,
    {
      client: projectClient,
      variables: { id: modelId },
      context: projectScopedContext,
    }
  )

  const model = modelData?.model?.model
  const modelError = modelData?.model?.error
  const modelName = model?.name
  const isManagedReadOnlyModel = model?.createdVia === 'IMPORTED'

  // develop workspace 始终拥有字段生命周期能力（托管模型除外）
  const canManageFieldLifecycle = !isManagedReadOnlyModel
  const canCreateRecord = !isManagedReadOnlyModel
  const canEditRecord = !isManagedReadOnlyModel
  const canDeleteRecord = !isManagedReadOnlyModel

  const showManagedReadonlyToast = useCallback(() => {
    toast.warning('托管模型仅支持查看，不支持写入或结构修改')
  }, [])

  // develop 场景：runtime client 使用 model-specific endpoint
  const runtimeClient = useMemo(() => {
    if (!model?.databaseName || !model?.name) return null
    return createModelRuntimeClient(orgName, projectSlug, model.databaseName, model.name)
  }, [orgName, projectSlug, model?.databaseName, model?.name])

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
      const deprecatedFieldNames = new Set(
        (model?.fields ?? [])
          .filter((field) => field?.isDeprecated === true)
          .map((field) => field.name)
      )

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
          isDeprecated: deprecatedFieldNames.has(name),
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
            isDeprecated: deprecatedFieldNames.has(name),
            storageHint: 'TEXT',
            schemaType: 'STRING',
          },
        ]
      })
    }

    return []
  }, [jsonSchema, model?.fields])

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

  const findUniqueQuery = useMemo(() => {
    if (!modelName) return null
    return buildFindUniqueQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  // AI コントロール — WorkspaceAIRefContext に命令型インタフェースを公開
  const workspaceAiRef = useWorkspaceAIRef()

  if (workspaceAiRef) {
    workspaceAiRef.current = {
      openCreate: (prefill) => {
        setCreatePrefill(prefill)
        setCreateDataOpen(true)
      },
      openEdit: async (id, patch) => {
        setEditItemId(id)
        setEditLoading(true)
        setEditDataOpen(true)
        try {
          if (!runtimeClient || !findUniqueQuery) return
          const queryResult = await runtimeClient.query({
            query: findUniqueQuery,
            variables: { where: { id } },
            fetchPolicy: 'network-only',
          })
          const rawData = queryResult.data as Record<string, Record<string, unknown>>
          const item = rawData?.findUnique?.item
          if (isRecord(item)) {
            const base = writableFieldNames.reduce<Record<string, unknown>>((acc, f) => {
              acc[f] = item[f] ?? ''
              return acc
            }, {})
            setEditFormData({ ...base, ...patch })
          }
        } catch {
          toast.error('获取数据失败')
          setEditDataOpen(false)
        } finally {
          setEditLoading(false)
        }
      },
      setHighlight: (ids, reasons) => {
        setHighlightedIds(ids)
        setHighlightReasons(reasons)
      },
    }
  }

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
      ...(whereFilter ? { where: whereFilter } : {}),
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

  // develop-only：字段生命周期 mutation，使用 projectClient
  const [deprecateField] = useMutation(DEPRECATE_FIELD, { client: projectClient })
  const [undeprecateField] = useMutation(UNDEPRECATE_FIELD, { client: projectClient })
  const [removeField] = useMutation(REMOVE_FIELD, { client: projectClient })

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

  // develop-only：字段废弃/取消废弃
  const handleToggleFieldDeprecated = useCallback(
    async (fieldInfo: ModelRecordTableFieldInfo) => {
      if (isManagedReadOnlyModel) {
        showManagedReadonlyToast()
        return
      }
      if (!model?.id) return

      try {
        const mutate = fieldInfo.isDeprecated ? undeprecateField : deprecateField
        await mutate({
          variables: {
            modelID: model.id,
            fieldName: fieldInfo.name,
          },
          context: projectScopedContext,
        })

        await refetchModel()
        toast.success(fieldInfo.isDeprecated ? '已取消废弃字段' : '字段已废弃')
      } catch (error) {
        console.error('Failed to toggle field deprecated state:', error)
        toast.error(fieldInfo.isDeprecated ? '取消废弃失败' : '废弃字段失败')
      }
    },
    [
      isManagedReadOnlyModel,
      showManagedReadonlyToast,
      model?.id,
      deprecateField,
      undeprecateField,
      projectScopedContext,
      refetchModel,
    ]
  )

  // develop-only：发起字段删除
  const handleRequestRemoveField = useCallback((fieldInfo: ModelRecordTableFieldInfo) => {
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }
    if (!fieldInfo.isDeprecated) {
      toast.error('请先废弃字段，再执行删除')
      return
    }

    setRemoveFieldTarget(fieldInfo)
    setRemoveFieldDialogOpen(true)
  }, [isManagedReadOnlyModel, showManagedReadonlyToast])

  // develop-only：确认删除字段
  const handleConfirmRemoveField = useCallback(async () => {
    if (isManagedReadOnlyModel) {
      showManagedReadonlyToast()
      return
    }
    if (!model?.id || !removeFieldTarget) return

    try {
      await removeField({
        variables: {
          modelID: model.id,
          fieldName: removeFieldTarget.name,
        },
        context: projectScopedContext,
      })

      setRemoveFieldDialogOpen(false)
      setRemoveFieldTarget(null)
      await Promise.all([refetchModel(), refetch()])
      toast.success('字段已删除')
    } catch (error) {
      console.error('Failed to remove field:', error)
      toast.error('删除字段失败')
    }
  }, [
    isManagedReadOnlyModel,
    showManagedReadonlyToast,
    model?.id,
    removeFieldTarget,
    removeField,
    projectScopedContext,
    refetchModel,
    refetch,
  ])

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

  if (modelLoading) {
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
        {quickNav ?? (
          <div className="flex items-center gap-2 overflow-x-auto border-b border-border bg-card p-2">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 shrink-0 text-muted-foreground"
              disabled
            >
              <TerminalSquare className="mr-1.5 size-3.5" />
              SQL 控制台（敬请期待）
            </Button>
          </div>
        )}

        {isManagedReadOnlyModel && (
          <div className="border-b border-border bg-card px-4 py-2">
            <Alert variant="warning" className="py-2">
              <AlertDescription className="text-xs">
                当前为托管模型，数据与结构均为只读模式，新增/编辑/删除操作已禁用。
              </AlertDescription>
            </Alert>
          </div>
        )}

        <FilterBar
          fields={runtimeFields}
          onApply={(whereJson) => setWhereFilter(JSON.parse(whereJson) as Record<string, unknown>)}
          onClear={() => setWhereFilter(null)}
          searchValue={searchKeyword}
          onSearchChange={setSearchKeyword}
          searchPlaceholder={`搜索 [${model.title || model.name}]...`}
          summaryText={pageCountText}
        />

        {/* 工具栏 */}
        <div className="flex h-10 items-center justify-between gap-2 overflow-x-auto border-b border-border bg-card p-1.5">
          <div className="flex items-center gap-2">
            <ModelRecordInsertMenu
              onCreateRecord={handleCreate}
              canCreateRecord={canCreateRecord}
            />
          </div>

          <div className="flex items-center gap-2">
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

        {/* 数据表格 — 包含字段生命周期控制 */}
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
          onManageRelations={(id) => {
            setRelationRecordId(id)
            setRelationManagerOpen(true)
          }}
          onToggleFieldDeprecated={handleToggleFieldDeprecated}
          onDeleteField={handleRequestRemoveField}
          canManageFieldLifecycle={canManageFieldLifecycle}
          canCreateRecord={canCreateRecord}
          canEditRecord={canEditRecord}
          canDeleteRecord={canDeleteRecord}
          highlightedIds={highlightedIds}
          highlightReasons={highlightReasons}
        />

        {/* 关系维护对话框 — develop workspace 始终可用 */}
        <RecordRelationManagerDialog
          open={relationManagerOpen}
          onOpenChange={(open) => {
            setRelationManagerOpen(open)
            if (!open) setRelationRecordId(null)
          }}
          jsonSchema={jsonSchema as import('@rjsf/utils').RJSFSchema | null}
          orgName={orgName}
          projectSlug={projectSlug}
          modelId={modelId}
          recordId={relationRecordId}
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
                initialValues={createPrefill}
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
                workspaceMode="design"
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
                workspaceMode="design"
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

        {/* 删除字段确认对话框 — develop-only */}
        <Dialog
          open={removeFieldDialogOpen}
          onOpenChange={(open) => {
            setRemoveFieldDialogOpen(open)
            if (!open) setRemoveFieldTarget(null)
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>确认删除字段</DialogTitle>
              <DialogDescription>
                确定要删除字段 <span className="font-mono">{removeFieldTarget?.name}</span> 吗？此操作不可撤销。
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => {
                  setRemoveFieldDialogOpen(false)
                  setRemoveFieldTarget(null)
                }}
              >
                取消
              </Button>
              <Button
                variant="destructive"
                onClick={handleConfirmRemoveField}
                disabled={isManagedReadOnlyModel}
              >
                删除字段
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </RecordAccessAdapterProvider>
  )
}
