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

// ── Types ────────────────────────────────────────────────────────────────────

interface EndUserNode {
  id: string
  username: string
  isForbidden: boolean
  createdBy?: string
  createdAt: string
  updatedAt: string
}

// 角色分配记录（来自 endUserRoleAssignments 查询）
interface RoleAssignmentEntry {
  projectSlug: string
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

interface RoleAssignmentsData {
  endUserRoleAssignments?: Array<{
    endUserId: string
    assignedAt: string
    role: {
      id: string
      name: string
      description?: string | null
    }
  }>
}

// ── Project Access Section ────────────────────────────────────────────────────

interface ProjectAccessSectionProps {
  orgName: string
  userId: string
}

function ProjectAccessSection({ orgName, userId }: ProjectAccessSectionProps) {
  const [assignments, setAssignments] = useState<RoleAssignmentEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    // endUserRoleAssignments 是 Project-scoped 查询，但实际 endpoint 基于 project context
    // 由于缺少通用 Org-level 查询，我们直接从 project GraphQL endpoint 查询
    // 这里用任意一个 project-scoped client 来查询（使用 "_" 作占位）
    // 实际上 endUserRoleAssignments 返回该 user 在 project 下的所有角色分配
    // TODO: 待后端提供 Org 级聚合查询后更新
    // 暂时从前端展示说明路径

    setLoading(false)
    setAssignments([])
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

  return (
    <div className="p-4">
      <p className="text-sm text-muted-foreground">
        用户在各项目下的角色分配请前往对应项目的「终端用户访问控制」页面查看和管理。
      </p>
    </div>
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
      {
        label: '状态',
        value: user.isForbidden ? (
          <Badge variant="secondary" className="text-xs text-muted-foreground">已禁用</Badge>
        ) : (
          <Badge variant="secondary" className="text-xs text-emerald-700">正常</Badge>
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

            {/* Project access card - 只读展示，引导到 Project 管理页 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <User className="size-4" />
                  项目访问
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <div className="border-t px-4 py-4">
                  <p className="text-sm text-muted-foreground">
                    用户的项目访问权限通过 Role Assignment 管理。
                    请前往对应项目的「终端用户访问控制」页面为用户分配或撤销角色。
                  </p>
                  <p className="mt-2 text-xs text-muted-foreground">
                    用户在某 Project 下拥有至少一个角色分配时，即可访问该 Project。
                  </p>
                  <Button
                    size="sm"
                    variant="outline"
                    className="mt-3"
                    onClick={() => router.push(`/org/${orgName}/projects`)}
                  >
                    <ArrowRight className="mr-1.5 size-4" />
                    前往项目列表
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </PageLayout>
    </AppLayout>
  )
}
