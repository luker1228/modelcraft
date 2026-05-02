'use client'

// src/web/components/features/end-users/EndUsersManagementTable.tsx
// Org 级终端用户管理表格（EndUser v2）

import { useState } from 'react'
import { Plus, MoreHorizontal, RefreshCw, Users, Search } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import { Input } from '@web/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
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
import { getOrgScopedClient } from '@api-client/apollo/public'
import { GET_PROJECTS } from '@api-client/project'

interface ProjectItem {
  slug: string
  title: string
}

interface GetProjectsData {
  projects: ProjectItem[]
}

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
  const { users, isLoading, error, search, setSearch, reload, createUser, toggleUserStatus, deleteUser } =
    useOrgEndUsers(orgName)
  const [createOpen, setCreateOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)
  const [assignOpen, setAssignOpen] = useState(false)
  const [selectedUser, setSelectedUser] = useState<OrgEndUser | null>(null)
  const [projects, setProjects] = useState<ProjectItem[]>([])
  const [selectedProjectSlug, setSelectedProjectSlug] = useState('')

  const handleToggleStatus = async (user: OrgEndUser) => {
    setActionError(null)
    try {
      await toggleUserStatus(user.id, !user.isForbidden)
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

  const handleOpenAssignDialog = async (user: OrgEndUser) => {
    setSelectedUser(user)
    setAssignOpen(true)
    setSelectedProjectSlug('')

    if (projects.length > 0) return

    try {
      const client = getOrgScopedClient()
      const { data } = await client.query<GetProjectsData>({
        query: GET_PROJECTS,
        variables: { input: { first: 100 } },
        fetchPolicy: 'network-only',
      })
      setProjects(data?.projects ?? [])
    } catch {
      setActionError('加载项目列表失败')
    }
  }

  const handleGoAssignProject = () => {
    if (!selectedUser || !selectedProjectSlug) return
    window.location.assign(
      `/org/${orgName}/project/${selectedProjectSlug}/end-user-access?grantUserId=${selectedUser.id}`
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex items-center gap-2">
        <div className="relative max-w-xs flex-1">
          <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input
            placeholder="搜索用户名…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-8 pl-8 text-sm"
          />
        </div>
        <div className="ml-auto flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={reload} disabled={isLoading}>
            <RefreshCw className={`mr-1.5 size-4 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1.5 size-4" />
            新增用户
          </Button>
        </div>
      </div>

      {actionError && (
        <p className="text-sm text-destructive">{actionError}</p>
      )}

      {/* Error State */}
      {error && !isLoading && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Users className="mb-3 size-8 text-muted-foreground/40" strokeWidth={1.5} />
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button size="sm" variant="outline" className="mt-3" onClick={reload}>
            重试
          </Button>
        </div>
      )}

      {/* Table */}
      {!error && (
        <div className="rounded-md border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>用户名</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建人</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="w-12" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading &&
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i}>
                    {Array.from({ length: 5 }).map((__, j) => (
                      <TableCell key={j}>
                        <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}

              {!isLoading && users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <Users className="mb-3 size-8 text-muted-foreground/40" strokeWidth={1.5} />
                      {search ? (
                        <p className="text-sm text-muted-foreground">未找到匹配「{search}」的用户</p>
                      ) : (
                        <>
                          <p className="text-sm text-muted-foreground">暂无终端用户</p>
                          <Button
                            size="sm"
                            variant="outline"
                            className="mt-3"
                            onClick={() => setCreateOpen(true)}
                          >
                            新增第一个用户
                          </Button>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!isLoading &&
                users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-2.5">
                        <div className="flex size-7 flex-shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium text-foreground">
                          {user.username.charAt(0).toUpperCase()}
                        </div>
                        <span className="text-sm font-medium text-foreground">
                          {user.username}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {user.isForbidden ? (
                        <Badge variant="secondary" className="text-xs text-muted-foreground">
                          已禁用
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="text-xs text-emerald-700">
                          正常
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {user.createdBy || <span className="text-muted-foreground/50">—</span>}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(user.createdAt)}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="size-7 text-muted-foreground/50 hover:text-foreground"
                          >
                            <MoreHorizontal className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleToggleStatus(user)}>
                            {user.isForbidden ? '启用账号' : '禁用账号'}
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => void handleOpenAssignDialog(user)}>
                            分配到项目
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
        <p className="text-sm text-muted-foreground">
          {search ? `找到 ${users.length} 名用户` : `共 ${users.length} 名终端用户`}
        </p>
      )}

      <CreateEndUserDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={createUser}
      />

      <Dialog open={assignOpen} onOpenChange={setAssignOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>分配用户到项目</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              用户: <span className="font-medium text-foreground">{selectedUser?.username ?? '-'}</span>
            </p>
            <div className="space-y-2">
              <Label>目标项目</Label>
              <Select value={selectedProjectSlug} onValueChange={setSelectedProjectSlug}>
                <SelectTrigger>
                  <SelectValue placeholder="选择项目" />
                </SelectTrigger>
                <SelectContent>
                  {projects.map((project) => (
                    <SelectItem key={project.slug} value={project.slug}>
                      {project.title} ({project.slug})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAssignOpen(false)}>
              取消
            </Button>
            <Button onClick={handleGoAssignProject} disabled={!selectedProjectSlug}>
              前往分配
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
