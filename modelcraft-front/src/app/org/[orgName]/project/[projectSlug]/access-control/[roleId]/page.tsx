'use client'

import * as React from 'react'
import Link from 'next/link'
import { useParams, useSearchParams, useRouter } from 'next/navigation'
import {
  ArrowLeft,
  KeyRound,
  Users,
  PackagePlus,
  X,
  Loader2,
  UserPlus,
  Search,
  Pencil,
  Check,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { toast } from 'sonner'
import { useQuery, useMutation } from '@apollo/client'

import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Skeleton } from '@web/components/ui/skeleton'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Checkbox } from '@web/components/ui/checkbox'
import { ScrollArea } from '@web/components/ui/scroll-area'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
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
import { PageLayout } from '@web/components/features/layout'

import { useRoleEdit } from '@/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/roles/useRoleEdit'
import { useRoleList } from '@/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/roles/useRoleList'
import { useProjectScopedClient } from '@api-client/apollo/public'
import { LIST_END_USERS } from '@/api-client/end-user'
import {
  ASSIGN_END_USER_ROLE_TO_USER,
  REVOKE_END_USER_ROLE_FROM_USER,
  UPDATE_END_USER_ROLE,
} from '@/api-client/rbac'
import type { EndUserPermissionAction, EndUserRowScope } from '@/types'

// ── BundlesTab ────────────────────────────────────────────────────────────────

interface BundlesTabProps {
  roleId: string
  orgName: string
  projectSlug: string
}

