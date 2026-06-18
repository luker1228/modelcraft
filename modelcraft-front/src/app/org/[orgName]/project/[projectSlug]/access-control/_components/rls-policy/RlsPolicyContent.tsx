'use client'
/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-return */

import * as React from 'react'
import { ArrowDown, ArrowUp, ArrowUpDown, Loader2, Plus, ShieldOff, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import { Button } from '@web/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Skeleton } from '@web/components/ui/skeleton'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@web/components/ui/table'
import { GET_MODEL, GET_MODELS_BY_DATABASE } from '@/api-client/model'
import { REGISTERED_DATABASES } from '@/api-client/cluster'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { useQuery } from '@apollo/client'
import { useRlsPolicyList } from '../../_hooks/rls-policy/useRlsPolicyList'
import { useRlsPolicyManage } from '../../_hooks/rls-policy/useRlsPolicyManage'
import { PolicyEditorDialog } from './PolicyEditorDialog'
import type { RlsAction, RlsPoliciesOrderBy } from '@/generated/graphql'
import { FIXED_RLS_AUTH_VARIABLES } from './rls-expression-utils'

interface RlsPolicyContentProps {
  orgName: string
  projectSlug: string
}

interface RlsPolicySelectionState {
  databaseName: string | null
  modelId: string | null
}

function getRlsPolicySelectionKey(orgName: string, projectSlug: string): string {
  return `modelcraft:access-control:rls-policy:${orgName}:${projectSlug}`
}

function readRlsPolicySelection(orgName: string, projectSlug: string): RlsPolicySelectionState {
  if (typeof window === 'undefined') return { databaseName: null, modelId: null }

  try {
    const raw = window.localStorage.getItem(getRlsPolicySelectionKey(orgName, projectSlug))
    if (!raw) return { databaseName: null, modelId: null }

    const parsed = JSON.parse(raw) as Partial<RlsPolicySelectionState>
    return {
      databaseName: typeof parsed.databaseName === 'string' ? parsed.databaseName : null,
      modelId: typeof parsed.modelId === 'string' ? parsed.modelId : null,
    }
  } catch {
    return { databaseName: null, modelId: null }
  }
}

function writeRlsPolicySelection(
  orgName: string,
  projectSlug: string,
  selection: RlsPolicySelectionState,
): void {
  if (typeof window === 'undefined') return

  window.localStorage.setItem(
    getRlsPolicySelectionKey(orgName, projectSlug),
    JSON.stringify(selection),
  )
}

export function RlsPolicyContent({ orgName, projectSlug }: RlsPolicyContentProps) {
  const [selectedDatabaseName, setSelectedDatabaseName] = React.useState<string | null>(null)
  const [selectedModelId, setSelectedModelId] = React.useState<string | null>(null)
  const [selectionLoaded, setSelectionLoaded] = React.useState(false)

  const client = useProjectScopedClient(projectSlug)
  const { data: catalogData, loading: databasesLoading } = useQuery(REGISTERED_DATABASES, {
    client,
    variables: { input: { pageSize: 100 } },
    skip: !projectSlug,
  })
  const databases = React.useMemo(
    () => (catalogData?.registeredDatabases?.data?.databases ?? []) as Array<{ name: string }>,
    [catalogData?.registeredDatabases?.data?.databases],
  )
  const [editorOpen, setEditorOpen] = React.useState(false)
  const [deleteTargetId, setDeleteTargetId] = React.useState<string | null>(null)
  const [orderBy, setOrderBy] = React.useState<RlsPoliciesOrderBy | null>(null)

  const { policies, loading } = useRlsPolicyList({
    projectSlug,
    modelId: selectedModelId,
    orderBy,
  })
  const { upsertPolicy, deletePolicy, validateRlsExpression, upserting, deleting } = useRlsPolicyManage({
    projectSlug,
    modelId: selectedModelId ?? '',
  })

  React.useEffect(() => {
    const selection = readRlsPolicySelection(orgName, projectSlug)
    setSelectedDatabaseName(selection.databaseName)
    setSelectedModelId(selection.modelId)
    setSelectionLoaded(true)
  }, [orgName, projectSlug])

  React.useEffect(() => {
    if (!selectionLoaded) return
    writeRlsPolicySelection(orgName, projectSlug, {
      databaseName: selectedDatabaseName,
      modelId: selectedModelId,
    })
  }, [orgName, projectSlug, selectedDatabaseName, selectedModelId, selectionLoaded])

  const { data: modelsData, loading: modelsLoading } = useQuery(GET_MODELS_BY_DATABASE, {
    client,
    variables: { input: { databaseName: selectedDatabaseName } },
    skip: !selectedDatabaseName || !projectSlug,
  })
  const { data: selectedModelData } = useQuery(GET_MODEL, {
    client,
    variables: { id: selectedModelId },
    skip: !selectedModelId,
  })
  const models = React.useMemo(
    () =>
      (modelsData?.models?.items ?? []) as Array<{
        id: string
        name: string
        title?: string | null
      }>,
    [modelsData?.models?.items],
  )
  const selectedModel = React.useMemo(
    () =>
      (selectedModelData?.model?.model ?? null) as null | {
        id: string
        name: string
        title?: string | null
        fields?: Array<{ name: string; title?: string | null }>
      },
    [selectedModelData?.model?.model],
  )
  const selectedModelExists = React.useMemo(
    () => models.some((model) => model.id === selectedModelId),
    [models, selectedModelId],
  )

  React.useEffect(() => {
    if (!selectedDatabaseName || databases.length === 0) return
    if (databases.some((db) => db.name === selectedDatabaseName)) return
    setSelectedDatabaseName(null)
    setSelectedModelId(null)
  }, [databases, selectedDatabaseName])

  React.useEffect(() => {
    if (!selectedDatabaseName) return
    if (models.length === 0) return
    if (selectedModelId && selectedModelExists) return
    if (selectedModelId && !selectedModelExists) {
      setSelectedModelId(null)
    }
  }, [models, selectedDatabaseName, selectedModelExists, selectedModelId])
  const docsHref = `/org/${orgName}/project/${projectSlug}/access-control/examples`

  const handleUpsert = async (data: {
    policyName: string
    action: RlsAction
    role: string
    usingExpr?: string
    withCheckExpr?: string
  }) => {
    const result = await upsertPolicy(data)
    if (result.success) {
      toast.success('策略已保存')
      setEditorOpen(false)
    } else {
      toast.error(result.errorMessage ?? '保存失败')
    }
  }

  const handleDelete = async () => {
    if (!deleteTargetId) return
    const result = await deletePolicy(deleteTargetId)
    if (result.success) {
      toast.success('策略已删除')
      setDeleteTargetId(null)
    } else {
      toast.error(result.errorMessage ?? '删除失败')
    }
  }

  const actionLabel = (a: string) => a

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Select
            value={selectedDatabaseName ?? ''}
            onValueChange={(v) => {
              setSelectedDatabaseName(v || null)
              setSelectedModelId(null)
            }}
          >
            <SelectTrigger className="w-[220px]">
              <SelectValue placeholder={
                databasesLoading ? '加载中...' : '选择数据库...'
              } />
            </SelectTrigger>
            <SelectContent>
              {databases.map((db) => (
                <SelectItem key={db.name} value={db.name}>{db.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            value={selectedModelId ?? ''}
            onValueChange={(v) => setSelectedModelId(v || null)}
            disabled={!selectedDatabaseName}
          >
            <SelectTrigger className="w-[220px]">
              <SelectValue placeholder={
                !selectedDatabaseName ? '请先选择数据库' :
                modelsLoading ? '加载中...' : '选择模型...'
              } />
            </SelectTrigger>
            <SelectContent>
              {models.map((m: { id: string; name: string }) => (
                <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button
          onClick={() => setEditorOpen(true)}
          size="sm"
          disabled={!selectedModelId}
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="size-4" strokeWidth={1.5} />
          添加策略
        </Button>
      </div>

      {!selectedDatabaseName && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <ShieldOff className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
          <p className="text-sm font-semibold text-foreground">请先选择数据库</p>
          <p className="mt-1 text-xs text-muted-foreground">
            选择数据库和模型后查看和管理其 RLS 策略
          </p>
        </div>
      )}

      {selectedDatabaseName && !selectedModelId && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <ShieldOff className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
          <p className="text-sm font-semibold text-foreground">请选择一个模型</p>
          <p className="mt-1 text-xs text-muted-foreground">
            选择模型后查看和管理其 RLS 策略
          </p>
        </div>
      )}

      {selectedModelId && loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {selectedModelId && !loading && (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow className="border-b-2 border-border bg-card hover:bg-card">
                <TableHead className="h-10 w-[100px]">
                  <button
                    type="button"
                    className="flex items-center gap-1 text-[11px] font-medium uppercase tracking-wider text-foreground hover:text-foreground/70"
                    onClick={() => setOrderBy((prev) =>
                      prev === 'ACTION_ASC' ? 'ACTION_DESC' : 'ACTION_ASC'
                    )}
                  >
                    Action
                    {orderBy === 'ACTION_ASC' ? (
                      <ArrowUp className="size-3" />
                    ) : orderBy === 'ACTION_DESC' ? (
                      <ArrowDown className="size-3" />
                    ) : (
                      <ArrowUpDown className="size-3 opacity-40" />
                    )}
                  </button>
                </TableHead>
                <TableHead className="h-10 w-[140px]">
                  <button
                    type="button"
                    className="flex items-center gap-1 text-[11px] font-medium uppercase tracking-wider text-foreground hover:text-foreground/70"
                    onClick={() => setOrderBy((prev) =>
                      prev === 'ROLE_ASC' ? 'ROLE_DESC' : 'ROLE_ASC'
                    )}
                  >
                    Role
                    {orderBy === 'ROLE_ASC' ? (
                      <ArrowUp className="size-3" />
                    ) : orderBy === 'ROLE_DESC' ? (
                      <ArrowDown className="size-3" />
                    ) : (
                      <ArrowUpDown className="size-3 opacity-40" />
                    )}
                  </button>
                </TableHead>
                <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">Using Expr</TableHead>
                <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">Check Expr</TableHead>
                <TableHead className="h-10 w-[80px] text-right text-[11px] font-medium uppercase tracking-wider text-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {policies.map((policy) => (
                <TableRow key={policy.id} className="group border-b border-border last:border-0 hover:bg-foreground/[0.015]">
                  <TableCell className="h-12 text-[13px]">
                    <span className="inline-flex items-center rounded-md bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground">
                      {actionLabel(policy.action)}
                    </span>
                  </TableCell>
                  <TableCell className="h-12 text-[13px] font-medium text-foreground">
                    {policy.role || <span className="text-muted-foreground/40">默认</span>}
                  </TableCell>
                  <TableCell className="h-12">
                    <code className="line-clamp-1 text-[11px] text-muted-foreground">
                      {policy.usingExpr || '—'}
                    </code>
                  </TableCell>
                  <TableCell className="h-12">
                    <code className="line-clamp-1 text-[11px] text-muted-foreground">
                      {policy.withCheckExpr || '—'}
                    </code>
                  </TableCell>
                  <TableCell className="h-12 text-right">
                    <div className="flex items-center justify-end opacity-0 transition-opacity group-hover:opacity-100">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTargetId(policy.id)}
                      >
                        <Trash2 className="size-3.5" strokeWidth={1.5} />
                        删除
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {policies.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center">
                    <div className="flex flex-col items-center justify-center py-8">
                      <ShieldOff className="mb-2 size-7 text-muted-foreground/25" strokeWidth={1} />
                      <p className="text-sm font-semibold text-foreground">暂无策略</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        点击「添加策略」为该模型创建 RLS 策略
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      )}

      <PolicyEditorDialog
        open={editorOpen}
        onOpenChange={setEditorOpen}
        onSave={handleUpsert}
        onDryRun={validateRlsExpression}
        saving={upserting}
        modelFields={selectedModel?.fields ?? []}
        authVariables={FIXED_RLS_AUTH_VARIABLES}
        docsHref={docsHref}
      />

      <AlertDialog open={!!deleteTargetId} onOpenChange={(open) => { if (!open) setDeleteTargetId(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除策略</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除此策略吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting ? <><Loader2 className="mr-2 size-4 animate-spin" />删除中...</> : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
