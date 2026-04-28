'use client'

import * as React from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  ShieldCheck,
  Loader2,
  Trash2,
  Plus,
  Pencil,
  Check,
  X,
  Users,
} from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
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
import { PageLayout } from '@web/components/features/layout'

import { useBundleManage } from '@/app/org/[orgName]/project/[projectSlug]/rbac/bundles/_hooks/useBundleManage'
import type { EndUserPermission } from '@/types'

// ── Label maps ───────────────────────────────────────────────────────────────

const ACTION_LABEL: Record<string, string> = {
  SELECT: '查询',
  INSERT: '新增',
  UPDATE: '更新',
  DELETE: '删除',
  EXPORT: '导出',
}

const ROW_SCOPE_LABEL: Record<string, string> = {
  ALL: '全部',
  SELF: '本人',
  DEPT: '本部门',
  DEPT_AND_CHILDREN: '部门及子部门',
}

const ACTION_VARIANT: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  SELECT: 'secondary',
  INSERT: 'default',
  UPDATE: 'outline',
  DELETE: 'destructive',
  EXPORT: 'outline',
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
      className: [
        'w-full rounded border border-border bg-background px-2 py-1 text-sm outline-none',
        'focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20',
        'disabled:opacity-60',
        inputClassName,
      ]
        .filter(Boolean)
        .join(' '),
    }

    return (
      <div className={['flex items-start gap-1.5', className].filter(Boolean).join(' ')}>
        {multiline ? (
          <textarea {...sharedProps} rows={2} />
        ) : (
          <input {...sharedProps} type="text" />
        )}
        <div className="mt-0.5 flex shrink-0 gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="size-6 text-indigo-600 hover:bg-indigo-50 hover:text-indigo-700"
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
      className={['group/field flex cursor-pointer items-start gap-1.5', className]
        .filter(Boolean)
        .join(' ')}
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

// ── BundleDetailPage ─────────────────────────────────────────────────────────

export default function BundleDetailPage() {
  const { orgName, projectSlug, bundleId } =
    useParams<{ orgName: string; projectSlug: string; bundleId: string }>()

  const backHref = `/org/${orgName}/project/${projectSlug}/roles?tab=bundles`
  const rolesTabHref = `/org/${orgName}/project/${projectSlug}/roles?tab=roles`

  const { bundle, loading, rolesLoading, error, removePermission, updateBundle, associatedRoles } =
    useBundleManage({
      orgName,
      projectSlug,
      bundleId,
    })

  const [removingId, setRemovingId] = React.useState<string | null>(null)

  const handleRemove = async (perm: EndUserPermission) => {
    setRemovingId(perm.id)
    try {
      const result = await removePermission(perm.id)
      if (result.success) {
        toast.success(`已移除「${perm.displayName ?? perm.modelId}」`)
      } else {
        toast.error(result.errorMessage ?? '移除失败，请重试')
      }
    } catch {
      toast.error('移除权限点时发生错误')
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

  const permissions = bundle?.permissions ?? []

  return (
    <PageLayout maxWidth="5xl">
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

      {/* Associated roles section */}
      <div className="mb-6 rounded-md border border-border">
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            关联的角色策略
          </span>
          {!rolesLoading && (
            <span className="ml-2 rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
              {associatedRoles.length}
            </span>
          )}
        </div>

        {rolesLoading ? (
          <div className="space-y-px">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-20" />
              </div>
            ))}
          </div>
        ) : associatedRoles.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 text-center">
            <Users className="mb-3 size-8 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm font-semibold text-foreground">暂未关联到任何角色策略</p>
            <p className="mt-1 text-xs text-muted-foreground">将权限包添加到角色，角色成员即可获得对应权限</p>
            <Button asChild variant="outline" size="sm" className="mt-4">
              <Link href={rolesTabHref}>
                <Plus className="size-4" strokeWidth={1.5} />
                去关联角色策略
              </Link>
            </Button>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {associatedRoles.map((role) => (
              <div
                key={role.id}
                className="flex items-center gap-3 px-4 py-3 hover:bg-foreground/[0.015]"
              >
                <Link
                  href={`/org/${orgName}/project/${projectSlug}/roles/${role.id}`}
                  className="text-sm text-indigo-600 transition-colors hover:text-indigo-700"
                >
                  {role.name}
                </Link>
                {role.isImplicit && (
                  <Badge variant="secondary" className="h-5 px-1.5 py-0 text-[11px]">
                    内置
                  </Badge>
                )}
                {role.description && (
                  <span className="text-xs text-muted-foreground">{role.description}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Permissions list */}
      <div className="rounded-md border border-border">
        {/* Header row */}
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            关联的权限点
          </span>
          {!loading && (
            <span className="ml-2 rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
              {permissions.length}
            </span>
          )}
        </div>

        {loading ? (
          <div className="space-y-px">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-4 w-28" />
                <Skeleton className="h-5 w-14" />
                <Skeleton className="h-5 w-16" />
              </div>
            ))}
          </div>
        ) : permissions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-14 text-center">
            <ShieldCheck className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm font-semibold text-foreground">暂无关联权限点</p>
            <p className="mt-1 text-xs text-muted-foreground">
              先创建权限点，再将其添加到权限包中
            </p>
            <Button
              asChild
              variant="outline"
              size="sm"
              className="mt-4"
            >
              <Link href={`/org/${orgName}/project/${projectSlug}/roles?tab=permissions`}>
                <Plus className="size-4" strokeWidth={1.5} />
                去管理权限点
              </Link>
            </Button>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {permissions.map((perm) => (
              <div
                key={perm.id}
                className="group flex items-center gap-3 px-4 py-3 hover:bg-foreground/[0.015]"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-foreground">
                      {perm.displayName || perm.modelDisplayName || perm.modelId}
                    </span>
                    {perm.displayName && (perm.modelDisplayName || perm.modelId) && (
                      <span className="font-mono text-[11px] text-muted-foreground/60">
                        {perm.modelDisplayName ?? perm.modelId}
                      </span>
                    )}
                  </div>
                  <div className="mt-1 flex items-center gap-1.5">
                    <Badge
                      variant={ACTION_VARIANT[perm.action] ?? 'secondary'}
                      className="h-5 px-1.5 py-0 text-[11px]"
                    >
                      {ACTION_LABEL[perm.action] ?? perm.action}
                    </Badge>
                    <Badge variant="outline" className="h-5 px-1.5 py-0 text-[11px]">
                      {ROW_SCOPE_LABEL[perm.rowScope] ?? perm.rowScope}
                    </Badge>
                  </div>
                </div>

                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`移除权限点${perm.displayName ?? perm.modelId}`}
                      className="size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-destructive"
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
                      <AlertDialogTitle>确认解绑权限点</AlertDialogTitle>
                      <AlertDialogDescription>
                        确定要从该权限包中移除「{perm.displayName ?? perm.modelId}」吗？
                        此操作不会删除权限点本身。
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
            ))}
          </div>
        )}
      </div>
    </PageLayout>
  )
}
