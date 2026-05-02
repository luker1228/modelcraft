'use client'

// src/app/org/[orgName]/end-users/[userId]/page.tsx
// Org 级终端用户详情页：基本信息、状态管理、项目访问控制

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  ArrowLeft,
  User,
  ShieldCheck,
  ShieldOff,
  Trash2,
  Plus,
  MoreHorizontal,
  RefreshCw,
  FolderOpen,
} from 'lucide-react'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { getOrgScopedClient, createProjectScopedClient } from '@api-client/apollo/public'
import { GET_PROJECTS } from '@api-client/project'
import {
  LIST_END_USERS,
  UPDATE_END_USER_STATUS,
  DELETE_END_USER,
} from '@api-client/end-user/graphql-docs'
import {
  LIST_PROJECT_END_USER_ACCESS,
  GRANT_PROJECT_END_USER_ACCESS,
  REVOKE_PROJECT_END_USER_ACCESS,
  GET_END_USER_BUNDLES,
} from '@api-client/rbac'

// ── Types ────────────────────────────────────────────────────────────────────

interface EndUserNode {
  id: string
  username: string
  isForbidden: boolean
  createdBy?: string
  createdAt: string
  updatedAt: string
}

interface ProjectItem {
  id: string
  slug: string
  title: string
}

interface ProjectAccessEntry {
  id: string
  permissionBundleId: string
  permissionBundleName: string
  grantedBy: string
  grantedAt: string
}

interface BundleItem {
  id: string
  name: string
  slug: string
}

// ── GraphQL response shapes ──────────────────────────────────────────────────

interface ListEndUsersData {
  listEndUsers: {
    connection?: { nodes: EndUserNode[] }
    error?: { message?: string }
  }
}

interface ProjectsData {
  projects: ProjectItem[]
}

interface UpdateStatusData {
  updateEndUserStatus: {
    endUser?: EndUserNode
    error?: { message?: string }
  }
}

interface DeleteEndUserData {
  deleteEndUser: {
    success: boolean
    error?: { message?: string }
  }
}

interface ListProjectAccessData {
  listProjectEndUserAccess: {
    connection?: {
      nodes: Array<{
        id: string
        endUser: { id: string; username: string }
        permissionBundleId: string
        permissionBundleName: string
        grantedBy: string
        grantedAt: string
      }>
    }
    error?: { message?: string }
  }
}

interface GrantAccessData {
  grantEndUserProjectAccess: {
    access?: { id: string }
    error?: { message?: string }
  }
}

interface RevokeAccessData {
  revokeEndUserProjectAccess: {
    success: boolean
    error?: { message?: string }
  }
}

interface BundlesData {
  endUserPermissionBundles: {
    edges: Array<{ node: BundleItem }>
  }
}

// ── Grant Dialog ─────────────────────────────────────────────────────────────

interface GrantDialogProps {
  open: boolean
  orgName: string
  projectSlug: string
  userId: string
  bundles: BundleItem[]
  onClose: () => void
  onGranted: () => void
}

