'use client'

import React, { useMemo, useState, useCallback, useEffect } from 'react'
import { useQuery, useMutation, gql, ApolloClient } from '@apollo/client'
import { toast } from 'sonner'
import { useProjectScopedClient, createModelRuntimeClient, buildRuntimeEndpoint } from '@bff/apollo/public'
import { InsertFieldSheet } from './InsertFieldSheet'
import { ModelRecordForm } from './ModelRecordForm'
import { ModelRecordTable } from './ModelRecordTable'
import { RecordRelationManagerDialog } from './RecordRelationManagerDialog'
import type { ModelRecordTableFieldInfo, ModelRecordTableRow } from './ModelRecordTable'
import { getFieldProtocols } from './fieldProtocol'
import {
  buildEditFormData,
  mapModelFieldsToRuntimeFields,
  mapModelFieldsToTableFieldInfos,
  type ModelField,
} from './modelFieldMapping'
import {
  buildFindManyQuery,
  buildFindUniqueQuery,
  buildDeleteMutation,
  buildCreateMutation,
  buildUpdateMutation,
  extractFieldsFromSchema,
  extractWritableFieldNamesFromSchema,
  sanitizeMutationInputData,
} from '@bff/cms/public'
import { REMOVE_FIELD } from '@web/graphql/mutations/model'
import { Button } from '@web/components/ui/button'
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  Filter,
  List,
  Plus,
  Edit,
  Loader2,
  Copy,
  Check,
  RefreshCw,
  ChevronDown,
  Columns,
} from 'lucide-react'

// Query to get model details with fields and JSON schema
const GET_MODEL_QUERY = gql`
  query GetModel($id: ID!) {
    model(id: $id) {
      model {
        id
        name
        title
        description
        databaseName
        jsonSchema
        fields {
          name
          title
          format
          schemaType
          storageHint
          isPrimary
          isDeprecated
          description
        }
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

const DEPRECATE_FIELD_MUTATION = gql`
  mutation DeprecateField($modelID: ID!, $fieldName: String!) {
    deprecateField(modelID: $modelID, fieldName: $fieldName) {
      id
    }
  }
`

const UNDEPRECATE_FIELD_MUTATION = gql`
  mutation UndeprecateField($modelID: ID!, $fieldName: String!) {
    undeprecateField(modelID: $modelID, fieldName: $fieldName) {
      id
    }
  }
