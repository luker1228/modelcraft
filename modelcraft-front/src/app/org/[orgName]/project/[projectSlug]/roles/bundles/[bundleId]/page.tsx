'use client'

import * as React from 'react'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import {
  ArrowLeft,
  Shield,
  Loader2,
  Trash2,
  Plus,
  Pencil,
  Check,
  X,
  Search,
  ChevronRight,
  ExternalLink,
  Clock,
  RotateCcw,
} from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Skeleton } from '@web/components/ui/skeleton'
import { Input } from '@web/components/ui/input'
import { ScrollArea } from '@web/components/ui/scroll-area'
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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { PageLayout } from '@web/components/features/layout'

import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type { EndUserPermission } from '@/types'
import { cn } from '@/shared/utils'

const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '更新',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部行',
  SELF: '本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '部门及子部门',
}

// ── Snapshot types ────────────────────────────────────────────────────────────

interface BundleSnapshotPermissionEntry {
  sortOrder: number
  permissionId: string
}

interface BundleSnapshot {
  version: number
  createdAt: string
  createdBy?: string | null
  restoredFrom?: number | null
  permissions: BundleSnapshotPermissionEntry[]
}

function formatTs(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

// ── VersionHistoryPanel ───────────────────────────────────────────────────────

interface VersionHistoryPanelProps {
  snapshots: BundleSnapshot[]
  currentVersion: number
  onRestore: (version: number) => Promise<void>
}

function VersionHistoryPanel({ snapshots, currentVersion, onRestore }: VersionHistoryPanelProps) {
  const [restoring, setRestoring] = React.useState<number | null>(null)

  const handleRestore = async (v: BundleSnapshot) => {
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
      <div className="mb-6 rounded-md border border-border">
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
    <div className="mb-6 rounded-md border border-border">
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
          return (
            <div
              key={v.version}
              className={cn(
                'group flex items-start gap-4 px-4 py-3',
                isCurrent ? 'bg-[rgba(79,70,229,0.03)]' : 'hover:bg-foreground/[0.015]',
              )}
            >
              {/* Version badge */}
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

              {/* Content */}
              <div className="min-w-0 flex-1">
                <p className="text-sm text-foreground">
                  共 {v.permissions.length} 个权限点
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

              {/* Restore action — non-current only */}
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
                        当前版本（v{currentVersion}）将保留为历史版本，并创建一个与 v{v.version} 内容相同的新版本作为当前版本。此操作不可撤销。
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

// ── Inline editable field ────────────────────────────────────────────────────

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

// ── AddStrategyDialog ────────────────────────────────────────────────────────
// Layout: left = searchable model list, right = permission picker

interface AddStrategyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bundleName: string
  allPermissions: EndUserPermission[]
  /** IDs of permissions already in this bundle */
  bundledPermissionIds: Set<string>
  /** Models already having a strategy in this bundle */
  bundledModelIds: Set<string>
  onAdd: (permissionId: string) => Promise<void>
  orgName: string
  projectSlug: string
}

// Group permissions by modelId
function groupByModel(permissions: EndUserPermission[]): Map<string, EndUserPermission[]> {
  const map = new Map<string, EndUserPermission[]>()
  for (const p of permissions) {
    const list = map.get(p.modelId) ?? []
    list.push(p)
    map.set(p.modelId, list)
  }
  return map
}

function AddStrategyDialog({
  open,
  onOpenChange,
  bundleName,
  allPermissions,
  bundledPermissionIds,
  bundledModelIds,
  onAdd,
  orgName,
  projectSlug,
}: AddStrategyDialogProps) {
  const [modelSearch, setModelSearch] = React.useState('')
  const [selectedModelId, setSelectedModelId] = React.useState<string | null>(null)
  const [selectedPermId, setSelectedPermId] = React.useState<string | null>(null)
  const [adding, setAdding] = React.useState(false)
  const router = useRouter()

  // Available permissions not yet in bundle
  const availablePerms = React.useMemo(
    () => allPermissions.filter((p) => !bundledPermissionIds.has(p.id)),
    [allPermissions, bundledPermissionIds],
  )

  // Group all permissions by model
  const allGrouped = React.useMemo(() => groupByModel(allPermissions), [allPermissions])
  // Group available permissions by model
  const availableGrouped = React.useMemo(() => groupByModel(availablePerms), [availablePerms])

  // Models filtered by search
  const modelList = React.useMemo(() => {
    return Array.from(allGrouped.entries())
      .map(([modelId, perms]) => ({
        modelId,
        displayName: perms[0]?.modelDisplayName ?? modelId,
      }))
      .filter(({ modelId, displayName }) => {
        if (!modelSearch) return true
        const q = modelSearch.toLowerCase()
        return displayName.toLowerCase().includes(q) || modelId.toLowerCase().includes(q)
      })
  }, [allGrouped, modelSearch])

  // Permissions available for selected model
  const permissionsForModel = React.useMemo(() => {
    if (!selectedModelId) return []
    return availableGrouped.get(selectedModelId) ?? []
  }, [availableGrouped, selectedModelId])

  const selectedModelConfigured = selectedModelId ? bundledModelIds.has(selectedModelId) : false

  const handleModelSelect = (modelId: string) => {
    if (bundledModelIds.has(modelId)) return
    setSelectedModelId(modelId)
    setSelectedPermId(null)
  }

  const handleConfirm = async () => {
    if (!selectedPermId) return
    setAdding(true)
    await onAdd(selectedPermId)
    setAdding(false)
    onOpenChange(false)
  }

  const handleClose = () => {
    setModelSearch('')
    setSelectedModelId(null)
    setSelectedPermId(null)
    onOpenChange(false)
  }

  const selectedModelLabel = selectedModelId
    ? (allGrouped.get(selectedModelId)?.[0]?.modelDisplayName ?? selectedModelId)
    : null

  const createHref = selectedModelId
    ? `/org/${orgName}/project/${projectSlug}/roles?tab=permissions&createFor=${selectedModelId}`
    : `/org/${orgName}/project/${projectSlug}/roles?tab=permissions`

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent
        className="flex max-h-[80vh] flex-col gap-0 overflow-hidden p-0 sm:max-w-[860px]"
        aria-describedby={undefined}
      >
        <DialogHeader className="shrink-0 border-b border-border px-5 py-4">
          <DialogTitle className="text-sm font-semibold text-foreground">
            为「{bundleName}」添加资源策略
          </DialogTitle>
        </DialogHeader>

        {/* Two-column body: model list | permission picker */}
        <div className="flex min-h-0 flex-1">
          {/* Col 1: Searchable model list */}
          <div className="flex w-52 shrink-0 flex-col border-r border-border">
            <div className="border-b border-border px-3 py-2">
              <div className="relative">
                <Search
                  className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground"
                  strokeWidth={1.5}
                />
                <Input
                  placeholder="搜索数据表…"
                  className="h-7 border-[#D8DDE5] bg-[#EBEEF2] pl-8 text-xs placeholder:text-muted-foreground focus-visible:ring-1"
                  value={modelSearch}
                  onChange={(e) => setModelSearch(e.target.value)}
                />
              </div>
            </div>
            <ScrollArea className="flex-1">
              <div className="py-1.5">
                {availablePerms.length === 0 ? (
                  <p className="px-3 py-6 text-center text-xs text-muted-foreground/60">
                    暂无可添加权限点
                  </p>
                ) : modelList.length === 0 ? (
                  <p className="px-3 py-4 text-center text-xs text-muted-foreground">
                    {modelSearch ? '无匹配数据表' : '暂无数据表'}
                  </p>
                ) : (
                  modelList.map(({ modelId, displayName }) => {
                    const isConfigured = bundledModelIds.has(modelId)
                    const isSelected = selectedModelId === modelId
                    return (
                      <button
                        key={modelId}
                        type="button"
                        disabled={isConfigured}
                        onClick={() => handleModelSelect(modelId)}
                        className={cn(
                          'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors',
                          isConfigured
                            ? 'cursor-not-allowed opacity-40'
                            : isSelected
                              ? 'bg-[rgba(79,70,229,0.08)] font-medium text-primary'
                              : 'text-foreground hover:bg-foreground/[0.04]',
                        )}
                      >
                        <span className="flex-1 truncate">{displayName}</span>
                        {isConfigured ? (
                          <span className="shrink-0 rounded bg-muted px-1 py-px font-mono text-[10px] text-muted-foreground">
                            已配置
                          </span>
                        ) : isSelected ? (
                          <ChevronRight className="size-3.5 shrink-0 text-primary/50" strokeWidth={1.5} />
                        ) : null}
                      </button>
                    )
                  })
                )}
              </div>
            </ScrollArea>
          </div>

          {/* Col 2: Permission picker */}
          <div className="flex flex-1 flex-col">
            <div className="border-b border-border px-4 py-2">
              <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                {selectedModelLabel ? `${selectedModelLabel} · 权限点` : '权限点'}
              </span>
            </div>

            {!selectedModelId ? (
              <div className="flex flex-1 flex-col items-center justify-center gap-2 px-8 py-12 text-center">
                <Shield className="size-8 text-muted-foreground/20" strokeWidth={1} />
                <p className="text-sm text-muted-foreground">从左侧选择数据表</p>
                <p className="text-xs text-muted-foreground/60">每个数据表只能配置一条权限策略</p>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="p-3">
                  {permissionsForModel.length === 0 ? (
                    <div className="flex flex-col items-center gap-3 px-4 py-10 text-center">
                      <Shield className="size-7 text-muted-foreground/20" strokeWidth={1} />
                      <div className="space-y-1">
                        <p className="text-sm font-medium text-foreground">暂无可用权限点</p>
                        <p className="text-xs text-muted-foreground">
                          {selectedModelConfigured
                            ? '该数据表已配置权限策略'
                            : '该数据表还没有自定义权限点，可以创建一个'}
                        </p>
                      </div>
                      {!selectedModelConfigured && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="mt-1 h-7 gap-1.5 text-xs"
                          onClick={() => {
                            handleClose()
                            router.push(createHref)
                          }}
                        >
                          <Plus className="size-3.5" strokeWidth={1.5} />
                          创建自定义权限点
                          <ExternalLink className="size-3" strokeWidth={1.5} />
                        </Button>
                      )}
                    </div>
                  ) : (
                    <div className="space-y-1">
                      {permissionsForModel.map((perm) => {
                        const isSelected = selectedPermId === perm.id
                        return (
                          <label
                            key={perm.id}
                            className={cn(
                              'flex cursor-pointer items-start gap-3 rounded-md border px-3 py-2.5 transition-colors',
                              isSelected
                                ? 'border-primary/30 bg-[rgba(79,70,229,0.06)]'
                                : 'border-transparent hover:border-border hover:bg-foreground/[0.02]',
                            )}
                          >
                            <input
                              type="radio"
                              name="perm-select"
                              value={perm.id}
                              checked={isSelected}
                              onChange={() => setSelectedPermId(perm.id)}
                              className="mt-0.5 size-3.5 accent-primary"
                            />
                            <div className="min-w-0 flex-1 space-y-1.5">
                              <span className="text-sm font-medium text-foreground">
                                {perm.displayName || (perm.modelDisplayName ?? perm.modelId)}
                              </span>
                              <div className="flex items-center gap-1.5">
                                <span className="rounded bg-[rgba(79,70,229,0.08)] px-1.5 py-px font-mono text-[11px] font-medium text-primary">
                                  {ACTION_LABEL[perm.action] ?? perm.action}
                                </span>
                                <span className="rounded border border-border bg-background px-1.5 py-px text-[11px] text-muted-foreground">
                                  {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                                </span>
                              </div>
                              {perm.description && (
                                <p className="text-xs text-muted-foreground">{perm.description}</p>
                              )}
                            </div>
                          </label>
                        )
                      })}

                      <div className="mt-2 border-t border-border pt-3">
                        <button
                          type="button"
                          className="flex w-full items-center gap-2 rounded p-2 text-xs text-muted-foreground transition-colors hover:bg-foreground/[0.03] hover:text-foreground"
                          onClick={() => {
                            handleClose()
                            router.push(createHref)
                          }}
                        >
                          <Plus className="size-3.5" strokeWidth={1.5} />
                          没有满意的？创建自定义权限点
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </ScrollArea>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="flex shrink-0 items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button variant="outline" size="sm" onClick={handleClose} disabled={adding}>
            取消
          </Button>
          <Button
            size="sm"
            onClick={() => void handleConfirm()}
            disabled={!selectedPermId || adding}
          >
            {adding && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
            确认添加
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ── BundleDetailPage ─────────────────────────────────────────────────────────

export default function BundleDetailPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`

  const { bundle, allPermissions, loading, error, removePermission, addPermission, updateBundle, restoreBundle } =
    useBundleManage({
      orgName,
      projectSlug,
      bundleId,
    })

  const [removingId, setRemovingId] = React.useState<string | null>(null)
  const [addDialogOpen, setAddDialogOpen] = React.useState(false)

  const permissions = React.useMemo(
    () => bundle?.permissions ?? [],
    [bundle?.permissions],
  )

  const bundledPermissionIds = React.useMemo(
    () => new Set(permissions.map((p) => p.id)),
    [permissions],
  )

  const bundledModelIds = React.useMemo(
    () => new Set(permissions.map((p) => p.modelId)),
    [permissions],
  )

  // Group current bundle permissions by modelId for display
  const permissionsByModel = React.useMemo(() => {
    const map = new Map<string, EndUserPermission[]>()
    for (const p of permissions) {
      const list = map.get(p.modelId) ?? []
      list.push(p)
      map.set(p.modelId, list)
    }
    return map
  }, [permissions])

  const handleAdd = async (permissionId: string) => {
    const result = await addPermission(permissionId)
    if (result.success) {
      toast.success('已添加资源策略')
    } else {
      toast.error(result.errorMessage ?? '添加失败，请重试')
    }
  }

  const handleRemove = async (perm: EndUserPermission) => {
    setRemovingId(perm.id)
    try {
      const result = await removePermission(perm.id)
      if (result.success) {
        toast.success(`已移除「${perm.displayName ?? perm.modelDisplayName ?? perm.modelId}」策略`)
      } else {
        toast.error(result.errorMessage ?? '移除失败，请重试')
      }
    } catch {
      toast.error('移除资源策略时发生错误')
    } finally {
      setRemovingId(null)
    }
  }

  const handleSaveName = async (name: string) => {
    if (!name) return
    const result = await updateBundle(name, bundle?.description)
    if (result.success) {
      toast.success('已更新名称')
    } else {
      toast.error(result.errorMessage ?? '更新失败')
    }
  }

  const handleSaveDescription = async (description: string) => {
    if (!bundle?.name) return
    const result = await updateBundle(bundle.name, description || undefined)
    if (result.success) {
      toast.success('已更新描述')
    } else {
      toast.error(result.errorMessage ?? '更新失败')
    }
  }

  return (
    <PageLayout maxWidth="7xl">
      {/* Back nav */}
      <div className="mb-6">
        <Link
          href={backHref}
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          返回权限包列表
        </Link>
      </div>

      {/* Header */}
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

      {/* Error */}
      {error && (
        <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          加载失败：{error.message}
        </div>
      )}

      {/* Tabs */}
      {(() => {
        const tabs = [
          { key: 'strategies', label: '资源策略' },
          { key: 'versions',   label: '版本历史' },
        ] as const
        type TabKey = typeof tabs[number]['key']
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const [activeTab, setActiveTab] = React.useState<TabKey>('strategies')

        return (
          <>
            {/* Tab nav */}
            <div className="mb-5 flex items-end gap-0 border-b border-border">
              {tabs.map((t) => (
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
                      {permissionsByModel.size}
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

            {/* Tab: resource strategies */}
            {activeTab === 'strategies' && (
              <div className="rounded-md border border-border">
                <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
                  <div className="ml-auto">
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 gap-1.5 text-xs"
                      onClick={() => setAddDialogOpen(true)}
                      disabled={loading}
                    >
                      <Plus className="size-3.5" strokeWidth={1.5} />
                      添加资源策略
                    </Button>
                  </div>
                </div>

                {loading ? (
                  <div className="space-y-px">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="px-4 py-3">
                        <Skeleton className="mb-2 h-4 w-24" />
                        <div className="flex items-center gap-2">
                          <Skeleton className="h-5 w-12" />
                          <Skeleton className="h-5 w-16" />
                        </div>
                      </div>
                    ))}
                  </div>
                ) : permissionsByModel.size === 0 ? (
                  <div className="flex flex-col items-center justify-center py-16 text-center">
                    <Shield className="mb-3 size-9 text-muted-foreground/20" strokeWidth={1} />
                    <p className="text-sm font-semibold text-foreground">尚未配置任何资源策略</p>
                    <p className="mt-1 max-w-xs text-xs text-muted-foreground">
                      为每个需要访问控制的资源添加一条权限策略
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      className="mt-5 gap-1.5"
                      onClick={() => setAddDialogOpen(true)}
                    >
                      <Plus className="size-4" strokeWidth={1.5} />
                      添加资源策略
                    </Button>
                  </div>
                ) : (
                  <div className="divide-y divide-border">
                    {Array.from(permissionsByModel.entries()).map(([modelId, perms]) => {
                      const modelLabel = perms[0]?.modelDisplayName ?? modelId
                      return perms.map((perm) => (
                        <div
                          key={perm.id}
                          className="group flex items-center gap-4 px-4 py-2.5 hover:bg-foreground/[0.015]"
                        >
                          <div className="w-48 shrink-0">
                            <span className="font-mono text-xs text-foreground">
                              {modelLabel}
                            </span>
                          </div>
                          <div className="flex min-w-0 flex-1 flex-col">
                            <span className="truncate text-sm text-foreground">
                              {perm.displayName || '—'}
                            </span>
                            {perm.description && (
                              <span className="truncate text-xs text-muted-foreground">
                                {perm.description}
                              </span>
                            )}
                          </div>
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                aria-label={`移除 ${modelLabel} 策略`}
                                className="size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                                disabled={removingId === perm.id}
                              >
                                {removingId === perm.id ? (
                                  <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                  <Trash2 className="size-3.5" />
                                )}
                              </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>移除资源策略</AlertDialogTitle>
                                <AlertDialogDescription>
                                  确定要从该权限包中移除「{modelLabel}」的{ACTION_LABEL[perm.action] ?? perm.action}策略吗？此操作不会删除权限点本身。
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <AlertDialogFooter>
                                <AlertDialogCancel>取消</AlertDialogCancel>
                                <AlertDialogAction
                                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                  onClick={() => void handleRemove(perm)}
                                >
                                  确认移除
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        </div>
                      ))
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Tab: version history */}
            {activeTab === 'versions' && (
              <VersionHistoryPanel
                snapshots={(bundle?.snapshots ?? []) as BundleSnapshot[]}
                currentVersion={bundle?.currentVersion ?? 0}
                onRestore={async (version) => {
                  const result = await restoreBundle(version)
                  if (!result.success) {
                    throw new Error(result.errorMessage ?? '还原失败')
                  }
                }}
              />
            )}
          </>
        )
      })()}

      {/* Add strategy dialog */}
      <AddStrategyDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        bundleName={bundle?.name ?? ''}
        allPermissions={allPermissions}
        bundledPermissionIds={bundledPermissionIds}
        bundledModelIds={bundledModelIds}
        onAdd={handleAdd}
        orgName={orgName}
        projectSlug={projectSlug}
      />
    </PageLayout>
  )
}
