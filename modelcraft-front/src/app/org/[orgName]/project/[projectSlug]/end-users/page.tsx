'use client'

import { useState, useCallback } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, useMutation } from '@apollo/client'
import {
  Users,
  Plus,
  Search,
  Trash2,
  Ban,
  CheckCircle2,
  Eye,
  EyeOff,
  Wrench,
} from 'lucide-react'
import { toast } from 'sonner'
import { useProjectScopedClient } from '@bff/apollo/public'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Badge } from '@web/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
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
import { LIST_END_USERS } from '@web/graphql/queries/end-user'
import {
  CREATE_END_USER,
  UPDATE_END_USER_STATUS,
  DELETE_END_USER,
  INIT_PRIVATE_DB,
} from '@web/graphql/mutations/end-user'

// ── Types ────────────────────────────────────────────────────────────

interface EndUser {
  id: string
  username: string
  isForbidden: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

interface PageInfo {
  hasNextPage: boolean
  endCursor: string | null
}

interface EndUserConnection {
  nodes: EndUser[]
  pageInfo: PageInfo
  totalCount: number
}

interface ListEndUsersData {
  listEndUsers: {
    connection: EndUserConnection | null
    error: { __typename: string; message: string } | null
  }
}

interface CreateEndUserData {
  createEndUser: {
    endUser: EndUser | null
    error: {
      __typename: string
      message: string
      suggestion?: string
    } | null
  }
}

interface UpdateEndUserStatusData {
  updateEndUserStatus: {
    endUser: EndUser | null
    error: { __typename: string; message: string } | null
  }
}

interface DeleteEndUserData {
  deleteEndUser: {
    success: boolean
    error: { __typename: string; message: string } | null
  }
}

interface InitPrivateDBData {
  initPrivateDB: {
    success: boolean
    error: { __typename: string; message: string } | null
  }
}

// ── Page ─────────────────────────────────────────────────────────────

export default function EndUsersPage() {
  const params = useParams()
  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string

  const projectClient = useProjectScopedClient(projectSlug, orgName)

  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Create dialog state
  const [createOpen, setCreateOpen] = useState(false)
  const [newUsername, setNewUsername] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)

  // Delete confirm state
  const [deleteTarget, setDeleteTarget] = useState<EndUser | null>(null)

  // ── Queries & Mutations ───────────────────────────────────────────

  const { data, loading, refetch } = useQuery<ListEndUsersData>(LIST_END_USERS, {
    client: projectClient,
    skip: !projectSlug,
    variables: {
      input: {
        search: debouncedSearch || undefined,
        first: 50,
      },
    },
    onError: (err) => {
      toast.error('获取用户列表失败', { description: err.message })
    },
  })

  const [createEndUser, { loading: creating }] =
    useMutation<CreateEndUserData>(CREATE_END_USER, {
      client: projectClient,
      onCompleted: (d) => {
        const error = d.createEndUser.error
        if (error) {
          const desc =
            'suggestion' in error && error.suggestion
              ? `${error.message}。建议: ${error.suggestion}`
              : error.message
          toast.error('创建用户失败', { description: desc })
        } else {
          toast.success(`用户 ${d.createEndUser.endUser?.username} 已创建`)
          setCreateOpen(false)
          setNewUsername('')
          setNewPassword('')
          refetch()
        }
      },
      onError: (err) => {
        toast.error('创建用户失败', { description: err.message })
      },
    })

  const [updateStatus, { loading: updatingId }] =
    useMutation<UpdateEndUserStatusData>(UPDATE_END_USER_STATUS, {
      client: projectClient,
      onCompleted: (d) => {
        const error = d.updateEndUserStatus.error
        if (error) {
          toast.error('更新状态失败', { description: error.message })
        } else {
          const user = d.updateEndUserStatus.endUser
          toast.success(
            user?.isForbidden ? `用户 ${user.username} 已禁用` : `用户 ${user?.username} 已启用`
          )
          refetch()
        }
      },
      onError: (err) => {
        toast.error('更新状态失败', { description: err.message })
      },
    })

  const [deleteEndUser, { loading: deleting }] =
    useMutation<DeleteEndUserData>(DELETE_END_USER, {
      client: projectClient,
      onCompleted: (d) => {
        const error = d.deleteEndUser.error
        if (error) {
          toast.error('删除用户失败', { description: error.message })
        } else {
          toast.success(`用户已删除`)
          setDeleteTarget(null)
          refetch()
        }
      },
      onError: (err) => {
        toast.error('删除用户失败', { description: err.message })
      },
    })

  const [initPrivateDB, { loading: initMutationLoading }] =
    useMutation<InitPrivateDBData>(INIT_PRIVATE_DB, {
      client: projectClient,
      onCompleted: (d) => {
        const payload = d.initPrivateDB
        if (payload.error) {
          toast.error('初始化失败', { description: payload.error.message })
          return
        }
        toast.success('私有库初始化成功')
        refetch()
      },
      onError: (err) => {
        toast.error('初始化失败', { description: err.message })
      },
    })

