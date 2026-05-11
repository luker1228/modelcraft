'use client'

// src/app/org/[orgName]/end-users/[userId]/page.tsx
// Org 级终端用户详情页：基本信息、状态管理、项目导航（只读）

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
import { Skeleton } from '@web/components/ui/skeleton'
import { getOrgScopedClient } from '@api-client/apollo/public'
import {
  LIST_END_USERS,
  UPDATE_END_USER_STATUS,
  DELETE_END_USER,
} from '@api-client/end-user/graphql-docs'
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

// ── Project Nav Section ───────────────────────────────────────────────────────
// 纯导航：列出当前 Org 下所有项目，每项提供「前往管理」入口
// 不涉及授权状态——EndUser 与项目是多对多关系，授权在项目侧管理

interface ProjectNavSectionProps {
  orgName: string
}

function ProjectNavSection({ orgName }: ProjectNavSectionProps) {
  const router = useRouter()
  const [projects, setProjects] = useState<Array<{ id: string; slug: string; title: string }>>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    const client = getOrgScopedClient()
    client.query<GetProjectsData>({
      query: GET_PROJECTS,
      fetchPolicy: 'network-only',
    })
      .then(({ data }) => {
        if (cancelled) return
        setProjects(data?.projects ?? [])
        setLoading(false)
      })
      .catch((e: unknown) => {
        if (cancelled) return
        setError(e instanceof Error ? e.message : '加载失败')
        setLoading(false)
      })

    return () => { cancelled = true }
  }, [orgName])

  if (loading) {
    return (
      <div className="divide-y">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center justify-between px-5 py-3">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-4 w-16" />
          </div>
        ))}
      </div>
    )
  }

  if (error) {
    return <div className="px-5 py-4 text-sm text-destructive">{error}</div>
  }

  if (projects.length === 0) {
    return (
      <div className="px-5 py-8 text-center">
        <p className="text-sm text-muted-foreground">当前组织下尚未创建项目</p>
      </div>
    )
  }

  return (
    <div className="divide-y">
      {projects.map((project) => (
        <button
          key={project.id}
          type="button"
          className="group flex w-full items-center justify-between px-5 py-3 text-left transition-colors hover:bg-muted/40"
          onClick={() => router.push(`/org/${orgName}/project/${project.slug}/end-user-access`)}
        >
          <span className="text-sm font-medium text-foreground">{project.title}</span>
          <span className="flex items-center gap-1 text-xs text-muted-foreground transition-colors group-hover:text-foreground">
            前往管理
            <ArrowRight className="size-3.5 transition-transform group-hover:translate-x-0.5" />
          </span>
        </button>
      ))}
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

    const client = getOrgScopedClient()

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

            {/* Project nav card - 引导到各项目终端用户管理页 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <User className="size-4" />
                  项目管理
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <ProjectNavSection orgName={orgName} />
              </CardContent>
            </Card>
          </div>
        )}
      </PageLayout>
    </AppLayout>
  )
}
