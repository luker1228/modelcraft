'use client'

import * as React from 'react'
import { useQuery } from '@apollo/client'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  Shield,
  Loader2,
  Trash2,
  Plus,
  Pencil,
  Check,
  X,
  Clock,
  RotateCcw,
} from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Skeleton } from '@web/components/ui/skeleton'
import { ScrollArea } from '@web/components/ui/scroll-area'
import { Badge } from '@web/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@web/components/ui/alert-dialog'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import { Tabs, TabsList, TabsTrigger } from '@web/components/ui/tabs'
import { PageLayout } from '@web/components/features/layout'

import { useProjectScopedClient } from '@api-client/apollo/public'
import { GET_VIRTUAL_PRESETS_BY_MODEL } from '@/api-client/rbac'
import { GET_MODELS_BY_DATABASE } from '@/api-client/model'
import { DATABASE_CATALOG } from '@/api-client/cluster'
import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type {
  EndUserBundleDataPermissionItem,
  EndUserPermission,
  EndUserPermissionBundleSnapshot,
  EndUserPermissionPreset,
} from '@/types'
import { cn } from '@/shared/utils'

// ── 预设标签映射 ────────────────────────────────────────────────────────────

const PRESET_LABEL: Record<string, string> = {
  READ_WRITE_ALL: '读写全部',
  READ_ALL: '只读全部',
  READ_WRITE_OWNER: '读写本人',
  READ_ALL_WRITE_OWNER: '读所有写本人',
}

const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '更新',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部行',
  SELF: '仅本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '本部门及下级',
}