function BundlesTab({ roleId, orgName, projectSlug }: BundlesTabProps) {
  const [addDialogOpen, setAddDialogOpen] = React.useState(false)
  const [selectedBundleIds, setSelectedBundleIds] = React.useState<string[]>([])
  const [adding, setAdding] = React.useState(false)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set())

  const { role, allBundles, loading, assignBundle, revokeBundle } = useRoleEdit({
    orgName,
    projectSlug,
    roleId,
  })

  const assignedBundleIds = new Set(role?.permissionBundles.map((entry) => entry.bundle.id) ?? [])
  const assignedBundles = role?.permissionBundles.map((entry) => entry.bundle) ?? []
  const unassignedBundles = allBundles.filter((b) => !assignedBundleIds.has(b.id))

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const ACTION_LABEL: Record<EndUserPermissionAction, string> = {
    SELECT: '查询', INSERT: '新增', UPDATE: '更新', DELETE: '删除', EXPORT: '导出',
  }
  const ROW_SCOPE_LABEL: Record<EndUserRowScope, string> = {
    ALL: '全部', SELF: '本人', DEPT: '本部门', DEPT_AND_CHILDREN: '部门及子部门',
  }

  const handleAddConfirm = async () => {
    setAdding(true)
    let failCount = 0
    for (const bundleId of selectedBundleIds) {
      const result = await assignBundle(bundleId)
      if (!result.success) failCount++
    }
    setAdding(false)
    setSelectedBundleIds([])
    setAddDialogOpen(false)
    if (failCount > 0) toast.error(`${failCount} 个权限包添加失败`)
    else toast.success(`已添加 ${selectedBundleIds.length} 个权限包`)
  }

  const handleRevoke = async (bundleId: string, bundleName: string) => {
    setRevokingId(bundleId)
    const result = await revokeBundle(bundleId)
    setRevokingId(null)
    if (result.success) toast.success(`已移除权限包「${bundleName}」`)
    else toast.error(result.errorMessage ?? '移除失败')
  }

  return (
    <>
      {/* Toolbar */}
      <div className="mb-4">
        <Button
          variant="outline"
          size="sm"
          onClick={() => { setSelectedBundleIds([]); setAddDialogOpen(true) }}
          disabled={loading || unassignedBundles.length === 0}
        >
          <PackagePlus className="size-4" strokeWidth={1.5} />
          添加权限包
        </Button>
        {!loading && allBundles.length === 0 && (
          <p className="mt-2 text-xs text-muted-foreground">
            暂无可用权限包，
            <a
              href={`/org/${orgName}/project/${projectSlug}/access-control?tab=bundles`}
              className="text-primary underline-offset-2 hover:underline"
            >
              去创建权限包
            </a>
          </p>
        )}
      </div>

      {/* List */}
      <div className="rounded-md border border-border">
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            已关联权限包
          </span>
          {!loading && (
            <span className="ml-2 rounded-full bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
              {assignedBundles.length}
            </span>
          )}
        </div>

        {loading ? (
          <div className="space-y-px">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-4 w-36" />
                <Skeleton className="h-4 w-24" />
              </div>
            ))}
          </div>
        ) : assignedBundles.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <KeyRound className="mb-3 size-8 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm text-muted-foreground">尚未关联权限包</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {assignedBundles.map((bundle) => {
              const expanded = expandedIds.has(bundle.id)
              const perms = bundle.permissions ?? []
              return (
                <div key={bundle.id} className="group">
                  {/* Bundle row */}
                  <div className="flex items-center gap-2 px-4 py-3 hover:bg-muted/20">
                    <button
                      className="flex min-w-0 flex-1 items-center gap-2 text-left"
                      onClick={() => toggleExpand(bundle.id)}
                    >
                      {expanded
                        ? <ChevronDown className="size-3.5 shrink-0 text-muted-foreground" />
                        : <ChevronRight className="size-3.5 shrink-0 text-muted-foreground" />
                      }
                      <span className="text-sm font-medium text-foreground">{bundle.name}</span>
                      {bundle.description && (
                        <span className="truncate text-xs text-muted-foreground">{bundle.description}</span>
                      )}
                      <span className="shrink-0 text-xs text-muted-foreground/60">
                        {perms.length} 个权限点
                      </span>
                    </button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7 shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                      disabled={revokingId === bundle.id}
                      onClick={() => handleRevoke(bundle.id, bundle.name)}
                    >
                      {revokingId === bundle.id
                        ? <Loader2 className="size-4 animate-spin" />
                        : <X className="size-4" />
                      }
                    </Button>
                  </div>

                  {/* Expanded permissions */}
                  {expanded && (
                    <div className="border-t border-border bg-muted/10">
                      {perms.length === 0 ? (
                        <p className="px-10 py-3 text-xs text-muted-foreground">暂无权限点</p>
                      ) : (
                        <div className="divide-y divide-border/50">
                          {perms.map((item) => {
                            const p = item.permission
                            return (
                              <div key={p.id} className="flex items-center gap-3 px-10 py-2.5">
                                <span className="min-w-0 flex-1 text-sm text-foreground">
                                  {p.displayName || p.modelId}
                                </span>
                                <span className="shrink-0 rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                                  {ACTION_LABEL[p.action] ?? p.action}
                                </span>
                                <span className="shrink-0 text-[10px] text-muted-foreground/60">
                                  {ROW_SCOPE_LABEL[p.rowScope] ?? p.rowScope}
                                </span>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Add Dialog */}
      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>添加权限包</DialogTitle>
          </DialogHeader>
          <ScrollArea className="max-h-60 py-2">
            {unassignedBundles.length === 0 ? (
              <p className="text-sm text-muted-foreground">所有权限包已关联</p>
            ) : (
              <div className="space-y-1">
                {unassignedBundles.map((bundle) => (
                  <label
                    key={bundle.id}
                    className="flex cursor-pointer items-start gap-3 rounded-md p-2 hover:bg-muted"
                  >
                    <Checkbox
                      className="mt-0.5"
                      checked={selectedBundleIds.includes(bundle.id)}
                      onCheckedChange={(checked) =>
                        setSelectedBundleIds((prev) =>
                          checked ? [...prev, bundle.id] : prev.filter((id) => id !== bundle.id)
                        )
                      }
                    />
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-foreground">{bundle.name}</p>
                      {bundle.description && (
                        <p className="text-xs text-muted-foreground">{bundle.description}</p>
                      )}
                    </div>
                  </label>
                ))}
              </div>
            )}
          </ScrollArea>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>取消</Button>
            <Button
              onClick={handleAddConfirm}
              disabled={selectedBundleIds.length === 0 || adding}
              className="bg-primary text-primary-foreground hover:bg-primary/90"
            >
              {adding
                ? <><Loader2 className="mr-2 size-4 animate-spin" />添加中...</>
                : `添加${selectedBundleIds.length > 0 ? ` (${selectedBundleIds.length})` : ''}`
              }
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// ── UsersTab ──────────────────────────────────────────────────────────────────

interface UsersTabProps {
  roleId: string
  orgName: string
  projectSlug: string
}

type EndUserNode = { id: string; username: string; isForbidden: boolean }
type AssignPayload = { assignEndUserRole?: { error?: { message?: string } | null } }
type RevokePayload = { revokeEndUserRole?: { error?: { message?: string } | null } }

function UsersTab({ roleId, orgName, projectSlug }: UsersTabProps) {
  const client = useProjectScopedClient(projectSlug, orgName)
  const [search, setSearch] = React.useState('')
  const [assigningId, setAssigningId] = React.useState<string | null>(null)
  const [revokingId, setRevokingId] = React.useState<string | null>(null)
  const [assignedUserIds, setAssignedUserIds] = React.useState<Set<string>>(new Set())

  const { data: usersData, loading: usersLoading } = useQuery<{
    listEndUsers?: { connection?: { nodes?: EndUserNode[] } }
  }>(LIST_END_USERS, {
    client,
    variables: { input: { first: 100 } },
  })

  const filteredUsers = React.useMemo(() => {
    const nodes = usersData?.listEndUsers?.connection?.nodes ?? []
    return nodes.filter((u) =>
      u.username.toLowerCase().includes(search.toLowerCase())
    )
  }, [usersData, search])

  const [assignRoleMutation] = useMutation(ASSIGN_END_USER_ROLE_TO_USER, { client })
  const [revokeRoleMutation] = useMutation(REVOKE_END_USER_ROLE_FROM_USER, { client })

  const handleAssign = React.useCallback(async (userId: string) => {
    setAssigningId(userId)
    try {
      const result = await assignRoleMutation({
        variables: { input: { endUserId: userId, roleId } },
      })
      const payload = (result.data as AssignPayload | undefined)?.assignEndUserRole
      if (payload?.error) {
        toast.error(payload.error.message ?? '分配失败')
      } else {
        setAssignedUserIds((prev) => new Set([...prev, userId]))
        toast.success('角色已分配')
      }
    } catch {
      toast.error('分配失败')
    } finally {
      setAssigningId(null)
    }
  }, [assignRoleMutation, roleId])

  const handleRevoke = React.useCallback(async (userId: string, username: string) => {
    setRevokingId(userId)
    try {
      const result = await revokeRoleMutation({
        variables: { input: { endUserId: userId, roleId } },
      })
      const payload = (result.data as RevokePayload | undefined)?.revokeEndUserRole
      if (payload?.error) {
        toast.error(payload.error.message ?? '撤销失败')
      } else {
        setAssignedUserIds((prev) => {
          const next = new Set(prev)
          next.delete(userId)
          return next
        })
        toast.success(`已撤销用户「${username}」的角色`)
      }
    } catch {
      toast.error('撤销失败')
    } finally {
      setRevokingId(null)
    }
  }, [revokeRoleMutation, roleId])

  return (
    <div>
      {/* Search */}
      <div className="relative mb-4 max-w-xs">
        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
        <Input
          placeholder="搜索用户..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* List */}
      <div className="rounded-md border border-border">
        <div className="flex items-center border-b border-border bg-muted/30 px-4 py-2.5">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            终端用户
          </span>
        </div>

        {usersLoading ? (
          <div className="space-y-px">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <Skeleton className="h-4 w-32" />
              </div>
            ))}
          </div>
        ) : filteredUsers.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Users className="mb-3 size-8 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm text-muted-foreground">
              {search ? '未找到匹配用户' : '暂无终端用户'}
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {filteredUsers.map((user) => {
              const assigned = assignedUserIds.has(user.id)
              const isAssigning = assigningId === user.id
              const isRevoking = revokingId === user.id
              return (
                <div
                  key={user.id}
                  className="flex items-center justify-between gap-3 px-4 py-3 hover:bg-muted/20"
                >
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-foreground">{user.username}</p>
                    {user.isForbidden && (
                      <p className="mt-0.5 text-xs text-destructive">已禁用</p>
                    )}
                  </div>
                  {assigned ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 shrink-0 gap-1 text-xs text-muted-foreground hover:text-destructive"
                      disabled={isRevoking}
                      onClick={() => handleRevoke(user.id, user.username)}
                    >
                      {isRevoking
                        ? <Loader2 className="size-3.5 animate-spin" />
                        : <X className="size-3.5" />
                      }
                      撤销
                    </Button>
                  ) : (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 shrink-0 gap-1 text-xs text-muted-foreground hover:text-foreground"
                      disabled={isAssigning || user.isForbidden}
                      onClick={() => handleAssign(user.id)}
                    >
                      {isAssigning
                        ? <Loader2 className="size-3.5 animate-spin" />
                        : <UserPlus className="size-3.5" />
                      }
                      分配
                    </Button>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

// ── RoleDetailPage ────────────────────────────────────────────────────────────

export default function RoleDetailPage() {
  const { orgName, projectSlug, roleId } =
    useParams<{ orgName: string; projectSlug: string; roleId: string }>()
  const searchParams = useSearchParams()
  const router = useRouter()

  const rawTab = searchParams.get('tab')
  const activeTab = rawTab === 'users' ? 'users' : 'bundles'

  const backHref = `/org/${orgName}/project/${projectSlug}/access-control?tab=roles`

  const { role, loading } = useRoleEdit({ orgName, projectSlug, roleId })
  const { deleteRole } = useRoleList({ orgName, projectSlug })
  const client = useProjectScopedClient(projectSlug, orgName)

  const [deleteOpen, setDeleteOpen] = React.useState(false)
  const [deleting, setDeleting] = React.useState(false)

  // Description edit state
  const [editingDesc, setEditingDesc] = React.useState(false)
  const [descValue, setDescValue] = React.useState('')
  const [savingDesc, setSavingDesc] = React.useState(false)

  const [updateRole] = useMutation(UPDATE_END_USER_ROLE, {
    client,
    refetchQueries: ['GetEndUserRole'],
  })

  const handleEditDesc = () => {
    setDescValue(role?.description ?? '')
    setEditingDesc(true)
  }

  const handleSaveDesc = async () => {
    setSavingDesc(true)
    try {
      const result = await updateRole({
        variables: { id: roleId, input: { description: descValue } },
      })
      const data = result.data as Record<string, { error?: { message?: string } } | null | undefined> | null | undefined
      const payload = data?.['updateEndUserRole']
      if (payload?.error) {
        toast.error(payload.error.message ?? '保存失败')
      } else {
        toast.success('描述已更新')
        setEditingDesc(false)
      }
    } catch {
      toast.error('保存失败')
    } finally {
      setSavingDesc(false)
    }
  }

  const handleDelete = async () => {
    if (!role) return
    setDeleting(true)
    const result = await deleteRole(role)
    setDeleting(false)
    if (result.success) {
      toast.success(`已删除角色「${role.name}」`)
      router.push(backHref)
    } else {
      toast.error(result.errorMessage ?? '删除失败')
    }
  }

  const tabs = [
    { key: 'bundles', label: '权限包', icon: KeyRound },
    { key: 'users', label: '用户管理', icon: Users },
  ]

  return (
    <PageLayout maxWidth="5xl">
      {/* Back nav */}
      <div className="mb-6">
        <Link
          href={backHref}
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          返回角色列表
        </Link>
      </div>

      {/* Header */}
      <div className="mb-6 flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          {loading ? (
            <>
              <Skeleton className="h-7 w-40" />
              <Skeleton className="mt-2 h-4 w-56" />
            </>
          ) : (
            <>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-semibold tracking-tight text-foreground">
                  {role?.name ?? '角色详情'}
                </h1>
                {role?.isImplicit && (
                  <Badge variant="secondary" className="text-xs">内置</Badge>
                )}
              </div>

              {/* Role ID */}
              <p className="mt-1 font-mono text-xs text-muted-foreground/60">{roleId}</p>

              {/* Description */}
              {editingDesc ? (
                <div className="mt-3 flex items-start gap-2">
                  <Textarea
                    value={descValue}
                    onChange={(e) => setDescValue(e.target.value)}
                    placeholder="描述该角色的用途"
                    rows={2}
                    className="resize-none text-sm"
                    autoFocus
                  />
                  <div className="flex flex-col gap-1">
                    <Button
                      size="icon"
                      className="size-7 bg-primary text-primary-foreground hover:bg-primary/90"
                      disabled={savingDesc}
                      onClick={handleSaveDesc}
                    >
                      {savingDesc ? <Loader2 className="size-3.5 animate-spin" /> : <Check className="size-3.5" />}
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-7"
                      onClick={() => setEditingDesc(false)}
                    >
                      <X className="size-3.5" />
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="group mt-2 flex items-center gap-1.5">
                  <p className="text-sm text-muted-foreground">
                    {role?.description || <span className="text-muted-foreground/40">暂无描述</span>}
                  </p>
                  {!role?.isImplicit && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-6 opacity-0 transition-opacity group-hover:opacity-100"
                      onClick={handleEditDesc}
                    >
                      <Pencil className="size-3" />
                    </Button>
                  )}
                </div>
              )}
            </>
          )}
        </div>

        {!loading && role && !role.isImplicit && (
          <Button
            variant="ghost"
            size="sm"
            className="shrink-0 text-muted-foreground hover:text-destructive"
            onClick={() => setDeleteOpen(true)}
          >
            删除角色
          </Button>
        )}
      </div>

      {/* Tab nav */}
      <div className="mb-6 border-b border-border">
        <div className="flex gap-0">
          {tabs.map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              onClick={() => {
                const url = key === 'bundles'
                  ? `/org/${orgName}/project/${projectSlug}/access-control/${roleId}`
                  : `/org/${orgName}/project/${projectSlug}/access-control/${roleId}?tab=${key}`
                router.push(url)
              }}
              className={[
                'flex items-center gap-1.5 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors',
                activeTab === key
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground',
              ].join(' ')}
            >
              <Icon className="size-3.5" strokeWidth={1.5} />
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab content */}
      {activeTab === 'bundles' && (
        <BundlesTab roleId={roleId} orgName={orgName} projectSlug={projectSlug} />
      )}
      {activeTab === 'users' && (
        <UsersTab roleId={roleId} orgName={orgName} projectSlug={projectSlug} />
      )}

      {/* Delete Dialog */}
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除角色</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除角色「{role?.name}」吗？此操作不可撤销，该角色的所有用户将失去对应权限。
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
    </PageLayout>
  )
}
