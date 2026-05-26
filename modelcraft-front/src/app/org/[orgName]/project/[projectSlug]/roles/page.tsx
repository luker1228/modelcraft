'use client'

import { useMemo, useState } from 'react'
import Link from 'next/link'
import { useParams, useSearchParams } from 'next/navigation'
import { Plus, Trash2, Search, ShieldOff, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { Skeleton } from '@web/components/ui/skeleton'
import { useRoleList } from '@/app/org/[orgName]/project/[projectSlug]/roles/_hooks/useRoleList'
import {
  BundlesTab,
  PermissionsTab,
} from '@web/components/features/rbac'
import { PageLayout, PageHeader } from '@web/components/features/layout'
import type { EndUserRole } from '@/types'

// ── RolesContent ─────────────────────────────────────────────────────

interface RolesContentProps {
  orgName: string
  projectSlug: string
}

function RolesContent({ orgName, projectSlug }: RolesContentProps) {
  const { roles, loading, createRole, deleteRole } = useRoleList({ orgName, projectSlug })

  const [search, setSearch] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDescription, setNewDescription] = useState('')
  const [creating, setCreating] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<EndUserRole | null>(null)
  const [deleting, setDeleting] = useState(false)

  const filteredRoles = useMemo(
    () => roles.filter(
      (r) =>
        r.name.toLowerCase().includes(search.toLowerCase()) ||
        (r.description ?? '').toLowerCase().includes(search.toLowerCase())
    ),
    [roles, search]
  )

  const handleCreate = async () => {
    if (!newName.trim()) { toast.error('请输入角色名称'); return }
    setCreating(true)
    const result = await createRole({ name: newName.trim(), description: newDescription.trim() || undefined })
    setCreating(false)
    if (result.success) {
      toast.success('角色已创建')
      setCreateOpen(false)
      setNewName('')
      setNewDescription('')
    } else {
      toast.error(result.errorMessage ?? '创建角色失败')
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    const result = await deleteRole(deleteTarget)
    setDeleting(false)
    if (result.success) {
      toast.success(`已删除角色「${deleteTarget.name}」`)
      setDeleteTarget(null)
    } else {
      toast.error(result.errorMessage ?? '删除失败')
    }
  }

  if (loading) {
    return (
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <Skeleton className="h-9 w-48" />
          <Skeleton className="h-9 w-24" />
        </div>
        <div className="overflow-hidden rounded-lg border border-border">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-12 w-full border-b last:border-0" />)}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="relative max-w-xs flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
          <Input
            placeholder="搜索角色..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button onClick={() => setCreateOpen(true)} size="sm" className="bg-primary text-primary-foreground hover:bg-primary/90">
          <Plus className="size-4" strokeWidth={1.5} />
          新建角色
        </Button>
      </div>

      {/* Table */}
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow className="border-b-2 border-border bg-card hover:bg-card">
              <TableHead className="h-10 w-[200px] text-[11px] font-medium uppercase tracking-wider text-foreground">角色名称</TableHead>
              <TableHead className="h-10 w-[260px] text-[11px] font-medium uppercase tracking-wider text-foreground">角色 ID</TableHead>
              <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">描述</TableHead>
              <TableHead className="h-10 w-[160px] text-[11px] font-medium uppercase tracking-wider text-foreground">修改时间</TableHead>
              <TableHead className="h-10 w-[100px] text-right text-[11px] font-medium uppercase tracking-wider text-foreground">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredRoles.map((role) => (
              <TableRow key={role.id} className="group border-b border-border last:border-0 hover:bg-foreground/[0.015]">
                <TableCell className="h-12 text-[13px]">
                  <div className="flex items-center gap-2">
                    <Link
                      href={`/org/${orgName}/project/${projectSlug}/roles/${role.id}`}
                      className="font-medium text-foreground underline-offset-2 hover:underline"
                    >
                      {role.name}
                    </Link>
                    {role.isImplicit && (
                      <Badge variant="secondary" className="text-xs">内置</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="h-12 text-[13px]">
                  <span className="font-mono text-[11px] text-muted-foreground">{role.id}</span>
                </TableCell>
                <TableCell className="h-12 text-[13px] text-muted-foreground">
                  {role.description || <span className="text-muted-foreground/40">—</span>}
                </TableCell>
                <TableCell className="h-12 text-[13px] text-muted-foreground">
                  {new Date(role.updatedAt).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}
                </TableCell>
                <TableCell className="h-12 text-right">
                  {!role.isImplicit && (
                    <div className="flex items-center justify-end opacity-0 transition-opacity group-hover:opacity-100">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTarget(role)}
                      >
                        <Trash2 className="size-3.5" strokeWidth={1.5} />
                        删除
                      </Button>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        {filteredRoles.length === 0 && (
          <div className="flex flex-col items-center justify-center py-14 text-center">
            <ShieldOff className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
            <p className="text-sm font-semibold text-foreground">
              {search ? '未找到匹配的角色' : '暂无角色'}
            </p>
            {!search && (
              <p className="mt-1 text-xs text-muted-foreground">
                点击「新建角色」开始配置权限
              </p>
            )}
            {!search && (
              <Button variant="outline" size="sm" className="mt-4" onClick={() => setCreateOpen(true)}>
                <Plus className="size-4" strokeWidth={1.5} />
                新建角色
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>新建角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">
                角色名称 <span className="text-destructive">*</span>
              </label>
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="例如：Editor"
                onKeyDown={(e) => e.key === 'Enter' && !creating && void handleCreate()}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">描述</label>
              <Textarea
                value={newDescription}
                onChange={(e) => setNewDescription(e.target.value)}
                placeholder="可选，描述该角色的用途"
                rows={2}
                className="resize-none"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
            <Button onClick={handleCreate} disabled={creating} className="bg-primary text-primary-foreground hover:bg-primary/90">
              {creating ? <><Loader2 className="mr-2 size-4 animate-spin" />创建中...</> : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Alert */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除角色</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除角色「{deleteTarget?.name}」吗？此操作不可撤销，该角色的所有用户将失去对应权限。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={handleDeleteConfirm}
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

// ── RolesPage ─────────────────────────────────────────────────────────────────

export default function RolesPage() {
  const params = useParams()
  const searchParams = useSearchParams()

  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const rawTab = searchParams.get('tab')
  const activeTab = rawTab === 'bundles' ? 'bundles' : rawTab === 'permissions' ? 'permissions' : 'roles'

  const tabTitle = activeTab === 'bundles' ? '权限包' : activeTab === 'permissions' ? '权限点' : '角色'

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title={tabTitle} />

      <div className="mt-6">
        {activeTab === 'roles' && <RolesContent orgName={orgName} projectSlug={projectSlug} />}
        {activeTab === 'bundles' && <BundlesTab orgName={orgName} projectSlug={projectSlug} />}
        {activeTab === 'permissions' && <PermissionsTab orgName={orgName} projectSlug={projectSlug} />}
      </div>
    </PageLayout>
  )
}
