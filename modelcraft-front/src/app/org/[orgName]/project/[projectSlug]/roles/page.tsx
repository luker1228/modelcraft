'use client'

import { useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Shield, Users, Plus, Trash2, Search, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@web/components/ui/input'
import { Badge } from '@web/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@web/components/ui/sheet'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@web/components/ui/select'
import { usePermission } from '@web/hooks/auth/use-permission'
import type { Role } from '@/types'

// ── Types ────────────────────────────────────────────────────────────

type RLSPreset = 'READ_WRITE_OWNER' | 'READ_ALL' | 'READ_WRITE_ALL' | 'NO_ACCESS'

interface MockUser {
  id: string
  name: string
  email: string
}

interface RoleRelationBinding {
  id: string
  tableName: string
  preset: RLSPreset
}

interface RoleMock extends Role {
  userIds: string[]
  relationBindings: RoleRelationBinding[]
}

// ── Constants ────────────────────────────────────────────────────────

const MOCK_USERS: MockUser[] = [
  { id: 'u-1', name: 'Alice', email: 'alice@demo.io' },
  { id: 'u-2', name: 'Bob', email: 'bob@demo.io' },
  { id: 'u-3', name: 'Carol', email: 'carol@demo.io' },
  { id: 'u-4', name: 'David', email: 'david@demo.io' },
]

const MOCK_TABLES = ['orders', 'customers', 'invoices', 'payments', 'tickets']

const RLS_PRESET_LABEL: Record<RLSPreset, string> = {
  READ_WRITE_OWNER: '仅所有者可读写',
  READ_ALL: '全员只读',
  READ_WRITE_ALL: '全员读写',
  NO_ACCESS: '禁止访问',
}

function nowISO(): string {
  return new Date().toISOString()
}

// ── Page ─────────────────────────────────────────────────────────────

export default function RolesPage() {
  const params = useParams()
  const router = useRouter()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string
  const canManageRoles = usePermission('*')

  const [roles, setRoles] = useState<RoleMock[]>([
    {
      id: 'r-admin',
      name: 'Admin',
      description: '系统管理员，拥有全部权限',
      permissions: [],
      isSystem: true,
      createdAt: nowISO(),
      updatedAt: nowISO(),
      userIds: ['u-1'],
      relationBindings: [
        { id: 'rel-1', tableName: 'orders', preset: 'READ_WRITE_OWNER' },
        { id: 'rel-2', tableName: 'customers', preset: 'READ_ALL' },
      ],
    },
    {
      id: 'r-ops',
      name: 'Ops',
      description: '运营角色，只读权限',
      permissions: [],
      isSystem: false,
      createdAt: nowISO(),
      updatedAt: nowISO(),
      userIds: ['u-2', 'u-3'],
      relationBindings: [{ id: 'rel-3', tableName: 'tickets', preset: 'READ_WRITE_OWNER' }],
    },
  ])

  const [search, setSearch] = useState('')

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDescription, setNewDescription] = useState('')

  // Users sheet
  const [sheetRole, setSheetRole] = useState<RoleMock | null>(null)
  const [userToAssign, setUserToAssign] = useState('')
  const [tableToBind, setTableToBind] = useState('')
  const [presetToBind, setPresetToBind] = useState<RLSPreset>('READ_WRITE_OWNER')

  // ── Derived ───────────────────────────────────────────────────────

  const filteredRoles = useMemo(
    () =>
      roles.filter(
        (r) =>
          r.name.toLowerCase().includes(search.toLowerCase()) ||
          (r.description ?? '').toLowerCase().includes(search.toLowerCase())
      ),
    [roles, search]
  )

  const liveSheetRole = useMemo(
    () => (sheetRole ? roles.find((r) => r.id === sheetRole.id) ?? null : null),
    [roles, sheetRole]
  )

  const sheetUsers = useMemo(
    () => (liveSheetRole ? MOCK_USERS.filter((u) => liveSheetRole.userIds.includes(u.id)) : []),
    [liveSheetRole]
  )

  const assignableUsers = useMemo(
    () => (liveSheetRole ? MOCK_USERS.filter((u) => !liveSheetRole.userIds.includes(u.id)) : []),
    [liveSheetRole]
  )

  // ── Handlers ─────────────────────────────────────────────────────

  const handleCreate = () => {
    if (!newName.trim()) { toast.error('请输入角色名称'); return }
    if (roles.some((r) => r.name.toLowerCase() === newName.trim().toLowerCase())) {
      toast.error('角色名称已存在'); return
    }
    const id = `r-${Date.now()}`
    setRoles((prev) => [
      {
        id,
        name: newName.trim(),
        description: newDescription.trim() || undefined,
        permissions: [],
        isSystem: false,
        createdAt: nowISO(),
        updatedAt: nowISO(),
        userIds: [],
        relationBindings: [],
      },
      ...prev,
    ])
    toast.success('角色已创建')
    setCreateOpen(false)
    setNewName('')
    setNewDescription('')
  }

  const handleDelete = (role: RoleMock) => {
    if (!confirm(`确定删除角色 "${role.name}" 吗？`)) return
    setRoles((prev) => prev.filter((r) => r.id !== role.id))
    toast.success(`已删除角色 ${role.name}`)
  }

  const assignUser = () => {
    if (!liveSheetRole || !userToAssign) return
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? { ...r, userIds: [...r.userIds, userToAssign], updatedAt: nowISO() }
          : r
      )
    )
    setUserToAssign('')
    toast.success('用户已分配')
  }

  const removeUser = (userId: string) => {
    if (!liveSheetRole) return
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? { ...r, userIds: r.userIds.filter((id) => id !== userId), updatedAt: nowISO() }
          : r
      )
    )
  }

  const addRelationBinding = () => {
    if (!liveSheetRole || !tableToBind) { toast.error('请选择表'); return }
    if (liveSheetRole.relationBindings.some((b) => b.tableName === tableToBind)) {
      toast.error('该表已存在 RLS 关系'); return
    }
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? {
              ...r,
              relationBindings: [
                ...r.relationBindings,
                { id: `rel-${Date.now()}`, tableName: tableToBind, preset: presetToBind },
              ],
              updatedAt: nowISO(),
            }
          : r
      )
    )
    setTableToBind('')
    toast.success('已添加 RLS 关系')
  }

  const removeRelationBinding = (bindingId: string) => {
    if (!liveSheetRole) return
    setRoles((prev) =>
      prev.map((r) =>
        r.id === liveSheetRole.id
          ? { ...r, relationBindings: r.relationBindings.filter((b) => b.id !== bindingId), updatedAt: nowISO() }
          : r
      )
    )
  }

  // ── Render ───────────────────────────────────────────────────────

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      {/* Header */}
      <section className="rounded-lg border border-border bg-background p-6 shadow-sm">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-md bg-blue-100 p-2 text-blue-600">
              <Shield className="size-5" strokeWidth={1.5} />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-foreground">角色管理</h1>
              <p className="text-sm text-muted-foreground">管理项目角色，配置用户与权限</p>
            </div>
          </div>
          <button
            onClick={() => setCreateOpen(true)}
            disabled={!canManageRoles}
            className="inline-flex h-9 items-center gap-2 rounded-md bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-40"
          >
            <Plus className="size-4" strokeWidth={1.5} />
            新建角色
          </button>
        </div>

        <div className="relative max-w-xs">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
          <Input
            placeholder="搜索角色..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 border border-slate-200 bg-white pl-9 text-sm focus:border-blue-600 focus:ring-1 focus:ring-blue-100"
          />
        </div>
      </section>

      {/* Table */}
      <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-sm">
            <thead className="border-b border-slate-200 bg-slate-50">
              <tr>
                <th className="h-10 px-4 py-3 text-left font-semibold text-foreground">角色名称</th>
                <th className="h-10 px-4 py-3 text-left font-semibold text-foreground">描述</th>
                <th className="h-10 px-4 py-3 text-right font-semibold text-foreground">操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredRoles.map((role, idx) => (
                <tr
                  key={role.id}
                  className={`border-b border-slate-200 last:border-0 transition-colors hover:bg-slate-50 ${idx % 2 !== 0 ? 'bg-muted/20' : 'bg-background'}`}
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">{role.name}</span>
                      {role.isSystem && (
                        <Badge variant="secondary" className="text-xs">系统</Badge>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {role.description || <span className="text-slate-300">—</span>}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <button
                        onClick={() => { setSheetRole(role); setUserToAssign('') }}
                        className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-slate-100 hover:text-foreground"
                      >
                        <Users className="size-4" strokeWidth={1.5} />
                        用户管理
                      </button>
                      <button
                        onClick={() =>
                          router.push(`/org/${orgName}/project/${projectSlug}/roles/${role.id}/permissions`)
                        }
                        className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-slate-100 hover:text-foreground"
                      >
                        <KeyRound className="size-4" strokeWidth={1.5} />
                        权限管理
                      </button>
                      <button
                        onClick={() => handleDelete(role)}
                        disabled={role.isSystem || !canManageRoles}
                        className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-red-50 hover:text-red-600 disabled:pointer-events-none disabled:opacity-30"
                      >
                        <Trash2 className="size-4" strokeWidth={1.5} />
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredRoles.length === 0 && (
          <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
            {search ? '未找到匹配的角色' : '暂无角色'}
          </div>
        )}
      </div>

      {/* ── Create Dialog ───────────────────────────────────────────── */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>新建角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-foreground">
                角色名称 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="例如：Editor"
                className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-foreground">描述</label>
              <textarea
                value={newDescription}
                onChange={(e) => setNewDescription(e.target.value)}
                placeholder="可选，描述该角色的用途"
                rows={2}
                className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
              />
            </div>
          </div>
          <DialogFooter>
            <button
              onClick={() => setCreateOpen(false)}
              className="h-9 rounded-md border border-slate-200 bg-white px-4 text-sm font-medium text-foreground hover:bg-slate-50"
            >
              取消
            </button>
            <button
              onClick={handleCreate}
              className="h-9 rounded-md bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700"
            >
              创建
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Users Sheet ─────────────────────────────────────────────── */}
      <Sheet open={!!sheetRole} onOpenChange={(open) => { if (!open) setSheetRole(null) }}>
        <SheetContent className="w-full overflow-y-auto sm:max-w-lg">
          {liveSheetRole && (
            <>
              <SheetHeader className="mb-6">
                <SheetTitle className="flex items-center gap-2">
                  <Users className="size-5" strokeWidth={1.5} />
                  {liveSheetRole.name} · 用户管理
                  {liveSheetRole.isSystem && <Badge variant="secondary">系统</Badge>}
                </SheetTitle>
                <SheetDescription>{liveSheetRole.description || '无描述'}</SheetDescription>
              </SheetHeader>

              {/* Assign */}
              <div className="mb-4 flex gap-2">
                <Select value={userToAssign} onValueChange={setUserToAssign}>
                  <SelectTrigger className="flex-1 text-sm">
                    <SelectValue placeholder="选择用户分配到此角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {assignableUsers.map((u) => (
                      <SelectItem key={u.id} value={u.id}>
                        {u.name} · {u.email}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <button
                  onClick={assignUser}
                  disabled={!userToAssign || !canManageRoles}
                  className="h-9 rounded-md bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-40"
                >
                  分配
                </button>
              </div>

              {/* User list */}
              <div className="mb-6 overflow-hidden rounded-md border border-slate-200">
                <table className="w-full border-collapse text-sm">
                  <thead className="border-b border-slate-200 bg-slate-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">用户</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">邮箱</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sheetUsers.map((user) => (
                      <tr key={user.id} className="border-b border-slate-100 last:border-0">
                        <td className="px-3 py-2.5 font-medium text-foreground">{user.name}</td>
                        <td className="px-3 py-2.5 text-muted-foreground">{user.email}</td>
                        <td className="px-3 py-2.5 text-right">
                          <button
                            onClick={() => removeUser(user.id)}
                            disabled={!canManageRoles}
                            className="inline-flex h-7 items-center gap-1 rounded px-2 text-xs text-muted-foreground hover:bg-red-50 hover:text-red-600 disabled:opacity-40"
                          >
                            移除
                          </button>
                        </td>
                      </tr>
                    ))}
                    {sheetUsers.length === 0 && (
                      <tr>
                        <td colSpan={3} className="px-3 py-6 text-center text-sm text-muted-foreground">
                          暂无绑定用户
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>

              {/* RLS */}
              <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                关系管理（RLS）
              </p>
              <div className="mb-3 flex gap-2">
                <Select value={tableToBind} onValueChange={setTableToBind}>
                  <SelectTrigger className="flex-1 text-sm">
                    <SelectValue placeholder="选择表" />
                  </SelectTrigger>
                  <SelectContent>
                    {MOCK_TABLES.map((t) => (
                      <SelectItem key={t} value={t}>{t}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Select value={presetToBind} onValueChange={(v) => setPresetToBind(v as RLSPreset)}>
                  <SelectTrigger className="w-40 text-sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {(Object.keys(RLS_PRESET_LABEL) as RLSPreset[]).map((p) => (
                      <SelectItem key={p} value={p}>{RLS_PRESET_LABEL[p]}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <button
                  onClick={addRelationBinding}
                  disabled={!canManageRoles}
                  className="h-9 rounded-md bg-blue-600 px-3 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-40"
                >
                  添加
                </button>
              </div>

              {liveSheetRole.relationBindings.length > 0 && (
                <div className="overflow-hidden rounded-md border border-slate-200">
                  <table className="w-full border-collapse text-sm">
                    <thead className="border-b border-slate-200 bg-slate-50">
                      <tr>
                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">表名</th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">策略</th>
                        <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {liveSheetRole.relationBindings.map((b) => (
                        <tr key={b.id} className="border-b border-slate-100 last:border-0">
                          <td className="px-3 py-2.5 font-mono text-xs text-foreground">{b.tableName}</td>
                          <td className="px-3 py-2.5">
                            <Badge variant="outline" className="text-xs">{RLS_PRESET_LABEL[b.preset]}</Badge>
                          </td>
                          <td className="px-3 py-2.5 text-right">
                            <button
                              onClick={() => removeRelationBinding(b.id)}
                              className="inline-flex h-7 items-center gap-1 rounded px-2 text-xs text-muted-foreground hover:bg-red-50 hover:text-red-600"
                            >
                              移除
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          )}
        </SheetContent>
      </Sheet>
    </div>
  )
}
