'use client'

import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { useParams } from 'next/navigation'
import {
  GET_API_KEYS,
  GET_ROLES,
  CREATE_API_KEY,
  UPDATE_API_KEY,
  REVOKE_API_KEY,
} from '@web/graphql'
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
import { Badge } from '@web/components/ui/badge'
import { Button } from '@web/components/ui/button'
import { Checkbox } from '@web/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { ScrollArea } from '@web/components/ui/scroll-area'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { KeyRound, Plus, Shield, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import type { ApiKey, Role } from '@/types'

interface ApiKeysQueryData {
  apiKeys: ApiKey[]
}

interface RolesQueryData {
  roles: Role[]
}

interface PayloadError {
  __typename?: string
  message: string
  suggestion?: string | null
}

interface CreateApiKeyMutationResult {
  createApiKey: {
    result?: {
      id: string
      name: string
      key: string
      keyPrefix: string
      roleIDs: string[]
      createdAt: string
    } | null
    error?: PayloadError | null
  }
}

interface UpdateApiKeyMutationResult {
  updateApiKey: {
    apiKey?: ApiKey | null
    error?: PayloadError | null
  }
}

interface RevokeApiKeyMutationResult {
  revokeApiKey: {
    apiKey?: Pick<ApiKey, 'id' | 'revokedAt'> | null
    error?: PayloadError | null
  }
}

interface ApiKeyFormState {
  name: string
  expiresAt: string
  roleIds: string[]
}

function formatDateTime(dateStr?: string | null): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString()
}

function toDatetimeLocalValue(dateStr?: string | null): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

function toISOFromDatetimeLocal(value: string): string | null {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return date.toISOString()
}

function includesRoleInputError(messages: string[]): boolean {
  return messages.some((message) =>
    (message.toLowerCase().includes('roleids') || message.toLowerCase().includes('roles')) &&
    (message.includes('CreateApiKeyInput') || message.includes('UpdateApiKeyInput'))
  )
}

function extractErrorMessages(error: unknown): string[] {
  if (!error || typeof error !== 'object') return []
  const maybeGraphQLErrors = (error as { graphQLErrors?: Array<{ message?: string }> }).graphQLErrors
  if (!Array.isArray(maybeGraphQLErrors)) return []
  return maybeGraphQLErrors.map((item) => item?.message ?? '').filter(Boolean)
}

const EMPTY_FORM: ApiKeyFormState = {
  name: '',
  expiresAt: '',
  roleIds: [],
}