const COLUMN_MODE_LABEL: Record<string, string> = {
  VISIBLE: '全列可见',
  READONLY: '只读',
  MASKED: '脱敏',
  HIDDEN: '隐藏',
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

// ── ItemGrantTypeBadge ───────────────────────────────────────────────────────

function ItemGrantTypeBadge({ item }: { item: EndUserBundleDataPermissionItem }) {
  if (item.grantType === 'PRESET') {
    return (
      <span className="rounded bg-[rgba(79,70,229,0.08)] px-1.5 py-px font-mono text-[11px] font-medium text-primary">
        {PRESET_LABEL[item.preset ?? ''] ?? item.preset ?? 'PRESET'}
      </span>
    )
  }
  const perm = item.customPermission
  return (
    <span className="rounded border border-border bg-background px-1.5 py-px text-[11px] text-muted-foreground">
      {perm?.displayName ?? `自定义: ${item.customPermissionId?.slice(0, 8) ?? '—'}`}
    </span>
  )
}

// ── VersionHistoryPanel ───────────────────────────────────────────────────────

interface VersionHistoryPanelProps {
  snapshots: EndUserPermissionBundleSnapshot[]
  currentVersion: number
  onRestore: (version: number) => Promise<void>
}

function VersionHistoryPanel({ snapshots, currentVersion, onRestore }: VersionHistoryPanelProps) {
  const [restoring, setRestoring] = React.useState<number | null>(null)

  const handleRestore = async (v: EndUserPermissionBundleSnapshot) => {
    setRestoring(v.version)
    try {
      await onRestore(v.version)
      toast.success(`已还原到版本 v${v.version}`)
    } catch {
      toast.error('还原失败，请重试')
    } finally {
      setRestoring(null)
    }
  }

  if (snapshots.length === 0) {
    return (
      <div className="rounded-md border border-border">
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <Clock className="mr-1.5 size-3.5 text-muted-foreground" strokeWidth={1.5} />
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            版本历史
          </span>
        </div>
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Clock className="mb-3 size-8 text-muted-foreground/20" strokeWidth={1} />
          <p className="text-sm text-muted-foreground">暂无历史版本</p>
          <p className="mt-1 text-xs text-muted-foreground/60">修改权限点后将自动生成版本快照</p>
        </div>
      </div>
    )
  }

  return (
    <div className="rounded-md border border-border">
      <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
        <Clock className="mr-1.5 size-3.5 text-muted-foreground" strokeWidth={1.5} />
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          版本历史
        </span>
        <span className="ml-2 rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
          {snapshots.length} / 5
        </span>
      </div>

      <div className="divide-y divide-border">
        {snapshots.map((v) => {
          const isCurrent = v.version === currentVersion
          const itemCount = v.items?.length ?? v.permissions?.length ?? 0
          return (
            <div
              key={v.version}
              className={cn(
                'group flex items-start gap-4 px-4 py-3',
                isCurrent ? 'bg-[rgba(79,70,229,0.03)]' : 'hover:bg-foreground/[0.015]',
              )}
            >
              <div className="flex w-10 shrink-0 flex-col items-center gap-1 pt-0.5">
                <span
                  className={cn(
                    'rounded px-1.5 py-px font-mono text-[11px] font-medium',
                    isCurrent
                      ? 'bg-[rgba(79,70,229,0.12)] text-primary'
                      : 'bg-muted text-muted-foreground',
                  )}
                >
                  v{v.version}
                </span>
                {isCurrent && (
                  <span className="rounded bg-[rgba(79,70,229,0.08)] px-1 py-px text-[10px] font-medium text-primary">
                    当前
                  </span>
                )}
              </div>

              <div className="min-w-0 flex-1">
                <p className="text-sm text-foreground">
                  共 {itemCount} 条数据权限配置
                  {v.restoredFrom != null && (
                    <span className="ml-2 text-xs text-muted-foreground">
                      （由 v{v.restoredFrom} 还原）
                    </span>
                  )}
                </p>
                <div className="mt-1 flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{formatTs(v.createdAt)}</span>
                  {v.createdBy && (
                    <>
                      <span className="text-muted-foreground/40">·</span>
                      <span>{v.createdBy}</span>
                    </>
                  )}
                </div>
              </div>

              {!isCurrent && (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 shrink-0 gap-1.5 px-2 text-xs text-muted-foreground opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100"
                      disabled={restoring === v.version}
                    >
                      {restoring === v.version ? (
                        <Loader2 className="size-3.5 animate-spin" />
                      ) : (
                        <RotateCcw className="size-3.5" strokeWidth={1.5} />
                      )}
                      还原
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>还原到 v{v.version}</AlertDialogTitle>
                      <AlertDialogDescription>
                        当前版本（v{currentVersion}）将保留为历史版本，并创建一个与 v{v.version}{' '}
                        内容相同的新版本作为当前版本。此操作不可撤销。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction onClick={() => void handleRestore(v)}>
                        确认还原
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ── InlineEditField ──────────────────────────────────────────────────────────

interface InlineEditFieldProps {
  value: string
  placeholder?: string
  onSave: (value: string) => Promise<void>
  className?: string
  inputClassName?: string
  multiline?: boolean
}

function InlineEditField({
  value,
  placeholder,
  onSave,
  className,
  inputClassName,
  multiline = false,
}: InlineEditFieldProps) {
  const [editing, setEditing] = React.useState(false)
  const [draft, setDraft] = React.useState(value)
  const [saving, setSaving] = React.useState(false)
  const inputRef = React.useRef<HTMLInputElement & HTMLTextAreaElement>(null)

  React.useEffect(() => {
    setDraft(value)
  }, [value])

  const startEdit = () => {
    setDraft(value)
    setEditing(true)
    requestAnimationFrame(() => inputRef.current?.focus())
  }

  const cancel = () => {
    setDraft(value)
    setEditing(false)
  }

  const save = async () => {
    const trimmed = draft.trim()
    if (trimmed === value) {
      setEditing(false)
      return
    }
    setSaving(true)
    await onSave(trimmed)
    setSaving(false)
    setEditing(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !multiline) {
      e.preventDefault()
      void save()
    }
    if (e.key === 'Escape') cancel()
  }

  if (editing) {
    const sharedProps = {
      ref: inputRef as React.Ref<HTMLInputElement & HTMLTextAreaElement>,
      value: draft,
      onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
        setDraft(e.target.value),
      onKeyDown: handleKeyDown,
      disabled: saving,
      className: cn(
        'w-full rounded border border-border bg-background px-2 py-1 text-sm outline-none',
        'focus:border-primary focus:ring-1 focus:ring-primary/20',
        'disabled:opacity-60',
        inputClassName,
      ),
    }

    return (
      <div className={cn('flex items-start gap-1.5', className)}>
        {multiline ? (
          <textarea {...sharedProps} rows={2} />
        ) : (
          <input {...sharedProps} type="text" />
        )}
        <div className="mt-0.5 flex shrink-0 gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="hover:bg-primary/8 size-6 text-primary hover:text-primary"
            onClick={() => void save()}
            disabled={saving}
            aria-label="保存"
          >
            {saving ? <Loader2 className="size-3.5 animate-spin" /> : <Check className="size-3.5" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-6 text-muted-foreground hover:text-foreground"
            onClick={cancel}
            disabled={saving}
            aria-label="取消"
          >
            <X className="size-3.5" />
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div
      className={cn('group/field flex cursor-pointer items-start gap-1.5', className)}
      onClick={startEdit}
    >
      <span className="flex-1">{value || <span className="text-muted-foreground/50">{placeholder}</span>}</span>
      <Pencil
        className="mt-0.5 size-3.5 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover/field:opacity-100"
        strokeWidth={1.5}
      />
    </div>
  )
}

// ── AddItemDialog ────────────────────────────────────────────────────────────
// 两路径：preset 绑定 / custom 绑定

interface ModelOption {
  id: string
  name: string
  title?: string
  databaseName?: string
}

interface AddItemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bundleName: string
  /** 已配置了 item 的 modelId 集合（用于展示替换提示） */
  configuredModelIds: Set<string>
  orgName: string
  projectSlug: string
  onBindPreset: (modelId: string, preset: EndUserPermissionPreset) => Promise<MutationResultMin>
  onBindCustom: (modelId: string, customPermissionId: string) => Promise<MutationResultMin>
  allPermissions: EndUserPermission[]
}

interface MutationResultMin {
  success: boolean
  errorMessage?: string
}

interface DatabaseCatalogData {
  modelDatabaseCatalog?: {
    data?: { databases: Array<{ name: string }> } | null
    error?: unknown
  } | null
}

interface ModelsQueryData {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  models?: { items?: Array<ModelOption | null> }
}

function AddItemDialog({
  open,
  onOpenChange,
  bundleName,
  configuredModelIds,
  orgName,
  projectSlug,
  onBindPreset,
  onBindCustom,
  allPermissions,
}: AddItemDialogProps) {
  const projectClient = useProjectScopedClient(projectSlug, orgName)
  const [bindMode, setBindMode] = React.useState<'preset' | 'custom'>('preset')
  const [selectedDatabase, setSelectedDatabase] = React.useState<string | null>(null)
  const [selectedModelId, setSelectedModelId] = React.useState<string | null>(null)
  const [selectedPreset, setSelectedPreset] = React.useState<EndUserPermissionPreset | null>(null)
  const [selectedCustomPermId, setSelectedCustomPermId] = React.useState<string | null>(null)
  const [submitting, setSubmitting] = React.useState(false)

  // Step 1：查数据库列表
  const { data: catalogData, loading: catalogLoading } = useQuery<DatabaseCatalogData>(
    DATABASE_CATALOG,
    {
      client: projectClient,
      skip: !open,
    },
  )

  const databaseOptions = React.useMemo(() => {
    return catalogData?.modelDatabaseCatalog?.data?.databases ?? []
  }, [catalogData])

  // Step 2：选择数据库后查模型列表
  const { data: modelsData, loading: modelsLoading } = useQuery<ModelsQueryData>(
    GET_MODELS_BY_DATABASE,
    {
      client: projectClient,
      variables: { input: { databaseName: selectedDatabase, pageSize: 200 } },
      skip: !open || !selectedDatabase,
    },
  )

  const modelOptions = React.useMemo<ModelOption[]>(() => {
    const items = modelsData?.models?.items ?? []
    return items
      .filter((n): n is ModelOption => Boolean(n?.id && n?.name))
      .sort((a, b) => (a.title ?? a.name).localeCompare(b.title ?? b.name))
  }, [modelsData])

  // Step 3：选择模型后查虚拟预设列表
  const { data: presetsData, loading: presetsLoading } = useQuery<{ virtualPresetsByModel: EndUserPermissionPreset[] }>(
    GET_VIRTUAL_PRESETS_BY_MODEL,
    {
      client: projectClient,
      variables: { modelId: selectedModelId ?? '' },
      skip: !selectedModelId || bindMode !== 'preset',
    },
  )

  const availablePresets: EndUserPermissionPreset[] = presetsData?.virtualPresetsByModel ?? []

  // custom permissions for selected model（not yet in bundle）
  const customPermsForModel = React.useMemo(() => {
    if (!selectedModelId) return []
    return allPermissions.filter((p) => p.modelId === selectedModelId)
  }, [allPermissions, selectedModelId])

  const canSubmit =
    selectedModelId &&
    (bindMode === 'preset' ? Boolean(selectedPreset) : Boolean(selectedCustomPermId))

  const handleClose = () => {
    setSelectedDatabase(null)
    setSelectedModelId(null)
    setSelectedPreset(null)
    setSelectedCustomPermId(null)
    setBindMode('preset')
    onOpenChange(false)
  }

  const handleSubmit = async () => {
    if (!canSubmit || !selectedModelId) return
    setSubmitting(true)
    let result: MutationResultMin
    if (bindMode === 'preset' && selectedPreset) {
      result = await onBindPreset(selectedModelId, selectedPreset)
    } else if (bindMode === 'custom' && selectedCustomPermId) {
      result = await onBindCustom(selectedModelId, selectedCustomPermId)
    } else {
      setSubmitting(false)
      return
    }
    setSubmitting(false)
    if (result.success) {
      handleClose()
    } else {
      toast.error(result.errorMessage ?? '操作失败，请重试')
    }
  }

  return (
    <Sheet open={open} onOpenChange={handleClose}>
      <SheetContent
        side="right"
        className="flex w-full max-w-[860px] flex-col gap-0 p-0 sm:max-w-[860px]"
      >
        {/* Header */}
        <SheetHeader className="shrink-0 border-b border-border px-5 py-4">
          <SheetTitle className="text-sm font-semibold text-foreground">
            为「{bundleName}」添加数据权限配置
          </SheetTitle>
        </SheetHeader>

        {/* 三列主体 */}
        <div className="flex min-h-0 flex-1 overflow-hidden">

          {/* 列 1：数据库 */}
          <div className="flex w-44 shrink-0 flex-col border-r border-border">
            <p className="shrink-0 px-3 py-2.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
              数据库
            </p>
            <ScrollArea className="flex-1">
              <div className="pb-2">
                {catalogLoading ? (
                  <div className="space-y-1 px-3 py-2">
                    {[1, 2, 3].map((i) => <Skeleton key={i} className="h-8 w-full" />)}
                  </div>
                ) : databaseOptions.length === 0 ? (
                  <p className="px-3 py-5 text-center text-xs text-muted-foreground/50">暂无数据库</p>
                ) : (
                  databaseOptions.map((db) => {
                    const isSelected = selectedDatabase === db.name
                    return (
                      <button
                        key={db.name}
                        type="button"
                        onClick={() => {
                          setSelectedDatabase(db.name)
                          setSelectedModelId(null)
                          setSelectedPreset(null)
                          setSelectedCustomPermId(null)
                        }}
                        className={cn(
                          'flex w-full items-center px-3 py-2 text-left text-xs transition-colors',
                          isSelected
                            ? 'bg-[rgba(79,70,229,0.08)] font-medium text-primary'
                            : 'text-foreground hover:bg-foreground/[0.04]',
                        )}
                      >
                        <span className="truncate font-mono">{db.name}</span>
                      </button>
                    )
                  })
                )}
              </div>
            </ScrollArea>
          </div>

          {/* 列 2：数据表 */}
          <div className="flex w-52 shrink-0 flex-col border-r border-border">
            <p className="shrink-0 px-3 py-2.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
              数据表
            </p>
            <ScrollArea className="flex-1">
              <div className="pb-2">
                {!selectedDatabase ? (
                  <p className="px-3 py-5 text-center text-xs text-muted-foreground/50">先选择数据库</p>
                ) : modelsLoading ? (
                  <div className="space-y-1 px-3 py-2">
                    {[1, 2, 3].map((i) => <Skeleton key={i} className="h-9 w-full" />)}
                  </div>
                ) : modelOptions.length === 0 ? (
                  <p className="px-3 py-5 text-center text-xs text-muted-foreground/50">该数据库下暂无数据表</p>
                ) : (
                  modelOptions.map((model) => {
                    const isConfigured = configuredModelIds.has(model.id)
                    const isSelected = selectedModelId === model.id
                    return (
                      <button
                        key={model.id}
                        type="button"
                        disabled={isConfigured}
                        onClick={() => {
                          if (isConfigured) return
                          setSelectedModelId(model.id)
                          setSelectedPreset(null)
                          setSelectedCustomPermId(null)
                        }}
                        className={cn(
                          'flex w-full items-start gap-2 px-3 py-2 text-left transition-colors',
                          isConfigured
                            ? 'cursor-not-allowed opacity-40'
                            : isSelected
                              ? 'bg-[rgba(79,70,229,0.08)] text-primary'
                              : 'text-foreground hover:bg-foreground/[0.04]',
                        )}
                      >
                        <div className="min-w-0 flex-1">
                          <p className={cn('truncate text-xs', isSelected && 'font-medium')}>
                            {model.title ?? model.name}
                          </p>
                          <p className="truncate font-mono text-[11px] text-muted-foreground">
                            {model.name}
                          </p>
                        </div>
                        {isConfigured && (
                          <span className="mt-0.5 shrink-0 rounded bg-muted px-1 py-px font-mono text-[10px] text-muted-foreground">
                            已配置
                          </span>
                        )}
                      </button>
                    )
                  })
                )}
              </div>
            </ScrollArea>
          </div>

          {/* 列 3：授权方式选择 */}
          <div className="flex flex-1 flex-col overflow-hidden">
            {/* 授权模式切换 */}
            <div className="shrink-0 border-b border-border px-4 py-2.5">
              <Tabs
                value={bindMode}
                onValueChange={(v) => {
                  setBindMode(v as 'preset' | 'custom')
                  setSelectedPreset(null)
                  setSelectedCustomPermId(null)
                }}
              >
                <TabsList className="h-7">
                  <TabsTrigger value="preset" className="text-xs">预设模板</TabsTrigger>
                  <TabsTrigger value="custom" className="text-xs">自定义策略</TabsTrigger>
                </TabsList>
              </Tabs>
            </div>

            {!selectedModelId ? (
              <div className="flex flex-1 flex-col items-center justify-center gap-2 px-8 text-center">
                <Shield className="size-8 text-muted-foreground/20" strokeWidth={1} />
                <p className="text-sm text-muted-foreground">从左侧选择数据表</p>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="space-y-2 p-4">
                  {bindMode === 'preset' ? (
                    presetsLoading ? (
                      <div className="space-y-2 pt-2">
                        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-14 w-full" />)}
                      </div>
                    ) : availablePresets.length === 0 ? (
                      <p className="py-6 text-center text-xs text-muted-foreground">该模型暂无可用预设</p>
                    ) : (
                      availablePresets.map((preset) => (
                        <label
                          key={preset}
                          className={cn(
                            'flex cursor-pointer items-center gap-3 rounded-md border px-4 py-3 transition-colors',
                            selectedPreset === preset
                              ? 'border-primary/30 bg-[rgba(79,70,229,0.06)]'
                              : 'border-transparent hover:border-border hover:bg-foreground/[0.02]',
                          )}
                        >
                          <input
                            type="radio"
                            name="preset-select"
                            value={preset}
                            checked={selectedPreset === preset}
                            onChange={() => setSelectedPreset(preset)}
                            className="size-3.5 accent-primary"
                          />
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-medium text-foreground">
                              {PRESET_LABEL[preset] ?? preset}
                            </p>
                            <p className="mt-0.5 text-xs text-muted-foreground">{preset}</p>
                          </div>
                        </label>
                      ))
                    )
                  ) : (
                    customPermsForModel.length === 0 ? (
                      <p className="py-6 text-center text-xs text-muted-foreground">该模型暂无自定义策略</p>
                    ) : (
                      customPermsForModel.map((perm) => (
                        <label
                          key={perm.id}
                          className={cn(
                            'flex cursor-pointer items-center gap-3 rounded-md border px-4 py-3 transition-colors',
                            selectedCustomPermId === perm.id
                              ? 'border-primary/30 bg-[rgba(79,70,229,0.06)]'
                              : 'border-transparent hover:border-border hover:bg-foreground/[0.02]',
                          )}
                        >
                          <input
                            type="radio"
                            name="custom-select"
                            value={perm.id}
                            checked={selectedCustomPermId === perm.id}
                            onChange={() => setSelectedCustomPermId(perm.id)}
                            className="size-3.5 accent-primary"
                          />
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-medium text-foreground">
                              {perm.displayName ?? '—'}
                            </p>
                            <div className="mt-1 flex items-center gap-1.5">
                              <Badge variant="outline" className="text-[11px]">
                                {ACTION_LABEL[perm.action] ?? perm.action}
                              </Badge>
                            </div>
                          </div>
                        </label>
                      ))
                    )
                  )}
                </div>
              </ScrollArea>
            )}

            {/* Footer 固定在第三列底部 */}
            <div className="flex shrink-0 items-center justify-end gap-2 border-t border-border px-4 py-3">
              <Button variant="outline" size="sm" onClick={handleClose} disabled={submitting}>
                取消
              </Button>
              <Button
                size="sm"
                onClick={() => void handleSubmit()}
                disabled={!canSubmit || submitting}
              >
                {submitting && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
                确认添加
              </Button>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}

// ── BundleDetailPage ─────────────────────────────────────────────────────────

export default function BundleDetailPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`

  const {
    bundle,
    allPermissions,
    loading,
    error,
    removeItemByModelId,
    bindPresetItem,
    bindCustomItem,
    updateBundle,
    restoreBundle,
  } = useBundleManage({ orgName, projectSlug, bundleId })

  const [removingModelId, setRemovingModelId] = React.useState<string | null>(null)
  const [addDialogOpen, setAddDialogOpen] = React.useState(false)
  const [activeTab, setActiveTab] = React.useState<'strategies' | 'versions'>('strategies')
  const [filterDatabase, setFilterDatabase] = React.useState<string>('__all__')

  // dataPermissionItems（优先）
  const items: EndUserBundleDataPermissionItem[] = React.useMemo(
    () => bundle?.dataPermissionItems ?? [],
    [bundle?.dataPermissionItems],
  )

  const configuredModelIds = React.useMemo(
    () => new Set(items.map((i) => i.modelId)),
    [items],
  )

  const databaseOptions = React.useMemo(
    () => [...new Set(items.map((i) => i.databaseName).filter((d): d is string => Boolean(d)))],
    [items],
  )

  const filteredItems = React.useMemo(
    () =>
      filterDatabase === '__all__'
        ? items
        : items.filter((i) => i.databaseName === filterDatabase),
    [items, filterDatabase],
  )

  const handleRemoveItem = async (item: EndUserBundleDataPermissionItem) => {
    setRemovingModelId(item.modelId)
    try {
      const result = await removeItemByModelId(item.modelId)
      if (result.success) {
        toast.success('已移除数据权限配置')
      } else {
        toast.error(result.errorMessage ?? '移除失败，请重试')
      }
    } catch {
      toast.error('移除时发生错误')
    } finally {
      setRemovingModelId(null)
    }
  }

  const handleBindPreset = async (
    modelId: string,
    preset: EndUserPermissionPreset,
  ) => {
    const result = await bindPresetItem(modelId, preset)
    if (result.success) {
      toast.success('已添加数据权限配置')
    }
    return result
  }

  const handleBindCustom = async (modelId: string, customPermissionId: string) => {
    const result = await bindCustomItem(modelId, customPermissionId)
    if (result.success) {
      toast.success('已添加数据权限配置')
    }
    return result
  }

  const handleSaveName = async (name: string) => {
    if (!name) return
    const result = await updateBundle(name, bundle?.description)
    if (result.success) toast.success('已更新名称')
    else toast.error(result.errorMessage ?? '更新失败')
  }

  const handleSaveDescription = async (description: string) => {
    if (!bundle?.name) return
    const result = await updateBundle(bundle.name, description || undefined)
    if (result.success) toast.success('已更新描述')
    else toast.error(result.errorMessage ?? '更新失败')
  }

  return (
    <PageLayout maxWidth="7xl">
      {/* 返回导航 */}
      <div className="mb-6">
        <Link
          href={backHref}
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          返回权限包列表
        </Link>
      </div>

      {/* 标题 */}
      <div className="mb-8">
        {loading ? (
          <>
            <Skeleton className="h-7 w-48" />
            <Skeleton className="mt-2 h-4 w-64" />
          </>
        ) : (
          <>
            <InlineEditField
              value={bundle?.name ?? ''}
              placeholder="权限包名称"
              onSave={handleSaveName}
              className="text-xl font-semibold tracking-tight text-foreground"
              inputClassName="text-xl font-semibold"
            />
            <InlineEditField
              value={bundle?.description ?? ''}
              placeholder="添加描述…"
              onSave={handleSaveDescription}
              className="mt-1.5 text-sm text-muted-foreground"
              multiline
            />
          </>
        )}
      </div>

      {error && (
        <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          加载失败：{error.message}
        </div>
      )}

      {/* Tab 导航 */}
      <div className="mb-5 flex items-end gap-0 border-b border-border">
        {(
          [
            { key: 'strategies', label: '数据权限配置' },
            { key: 'versions', label: '版本历史' },
          ] as const
        ).map((t) => (
          <button
            key={t.key}
            type="button"
            onClick={() => setActiveTab(t.key)}
            className={cn(
              'relative -mb-px px-4 pb-2.5 pt-1 text-sm transition-colors',
              activeTab === t.key
                ? 'border-b-2 border-primary font-medium text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {t.label}
            {t.key === 'strategies' && !loading && (
              <span className="ml-1.5 rounded bg-muted px-1.5 py-px text-[11px] text-muted-foreground">
                {items.length}
              </span>
            )}
            {t.key === 'versions' && !loading && (
              <span className="ml-1.5 rounded bg-muted px-1.5 py-px text-[11px] text-muted-foreground">
                {bundle?.snapshots?.length ?? 0} / 5
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Tab: 数据权限配置 */}
      {activeTab === 'strategies' && (
        <div className="rounded-md border border-border">
          <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
            <div className="ml-auto flex items-center gap-2">
              {databaseOptions.length > 0 && (
                <Select value={filterDatabase} onValueChange={setFilterDatabase}>
                  <SelectTrigger className="h-7 w-44 text-xs">
                    <SelectValue placeholder="全部数据库" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__all__">全部数据库</SelectItem>
                    {databaseOptions.map((db) => (
                      <SelectItem key={db} value={db}>
                        {db}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
              <Button
                variant="outline"
                size="sm"
                className="h-7 gap-1.5 text-xs"
                onClick={() => setAddDialogOpen(true)}
                disabled={loading}
              >
                <Plus className="size-3.5" strokeWidth={1.5} />
                添加数据权限
              </Button>
            </div>
          </div>

          {loading ? (
            <div className="space-y-px">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="px-4 py-3">
                  <Skeleton className="mb-2 h-4 w-24" />
                  <Skeleton className="h-5 w-20" />
                </div>
              ))}
            </div>
          ) : items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Shield className="mb-3 size-9 text-muted-foreground/20" strokeWidth={1} />
              <p className="text-sm font-semibold text-foreground">尚未配置任何数据权限</p>
              <p className="mt-1 max-w-xs text-xs text-muted-foreground">
                为每个需要访问控制的数据表绑定预设模板或自定义策略
              </p>
              <Button
                variant="outline"
                size="sm"
                className="mt-5 gap-1.5"
                onClick={() => setAddDialogOpen(true)}
              >
                <Plus className="size-4" strokeWidth={1.5} />
                添加数据权限
              </Button>
            </div>
          ) : (
            <div>
              {/* 列标题 */}
              <div className="flex items-center gap-4 border-b border-border px-4 py-2">
                <span className="w-36 shrink-0 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  数据库
                </span>
                <span className="w-40 shrink-0 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  模型标识
                </span>
                <span className="min-w-0 flex-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  模型名称
                </span>
                <span className="w-36 shrink-0 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  创建时间
                </span>
                <span className="w-44 shrink-0 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  策略
                </span>
                <span className="size-7 shrink-0" />
              </div>

              <div className="divide-y divide-border">
              {filteredItems.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-10 text-center">
                  <p className="text-sm text-muted-foreground">该数据库下暂无配置项</p>
                </div>
              ) : filteredItems.map((item) => {
                return (
                  <div
                    key={item.id}
                    className="group flex items-center gap-4 px-4 py-3 hover:bg-foreground/[0.015]"
                  >
                    {/* Col 1: 数据库 */}
                    <div className="w-36 min-w-0 shrink-0">
                      <p className="truncate font-mono text-xs text-foreground">
                        {item.databaseName ?? <span className="text-muted-foreground/40">—</span>}
                      </p>
                    </div>

                    {/* Col 2: 模型标识 */}
                    <div className="w-40 min-w-0 shrink-0">
                      <p className="truncate font-mono text-xs text-foreground">
                        {item.modelName ?? <span className="text-muted-foreground/40">—</span>}
                      </p>
                    </div>

                    {/* Col 3: 模型名称 */}
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm text-foreground">
                        {item.modelTitle ?? <span className="text-muted-foreground/40">—</span>}
                      </p>
                    </div>

                    {/* Col 4: 创建时间 */}
                    <div className="w-36 min-w-0 shrink-0">
                      <p className="truncate text-xs text-muted-foreground">
                        {formatTs(item.createdAt)}
                      </p>
                    </div>

                    {/* Col 5: 策略 */}
                    <div className="flex w-44 shrink-0 flex-wrap items-center gap-1.5">
                      <ItemGrantTypeBadge item={item} />
                      {item.grantType === 'CUSTOM' && item.customPermission && (
                        <>
                          {item.customPermission.action && (
                            <span className="rounded border border-border px-1.5 py-px text-[11px] text-muted-foreground">
                              {ACTION_LABEL[item.customPermission.action] ?? item.customPermission.action}
                            </span>
                          )}
                          {item.customPermission.rowScope && (
                            <span className="rounded border border-border px-1.5 py-px text-[11px] text-muted-foreground">
                              {ROW_SCOPE_LABEL[item.customPermission.rowScope] ?? item.customPermission.rowScope}
                            </span>
                          )}
                          {item.customPermission.columnPolicy && (
                            <span className="rounded border border-border px-1.5 py-px text-[11px] text-muted-foreground">
                              列: {COLUMN_MODE_LABEL[item.customPermission.columnPolicy.defaultMode] ?? item.customPermission.columnPolicy.defaultMode}
                              {item.customPermission.columnPolicy.rules.length > 0 && ` +${item.customPermission.columnPolicy.rules.length} 规则`}
                            </span>
                          )}
                        </>
                      )}
                    </div>

                    {/* Col 5: 删除按钮 */}
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label="移除配置"
                          className="mt-0.5 size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                          disabled={removingModelId === item.modelId}
                        >
                          {removingModelId === item.modelId ? (
                            <Loader2 className="size-3.5 animate-spin" />
                          ) : (
                            <Trash2 className="size-3.5" />
                          )}
                        </Button>
                      </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>移除数据权限配置</AlertDialogTitle>
                        <AlertDialogDescription>
                          确定要从该权限包中移除此数据权限配置吗？此操作不可撤销。
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                          className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          onClick={() => void handleRemoveItem(item)}
                        >
                          确认移除
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              )
              })}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab: 版本历史 */}
      {activeTab === 'versions' && (
        <VersionHistoryPanel
          snapshots={(bundle?.snapshots ?? []) as EndUserPermissionBundleSnapshot[]}
          currentVersion={bundle?.currentVersion ?? 0}
          onRestore={async (version) => {
            const result = await restoreBundle(version)
            if (!result.success) throw new Error(result.errorMessage ?? '还原失败')
          }}
        />
      )}

      {/* 添加弹窗 */}
      <AddItemDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        bundleName={bundle?.name ?? ''}
        configuredModelIds={configuredModelIds}
        orgName={orgName}
        projectSlug={projectSlug}
        onBindPreset={handleBindPreset}
        onBindCustom={handleBindCustom}
        allPermissions={allPermissions}
      />
    </PageLayout>
  )
}
