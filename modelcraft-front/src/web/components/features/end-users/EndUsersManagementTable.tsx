'use client'

// src/web/components/features/end-users/EndUsersManagementTable.tsx
// Org 级终端用户管理表格（EndUser v2）

import { useState, useEffect } from 'react'
import { Plus, MoreHorizontal, RefreshCw, Users } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { CreateEndUserDialog } from './CreateEndUserDialog'
import { useOrgEndUsers, type OrgEndUser } from '@web/hooks/end-users/useOrgEndUsers'

// ============================================================================
// UserProjectBadges — fetches and renders the project list for one user
// ============================================================================

interface AccessibleProject {
  slug: string
  title: string
}

interface AccessibleProjectsBffResponse {
  projects?: AccessibleProject[]
}

interface UserProjectBadgesProps {
  orgName: string
  userId: string
}

function UserProjectBadges({ orgName, userId }: UserProjectBadgesProps) {
  const [projects, setProjects] = useState<AccessibleProject[] | null>(null)

  useEffect(() => {
    fetch(`/api/bff/org/${orgName}/end-user/users/${userId}/accessible-projects`)
      .then((r) => r.json() as Promise<AccessibleProjectsBffResponse>)
      .then((d) => setProjects(d.projects ?? []))
      .catch(() => setProjects([]))
  }, [orgName, userId])

  if (projects === null) {
    return <div className="h-4 w-20 animate-pulse rounded bg-muted" />
  }

  if (projects.length === 0) {
    return <span className="text-xs text-muted-foreground">暂无</span>
  }

  return (
    <div className="flex flex-wrap gap-1">
      {projects.map((p) => (
        <Badge key={p.slug} variant="outline" className="text-xs font-medium text-foreground">
          {p.title}
        </Badge>
      ))}
    </div>
  )
}

// ============================================================================
// Helpers
// ============================================================================

function formatDate(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleDateString('zh-CN')
  } catch {
    return '-'
  }
}

interface EndUsersManagementTableProps {
  orgName: string
}

export function EndUsersManagementTable({ orgName }: EndUsersManagementTableProps) {
  const { users, isLoading, error, reload, createUser, toggleUserStatus, deleteUser } =
    useOrgEndUsers(orgName)
  const [createOpen, setCreateOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const handleToggleStatus = async (user: OrgEndUser) => {
    setActionError(null)
    const newStatus = user.status === 'ACTIVE' ? 'DISABLED' : 'ACTIVE'
    try {
      await toggleUserStatus(user.id, newStatus)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '操作失败')
    }
  }

  const handleDelete = async (userId: string) => {
    if (!confirm('确认删除此用户？此操作不可恢复。')) return
    setActionError(null)
    try {
      await deleteUser(userId)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '删除失败')
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex items-center gap-2">
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1.5 size-4" />
          新增用户
        </Button>
        <Button size="sm" variant="outline" onClick={reload} disabled={isLoading}>
          <RefreshCw className={`mr-1.5 size-4 ${isLoading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>
      {actionError && (
        <p className="text-sm text-destructive">{actionError}</p>
      )}

      {/* Error State */}
      {error && !isLoading && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Users className="mb-3 size-10 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button size="sm" variant="outline" className="mt-3" onClick={reload}>
            重试
          </Button>
        </div>
      )}

      {/* Table */}
      {!error && (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="font-semibold">用户名</TableHead>
                <TableHead className="font-semibold">显示名称</TableHead>
                <TableHead className="w-56 font-semibold">关联项目</TableHead>
                <TableHead className="font-semibold">状态</TableHead>
                <TableHead className="font-semibold">创建时间</TableHead>
                <TableHead className="w-12" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading &&
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i}>
                    {Array.from({ length: 6 }).map((__, j) => (
                      <TableCell key={j}>
                        <div className="h-4 w-24 animate-pulse rounded bg-slate-200" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}

              {!isLoading && users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6}>
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <Users className="mb-3 size-10 text-muted-foreground/50" />
                      <p className="text-sm text-muted-foreground">暂无终端用户</p>
                      <Button
                        size="sm"
                        variant="outline"
                        className="mt-3"
                        onClick={() => setCreateOpen(true)}
                      >
                        新增第一个用户
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!isLoading &&
                users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex size-8 flex-shrink-0 items-center justify-center rounded-full bg-muted text-sm font-semibold text-foreground">
                          {user.username.charAt(0).toUpperCase()}
                        </div>
                        <span className="text-sm font-semibold">{user.username}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {user.displayName || '-'}
                    </TableCell>
                    <TableCell>
                      <UserProjectBadges orgName={orgName} userId={user.id} />
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={user.status === 'ACTIVE' ? 'outline' : 'destructive'}
                        className={user.status === 'ACTIVE' ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : ''}
                      >
                        {user.status === 'ACTIVE' ? '正常' : '已禁用'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(user.createdAt)}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button size="icon" variant="ghost" className="size-8">
                            <MoreHorizontal className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleToggleStatus(user)}>
                            {user.status === 'ACTIVE' ? '禁用账号' : '启用账号'}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => handleDelete(user.id)}
                          >
                            删除用户
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </div>
      )}

      {!isLoading && !error && users.length > 0 && (
        <p className="text-sm text-muted-foreground">共 {users.length} 名终端用户</p>
      )}

      <CreateEndUserDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={createUser}
      />
    </div>
  )
}
