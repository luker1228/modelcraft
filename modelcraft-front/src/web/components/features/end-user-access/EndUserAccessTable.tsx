'use client'

// src/web/components/features/end-user-access/EndUserAccessTable.tsx
// Project 级终端用户访问控制表格（EndUser v2）

import { useState } from 'react'
import { Plus, MoreHorizontal, RefreshCw, ShieldOff } from 'lucide-react'
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
import { GrantEndUserAccessDialog } from './GrantEndUserAccessDialog'
import {
  useProjectEndUserAccess,
  type EndUserProjectAccessEntry,
} from '@web/hooks/end-user-access/useProjectEndUserAccess'

const PERMISSION_BUNDLE_LABELS: Record<string, string> = {
  viewer: '查看者',
  editor: '编辑者',
  admin: '管理员',
}

function formatDate(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleDateString('zh-CN')
  } catch {
    return '-'
  }
}

interface EndUserAccessTableProps {
  orgName: string
  projectSlug: string
}

export function EndUserAccessTable({ orgName, projectSlug }: EndUserAccessTableProps) {
  const { accesses, isLoading, error, reload, grantAccess, revokeAccess, updatePermissionBundle } =
    useProjectEndUserAccess(orgName, projectSlug)

  const [grantOpen, setGrantOpen] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)
  const [updatingUserId, setUpdatingUserId] = useState<string | null>(null)

  const existingUserIds = accesses.map((a) => a.userId)

  const handleRevoke = async (entry: EndUserProjectAccessEntry) => {
    if (!confirm(`确认撤销 ${entry.username} 的项目访问权限？`)) return
    setActionError(null)
    try {
      await revokeAccess(entry.accessId)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '撤销失败')
    }
  }

  const handleUpdateBundle = async (accessId: string, bundle: string) => {
    setUpdatingUserId(accessId)
    setActionError(null)
    try {
      await updatePermissionBundle(accessId, bundle)
    } catch (e: unknown) {
      setActionError(e instanceof Error ? e.message : '更新权限失败')
    } finally {
      setUpdatingUserId(null)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button size="sm" onClick={() => setGrantOpen(true)}>
            <Plus className="mr-1.5 size-4" />
            授权用户
          </Button>
          <Button size="sm" variant="outline" onClick={reload} disabled={isLoading}>
            <RefreshCw className={`mr-1.5 size-4 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>
        {actionError && <p className="text-sm text-destructive">{actionError}</p>}
      </div>

      {/* Error State */}
      {error && !isLoading && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <ShieldOff className="mb-3 size-10 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button size="sm" variant="outline" className="mt-3" onClick={reload}>
            重试
          </Button>
        </div>
      )}

      {/* Table */}
      {!error && (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="font-semibold">用户名</TableHead>
                <TableHead className="font-semibold">显示名称</TableHead>
                <TableHead className="font-semibold">权限包</TableHead>
                <TableHead className="font-semibold">授权时间</TableHead>
                <TableHead className="w-12" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading &&
                Array.from({ length: 4 }).map((_, i) => (
                  <TableRow key={i}>
                    {Array.from({ length: 5 }).map((__, j) => (
                      <TableCell key={j}>
                        <div className="h-4 w-24 animate-pulse rounded bg-slate-200" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))}

              {!isLoading && accesses.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <ShieldOff className="mb-3 size-10 text-muted-foreground/50" />
                      <p className="text-sm text-muted-foreground">暂无已授权的终端用户</p>
                      <Button
                        size="sm"
                        variant="outline"
                        className="mt-3"
                        onClick={() => setGrantOpen(true)}
                      >
                        授权第一个用户
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              )}

              {!isLoading &&
                accesses.map((entry) => (
                  <TableRow key={entry.userId}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex size-8 flex-shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
                          {entry.username.charAt(0).toUpperCase()}
                        </div>
                        <span className="text-sm font-semibold">{entry.username}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {entry.displayName || '-'}
                    </TableCell>
                    <TableCell>
                      {entry.permissionBundle ? (
                        <Badge variant="secondary">
                          {PERMISSION_BUNDLE_LABELS[entry.permissionBundle] ??
                            entry.permissionBundle}
                        </Badge>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(entry.grantedAt)}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="size-8"
                            disabled={updatingUserId === entry.accessId}
                          >
                            <MoreHorizontal className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => handleUpdateBundle(entry.accessId, 'viewer')}
                            disabled={entry.permissionBundle === 'viewer'}
                          >
                            设为查看者
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleUpdateBundle(entry.accessId, 'editor')}
                            disabled={entry.permissionBundle === 'editor'}
                          >
                            设为编辑者
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleUpdateBundle(entry.accessId, 'admin')}
                            disabled={entry.permissionBundle === 'admin'}
                          >
                            设为管理员
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => handleRevoke(entry)}
                          >
                            撤销授权
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

      {!isLoading && !error && accesses.length > 0 && (
        <p className="text-sm text-muted-foreground">共 {accesses.length} 名已授权终端用户</p>
      )}

      <GrantEndUserAccessDialog
        open={grantOpen}
        orgName={orgName}
        onClose={() => setGrantOpen(false)}
        onGrant={async (payload) => {
          await grantAccess(payload)
          setGrantOpen(false)
        }}
        existingUserIds={existingUserIds}
      />
    </div>
  )
}