function GrantAccessDialog({
  open,
  orgName,
  projectSlug,
  userId,
  bundles,
  onClose,
  onGranted,
}: GrantDialogProps) {
  const [selectedBundle, setSelectedBundle] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleGrant = async () => {
    if (!selectedBundle) return
    setLoading(true)
    setError(null)
    try {
      const client = createProjectScopedClient(orgName, projectSlug)
      const { data } = await client.mutate<GrantAccessData>({
        mutation: GRANT_PROJECT_END_USER_ACCESS,
        variables: { input: { endUserId: userId, permissionBundleId: selectedBundle } },
      })
      const err = data?.grantEndUserProjectAccess?.error
      if (err?.message) throw new Error(err.message)
      onGranted()
      onClose()
      setSelectedBundle('')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '授权失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenChange = (v: boolean) => {
    if (!v) {
      onClose()
      setSelectedBundle('')
      setError(null)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>授权项目访问</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <p className="text-sm text-muted-foreground">
            为用户授权访问项目 <span className="font-medium text-foreground">{projectSlug}</span>
          </p>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">权限包</label>
            <Select value={selectedBundle} onValueChange={setSelectedBundle}>
              <SelectTrigger>
                <SelectValue placeholder="选择权限包" />
              </SelectTrigger>
              <SelectContent>
                {bundles.map((b) => (
                  <SelectItem key={b.id} value={b.id}>
                    {b.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>
            取消
          </Button>
          <Button onClick={handleGrant} disabled={!selectedBundle || loading}>
            {loading ? '授权中…' : '确认授权'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Project Row ───────────────────────────────────────────────────────────────

interface ProjectRowProps {
  project: ProjectItem
  orgName: string
  userId: string
  onAccessChange: () => void
}

function ProjectAccessRow({ project, orgName, userId, onAccessChange }: ProjectRowProps) {
  const [access, setAccess] = useState<ProjectAccessEntry | null | undefined>(undefined) // undefined = loading
  const [bundles, setBundles] = useState<BundleItem[]>([])
  const [grantOpen, setGrantOpen] = useState(false)
  const [revoking, setRevoking] = useState(false)
  const [rowError, setRowError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  // Load access for this project + userId
  useEffect(() => {
    let cancelled = false
    setAccess(undefined)
    setRowError(null)

    const client = createProjectScopedClient(orgName, project.slug)

    Promise.all([
      client.query<ListProjectAccessData>({
        query: LIST_PROJECT_END_USER_ACCESS,
        variables: { input: { first: 200 } },
        fetchPolicy: 'network-only',
      }),
      client.query<BundlesData>({
        query: GET_END_USER_BUNDLES,
        fetchPolicy: 'cache-first',
      }),
    ])
      .then(([{ data: accessData }, { data: bundleData }]) => {
        if (cancelled) return
        const err = accessData?.listProjectEndUserAccess?.error
        if (err?.message) {
          setRowError(err.message)
          setAccess(null)
          return
        }
        const nodes = accessData?.listProjectEndUserAccess?.connection?.nodes ?? []
        const matched = nodes.find((n) => n.endUser.id === userId) ?? null
        setAccess(
          matched
            ? {
                id: matched.id,
                permissionBundleId: matched.permissionBundleId,
                permissionBundleName: matched.permissionBundleName,
                grantedBy: matched.grantedBy,
                grantedAt: matched.grantedAt,
              }
            : null
        )
        const bEdges = bundleData?.endUserPermissionBundles?.edges ?? []
        setBundles(bEdges.map((e) => e.node))
      })
      .catch((e: unknown) => {
        if (cancelled) return
        setRowError(e instanceof Error ? e.message : '加载失败')
        setAccess(null)
      })

    return () => {
      cancelled = true
    }
  }, [orgName, project.slug, userId, version])

  const handleRevoke = async () => {
    if (!access) return
    if (!confirm(`确认撤销该用户对项目「${project.title}」的访问权限？`)) return
    setRevoking(true)
    setRowError(null)
    try {
      const client = createProjectScopedClient(orgName, project.slug)
      const { data } = await client.mutate<RevokeAccessData>({
        mutation: REVOKE_PROJECT_END_USER_ACCESS,
        variables: { input: { accessId: access.id } },
      })
      const err = data?.revokeEndUserProjectAccess?.error
      if (err?.message) throw new Error(err.message)
      setVersion((v) => v + 1)
      onAccessChange()
    } catch (e: unknown) {
      setRowError(e instanceof Error ? e.message : '撤销失败')
    } finally {
      setRevoking(false)
    }
  }

  const handleGranted = () => {
    setVersion((v) => v + 1)
    onAccessChange()
  }

  return (
    <>
      <TableRow>
        <TableCell>
          <div className="flex items-center gap-2">
            <FolderOpen className="size-3.5 shrink-0 text-muted-foreground/60" />
            <span className="text-sm font-medium text-foreground">{project.title}</span>
            <span className="text-xs text-muted-foreground/60">{project.slug}</span>
          </div>
        </TableCell>
        <TableCell>
          {access === undefined ? (
            <div className="h-4 w-20 animate-pulse rounded bg-muted" />
          ) : access ? (
            <Badge variant="secondary" className="text-xs text-emerald-700">
              已授权
            </Badge>
          ) : (
            <Badge variant="outline" className="text-xs text-muted-foreground">
              未授权
            </Badge>
          )}
        </TableCell>
        <TableCell className="text-sm text-muted-foreground">
          {access === undefined ? (
            <div className="h-4 w-24 animate-pulse rounded bg-muted" />
          ) : access ? (
            access.permissionBundleName
          ) : (
            <span className="text-muted-foreground/40">—</span>
          )}
        </TableCell>
        <TableCell className="text-sm text-muted-foreground">
          {access === undefined ? (
            <div className="h-4 w-20 animate-pulse rounded bg-muted" />
          ) : access ? (
            new Date(access.grantedAt).toLocaleDateString('zh-CN')
          ) : (
            <span className="text-muted-foreground/40">—</span>
          )}
        </TableCell>
        <TableCell className="text-right">
          {access === undefined ? null : access ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  size="icon"
                  variant="ghost"
                  className="size-7 text-muted-foreground/50 hover:text-foreground"
                  disabled={revoking}
                >
                  <MoreHorizontal className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  className="text-destructive"
                  onClick={handleRevoke}
                  disabled={revoking}
                >
                  撤销访问权限
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <Button
              size="sm"
              variant="outline"
              className="h-7 text-xs"
              onClick={() => setGrantOpen(true)}
              disabled={bundles.length === 0}
            >
              <Plus className="mr-1 size-3" />
              授权
            </Button>
          )}
          {rowError && <p className="mt-1 text-xs text-destructive">{rowError}</p>}
        </TableCell>
      </TableRow>

      <GrantAccessDialog
        open={grantOpen}
        orgName={orgName}
        projectSlug={project.slug}
        userId={userId}
        bundles={bundles}
        onClose={() => setGrantOpen(false)}
        onGranted={handleGranted}
      />
    </>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function OrgEndUserDetailPage() {
  const params = useParams<{ orgName: string; userId: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const userId = params.userId

  const [user, setUser] = useState<EndUserNode | null | undefined>(undefined) // undefined = loading
  const [projects, setProjects] = useState<ProjectItem[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)
  const [statusLoading, setStatusLoading] = useState(false)
  const [statusError, setStatusError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)
  const [accessVersion, setAccessVersion] = useState(0)

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    let cancelled = false
    setLoadError(null)

    const client = getOrgScopedClient()

    Promise.all([
      client.query<ListEndUsersData>({
        query: LIST_END_USERS,
        variables: { input: { first: 500 } },
        fetchPolicy: 'network-only',
      }),
      client.query<ProjectsData>({
        query: GET_PROJECTS,
        variables: { input: {} },
        fetchPolicy: 'network-only',
      }),
    ])
      .then(([{ data: usersData }, { data: projectsData }]) => {
        if (cancelled) return
        const err = usersData?.listEndUsers?.error
        if (err?.message) {
          setLoadError(err.message)
          setUser(null)
          return
        }
        const nodes = usersData?.listEndUsers?.connection?.nodes ?? []
        const matched = nodes.find((u) => u.id === userId) ?? null
        setUser(matched)
        setProjects(projectsData?.projects ?? [])
      })
      .catch((e: unknown) => {
        if (cancelled) return
        setLoadError(e instanceof Error ? e.message : '加载用户详情失败')
        setUser(null)
      })

    return () => {
      cancelled = true
    }
  }, [userId, version])

  const handleToggleStatus = async () => {
    if (!user) return
    setStatusError(null)
    setStatusLoading(true)
    try {
      const client = getOrgScopedClient()
      const { data } = await client.mutate<UpdateStatusData>({
        mutation: UPDATE_END_USER_STATUS,
        variables: { input: { userId: user.id, isForbidden: !user.isForbidden } },
      })
      const err = data?.updateEndUserStatus?.error
      if (err?.message) throw new Error(err.message)
      const updated = data?.updateEndUserStatus?.endUser
      if (updated) {
        setUser((prev) => (prev ? { ...prev, isForbidden: updated.isForbidden } : prev))
      }
    } catch (e: unknown) {
      setStatusError(e instanceof Error ? e.message : '操作失败')
    } finally {
      setStatusLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!user) return
    if (!confirm(`确认删除用户「${user.username}」？此操作不可恢复。`)) return
    try {
      const client = getOrgScopedClient()
      const { data } = await client.mutate<DeleteEndUserData>({
        mutation: DELETE_END_USER,
        variables: { input: { userId: user.id } },
      })
      const err = data?.deleteEndUser?.error
      if (err?.message) throw new Error(err.message)
      router.push(`/org/${orgName}/end-users`)
    } catch (e: unknown) {
      setStatusError(e instanceof Error ? e.message : '删除失败')
    }
  }

  const isLoading = user === undefined

  const infoRows = useMemo(() => {
    if (!user) return []
    return [
      { label: '用户名', value: user.username },
      {
        label: '状态',
        value: user.isForbidden ? (
          <Badge variant="secondary" className="text-xs text-muted-foreground">
            已禁用
          </Badge>
        ) : (
          <Badge variant="secondary" className="text-xs text-emerald-700">
            正常
          </Badge>
        ),
      },
      { label: '创建人', value: user.createdBy ?? '—' },
      { label: '创建时间', value: new Date(user.createdAt).toLocaleString('zh-CN') },
      { label: '更新时间', value: new Date(user.updatedAt).toLocaleString('zh-CN') },
    ]
  }, [user])

  return (
    <AppLayout pageTitle="终端用户详情">
      <PageLayout maxWidth="5xl">
        <PageHeader title="终端用户详情" bordered />

        {/* Back button */}
        <div className="mb-4">
          <Button variant="outline" size="sm" onClick={() => router.push(`/org/${orgName}/end-users`)}>
            <ArrowLeft className="mr-1.5 size-4" />
            返回用户列表
          </Button>
        </div>

        {/* Loading skeleton */}
        {isLoading && (
          <div className="space-y-4">
            <Card>
              <CardContent className="space-y-3 py-6">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="h-4 w-64 animate-pulse rounded bg-muted" />
                ))}
              </CardContent>
            </Card>
          </div>
        )}

        {/* Load error */}
        {!isLoading && loadError && (
          <Card>
            <CardContent className="flex flex-col items-center gap-3 py-12 text-center">
              <p className="text-sm text-destructive">{loadError}</p>
              <Button size="sm" variant="outline" onClick={reload}>
                <RefreshCw className="mr-1.5 size-4" />
                重试
              </Button>
            </CardContent>
          </Card>
        )}

        {/* User not found */}
        {!isLoading && !loadError && !user && (
          <Card>
            <CardContent className="py-10 text-sm text-muted-foreground">
              用户不存在或已删除
            </CardContent>
          </Card>
        )}

        {/* Main content */}
        {!isLoading && !loadError && user && (
          <div className="space-y-5">
            {/* User info card */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2 text-base">
                    <div className="flex size-8 items-center justify-center rounded-full bg-muted text-sm font-medium text-foreground">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <span>{user.username}</span>
                    {user.isForbidden ? (
                      <Badge variant="secondary" className="text-xs text-muted-foreground">
                        已禁用
                      </Badge>
                    ) : (
                      <Badge variant="secondary" className="text-xs text-emerald-700">
                        正常
                      </Badge>
                    )}
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={handleToggleStatus}
                      disabled={statusLoading}
                    >
                      {user.isForbidden ? (
                        <>
                          <ShieldCheck className="mr-1.5 size-4 text-emerald-600" />
                          启用账号
                        </>
                      ) : (
                        <>
                          <ShieldOff className="mr-1.5 size-4 text-amber-600" />
                          禁用账号
                        </>
                      )}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-destructive hover:text-destructive"
                      onClick={handleDelete}
                      disabled={statusLoading}
                    >
                      <Trash2 className="mr-1.5 size-4" />
                      删除用户
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                {statusError && (
                  <p className="mb-3 text-sm text-destructive">{statusError}</p>
                )}
                <dl className="grid grid-cols-2 gap-x-8 gap-y-2 text-sm sm:grid-cols-3">
                  {infoRows.map((row) => (
                    <div key={row.label}>
                      <dt className="text-muted-foreground">{row.label}</dt>
                      <dd className="mt-0.5 font-medium text-foreground">{row.value}</dd>
                    </div>
                  ))}
                </dl>
              </CardContent>
            </Card>

            {/* Project access card */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <User className="size-4" />
                  项目访问权限
                  <span className="ml-1 text-xs font-normal text-muted-foreground">
                    （共 {projects.length} 个项目）
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {projects.length === 0 ? (
                  <div className="py-8 text-center text-sm text-muted-foreground">
                    暂无项目
                  </div>
                ) : (
                  <div className="rounded-b-md border-t border-border">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>项目</TableHead>
                          <TableHead>状态</TableHead>
                          <TableHead>权限包</TableHead>
                          <TableHead>授权时间</TableHead>
                          <TableHead className="w-20" />
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {projects.map((project) => (
                          <ProjectAccessRow
                            key={`${project.slug}-${accessVersion}`}
                            project={project}
                            orgName={orgName}
                            userId={userId}
                            onAccessChange={() => setAccessVersion((v) => v + 1)}
                          />
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )}
      </PageLayout>
    </AppLayout>
  )
}
