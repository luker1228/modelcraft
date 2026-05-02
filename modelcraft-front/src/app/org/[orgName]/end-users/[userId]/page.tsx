'use client'

// src/app/org/[orgName]/end-users/[userId]/page.tsx
// Org 级终端用户详情页：基本信息、状态管理、项目访问（只读）

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  ArrowLeft,
  User,
  ShieldCheck,
  ShieldOff,
  Trash2,
  RefreshCw,
  ArrowRight,
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
import { Skeleton } from '@web/components/ui/skeleton'
import { getOrgScopedClient, createProjectScopedClient } from '@api-client/apollo/public'
import {
  LIST_END_USERS,
  UPDATE_END_USER_STATUS,
  DELETE_END_USER,
} from '@api-client/end-user/graphql-docs'
import { GET_END_USER_ROLE_ASSIGNMENTS } from '@api-client/rbac'
import { GET_PROJECTS } from '@api-client/project/graphql-docs'

// ── Types ────────────────────────────────────────────────────────────────────

interface EndUserNode {
  id: string
  username: string
  isForbidden: boolean
  createdBy?: string
  createdAt: string
  updatedAt: string
}

// 角色分配记录（按项目聚合）
interface RoleAssignmentEntry {
  projectSlug: string
  projectTitle: string
  roleName: string
  assignedAt: string
}

// ── GraphQL response shapes ──────────────────────────────────────────────────