`

interface DynamicModelTableProps {
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
      jsonSchema?: string | null
      fields: ModelField[]
    } | null
    error?: {
      message?: string | null
    } | null
  } | null
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export default function DynamicModelTable({
  modelId,
  projectSlug,
  orgName,
  refreshToken = 0,
}: DynamicModelTableProps) {
  const projectClient = useProjectScopedClient(projectSlug)

  // 创建 project-scoped context
  const projectScopedContext = useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null)
  const [removeFieldDialogOpen, setRemoveFieldDialogOpen] = useState(false)
  const [removeFieldTarget, setRemoveFieldTarget] = useState<ModelRecordTableFieldInfo | null>(null)
  const [relationManagerOpen, setRelationManagerOpen] = useState(false)
  const [relationRecordId, setRelationRecordId] = useState<string | null>(null)

  // 添加数据状态
  const [createDataOpen, setCreateDataOpen] = useState(false)
  const [createSaving, setCreateSaving] = useState(false)

  // 编辑数据状态
  const [editDataOpen, setEditDataOpen] = useState(false)
  const [editItemId, setEditItemId] = useState<string | null>(null)
  const [editFormData, setEditFormData] = useState<Record<string, unknown>>({})
  const [editSaving, setEditSaving] = useState(false)
  const [editLoading, setEditLoading] = useState(false)

  // 插入列状态
  const [insertColumnOpen, setInsertColumnOpen] = useState(false)

  // 复制端点状态
  const [endpointCopied, setEndpointCopied] = useState(false)
  // 复制 name 状态
  const [nameCopied, setNameCopied] = useState(false)

  // Fetch model details with fields and JSON schema
  const { data: modelData, loading: modelLoading, refetch: refetchModel } = useQuery<GetModelQueryData, { id: string }>(GET_MODEL_QUERY, {
    client: projectClient,
    variables: { id: modelId },
    context: projectScopedContext,
  })

  // 注意：model query 返回 GetModelPayload，model 在 payload.model 中
  const model = modelData?.model?.model
  const modelError = modelData?.model?.error
  const modelName = model?.name
  const fields = useMemo<ModelField[]>(() => model?.fields ?? [], [model?.fields])

  // 创建动态的 Runtime Client，使用模型特定的 GraphQL 端点
  const runtimeClient = useMemo(() => {
    if (!model?.databaseName || !model?.name) return null
    return createModelRuntimeClient(orgName, projectSlug, model.databaseName, model.name)
  }, [orgName, projectSlug, model?.databaseName, model?.name]) as ApolloClient<object> | null

  // Parse schema to get runtime fields
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
    if (fields.length > 0) {
      return mapModelFieldsToRuntimeFields(fields)
    }
    if (!jsonSchema) return ['id']
    return extractFieldsFromSchema(jsonSchema as { properties?: Record<string, unknown> })
  }, [fields, jsonSchema])

  const writableFieldNames = useMemo(
    () => extractWritableFieldNamesFromSchema(jsonSchema as { properties?: Record<string, unknown> } | null | undefined),
    [jsonSchema]
  )

  const tableFieldInfos = useMemo<ModelRecordTableFieldInfo[]>(() => {
    if (fields.length === 0) {
      return []
    }

    return mapModelFieldsToTableFieldInfos(fields)
  }, [fields])

  // 使用 model.fields 作为表头字段来源（更可靠）
  const displayFields = useMemo(() => {
    if (tableFieldInfos.length > 0) {
      return tableFieldInfos.map((field) => field.name)
    }
    return runtimeFields.map((f) => typeof f === 'string' ? f : f.name)
  }, [tableFieldInfos, runtimeFields])

  // Build a name→prop lookup from JSON Schema for cell rendering
  const propByName = useMemo(() => {
    if (!jsonSchema) return {}
    const protocols = getFieldProtocols(jsonSchema as import('@rjsf/utils').RJSFSchema)
    return Object.fromEntries(protocols.map(({ name, prop }) => [name, prop]))
  }, [jsonSchema])

  // 辅助函数：根据字段名获取字段信息
  const getFieldInfo = useCallback(
    (fieldName: string): ModelRecordTableFieldInfo | null => {
      return tableFieldInfos.find((field) => field.name === fieldName) ?? null
    },
    [tableFieldInfos]
  )

  // 辅助函数：获取字段类型的简短显示
  const getFieldTypeDisplay = useCallback((fieldInfo: ModelRecordTableFieldInfo | null) => {
    if (!fieldInfo) return ''
    // 优先使用 storageHint，其次使用 schemaType
    return fieldInfo.storageHint || fieldInfo.schemaType || ''
  }, [])

  // Build dynamic query for fetching content
  const findManyQuery = useMemo(() => {
    if (!modelName) return null
    return buildFindManyQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  // Fetch content list
  const {
    data: contentData,
    loading: contentLoading,
    refetch,
  } = useQuery<Record<string, unknown>>(findManyQuery || gql`query { __typename }`, {
    client: runtimeClient!,
    skip: !findManyQuery || !runtimeClient,
    variables: {
      take: 50,
      skip: 0,
    },
  })

  // Build delete mutation
  const deleteMutation = useMemo(() => {
    if (!modelName) return null
    return buildDeleteMutation(modelName)
  }, [modelName])

  const [deleteContent] = useMutation(deleteMutation || gql`mutation { __typename }`, {
    client: runtimeClient!,
    onCompleted: () => {
      refetch()
      setDeleteDialogOpen(false)
      setDeleteItemId(null)
    },
  })

  const [deprecateField] = useMutation(DEPRECATE_FIELD_MUTATION, {
    client: projectClient,
  })

  const [undeprecateField] = useMutation(UNDEPRECATE_FIELD_MUTATION, {
    client: projectClient,
  })

  const [removeField] = useMutation(REMOVE_FIELD, {
    client: projectClient,
  })

  // Build create mutation
  const createMutation = useMemo(() => {
    if (!modelName) return null
    return buildCreateMutation(modelName)
  }, [modelName])

  const [createContent] = useMutation(createMutation || gql`mutation { __typename }`, {
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

  // Build findUnique query for editing
  const findUniqueQuery = useMemo(() => {
    if (!modelName) return null
    return buildFindUniqueQuery(modelName, runtimeFields)
  }, [modelName, runtimeFields])

  // Build update mutation
  const updateMutation = useMemo(() => {
    if (!modelName) return null
    return buildUpdateMutation(modelName)
  }, [modelName])

  const [updateContent] = useMutation(updateMutation || gql`mutation { __typename }`, {
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

  const contentList: ModelRecordTableRow[] = Array.isArray((contentData as Record<string, unknown> | undefined)?.findMany && ((contentData as Record<string, unknown>).findMany as Record<string, unknown>)?.items)
    ? ((contentData as Record<string, unknown>).findMany as Record<string, unknown>).items as ModelRecordTableRow[]
    : []

  useEffect(() => {
    if (!refreshToken) return
    void refetchModel()
    void refetch()
  }, [refreshToken, refetchModel, refetch])

  const handleDelete = async () => {
    if (!deleteItemId) return

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
    if (!runtimeClient || !findUniqueQuery) return

    setEditItemId(id)
    setEditLoading(true)
    setEditDataOpen(true)

    try {
      // 获取当前数据
      const { data } = await runtimeClient.query<Record<string, unknown>>({
        query: findUniqueQuery,
        variables: { where: { id } },
        fetchPolicy: 'network-only',
      })

      const item = (data as Record<string, unknown> | undefined)?.findUnique && ((data as Record<string, unknown>).findUnique as Record<string, unknown>)?.item
      if (isRecord(item)) {
        setEditFormData(buildEditFormData(fields, item))
      }
    } catch (error) {
      console.error('Failed to fetch content:', error)
      toast.error('获取数据失败')
      setEditDataOpen(false)
    } finally {
      setEditLoading(false)
    }
  }

  // 打开添加数据弹窗
  const handleCreate = () => {
    setCreateDataOpen(true)
  }

  const confirmDelete = (id: string) => {
    setDeleteItemId(id)
    setDeleteDialogOpen(true)
  }

  const handleToggleFieldDeprecated = useCallback(
    async (fieldInfo: ModelRecordTableFieldInfo) => {
      if (!model?.id) {
        return
      }

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
    [model?.id, deprecateField, undeprecateField, projectScopedContext, refetchModel]
  )

  const handleRequestRemoveField = useCallback((fieldInfo: ModelRecordTableFieldInfo) => {
    if (!fieldInfo.isDeprecated) {
      toast.error('请先废弃字段，再执行删除')
      return
    }

    setRemoveFieldTarget(fieldInfo)
    setRemoveFieldDialogOpen(true)
  }, [])

  const handleConfirmRemoveField = useCallback(async () => {
    if (!model?.id || !removeFieldTarget) {
      return
    }

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
  }, [model?.id, removeFieldTarget, removeField, projectScopedContext, refetchModel, refetch])

  // 构建 GraphQL 端点 URL
  const graphqlEndpoint = model
    ? buildRuntimeEndpoint(orgName, projectSlug, model.databaseName ?? '', model.name ?? '')
    : ''

  // 复制端点到剪贴板
  const handleCopyEndpoint = async () => {
    try {
      const fullUrl = `${window.location.origin}${graphqlEndpoint}`
      await navigator.clipboard.writeText(fullUrl)
      setEndpointCopied(true)
      setTimeout(() => setEndpointCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy endpoint:', error)
    }
  }

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
    <div className="flex h-full flex-col">
      {/* 顶部标题栏 */}
      <div className="flex items-center justify-between border-b border-border bg-sidebar px-4 py-3">
        <div className="flex items-center gap-4">
          <div>
            <h1 className="flex items-center gap-2 font-heading text-base font-semibold text-foreground">
              {model.title || model.name}
              {model.title && model.name && (
                <button
                  onClick={async () => {
                    try {
                      await navigator.clipboard.writeText(model.name)
                      setNameCopied(true)
                      setTimeout(() => setNameCopied(false), 2000)
                    } catch (error) {
                      console.error('Failed to copy name:', error)
                    }
                  }}
                  className={`cursor-pointer text-sm font-normal transition-colors ${
                    nameCopied ? 'text-emerald-600' : 'text-muted-foreground hover:text-primary'
                  }`}
                  title={nameCopied ? '已复制!' : '点击复制 name'}
                >
                  ({model.name})
                  {nameCopied && <Check className="ml-1 inline size-3" />}
                </button>
              )}
            </h1>
            <div className="mt-0.5 flex items-center gap-2">
              <code className="max-w-[300px] truncate rounded border border-border/50 bg-muted/60 px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
                {graphqlEndpoint}
              </code>
              <button
                onClick={handleCopyEndpoint}
                className="flex-shrink-0 text-muted-foreground transition-colors hover:text-primary"
                title="复制端点"
              >
                {endpointCopied ? (
                  <Check className="size-3.5 text-emerald-600" />
                ) : (
                  <Copy className="size-3.5" />
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 紧凑工具栏 - 数据列表上方 */}
      <div className="flex h-10 items-center justify-between gap-2 overflow-x-auto border-b border-border bg-sidebar p-1.5">
        <div className="flex items-center gap-4">
          {/* Filter & Sort */}
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-[26px] border-transparent px-2.5 text-xs font-normal text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <Filter className="mr-1.5 size-3.5" />
              <span>筛选</span>
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

          {/* 分隔线 */}
          <div className="h-5 w-px bg-border" />

          {/* Insert 下拉菜单 */}
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="sm"
                  className="h-[26px] border-0 bg-primary px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-primary/90"
                >
                  <Plus className="mr-1.5 size-3.5" />
                  <span>插入</span>
                  <ChevronDown className="ml-1.5 size-3 opacity-70" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-40 border border-slate-200 shadow-lg">
                <DropdownMenuItem onClick={handleCreate} className="cursor-pointer text-xs focus:bg-selected focus:text-foreground">
                  <Plus className="mr-2 size-3.5" />
                  插入数据
                </DropdownMenuItem>
                <DropdownMenuItem className="cursor-pointer text-xs focus:bg-selected focus:text-foreground" onClick={() => setInsertColumnOpen(true)}>
                  <Columns className="mr-2 size-3.5" />
                  插入列
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* 右侧工具 */}
        <div className="flex items-center gap-2">
          {!contentLoading && (
            <span className="text-xs text-muted-foreground">
              {contentList.length} 条记录
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

      {/* 主内容区 - 数据列表 */}
      <ModelRecordTable
        contentLoading={contentLoading}
        contentList={contentList}
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
      />

      <RecordRelationManagerDialog
        open={relationManagerOpen}
        onOpenChange={(open) => {
          setRelationManagerOpen(open)
          if (!open) {
            setRelationRecordId(null)
          }
        }}
        jsonSchema={jsonSchema as import('@rjsf/utils').RJSFSchema | null}
        orgName={orgName}
        projectSlug={projectSlug}
        modelId={modelId}
        recordId={relationRecordId}
      />

      {/* 添加数据侧边栏 */}
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
                setCreateSaving(true)
                try {
                  const sanitizedData = sanitizeMutationInputData(data, writableFieldNames)
                  await createContent({ variables: { data: sanitizedData } })
                  setCreateDataOpen(false)
                } catch (error) {
                  throw error  // ModelRecordForm catches and toasts
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
                setEditSaving(true)
                try {
                  const sanitizedData = sanitizeMutationInputData(data, writableFieldNames)
                  await updateContent({ variables: { where: { id: editItemId }, data: sanitizedData } })
                  setEditDataOpen(false)
                } catch (error) {
                  throw error  // ModelRecordForm catches and toasts
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
            />
          )}
        </SheetContent>
      </Sheet>

      {/* 插入列侧边栏 */}
      <InsertFieldSheet
        open={insertColumnOpen}
        onOpenChange={setInsertColumnOpen}
        modelId={modelId}
        modelName={model?.name}
        projectSlug={projectSlug}
        orgName={orgName}
        existingFieldNames={(model?.fields ?? []).map((f) => f.name)}
        onSuccess={() => {
          void refetch()
          void refetchModel()
        }}
      />

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
            <Button variant="destructive" onClick={handleDelete}>
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除字段确认对话框 */}
      <Dialog
        open={removeFieldDialogOpen}
        onOpenChange={(open) => {
          setRemoveFieldDialogOpen(open)
          if (!open) {
            setRemoveFieldTarget(null)
          }
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
            <Button variant="destructive" onClick={handleConfirmRemoveField}>
              删除字段
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