  // ── Handlers ─────────────────────────────────────────────────────

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value
      setSearch(val)
      // Simple debounce via setTimeout
      const timer = setTimeout(() => setDebouncedSearch(val), 300)
      return () => clearTimeout(timer)
    },
    []
  )

  const handleCreate = () => {
    if (!newUsername.trim() || !newPassword.trim()) return
    createEndUser({
      variables: { input: { username: newUsername.trim(), password: newPassword } },
    })
  }

  const handleToggleStatus = (user: EndUser) => {
    updateStatus({
      variables: { input: { userId: user.id, isForbidden: !user.isForbidden } },
    })
  }

  const handleDelete = () => {
    if (!deleteTarget) return
    deleteEndUser({ variables: { input: { userId: deleteTarget.id } } })
  }

  const handleInitPrivateDB = () => {
    void initPrivateDB()
  }

  // ── Data ─────────────────────────────────────────────────────────

  const connection = data?.listEndUsers?.connection
  const users = connection?.nodes ?? []
  const totalCount = connection?.totalCount ?? 0
  const listError = data?.listEndUsers?.error

  const showInitPrivateDBAction = Boolean(
    listError
      && listError.__typename === 'ClusterNotFound'
      && (
        listError.message.includes('私有库尚未初始化')
        || /unknown database/i.test(listError.message)
        || listError.message.includes('mc_private_')
        || listError.message.includes('PRIVATE_DB_NOT_INITIALIZED')
      )
  )

  const displayListErrorMessage = (() => {
    if (!listError) return ''
    if (showInitPrivateDBAction) {
      return '项目私有库尚未初始化，初始化后即可管理终端用户。'
    }
    return listError.message
  })()


  // ── Render ───────────────────────────────────────────────────────

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      {/* Header */}
      <section className="rounded-lg border border-border bg-background p-6 shadow-sm">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-md bg-blue-100 p-2 text-blue-600">
              <Users className="size-5" strokeWidth={1.5} />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-foreground">用户管理</h1>
              <p className="text-sm text-muted-foreground">
                管理项目的终端用户账号
              </p>
            </div>
          </div>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 size-4" strokeWidth={1.5} />
            新建用户
          </Button>
        </div>

        {/* Search & stats */}
        <div className="mb-4 flex items-center gap-3">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
            <Input
              className="pl-8"
              placeholder="搜索用户名..."
              value={search}
              onChange={handleSearchChange}
            />
          </div>
          <span className="text-sm text-muted-foreground">
            共 {totalCount} 个用户
          </span>
        </div>

        {/* Error state */}
        {listError && (
          <div className="space-y-3 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            <div>{displayListErrorMessage}</div>
            {showInitPrivateDBAction && (
              <div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void handleInitPrivateDB()}
                  disabled={initMutationLoading}
                  className="h-8 border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100"
                >
                  <Wrench className="mr-1 size-3.5" />
                  {initMutationLoading ? '初始化中...' : '初始化私有库'}
                </Button>
              </div>
            )}
          </div>
        )}

        {/* User table */}
        {loading ? (
          <div className="py-12 text-center text-sm text-muted-foreground">加载中...</div>
        ) : users.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            {debouncedSearch ? '未找到匹配用户' : '暂无终端用户'}
          </div>
        ) : (
          <div className="overflow-hidden rounded-md border border-border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/40">
                  <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">用户名</th>
                  <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">状态</th>
                  <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">创建时间</th>
                  <th className="px-4 py-2.5 text-right font-medium text-muted-foreground">操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user, idx) => (
                  <tr
                    key={user.id}
                    className={idx % 2 === 0 ? 'bg-background' : 'bg-muted/20'}
                  >
                    <td className="px-4 py-3 font-mono font-medium text-foreground">
                      {user.username}
                    </td>
                    <td className="px-4 py-3">
                      {user.isForbidden ? (
                        <Badge variant="destructive" className="gap-1">
                          <Ban className="size-3" />
                          已禁用
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="gap-1 bg-green-100 text-green-700 hover:bg-green-100">
                          <CheckCircle2 className="size-3" />
                          正常
                        </Badge>
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {new Date(user.createdAt).toLocaleString('zh-CN', {
                        year: 'numeric',
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 px-2 text-xs"
                          onClick={() => handleToggleStatus(user)}
                          disabled={!!updatingId}
                        >
                          {user.isForbidden ? (
                            <>
                              <CheckCircle2 className="mr-1 size-3" />
                              启用
                            </>
                          ) : (
                            <>
                              <Ban className="mr-1 size-3" />
                              禁用
                            </>
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                          onClick={() => setDeleteTarget(user)}
                        >
                          <Trash2 className="mr-1 size-3" />
                          删除
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>新建终端用户</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                placeholder="3–64 位字母、数字、下划线或连字符"
                value={newUsername}
                onChange={(e) => setNewUsername(e.target.value)}
                autoComplete="off"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="至少 8 位，须含字母和数字"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  autoComplete="new-password"
                  className="pr-9"
                />
                <button
                  type="button"
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  onClick={() => setShowPassword((v) => !v)}
                >
                  {showPassword ? (
                    <EyeOff className="size-4" strokeWidth={1.5} />
                  ) : (
                    <Eye className="size-4" strokeWidth={1.5} />
                  )}
                </button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">取消</Button>
            </DialogClose>
            <Button
              onClick={handleCreate}
              disabled={creating || !newUsername.trim() || !newPassword.trim()}
            >
              {creating ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除用户</AlertDialogTitle>
            <AlertDialogDescription>
              将永久删除用户{' '}
              <span className="font-mono font-medium text-foreground">
                {deleteTarget?.username}
              </span>
              ，此操作无法撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleting ? '删除中...' : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