interface ListEndUsersData {
  listEndUsers: {
    connection?: { nodes: EndUserNode[] }
    error?: { message?: string }
  }
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

interface GetProjectsData {
  projects: Array<{
    id: string
    slug: string
    title: string
  }>
}

interface RoleAssignmentsData {
  endUserRoleAssignments?: Array<{
    endUserId: string
    assignedAt: string
    role: {
      id: string
      name: string
      isImplicit: boolean
    }
  }>
}

// ── Project Access Section ────────────────────────────────────────────────────

interface ProjectAccessSectionProps {
  orgName: string
  userId: string
}

function ProjectAccessSection({ orgName, userId }: ProjectAccessSectionProps) {
  const router = useRouter()
  const [assignments, setAssignments] = useState<RoleAssignmentEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    async function fetchAssignments() {
      // Step 1: fetch all projects in this org
      const orgClient = getOrgScopedClient()
      const projectsResult = await orgClient.query<GetProjectsData>({
        query: GET_PROJECTS,
        fetchPolicy: 'network-only',
      })
      const projects = projectsResult.data?.projects ?? []

      if (projects.length === 0) {
        return []
      }

      // Step 2: for each project, query role assignments for this user
      const results = await Promise.allSettled(
        projects.map(async (project) => {
          const projectClient = createProjectScopedClient(orgName, project.slug)
          const { data } = await projectClient.query<RoleAssignmentsData>({
            query: GET_END_USER_ROLE_ASSIGNMENTS,
            variables: { endUserId: userId },
            fetchPolicy: 'network-only',
          })
          const rows: RoleAssignmentEntry[] = []
          for (const a of data?.endUserRoleAssignments ?? []) {
            if (!a.role.isImplicit) {
              rows.push({
                projectSlug: project.slug,
                projectTitle: project.title,
                roleName: a.role.name,
                assignedAt: a.assignedAt,
              })
            }
          }
          return rows
        })
      )

      const all: RoleAssignmentEntry[] = []
      for (const r of results) {
        if (r.status === 'fulfilled') {
          all.push(...r.value)
        }
      }
      return all
    }

    fetchAssignments()
      .then((rows) => {
        if (!cancelled) {
          setAssignments(rows)
          setLoading(false)
        }
      })
      .catch((e: unknown) => {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : '加载项目访问数据失败')
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [orgName, userId])

  if (loading) {
    return (
      <div className="space-y-2 p-4">
        {[1, 2].map((i) => <Skeleton key={i} className="h-10 w-full" />)}
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-sm text-destructive">{error}</div>
    )
  }

  if (assignments.length === 0) {
    return (
      <div className="flex flex-col items-start gap-4 px-6 py-8">
        <div className="space-y-1">
          <p className="text-sm font-medium text-foreground">尚未分配任何项目角色</p>
          <p className="text-sm text-muted-foreground">
            在各项目的「终端用户访问」页为该用户分配角色后，这里将显示汇总。
          </p>
        </div>
        <Button
          size="sm"
          variant="outline"
          onClick={() => router.push(`/org/${orgName}/workspace`)}
        >
          前往项目列表
          <ArrowRight className="ml-1.5 size-3.5" />
        </Button>
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>项目</TableHead>
          <TableHead>角色</TableHead>
          <TableHead>授权时间</TableHead>
          <TableHead className="w-24" />
        </TableRow>
      </TableHeader>
      <TableBody>
        {assignments.map((row, i) => (
          <TableRow key={i}>
            <TableCell className="font-medium">{row.projectTitle}</TableCell>
            <TableCell>{row.roleName}</TableCell>
            <TableCell className="text-muted-foreground">
              {new Date(row.assignedAt).toLocaleString('zh-CN')}
            </TableCell>
            <TableCell>
              <Button
                size="sm"
                variant="ghost"
                className="h-7 px-2 text-xs"
                onClick={() =>
                  router.push(`/org/${orgName}/project/${row.projectSlug}/end-user-access`)
                }
              >
                前往管理
                <ArrowRight className="ml-1 size-3" />
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function OrgEndUserDetailPage() {
  const params = useParams<{ orgName: string; userId: string }>()
  const router = useRouter()
  const orgName = params.orgName
  const userId = params.userId

  const [user, setUser] = useState<EndUserNode | null | undefined>(undefined)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [statusLoading, setStatusLoading] = useState(false)
  const [statusError, setStatusError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  const reload = useCallback(() => setVersion((v) => v + 1), [])

  useEffect(() => {
    let cancelled = false
    setLoadError(null)

    const client = getOrgScopedClient(orgName)

    client.query<ListEndUsersData>({
      query: LIST_END_USERS,
      variables: { input: { first: 500 } },
      fetchPolicy: 'network-only',
    })
      .then(({ data }) => {
        if (cancelled) return
        const err = data?.listEndUsers?.error
        if (err?.message) {
          setLoadError(err.message)
          setUser(null)
          return
        }
        const nodes = data?.listEndUsers?.connection?.nodes ?? []
        const matched = nodes.find((u) => u.id === userId) ?? null
        setUser(matched)
      })
      .catch((e: unknown) => {
        if (cancelled) return
        setLoadError(e instanceof Error ? e.message : '加载用户详情失败')
        setUser(null)
      })

    return () => { cancelled = true }
  }, [orgName, userId, version])

  const handleToggleStatus = async () => {
    if (!user) return
    setStatusError(null)
    setStatusLoading(true)
    try {
      const client = getOrgScopedClient(orgName)
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
      const client = getOrgScopedClient(orgName)
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
      { label: '创建人', value: user.createdBy ? <span className="font-mono text-xs text-muted-foreground">{user.createdBy.slice(0, 8)}…</span> : '—' },
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
                    <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-indigo-50 text-xs font-semibold text-indigo-700">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <span>{user.username}</span>
                    {user.isForbidden ? (
                      <Badge variant="secondary" className="text-xs text-muted-foreground">已禁用</Badge>
                    ) : (
                      <Badge variant="secondary" className="text-xs text-emerald-700">正常</Badge>
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
                          <ShieldCheck className="mr-1.5 size-4" />
                          启用账号
                        </>
                      ) : (
                        <>
                          <ShieldOff className="mr-1.5 size-4" />
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
                <dl className="grid grid-cols-2 gap-x-8 gap-y-4 text-sm">
                  {infoRows.map((row) => (
                    <div key={row.label}>
                      <dt className="text-xs text-muted-foreground">{row.label}</dt>
                      <dd className="mt-1 font-medium text-foreground">{row.value}</dd>
                    </div>
                  ))}
                </dl>
              </CardContent>
            </Card>

            {/* Project access card - 只读展示，引导到 Project 管理页 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <User className="size-4" />
                  项目访问
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <div className="border-t">
                  <ProjectAccessSection orgName={orgName} userId={userId} />
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </PageLayout>
    </AppLayout>
  )
}
