'use client'

import { useState, useMemo, useEffect, Suspense, lazy } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useMutation, useQuery, useLazyQuery } from '@apollo/client'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import {
  Table2,
  Search,
  Plus,
  Download,
  MoreVertical,
  ChevronsUpDown,
  X,
  Filter,
  Loader2,
  AlertTriangle,
  ExternalLink,
  Edit,
  Key,
  Settings,
  Trash2,
  Archive,
  Link2,
} from 'lucide-react'

// 懒加载动态表格组件
const DynamicModelTable = lazy(() => import('@web/components/model-editor/DynamicModelTable'))
import { InsertFieldSheet } from '@web/components/model-editor/InsertFieldSheet'
import { ImportModelDialog } from '@web/components/model-editor/ImportModelDialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
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
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@web/components/ui/drawer'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { TEST_CLUSTER_CONNECTION } from '@web/graphql/mutations/cluster'
import { CREATE_MODEL, UPDATE_MODEL, DELETE_MODEL, UPDATE_FIELD, REMOVE_FIELD, DEPRECATE_FIELD, UNDEPRECATE_FIELD } from '@web/graphql/mutations/model'
import { GET_LOGICAL_FOREIGN_KEYS, GET_MODEL } from '@web/graphql/queries/model'
import {
  CREATE_LOGICAL_FOREIGN_KEY,
  DELETE_LOGICAL_FOREIGN_KEY,
} from '@web/graphql/mutations/model'
import { GET_MODELS } from '@web/graphql/queries/model'
import { useDatabases } from '@web/hooks/useDatabases'
import { toast } from 'sonner'
import type { LogicalForeignKey, DbColumnInfo, CreateLogicalForeignKeyResult, DeleteLogicalForeignKeyResult } from '@/types'

// Model 类型定义
interface Model {
  id: string
  name: string
  title: string
  description?: string
  databaseName: string
  storageType?: string
}

// 模型详情类型（包含字段信息）
interface ModelField {
  name: string
  title: string
  format?: string
  schemaType?: string
  storageHint?: string
  nonNull?: boolean
  required?: boolean
  isPrimary?: boolean
  description?: string
  isDeprecated?: boolean
  dbColumn?: DbColumnInfo
}

interface ModelDetail extends Model {
  fields: ModelField[]
}

// GraphQL response types
interface ModelsEdge {
  node: Model
}

interface ModelsQueryData {
  models?: {
    edges: ModelsEdge[]
  }
}

interface ModelQueryData {
  model?: {
    model?: ModelDetail
    error?: { message: string }
  }
}

interface TestConnectionResult {
  testDatabaseConnection?: {
    success: boolean
    error?: {
      message: string
    }
  }
}

interface CreateModelResult {
  createModel?: {
    model?: { id: string }
    error?: { message: string }
  }
}

interface DeleteModelResult {
  deleteModel?: {
    success?: boolean
    error?: { message: string }
  }
}

interface UpdateModelMetaResult {
  updateModelMeta?: {
    model?: ModelDetail
    error?: { message: string }
  }
}

interface ForeignKeysQueryData {
  logicalForeignKeys?: LogicalForeignKey[]
}