export default function ApiKeysPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const orgScopedContext = useMemo(() => ({ uri: `/graphql/org/${orgName}/` }), [orgName])

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [createForm, setCreateForm] = useState<ApiKeyFormState>(EMPTY_FORM)
  const [editTarget, setEditTarget] = useState<ApiKey | null>(null)
  const [editForm, setEditForm] = useState<ApiKeyFormState>(EMPTY_FORM)
  const [revokeTarget, setRevokeTarget] = useState<ApiKey | null>(null)
  const [newPlainKey, setNewPlainKey] = useState<string | null>(null)
  const [newKeyName, setNewKeyName] = useState<string>('')

  const { data, loading, error } = useQuery<ApiKeysQueryData>(GET_API_KEYS, {
    skip: !orgName,
    context: orgScopedContext,
  })
  const { data: rolesData, loading: rolesLoading } = useQuery<RolesQueryData>(GET_ROLES, {
    skip: !orgName,
    context: orgScopedContext,
  })

  const [apiKeysLocal, setApiKeysLocal] = useState<ApiKey[]>([])

  useEffect(() => {
    if (data?.apiKeys) {
      setApiKeysLocal(data.apiKeys)
    }
  }, [data?.apiKeys])

  const [createApiKey, { loading: creating }] = useMutation<CreateApiKeyMutationResult>(CREATE_API_KEY, {
    context: orgScopedContext,
  })
  const [updateApiKey, { loading: updating }] = useMutation<UpdateApiKeyMutationResult>(UPDATE_API_KEY, {
    context: orgScopedContext,
  })
  const [revokeApiKey, { loading: revoking }] = useMutation<RevokeApiKeyMutationResult>(REVOKE_API_KEY, {
    context: orgScopedContext,
  })

  const apiKeys = apiKeysLocal
  const roles = useMemo(() => rolesData?.roles ?? [], [rolesData?.roles])
  const roleNameMap = useMemo(
    () => new Map(roles.map((role) => [role.id, role.name])),
    [roles]
  )

  const runCreate = async (input: Record<string, unknown>, withRoleInput: boolean) => {
    try {
      let result = await createApiKey({ variables: { input } })
      const messages = (result.errors ?? []).map((err) => err.message)
      if (withRoleInput && includesRoleInputError(messages)) {
        result = await createApiKey({
          variables: {
            input: {
              name: input.name,
              expiresAt: input.expiresAt ?? null,
            },
          },
        })
        toast.warning('当前后端版本尚未支持 API Key 角色绑定，已按基础模式创建')
      }
      return result
    } catch (err) {
      const messages = extractErrorMessages(err)
      if (withRoleInput && includesRoleInputError(messages)) {
        const result = await createApiKey({
          variables: {
            input: {
              name: input.name,
              expiresAt: input.expiresAt ?? null,
            },
          },
        })
        toast.warning('当前后端版本尚未支持 API Key 角色绑定，已按基础模式创建')
        return result
      }
      throw err
    }
  }

  const runUpdate = async (id: string, input: Record<string, unknown>, withRoleInput: boolean) => {
    try {
      let result = await updateApiKey({ variables: { id, input } })
      const messages = (result.errors ?? []).map((err) => err.message)
      if (withRoleInput && includesRoleInputError(messages)) {
        result = await updateApiKey({
          variables: {
            id,
            input: {
              name: input.name,
              expiresAt: input.expiresAt ?? null,
            },
          },
        })
        toast.warning('当前后端版本尚未支持 API Key 角色绑定，已按基础模式更新')
      }
      return result
    } catch (err) {
      const messages = extractErrorMessages(err)
      if (withRoleInput && includesRoleInputError(messages)) {
        const result = await updateApiKey({
          variables: {
            id,
            input: {
              name: input.name,
              expiresAt: input.expiresAt ?? null,
            },
          },
        })
        toast.warning('当前后端版本尚未支持 API Key 角色绑定，已按基础模式更新')
        return result
      }
      throw err
    }
  }

  const handleCreate = async () => {
    if (!createForm.name.trim()) {
      toast.error('请输入 API Key 名称')
      return
    }

    const expiresAt = toISOFromDatetimeLocal(createForm.expiresAt)
    const input: Record<string, unknown> = {
      name: createForm.name.trim(),
      expiresAt,
    }
    if (createForm.roleIds.length > 0) {
      input.roleIDs = createForm.roleIds
    }

    try {
      const result = await runCreate(input, createForm.roleIds.length > 0)
      if (result.data?.createApiKey.error) {
        toast.error(result.data.createApiKey.error.message)
        return
      }
      const created = result.data?.createApiKey.result
      if (!created) {
        toast.error('创建 API Key 失败')
        return
      }
      setApiKeysLocal((prev) => ([
        {
          id: created.id,
          name: created.name,
          keyPrefix: created.keyPrefix,
          roleIDs: created.roleIDs,
          createdAt: created.createdAt,
          expiresAt,
          lastUsedAt: null,
          revokedAt: null,
        },
        ...prev,
      ]))
      setNewPlainKey(created.key)
      setNewKeyName(created.name)
      setCreateForm(EMPTY_FORM)
      setShowCreateDialog(false)
      toast.success('API Key 创建成功')
    } catch (err) {
      const message = err instanceof Error ? err.message : '创建 API Key 失败'
      toast.error(message)
    }
  }

  const handleOpenEdit = (apiKey: ApiKey) => {
    setEditTarget(apiKey)
    setEditForm({
      name: apiKey.name,
      expiresAt: toDatetimeLocalValue(apiKey.expiresAt),
      roleIds: apiKey.roleIDs ?? [],
    })
  }

  const handleUpdate = async () => {
    if (!editTarget) return
    if (!editForm.name.trim()) {
      toast.error('请输入 API Key 名称')
      return
    }

    const expiresAt = toISOFromDatetimeLocal(editForm.expiresAt)
    const input: Record<string, unknown> = {
      name: editForm.name.trim(),
      expiresAt,
    }
    if (editForm.roleIds.length > 0) {
      input.roleIDs = editForm.roleIds
    }

    try {
      const result = await runUpdate(editTarget.id, input, editForm.roleIds.length > 0)
      if (result.data?.updateApiKey.error) {
        toast.error(result.data.updateApiKey.error.message)
        return
      }
      const updated = result.data?.updateApiKey.apiKey
      if (updated) {
        setApiKeysLocal((prev) => prev.map((key) => (key.id === updated.id ? updated : key)))
      } else {
        setApiKeysLocal((prev) => prev.map((key) => (
          key.id === editTarget.id
            ? { ...key, name: editForm.name.trim(), roleIDs: editForm.roleIds, expiresAt }
            : key
        )))
      }
      setEditTarget(null)
      toast.success('API Key 更新成功')
    } catch (err) {
      const message = err instanceof Error ? err.message : '更新 API Key 失败'
      toast.error(message)
    }
  }

  const handleRevoke = async () => {
    if (!revokeTarget) return
    try {
      const result = await revokeApiKey({ variables: { id: revokeTarget.id } })
      if (result.data?.revokeApiKey.error) {
        toast.error(result.data.revokeApiKey.error.message)
        return
      }
      const revokedAt = result.data?.revokeApiKey.apiKey?.revokedAt ?? new Date().toISOString()
      setApiKeysLocal((prev) => prev.map((key) => (
        key.id === revokeTarget.id ? { ...key, revokedAt } : key
      )))
      toast.success('API Key 已撤销')
      setRevokeTarget(null)
    } catch (err) {
      const message = err instanceof Error ? err.message : '撤销 API Key 失败'
      toast.error(message)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">API Keys</h2>
          <p className="text-sm text-muted-foreground">
            管理组织级 API Key，并为每个 Key 设置可用角色。
          </p>
        </div>
        <Button size="sm" onClick={() => setShowCreateDialog(true)}>
          <Plus className="mr-1 size-4" />
          Create API Key
        </Button>
      </div>

      {error && !loading && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <p className="text-sm text-destructive">Failed to load API keys: {error.message}</p>
        </div>
      )}

      {!error && (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Prefix</TableHead>
                <TableHead>Roles</TableHead>
                <TableHead>Expires At</TableHead>
                <TableHead>Last Used</TableHead>
                <TableHead>Created At</TableHead>
                <TableHead className="w-[140px]">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading && (
                <TableRow>
                  <TableCell colSpan={7} className="py-8 text-center text-sm text-muted-foreground">
                    加载中...
                  </TableCell>
                </TableRow>
              )}
              {!loading && apiKeys.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                      <KeyRound className="size-8 text-muted-foreground/50" />
                      <p className="text-sm">暂无 API Key</p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
              {!loading && apiKeys.map((apiKey) => {
                const assignedRoleIds = apiKey.roleIDs ?? []
                const isRevoked = !!apiKey.revokedAt
                return (
                  <TableRow key={apiKey.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <span>{apiKey.name}</span>
                        {isRevoked && (
                          <Badge variant="outline" className="text-xs text-muted-foreground">
                            Revoked
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{apiKey.keyPrefix}</code>
                    </TableCell>
                    <TableCell>
                      {assignedRoleIds.length === 0 ? (
                        <span className="text-sm text-muted-foreground">-</span>
                      ) : (
                        <div className="flex flex-wrap gap-1">
                          {assignedRoleIds.map((roleId) => (
                            <Badge key={roleId} variant="secondary" className="text-xs">
                              {roleNameMap.get(roleId) ?? roleId}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>{formatDateTime(apiKey.expiresAt)}</TableCell>
                    <TableCell>{formatDateTime(apiKey.lastUsedAt)}</TableCell>
                    <TableCell>{formatDateTime(apiKey.createdAt)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleOpenEdit(apiKey)}
                          disabled={isRevoked}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={() => setRevokeTarget(apiKey)}
                          disabled={isRevoked}
                        >
                          <Trash2 className="mr-1 size-4" />
                          Revoke
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}

      <Dialog
        open={showCreateDialog}
        onOpenChange={(open) => {
          setShowCreateDialog(open)
          if (!open) setCreateForm(EMPTY_FORM)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create API Key</DialogTitle>
            <DialogDescription>创建新 Key，并可绑定可用角色范围。</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="api-key-name">Name</Label>
              <Input
                id="api-key-name"
                placeholder="e.g. CI/CD Token"
                value={createForm.name}
                onChange={(e) => setCreateForm((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="api-key-expires-at">Expires At (Optional)</Label>
              <Input
                id="api-key-expires-at"
                type="datetime-local"
                value={createForm.expiresAt}
                onChange={(e) => setCreateForm((prev) => ({ ...prev, expiresAt: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label className="inline-flex items-center gap-1.5">
                <Shield className="size-4" />
                Roles
              </Label>
              <ScrollArea className="h-40 rounded-md border p-3">
                <div className="space-y-2">
                  {rolesLoading && (
                    <p className="text-sm text-muted-foreground">加载角色中...</p>
                  )}
                  {!rolesLoading && roles.length === 0 && (
                    <p className="text-sm text-muted-foreground">暂无可分配角色</p>
                  )}
                  {!rolesLoading && roles.map((role) => (
                    <div key={role.id} className="flex items-center gap-2">
                      <Checkbox
                        id={`create-role-${role.id}`}
                        checked={createForm.roleIds.includes(role.id)}
                        onCheckedChange={(checked) => {
                          setCreateForm((prev) => {
                            const hasRole = prev.roleIds.includes(role.id)
                            if (checked && !hasRole) {
                              return { ...prev, roleIds: [...prev.roleIds, role.id] }
                            }
                            if (!checked && hasRole) {
                              return { ...prev, roleIds: prev.roleIds.filter((id) => id !== role.id) }
                            }
                            return prev
                          })
                        }}
                      />
                      <Label htmlFor={`create-role-${role.id}`} className="font-normal">
                        {role.name}
                      </Label>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={creating}>
              {creating ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!editTarget}
        onOpenChange={(open) => {
          if (!open) {
            setEditTarget(null)
            setEditForm(EMPTY_FORM)
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit API Key</DialogTitle>
            <DialogDescription>更新名称、过期时间与角色设置。</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-api-key-name">Name</Label>
              <Input
                id="edit-api-key-name"
                value={editForm.name}
                onChange={(e) => setEditForm((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-api-key-expires-at">Expires At (Optional)</Label>
              <Input
                id="edit-api-key-expires-at"
                type="datetime-local"
                value={editForm.expiresAt}
                onChange={(e) => setEditForm((prev) => ({ ...prev, expiresAt: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label className="inline-flex items-center gap-1.5">
                <Shield className="size-4" />
                Roles
              </Label>
              <ScrollArea className="h-40 rounded-md border p-3">
                <div className="space-y-2">
                  {rolesLoading && (
                    <p className="text-sm text-muted-foreground">加载角色中...</p>
                  )}
                  {!rolesLoading && roles.length === 0 && (
                    <p className="text-sm text-muted-foreground">暂无可分配角色</p>
                  )}
                  {!rolesLoading && roles.map((role) => (
                    <div key={role.id} className="flex items-center gap-2">
                      <Checkbox
                        id={`edit-role-${role.id}`}
                        checked={editForm.roleIds.includes(role.id)}
                        onCheckedChange={(checked) => {
                          setEditForm((prev) => {
                            const hasRole = prev.roleIds.includes(role.id)
                            if (checked && !hasRole) {
                              return { ...prev, roleIds: [...prev.roleIds, role.id] }
                            }
                            if (!checked && hasRole) {
                              return { ...prev, roleIds: prev.roleIds.filter((id) => id !== role.id) }
                            }
                            return prev
                          })
                        }}
                      />
                      <Label htmlFor={`edit-role-${role.id}`} className="font-normal">
                        {role.name}
                      </Label>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTarget(null)} disabled={updating}>
              Cancel
            </Button>
            <Button onClick={handleUpdate} disabled={updating}>
              {updating ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!revokeTarget} onOpenChange={(open) => !open && setRevokeTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke API Key</AlertDialogTitle>
            <AlertDialogDescription>
              撤销后该 Key 将立即失效，且不可恢复。确定撤销 &quot;{revokeTarget?.name ?? '-'}&quot; 吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={revoking}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRevoke}
              disabled={revoking}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {revoking ? 'Revoking...' : 'Revoke'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Dialog open={!!newPlainKey} onOpenChange={(open) => !open && setNewPlainKey(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Key Created</DialogTitle>
            <DialogDescription>
              这是你唯一一次看到完整 Key。请立即保存到安全位置。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>{newKeyName}</Label>
            <div className="break-all rounded-md border bg-muted px-3 py-2 font-mono text-xs">
              {newPlainKey}
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={async () => {
                if (!newPlainKey) return
                await navigator.clipboard.writeText(newPlainKey)
                toast.success('API Key 已复制')
              }}
            >
              Copy
            </Button>
            <Button onClick={() => setNewPlainKey(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