export default function ModelEditorPage() {
  const params = useParams()
  const router = useRouter()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const [selectedDatabase, setSelectedDatabase] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null)
  const [databaseOpen, setDatabaseOpen] = useState(false)
  const [connectionChecking, setConnectionChecking] = useState(true)
  const [connectionFailed, setConnectionFailed] = useState(false)
  const [connectionError, setConnectionError] = useState<string>('')
  const [createModelOpen, setCreateModelOpen] = useState(false)
  const [newModelName, setNewModelName] = useState('')
  const [newModelTitle, setNewModelTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [importDialogOpen, setImportDialogOpen] = useState(false)

  // 编辑模型状态
  const [editModelOpen, setEditModelOpen] = useState(false)
  const [editModelId, setEditModelId] = useState<string | null>(null)
  const [editModelData, setEditModelData] = useState<ModelDetail | null>(null)
  const [editModelLoading, setEditModelLoading] = useState(false)

  // 删除模型确认对话框状态
  const [deleteModelDialogOpen, setDeleteModelDialogOpen] = useState(false)
  const [modelToDelete, setModelToDelete] = useState<Model | null>(null)
  const [deletingModel, setDeletingModel] = useState(false)

  // 元信息内联编辑状态
  const [metaTitle, setMetaTitle] = useState('')
  const [metaDescription, setMetaDescription] = useState('')
  const [metaSaving, setMetaSaving] = useState(false)
  const [metaEditMode, setMetaEditMode] = useState(false)

  // 插入字段状态
  const [insertFieldOpen, setInsertFieldOpen] = useState(false)

  // 编辑字段状态
  const [editFieldOpen, setEditFieldOpen] = useState(false)
  const [editingField, setEditingField] = useState<ModelField | null>(null)
  const [editFieldTitle, setEditFieldTitle] = useState('')
  const [editFieldDescription, setEditFieldDescription] = useState('')

  // 逻辑外键状态
  const [fkList, setFkList] = useState<LogicalForeignKey[]>([])
  const [fkLoading, setFkLoading] = useState(false)
  const [fkFormOpen, setFkFormOpen] = useState(false)
  const [fkRefModelId, setFkRefModelId] = useState('')
  const [fkRefModelDetail, setFkRefModelDetail] = useState<ModelDetail | null>(null)
  const [fkRefModelLoading, setFkRefModelLoading] = useState(false)
  const [fkMappings, setFkMappings] = useState<{ sourceField: string; targetField: string }[]>([
    { sourceField: '', targetField: '' },
  ])
  const [fkSubmitting, setFkSubmitting] = useState(false)
  const [fkDeleteConfirm, setFkDeleteConfirm] = useState<string | null>(null)

  // org-scoped context for operations that belong to the org schema (e.g. testDatabaseConnection)
  const orgScopedContext = useMemo(() => {
    if (!orgName) return undefined
    return { uri: `/graphql/org/${orgName}/` }
  }, [orgName])

  // project-scoped context for model/field/FK operations (project schema)
  const projectScopedContext = useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])

  const [testConnection] = useMutation<TestConnectionResult>(TEST_CLUSTER_CONNECTION)
  const [createModelMutation] = useMutation<CreateModelResult>(CREATE_MODEL, { context: projectScopedContext })
  const [updateModelMutation] = useMutation<UpdateModelMetaResult>(UPDATE_MODEL, { context: projectScopedContext })
  const [updateFieldMutation] = useMutation(UPDATE_FIELD, { context: projectScopedContext })
  const [deprecateFieldMutation] = useMutation(DEPRECATE_FIELD, { context: projectScopedContext })
  const [undeprecateFieldMutation] = useMutation(UNDEPRECATE_FIELD, { context: projectScopedContext })
  const [removeFieldMutation] = useMutation(REMOVE_FIELD, { context: projectScopedContext })
  const [deleteModelMutation] = useMutation<DeleteModelResult>(DELETE_MODEL, {
    context: projectScopedContext,
  })

  // 懒加载获取模型详情
  const [fetchModelDetail] = useLazyQuery<ModelQueryData>(GET_MODEL, {
    fetchPolicy: 'network-only',
    context: projectScopedContext,
  })

  // 逻辑外键 Apollo hooks
  const [fetchForeignKeys] = useLazyQuery<ForeignKeysQueryData>(GET_LOGICAL_FOREIGN_KEYS, {
    fetchPolicy: 'network-only',
    context: projectScopedContext,
  })

  const [createFKMutation] = useMutation<{ createLogicalForeignKey?: { result?: CreateLogicalForeignKeyResult } }>(CREATE_LOGICAL_FOREIGN_KEY, {
    context: projectScopedContext,
  })

  const [deleteFKMutation] = useMutation<{ deleteLogicalForeignKey?: { result?: DeleteLogicalForeignKeyResult } }>(DELETE_LOGICAL_FOREIGN_KEY, {
    context: projectScopedContext,
  })

  // Check cluster connectivity on mount
  useEffect(() => {
    if (!projectSlug || !orgName) return

    let cancelled = false

    const check = async () => {
      setConnectionChecking(true)
      try {
        const result = await testConnection({
          variables: { input: { projectSlug } },
          context: orgScopedContext,
        })
        if (cancelled) return
        const payload = result.data?.testDatabaseConnection
        if (!payload?.success) {
          const errNode = payload?.error
          const msg = errNode?.message ?? '数据库连接失败'
          setConnectionError(msg)
          setConnectionFailed(true)
        }
      } catch {
        if (cancelled) return
        setConnectionError('无法连接到数据库集群')
        setConnectionFailed(true)
      } finally {
        if (!cancelled) setConnectionChecking(false)
      }
    }

    check()
    return () => { cancelled = true }
  }, [projectSlug, orgName]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleGoToCluster = () => {
    router.push(`/org/${orgName}/projects/${projectSlug}/cluster`)
  }

  // Fetch databases using standardized hook (skip when connection checking or failed)
  const { databases, loading: databasesLoading } = useDatabases(
    !connectionChecking && !connectionFailed ? projectSlug : null,
    { initialLimit: 50 }
  )

  // Set default selected database when data loads
  useEffect(() => {
    if (databases.length > 0 && !selectedDatabase) {
      setSelectedDatabase(databases[0].name)
    }
  }, [databases, selectedDatabase])

  // 当选择引用模型时，懒加载引用模型的字段
  useEffect(() => {
    if (!fkRefModelId || !projectSlug) {
      setFkRefModelDetail(null)
      return
    }
    let cancelled = false
    setFkRefModelLoading(true)
    fetchModelDetail({ variables: { id: fkRefModelId } })
      .then(result => {
        if (cancelled) return
        const m = result.data?.model?.model
        if (m) setFkRefModelDetail(m as ModelDetail)
      })
      .finally(() => {
        if (!cancelled) setFkRefModelLoading(false)
      })
    return () => { cancelled = true }
  }, [fkRefModelId, projectSlug]) // eslint-disable-line react-hooks/exhaustive-deps

  // Fetch models from API
  const { data: modelsData, loading: modelsLoading, refetch: refetchModels } = useQuery<ModelsQueryData>(GET_MODELS, {
    variables: {
      input: {
        databaseName: selectedDatabase,
        limit: 100,
      },
    },
    skip: !projectSlug || !selectedDatabase || connectionChecking || connectionFailed,
    context: projectScopedContext,
  })

  // Extract models from query result
  const models: Model[] = useMemo(() => {
    if (!modelsData?.models?.edges) return []
    return modelsData.models.edges.map((edge) => edge.node)
  }, [modelsData])

  // Filter models by search query
  const filteredModels = useMemo(() => {
    if (!searchQuery) return models
    const query = searchQuery.toLowerCase()
    return models.filter(m =>
      m.name.toLowerCase().includes(query) ||
      m.title?.toLowerCase().includes(query)
    )
  }, [models, searchQuery])

  const handleModelClick = (modelId: string) => {
    setSelectedModelId(modelId)
    // 不再跳转到新页面，而是在当前页面右侧显示数据表格
    // router.push(`/org/${orgName}/projects/${projectSlug}/model-editor/${modelId}`)
  }

  const handleModelDetailClick = (modelId: string) => {
    handleModelClick(modelId)
  }

  const handleCreateModel = () => {
    if (!selectedDatabase) {
      alert('请先选择数据库')
      return
    }
    setCreateModelOpen(true)
  }

  const handleConfirmCreateModel = async () => {
    if (!newModelName.trim() || !newModelTitle.trim()) {
      alert('请填写模型标识和展示名称')
      return
    }

    if (!selectedDatabase || !projectSlug) {
      alert('缺少必要参数')
      return
    }

    setCreating(true)
    try {
      const result = await createModelMutation({
        variables: {
          input: {
            name: newModelName.trim(),
            title: newModelTitle.trim(),
            databaseName: selectedDatabase,
          },
        },
        context: projectScopedContext,
      })

      if (result.data?.createModel?.model) {
        const modelId = result.data.createModel.model.id
        setCreateModelOpen(false)
        setNewModelName('')
        setNewModelTitle('')
        // Refetch models list
        refetchModels()
        // 选择新创建的模型，在当前页面显示
        setSelectedModelId(modelId)
      } else if (result.data?.createModel?.error) {
        alert(result.data.createModel.error.message || '创建失败')
      }
    } catch (error) {
      console.error('创建模型失败:', error)
      alert('创建模型失败')
    } finally {
      setCreating(false)
    }
  }

  // 打开编辑模型弹窗
  const handleEditModel = async (modelId: string) => {
    setEditModelId(modelId)
    setEditModelOpen(true)
    setEditModelLoading(true)
    setEditModelData(null)

    try {
      const { data } = await fetchModelDetail({
        variables: { id: modelId, withActualSchema: true },
      })

      if (data?.model?.model) {
        setEditModelData(data.model.model as ModelDetail)
        setMetaTitle(data.model.model.title || '')
        setMetaDescription(data.model.model.description || '')
        setFkList([])
        setFkFormOpen(false)
        setFkMappings([{ sourceField: '', targetField: '' }])
        setFkRefModelId('')
        loadForeignKeys(modelId)
      } else if (data?.model?.error) {
        alert(data.model.error.message || '获取模型详情失败')
        setEditModelOpen(false)
      }
    } catch (error) {
      console.error('获取模型详情失败:', error)
      alert('获取模型详情失败')
      setEditModelOpen(false)
    } finally {
      setEditModelLoading(false)
    }
  }

  // 关闭编辑模型弹窗
  const handleCloseEditModel = () => {
    setEditModelOpen(false)
    setEditModelId(null)
    setEditModelData(null)
    setFkList([])
    setFkFormOpen(false)
    setFkRefModelId('')
    setFkMappings([{ sourceField: '', targetField: '' }])
    setFkDeleteConfirm(null)
    setFkRefModelDetail(null)
  }

  // 删除模型处理函数
  const handleDeleteModel = async () => {
    if (!modelToDelete) return
    setDeletingModel(true)
    try {
      const result = await deleteModelMutation({
        variables: {
          id: modelToDelete.id,
        },
      })

      if (result.data?.deleteModel?.success) {
        toast.success('模型删除成功')
        setDeleteModelDialogOpen(false)
        setModelToDelete(null)
        if (selectedModelId === modelToDelete.id) {
          setSelectedModelId(null)
          handleCloseEditModel()
        }
        refetchModels()
      } else {
        const err = result.data?.deleteModel?.error
        toast.error(err?.message || '删除模型失败')
      }
    } catch (error) {
      console.error('删除模型失败:', error)
      toast.error('删除模型失败')
    } finally {
      setDeletingModel(false)
    }
  }

  // 保存元信息
  const handleSaveMeta = async () => {
    if (!editModelId || !projectSlug) return
    setMetaSaving(true)
    try {
      const result = await updateModelMutation({
        variables: {
          id: editModelId,
          input: { title: metaTitle, description: metaDescription },
        },
      })
      if (result.data?.updateModelMeta?.model) {
        setEditModelData(prev => prev ? { ...prev, title: metaTitle, description: metaDescription } : prev)
        toast.success('保存成功')
      } else {
        toast.error(result.data?.updateModelMeta?.error?.message || '保存失败')
      }
    } catch {
      toast.error('保存失败，请重试')
    } finally {
      setMetaSaving(false)
    }
  }

  const loadForeignKeys = async (modelId: string) => {
    setFkLoading(true)
    try {
      const result = await fetchForeignKeys({
        variables: { modelId },
      })
      setFkList(result.data?.logicalForeignKeys ?? [])
    } catch {
      setFkList([])
    } finally {
      setFkLoading(false)
    }
  }

  const handleCreateFK = async () => {
    if (!editModelId || !fkRefModelId) return
    const validMappings = fkMappings.filter(m => m.sourceField && m.targetField)
    if (validMappings.length === 0) return

    setFkSubmitting(true)
    try {
      const result = await createFKMutation({
        variables: {
          input: {
            modelId: editModelId,
            refModelId: fkRefModelId,
            sourceFields: validMappings.map(m => m.sourceField),
            targetFields: validMappings.map(m => m.targetField),
          },
        },
      })
      const r = result.data?.createLogicalForeignKey?.result
      if (r?.__typename === 'LogicalForeignKey') {
        toast.success('外键创建成功')
        setFkFormOpen(false)
        setFkRefModelId('')
        setFkMappings([{ sourceField: '', targetField: '' }])
        await loadForeignKeys(editModelId)
      } else if (r?.__typename === 'FKColumnsNotFoundError') {
        toast.error(`字段不存在：${r.message}`)
      } else if (r?.__typename === 'FKFieldCountMismatchError') {
        toast.error('源字段与目标字段数量不匹配')
      }
    } catch {
      toast.error('创建外键失败，请重试')
    } finally {
      setFkSubmitting(false)
    }
  }

  const handleDeleteFK = async (pairId: string) => {
    try {
      const result = await deleteFKMutation({
        variables: { pairId },
      })
      const r = result.data?.deleteLogicalForeignKey?.result
      if (r?.__typename === 'DeleteLogicalForeignKeySuccess') {
        toast.success('外键已删除')
        if (editModelId) await loadForeignKeys(editModelId)
      } else if (r?.__typename === 'FKPairHasRelateFieldsError') {
        toast.error('该外键关联了关系字段，请先删除相关字段')
      } else if (r?.__typename === 'FKNotFoundError') {
        toast.error('外键不存在')
      }
    } catch {
      toast.error('删除外键失败，请重试')
    } finally {
      setFkDeleteConfirm(null)
    }
  }

  // 获取字段类型的图标
  const getFieldTypeIcon = (schemaType?: string) => {
    switch (schemaType?.toLowerCase()) {
      case 'integer':
      case 'number':
        return <span className="font-mono text-[10px] text-blue-600">#</span>
      case 'boolean':
        return <span className="font-mono text-[10px] text-emerald-600">✓</span>
      case 'datetime':
        return <span className="font-mono text-[10px] text-violet-600">◷</span>
      default:
        return <span className="font-mono text-[10px] text-muted-foreground">T</span>
    }
  }

  return (
    <div className="relative flex size-full">
      {/* Cluster connection failure dialog */}
      <Dialog open={connectionFailed} onOpenChange={() => {}}>
        <DialogContent
          className="sm:max-w-md"
          onInteractOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => e.preventDefault()}
        >
          <DialogHeader>
            <div className="mb-1 flex items-center gap-3">
              <div className="flex size-10 flex-shrink-0 items-center justify-center rounded-full bg-destructive/10">
                <AlertTriangle className="size-5 text-destructive" />
              </div>
              <DialogTitle className="font-heading text-base">数据库连接失败</DialogTitle>
            </div>
            <DialogDescription className="pl-[52px] text-sm leading-relaxed">
              {connectionError}
              <br />
              <span className="text-muted-foreground">请前往集群配置页检查数据库连接信息。</span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="flex gap-2 sm:justify-end">
            <Button variant="outline" size="sm" onClick={() => router.back()}>
              返回
            </Button>
            <Button size="sm" onClick={handleGoToCluster}>
              <ExternalLink className="mr-1.5 size-3.5" />
              前往集群配置
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Model Sheet */}
      <Sheet open={createModelOpen} onOpenChange={setCreateModelOpen}>
        <SheetContent className="w-[400px] sm:max-w-[400px]">
          <SheetHeader>
            <SheetTitle className="font-heading text-base">新建模型</SheetTitle>
            <SheetDescription className="text-sm">
              创建一个新的数据模型，用于定义数据结构。
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 py-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">模型标识 <span className="text-destructive">*</span></label>
              <Input
                placeholder="例如：user_profile"
                value={newModelName}
                onChange={(e) => setNewModelName(e.target.value)}
                className="text-sm"
              />
              <p className="text-xs text-muted-foreground">英文字母、数字、下划线组成，用于代码引用</p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">模型展示名称 <span className="text-destructive">*</span></label>
              <Input
                placeholder="例如：用户档案"
                value={newModelTitle}
                onChange={(e) => setNewModelTitle(e.target.value)}
                className="text-sm"
              />
              <p className="text-xs text-muted-foreground">中文显示名称，便于理解</p>
            </div>
          </div>
          <SheetFooter className="flex gap-2 sm:justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setCreateModelOpen(false)
                setNewModelName('')
                setNewModelTitle('')
              }}
            >
              取消
            </Button>
            <Button
              size="sm"
              className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
              onClick={handleConfirmCreateModel}
              disabled={creating}
            >
              {creating && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
              {creating ? '创建中...' : '创建'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      {/* Edit Model Drawer */}
      <Drawer open={editModelOpen} onOpenChange={handleCloseEditModel} direction="right">
        <DrawerContent direction="right" className="flex w-[680px] flex-col rounded-none">
          {/* Insert Field Sheet - nested inside Edit Model Drawer for correct z-index layering */}
          <InsertFieldSheet
            open={insertFieldOpen}
            onOpenChange={setInsertFieldOpen}
            modelId={editModelId || ''}
            modelName={editModelData?.name}
            projectSlug={projectSlug}
            orgName={orgName}
            onSuccess={async () => {
              if (editModelId) {
                const { data } = await fetchModelDetail({ variables: { id: editModelId, withActualSchema: true } })
                if (data?.model?.model) setEditModelData(data.model.model as ModelDetail)
              }
            }}
          />
          {/* Header */}
          <div className="flex shrink-0 items-start justify-between border-b border-border px-6 py-4">
            <div className="min-w-0">
              <h2 className="font-heading text-base font-semibold text-foreground">
                {editModelData?.title || editModelData?.name || '模型详情'}
              </h2>
              <p className="mt-0.5 font-mono text-xs text-muted-foreground">{editModelData?.name}</p>
            </div>
            <button
              onClick={handleCloseEditModel}
              className="ml-4 shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <X className="size-4" />
            </button>
          </div>

          {/* Scrollable body */}
          <div className="min-h-0 flex-1 overflow-y-auto">
          {editModelLoading ? (
            <div className="flex flex-col items-center justify-center py-24">
              <Loader2 className="mb-3 size-5 animate-spin text-muted-foreground" />
              <span className="text-sm text-muted-foreground">加载中...</span>
            </div>
          ) : editModelData ? (
            <div className="divide-y divide-border [&>div]:py-6">
              {/* ── 元信息 ─────────────────────────────── */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <span className="text-sm font-semibold text-foreground">元信息</span>
                  {!metaEditMode && (
                    <button
                      type="button"
                      title="编辑元信息"
                      className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      onClick={() => setMetaEditMode(true)}
                    >
                      <Edit className="size-3.5" />
                    </button>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-x-6 gap-y-3">
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">标识名称</label>
                    <Input
                      value={editModelData.name}
                      disabled
                      className="h-8 bg-muted/30 font-mono text-xs"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">显示标题</label>
                    <Input
                      value={metaTitle}
                      onChange={(e) => setMetaTitle(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入显示标题"
                      disabled={!metaEditMode}
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">数据库</label>
                    <Input
                      value={editModelData.databaseName}
                      disabled
                      className="h-8 bg-muted/30 font-mono text-xs"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">描述</label>
                    <Input
                      value={metaDescription}
                      onChange={(e) => setMetaDescription(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入模型描述"
                      disabled={!metaEditMode}
                    />
                  </div>
                </div>
                {metaEditMode && (
                  <div className="mt-4 flex items-center justify-end gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => {
                        setMetaTitle(editModelData.title || '')
                        setMetaDescription(editModelData.description || '')
                        setMetaEditMode(false)
                      }}
                    >
                      取消
                    </Button>
                    <Button
                      size="sm"
                      className="h-7 bg-[#2563eb] px-4 text-xs text-white hover:bg-[#1d4ed8]"
                      disabled={metaSaving}
                      onClick={async () => {
                        await handleSaveMeta()
                        setMetaEditMode(false)
                      }}
                    >
                      {metaSaving && <Loader2 className="mr-1.5 size-3 animate-spin" />}
                      保存更改
                    </Button>
                  </div>
                )}
              </div>

              {/* ── 字段定义 ─────────────────────────────── */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-foreground">字段定义</span>
                    <span className="rounded-full bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                      {editModelData.fields?.length || 0}
                    </span>
                  </div>
                  <button
                    type="button"
                    title="插入字段"
                    className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    onClick={() => setInsertFieldOpen(true)}
                  >
                    <Plus className="size-3.5" />
                  </button>
                </div>

                {/* Supabase 风格字段表格 */}
                <div className="overflow-hidden rounded-lg border border-border bg-card">
                  {editModelData.fields && editModelData.fields.length > 0 ? (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        {/* 表头 */}
                        <thead>
                          <tr className="border-b border-border bg-muted/30">
                            <th className="w-[180px] px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                              标识(名称)
                            </th>
                            <th className="w-[90px] px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                              格式
                            </th>
                            <th className="w-[90px] px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                              类型
                            </th>
                            <th className="w-[80px] px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                              默认值
                            </th>
                            <th className="w-[60px] px-3 py-2 text-center text-xs font-medium text-muted-foreground">
                              主键
                            </th>
                            <th className="w-[50px] px-3 py-2 text-center text-xs font-medium text-muted-foreground">
                              
                            </th>
                          </tr>
                        </thead>
                        {/* 表体 */}
                        <tbody className="divide-y divide-border">
                          {editModelData.fields.map((field) => (
                            <tr
                              key={field.name}
                              className="transition-colors hover:bg-muted/20"
                            >
                              {/* Name 列 - 包含 name 和 title */}
                              <td className="px-3 py-2">
                                <div className="flex flex-col">
                                  <div className="flex items-center gap-2">
                                    <span className={`font-mono text-sm ${field.isDeprecated ? 'text-muted-foreground line-through' : 'text-foreground'}`}>
                                      {field.name}
                                    </span>
                                    {field.isDeprecated && (
                                      <span className="inline-flex items-center rounded bg-orange-50 px-1.5 py-0.5 font-mono text-xs text-orange-700 dark:bg-orange-900/30 dark:text-orange-400">
                                        已废弃
                                      </span>
                                    )}
                                  </div>
                                  {field.title && (
                                    <span className="truncate text-xs text-muted-foreground">
                                      {field.title}
                                    </span>
                                  )}
                                </div>
                              </td>
                              {/* Format 列 */}
                              <td className="px-3 py-2">
                                <span className="inline-flex items-center rounded bg-emerald-50 px-1.5 py-0.5 font-mono text-xs text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400">
                                  {field.format || '-'}
                                </span>
                              </td>
                              {/* Type 列 */}
                              <td className="px-3 py-2">
                                <span className="inline-flex items-center rounded bg-blue-50 px-1.5 py-0.5 font-mono text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                                  {field.dbColumn?.columnType || field.storageHint || field.schemaType || 'String'}
                                </span>
                              </td>
                              {/* Default 列 */}
                              <td className="px-3 py-2">
                                <span className="font-mono text-xs text-muted-foreground">
                                  {field.dbColumn?.defaultValue !== undefined ? String(field.dbColumn.defaultValue) : '-'}
                                </span>
                              </td>
                              {/* Primary 列 */}
                              <td className="px-3 py-2 text-center">
                                {field.isPrimary ? (
                                  <span className="inline-flex size-5 items-center justify-center rounded bg-amber-100 dark:bg-amber-900/30">
                                    <Key className="size-3 text-amber-600 dark:text-amber-400" />
                                  </span>
                                ) : (
                                  <span className="text-muted-foreground/30">-</span>
                                )}
                              </td>
                              {/* 操作列 */}
                              <td className="p-2 text-center">
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="size-6 p-0 hover:bg-muted"
                                    >
                                      <MoreVertical className="size-3.5 text-muted-foreground" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent align="end" className="w-36">
                                    <DropdownMenuItem
                                      className="cursor-pointer text-xs"
                                      onClick={() => {
                                        setEditingField(field)
                                        setEditFieldTitle(field.title || '')
                                        setEditFieldDescription(field.description || '')
                                        setEditFieldOpen(true)
                                      }}
                                    >
                                      <Edit className="mr-2 size-3.5" />
                                      编辑
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                      className="cursor-pointer text-xs"
                                      onClick={async () => {
                                        if (!editModelData) return
                                        try {
                                          const newDeprecated = !field.isDeprecated
                                          const mutation = newDeprecated ? deprecateFieldMutation : undeprecateFieldMutation
                                          await mutation({
                                            variables: {
                                              modelID: editModelData.id,
                                              fieldName: field.name,
                                            },
                                          })
                                          setEditModelData({
                                            ...editModelData,
                                            fields: editModelData.fields.map(f =>
                                              f.name === field.name
                                                ? { ...f, isDeprecated: newDeprecated }
                                                : f
                                            )
                                          })
                                          toast.success(newDeprecated ? '字段已废弃' : '已取消废弃')
                                        } catch {
                                          toast.error('操作失败，请重试')
                                        }
                                      }}
                                    >
                                      <Archive className="mr-2 size-3.5" />
                                      {field.isDeprecated ? '取消废弃' : '废弃'}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                      className={`cursor-pointer text-xs ${
                                        field.isDeprecated
                                          ? 'text-destructive focus:text-destructive'
                                          : 'cursor-not-allowed text-muted-foreground/50'
                                      }`}
                                      onClick={async () => {
                                        if (!field.isDeprecated || !editModelData) return
                                        try {
                                          await removeFieldMutation({
                                            variables: {
                                              modelID: editModelData.id,
                                              fieldName: field.name,
                                            },
                                          })
                                          setEditModelData({
                                            ...editModelData,
                                            fields: editModelData.fields.filter(f => f.name !== field.name)
                                          })
                                          toast.success('字段已删除')
                                        } catch {
                                          toast.error('删除失败，请重试')
                                        }
                                      }}
                                      disabled={!field.isDeprecated}
                                    >
                                      <Trash2 className="mr-2 size-3.5" />
                                      删除
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <Table2 className="mb-2 size-8 opacity-30" />
                      <p className="text-sm">暂无字段</p>
                      <p className="mt-1 text-xs">点击上方按钮添加字段</p>
                    </div>
                  )}
                </div>
              </div>

              {/* 逻辑外键 */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-foreground">关联关系</span>
                    <span className="inline-flex items-center rounded-full bg-muted px-1.5 py-0.5 text-xs font-medium text-muted-foreground">
                      {fkList.length}
                    </span>
                  </div>
                  {!fkFormOpen && (
                    <button
                      type="button"
                      title="添加关系"
                      className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      onClick={() => {
                        setFkFormOpen(true)
                        setFkRefModelId('')
                        setFkMappings([{ sourceField: '', targetField: '' }])
                      }}
                    >
                      <Plus className="size-3.5" />
                    </button>
                  )}
                </div>

                {/* FK 列表表格 */}
                {fkLoading ? (
                  <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
                    <Loader2 className="size-4 animate-spin" />
                    加载中...
                  </div>
                ) : fkList.length > 0 ? (
                  <div className="overflow-hidden rounded-lg border border-border bg-card">
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b border-border bg-muted/30">
                            <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                              关系
                            </th>
                            <th className="w-[80px] px-3 py-2 text-center text-xs font-medium text-muted-foreground">
                              方向
                            </th>
                            <th className="w-[50px] px-3 py-2 text-center text-xs font-medium text-muted-foreground" />
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {fkList.map((fk) => {
                            const srcName = fk.modelName
                            const refName = fk.refModelName
                            const pairs = fk.sourceFields.map((sf, i) => {
                              const tf = fk.targetFields[i] ?? '?'
                              return `${srcName}.${sf} → ${refName}.${tf}`
                            })
                            const isNormal = fk.direction === 'NORMAL'
                            return (
                              <tr key={fk.id} className="transition-colors hover:bg-muted/20">
                                <td className="px-3 py-2">
                                  <div className="flex flex-col gap-0.5">
                                    {pairs.map((p, i) => (
                                      <span key={i} className="font-mono text-xs text-foreground">
                                        {p}
                                      </span>
                                    ))}
                                  </div>
                                </td>
                                <td className="px-3 py-2 text-center">
                                  <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${isNormal ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' : 'bg-orange-500/10 text-orange-600 dark:text-orange-400'}`}>
                                    {isNormal ? '多→一' : '一→多'}
                                  </span>
                                </td>
                                <td className="p-2 text-center">
                                  {fkDeleteConfirm === fk.pairId ? (
                                    <div className="flex items-center gap-1">
                                      <button
                                        className="rounded px-1.5 py-0.5 text-xs text-destructive hover:bg-destructive/10"
                                        onClick={() => handleDeleteFK(fk.pairId)}
                                      >
                                        确认
                                      </button>
                                      <button
                                        className="rounded px-1.5 py-0.5 text-xs text-muted-foreground hover:bg-muted"
                                        onClick={() => setFkDeleteConfirm(null)}
                                      >
                                        取消
                                      </button>
                                    </div>
                                  ) : (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="size-6 p-0 hover:bg-muted hover:text-destructive"
                                      onClick={() => setFkDeleteConfirm(fk.pairId)}
                                    >
                                      <Trash2 className="size-3.5 text-muted-foreground" />
                                    </Button>
                                  )}
                                </td>
                              </tr>
                            )
                          })}
                        </tbody>
                      </table>
                    </div>
                  </div>
                ) : !fkFormOpen ? (
                  <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border py-6 text-muted-foreground">
                    <Link2 className="mb-2 size-6 opacity-30" />
                    <p className="text-sm">暂无逻辑外键</p>
                    <p className="mt-1 text-xs">点击"添加关系"创建逻辑外键</p>
                  </div>
                ) : null}

                {/* 内联创建表单 */}
                {fkFormOpen && (
                  <div className="space-y-3 rounded-lg border border-border bg-muted/10 p-4">
                    <h4 className="text-xs font-semibold text-foreground">新建逻辑外键</h4>

                    {/* 引用模型选择 */}
                    <div className="flex items-center gap-2">
                      <label className="w-20 shrink-0 text-xs text-muted-foreground">引用模型</label>
                      <Select value={fkRefModelId} onValueChange={(v) => {
                        setFkRefModelId(v)
                        setFkMappings([{ sourceField: '', targetField: '' }])
                      }}>
                        <SelectTrigger className="h-7 text-xs">
                          <SelectValue placeholder="选择引用模型" />
                        </SelectTrigger>
                        <SelectContent>
                          {models
                            .filter(m => m.id !== editModelId)
                            .map(m => (
                              <SelectItem key={m.id} value={m.id} className="font-mono text-xs">
                                {m.name}{m.title && m.title !== m.name ? ` (${m.title})` : ''}
                              </SelectItem>
                            ))}
                        </SelectContent>
                      </Select>
                    </div>

                    {/* 字段映射 */}
                    <div className="space-y-2">
                      <label className="text-xs text-muted-foreground">字段映射（源字段 → 目标字段）</label>
                      {fkMappings.map((mapping, idx) => (
                        <div key={idx} className="flex items-center gap-2">
                          {/* 源字段（当前模型） */}
                          <Select
                            value={mapping.sourceField}
                            onValueChange={(v) => {
                              const next = [...fkMappings]
                              next[idx] = { ...next[idx], sourceField: v }
                              setFkMappings(next)
                            }}
                          >
                            <SelectTrigger className="h-7 text-xs">
                              <SelectValue placeholder="源字段" />
                            </SelectTrigger>
                            <SelectContent>
                              {editModelData?.fields?.map(f => (
                                <SelectItem key={f.name} value={f.name} className="font-mono text-xs">
                                  {f.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>

                          <span className="shrink-0 text-xs text-muted-foreground">→</span>

                          {/* 目标字段（引用模型） */}
                          <Select
                            value={mapping.targetField}
                            disabled={!fkRefModelId || fkRefModelLoading}
                            onValueChange={(v) => {
                              const next = [...fkMappings]
                              next[idx] = { ...next[idx], targetField: v }
                              setFkMappings(next)
                            }}
                          >
                            <SelectTrigger className="h-7 text-xs">
                              <SelectValue placeholder={fkRefModelLoading ? '加载中...' : '目标字段'} />
                            </SelectTrigger>
                            <SelectContent>
                              {fkRefModelDetail?.fields?.map((f: ModelField) => (
                                <SelectItem key={f.name} value={f.name} className="font-mono text-xs">
                                  {f.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>

                          {/* 删除此行映射 */}
                          {fkMappings.length > 1 && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="size-6 shrink-0 p-0 hover:text-destructive"
                              onClick={() => setFkMappings(fkMappings.filter((_, i) => i !== idx))}
                            >
                              <Trash2 className="size-3" />
                            </Button>
                          )}
                        </div>
                      ))}

                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs text-muted-foreground hover:text-foreground"
                        onClick={() => setFkMappings([...fkMappings, { sourceField: '', targetField: '' }])}
                      >
                        <Plus className="mr-1 size-3" />
                        添加映射
                      </Button>
                    </div>

                    {/* 操作按钮 */}
                    <div className="flex justify-end gap-2 pt-1">
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => {
                          setFkFormOpen(false)
                          setFkRefModelId('')
                          setFkMappings([{ sourceField: '', targetField: '' }])
                        }}
                      >
                        取消
                      </Button>
                      <Button
                        size="sm"
                        className="h-7 text-xs"
                        disabled={
                          fkSubmitting ||
                          !fkRefModelId ||
                          fkMappings.every(m => !m.sourceField || !m.targetField)
                        }
                        onClick={handleCreateFK}
                      >
                        {fkSubmitting && <Loader2 className="mr-1 size-3 animate-spin" />}
                        创建外键
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          ) : null}
          </div>

          <div className="flex shrink-0 justify-end border-t border-border px-6 py-4">
            <Button
              variant="outline"
              size="sm"
              onClick={handleCloseEditModel}
            >
              关闭
            </Button>
          </div>
        </DrawerContent>
      </Drawer>

      {/* Edit Field Sheet */}
      <Sheet open={editFieldOpen} onOpenChange={setEditFieldOpen}>
        <SheetContent className="w-[480px] overflow-y-auto sm:max-w-[480px]">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2 font-heading text-base">
              <Settings className="size-4 text-[#2563eb]" />
              编辑字段
            </SheetTitle>
            <SheetDescription className="text-sm">
              编辑字段 <span className="font-mono text-[#2563eb]">{editingField?.name}</span> 的配置
            </SheetDescription>
          </SheetHeader>

          {editingField && (
            <div className="space-y-4 py-4">
              {/* 字段名称 - 只读 */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">字段名称</label>
                <Input
                  value={editingField.name}
                  disabled
                  className="bg-muted/30 font-mono text-sm"
                />
              </div>

              {/* 显示名称 - 可编辑 */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">显示名称 (DisplayName)</label>
                <Input
                  value={editFieldTitle}
                  onChange={(e) => setEditFieldTitle(e.target.value)}
                  className="text-sm"
                  placeholder="字段显示名称"
                />
              </div>

              {/* Format - 只读 */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">Format</label>
                <Input
                  value={editingField.format || '-'}
                  disabled
                  className="bg-muted/30 font-mono text-sm"
                />
              </div>

              {/* Type - 只读 */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">Type</label>
                <Input
                  value={editingField.storageHint || editingField.schemaType || 'String'}
                  disabled
                  className="bg-muted/30 font-mono text-sm"
                />
              </div>

              {/* 属性标记 - 只读 */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-muted-foreground">属性</label>
                <div className="flex flex-wrap gap-2">
                  {editingField.isPrimary && (
                    <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fef3c7', color: '#d97706' }}>
                      <Key className="mr-1 size-3" />
                      主键
                    </span>
                  )}
                  {editingField.nonNull && (
                    <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fee2e2', color: '#ef4444' }}>
                      必填
                    </span>
                  )}
                  {editingField.required && (
                    <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fef3c7', color: '#d97706' }}>
                      必需
                    </span>
                  )}
                  {!editingField.isPrimary && !editingField.nonNull && !editingField.required && (
                    <span className="text-xs text-muted-foreground">无特殊属性</span>
                  )}
                </div>
              </div>

              {/* 描述 - 可编辑 */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">描述</label>
                <Textarea
                  value={editFieldDescription}
                  onChange={(e) => setEditFieldDescription(e.target.value)}
                  className="min-h-[80px] resize-none text-sm"
                  placeholder="字段描述"
                />
              </div>
            </div>
          )}

          <SheetFooter className="mt-4 flex gap-2 border-t border-border pt-4 sm:justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditFieldOpen(false)}
            >
              取消
            </Button>
            <Button
              size="sm"
              className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
              onClick={() => {
                // TODO: 保存字段修改
                console.log('保存字段:', {
                  name: editingField?.name,
                  title: editFieldTitle,
                  description: editFieldDescription,
                })
                setEditFieldOpen(false)
              }}
            >
              保存
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      {/* Import Model Dialog */}
      <ImportModelDialog
        open={importDialogOpen}
        onOpenChange={setImportDialogOpen}
        projectSlug={projectSlug}
        databaseName={selectedDatabase}
        onSuccess={() => refetchModels()}
      />

      {/* Connection checking overlay */}
      {connectionChecking && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
          <div className="flex flex-col items-center gap-3 text-muted-foreground">
            <Loader2 className="size-6 animate-spin" />
            <span className="text-sm">正在检查数据库连接...</span>
          </div>
        </div>
      )}

      {/* Left Sidebar - Model List */}
      <aside className="flex w-[260px] flex-shrink-0 flex-col border-r border-border bg-sidebar">
        {/* Header */}
        <header className="flex min-h-[var(--header-height,56px)] items-center border-b border-border px-6">
          <h1 className="font-heading text-lg font-semibold text-foreground">模型编辑器</h1>
        </header>

        {/* Content */}
        <div className="flex flex-1 flex-col gap-4 overflow-hidden pt-4">
          {/* Controls Section */}
          <div className="flex flex-col gap-2 px-4">
            {/* Database Selector */}
            <Popover open={databaseOpen} onOpenChange={setDatabaseOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-strong hover:border-stronger h-7 w-full justify-between bg-muted px-2.5 text-xs font-normal transition-colors hover:bg-accent"
                  disabled={databasesLoading}
                >
                  <span className="flex items-center gap-1.5 truncate">
                    <span className="text-muted-foreground">database</span>
                    {databasesLoading ? (
                      <Loader2 className="size-3 animate-spin" />
                    ) : (
                      <span className="text-foreground">{selectedDatabase || 'Select...'}</span>
                    )}
                  </span>
                  <ChevronsUpDown className="size-3.5 shrink-0 text-muted-foreground" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[228px] border border-slate-200 p-1 shadow-lg" align="start">
                {databases.length === 0 ? (
                  <div className="px-2.5 py-3 text-center text-sm text-muted-foreground">
                    No databases found
                  </div>
                ) : (
                  databases.map((db) => (
                    <button
                      key={db.name}
                      type="button"
                      className={cn(
                        'w-full text-left px-2.5 py-1.5 text-sm rounded-sm transition-colors cursor-pointer',
                        selectedDatabase === db.name
                          ? 'bg-selected text-foreground'
                          : 'text-muted-foreground hover:bg-selected hover:text-foreground'
                      )}
                      onClick={() => {
                        setSelectedDatabase(db.name)
                        setDatabaseOpen(false)
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <span>{db.name}</span>
                      </div>
                    </button>
                  ))
                )}
              </PopoverContent>
            </Popover>

            {/* New Model Button */}
            <Button
              size="sm"
              className="h-7 w-full justify-start border-0 bg-[#2563eb] px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
              onClick={handleCreateModel}
            >
              <Plus className="mr-1.5 size-3.5" />
              <span>新建模型</span>
            </Button>

            {/* Import Model Button */}
            <button
              className="border-strong hover:border-stronger inline-flex h-7 w-full items-center justify-start gap-2 rounded-md border bg-muted px-2.5 text-xs font-normal shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
              onClick={() => setImportDialogOpen(true)}
              disabled={!selectedDatabase}
            >
              <Download className="mr-1.5 size-3.5" strokeWidth={1.5} />
              <span>导入模型</span>
            </button>
          </div>

          {/* Search & Filter */}
          <div className="flex items-center gap-2 px-4">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="text"
                placeholder="查询模型..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="border-control focus-visible:ring-background-control h-7 bg-foreground/[.026] px-8 text-xs focus-visible:ring-2 md:h-7"
              />
              {searchQuery && (
                <button
                  type="button"
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                  onClick={() => setSearchQuery('')}
                >
                  <X className="size-3.5" />
                </button>
              )}
            </div>
            <Button
              variant="outline"
              size="icon"
              className="border-strong hover:border-stronger size-7 shrink-0 border-dashed bg-transparent transition-colors hover:bg-accent"
            >
              <Filter className="size-3.5 text-muted-foreground" />
            </Button>
          </div>

          {/* Model List */}
          <nav className="min-h-0 flex-1 overflow-y-auto px-2 pb-4">
            <div className="space-y-px">
              {modelsLoading && (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <Loader2 className="mb-3 size-6 animate-spin" />
                  <p className="text-sm">加载模型中...</p>
                </div>
              )}

              {!modelsLoading && filteredModels.map((model) => (
                <div
                  key={model.id}
                  role="button"
                  tabIndex={0}
                  onClick={() => handleModelDetailClick(model.id)}
                  onKeyDown={(e) => e.key === 'Enter' && handleModelDetailClick(model.id)}
                  className={cn(
                    'group relative flex items-center gap-3 h-7 pl-4 pr-1 rounded-sm cursor-pointer text-sm transition-colors select-none',
                    selectedModelId === model.id
                      ? 'bg-selected text-foreground'
                      : 'text-muted-foreground hover:bg-selected/50 hover:text-foreground'
                  )}
                >
                  {/* Active indicator */}
                  {selectedModelId === model.id && (
                    <div className="absolute inset-y-0 left-0 w-0.5 bg-foreground" />
                  )}

                  {/* Icon */}
                  <Table2 className="group-hover:text-foreground-lighter size-[15px] shrink-0 text-muted-foreground transition-colors" />

                  {/* Name */}
                  <span className={cn(
                    'truncate flex-1 text-sm transition-colors',
                    selectedModelId === model.id ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'
                  )}>
                    {model.name}
                  </span>

                  {/* Title tooltip */}
                  {model.title && model.title !== model.name && (
                    <span className="max-w-[60px] truncate text-xs text-muted-foreground" title={model.title}>
                      {model.title}
                    </span>
                  )}

                  {/* More menu */}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <button
                        type="button"
                        className={cn(
                          'opacity-0 group-hover:opacity-100 transition-opacity w-6 h-6 flex items-center justify-center hover:bg-accent rounded',
                          selectedModelId === model.id && 'opacity-100'
                        )}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="size-3.5" />
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-40 border border-slate-200 shadow-lg">
                      <DropdownMenuItem 
                        className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleEditModel(model.id)
                        }}
                      >
                        <Edit className="mr-2 size-3.5" />
                        编辑模型
                      </DropdownMenuItem>
                      <DropdownMenuItem 
                        className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
                        onClick={async (e) => {
                          e.stopPropagation()
                          try {
                            await navigator.clipboard.writeText(model.name)
                          } catch (err) {
                            console.error('复制失败:', err)
                          }
                        }}
                      >
                        <X className="mr-2 size-3.5 opacity-0" />
                        复制名称
                      </DropdownMenuItem>
                      <DropdownMenuItem 
                        className="cursor-pointer text-xs text-destructive focus:bg-selected focus:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation()
                          setModelToDelete(model)
                          setDeleteModelDialogOpen(true)
                        }}
                      >
                        <X className="mr-2 size-3.5 opacity-0" />
                        删除模型
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              ))}

              {!modelsLoading && filteredModels.length === 0 && selectedDatabase && (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <Table2 className="mb-3 size-10 opacity-20" />
                  <p className="text-sm">暂无模型</p>
                </div>
              )}

              {!selectedDatabase && !databasesLoading && (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <Table2 className="mb-3 size-10 opacity-20" />
                  <p className="text-sm">请先选择数据库</p>
                </div>
              )}
            </div>
          </nav>
        </div>
      </aside>

      {/* Right Content Area */}
      <main className="flex min-w-0 flex-1 flex-col bg-sidebar">
        {selectedModelId ? (
          <Suspense
            fallback={
              <div className="flex flex-1 items-center justify-center">
                <div className="flex flex-col items-center gap-3 text-muted-foreground">
                  <Loader2 className="size-6 animate-spin" />
                  <span className="text-sm">加载中...</span>
                </div>
              </div>
            }
          >
            <DynamicModelTable
              modelId={selectedModelId}
              projectSlug={projectSlug}
              orgName={orgName}
            />
          </Suspense>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <div className="px-6 text-center text-muted-foreground">
              <div className="mx-auto mb-5 flex size-20 items-center justify-center rounded-2xl bg-muted/30">
                <Table2 className="size-10 opacity-30" />
              </div>
              <h2 className="mb-2 font-heading text-base font-semibold text-foreground">选择一个模型</h2>
              <p className="mx-auto max-w-[280px] text-sm leading-relaxed">
                从左侧选择一个模型开始编辑，或点击 &ldquo;新建模型&rdquo; 创建新模型
              </p>
            </div>
          </div>
        )}
      </main>

      {/* Delete Model Confirmation Dialog */}
      <Dialog open={deleteModelDialogOpen} onOpenChange={setDeleteModelDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>删除模型</DialogTitle>
            <DialogDescription>
              确定要删除模型 <span className="font-mono font-semibold text-foreground">{modelToDelete?.name}</span> 吗？此操作无法撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={() => setDeleteModelDialogOpen(false)}
              disabled={deletingModel}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteModel}
              disabled={deletingModel}
            >
              {deletingModel ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  删除中...
                </>
              ) : (
                '删除'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
