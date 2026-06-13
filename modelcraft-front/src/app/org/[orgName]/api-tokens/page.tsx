'use client'

import { useEffect, useState } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { useParams } from 'next/navigation'
import {
  CREATE_API_TOKEN,
  GET_API_TOKENS,
  REVOKE_API_TOKEN,
} from '@/api-client/user'
import { useOrgScopedContext } from '@api-client/apollo/context'
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
import { Button } from '@web/components/ui/button'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { AppLayout, PageHeader, PageLayout } from '@web/components/features/layout'
import { Copy, KeyRound, Loader2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import type { ApiToken } from '@/types'

interface ApiTokensQueryData {
  endUserAPITokens: ApiToken[]
}

interface PayloadError {
  __typename?: string
  message: string
  suggestion?: string | null
}

interface CreateApiTokenMutationResult {
  createEndUserAPIToken: {
    token?: ApiToken | null
    plaintext?: string | null
    error?: PayloadError | null
  }
}

interface RevokeApiTokenMutationResult {
  revokeEndUserAPIToken: {
    success?: boolean | null
    error?: PayloadError | null
  }
}

function formatDateTime(value?: string | null): string {
  if (!value) return '永不过期'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString()
}

function formatLastUsedAt(value?: string | null): string {
  if (!value) return '从未使用'
  return formatDateTime(value)
}

function toISOFromDatetimeLocal(value: string): string | null {
  if (!value) return null
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return date.toISOString()
}

export default function ApiTokensPage() {
  const params = useParams()
  const orgName = params.orgName as string
  const orgScopedContext = useOrgScopedContext(orgName)

  const [tokens, setTokens] = useState<ApiToken[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [revokeTarget, setRevokeTarget] = useState<ApiToken | null>(null)
  const [plainToken, setPlainToken] = useState<string | null>(null)
  const [plainTokenName, setPlainTokenName] = useState('')

  const { data, loading, error } = useQuery<ApiTokensQueryData>(GET_API_TOKENS, {
    skip: !orgName,
    context: orgScopedContext,
  })

  useEffect(() => {
    if (data?.endUserAPITokens) {
      setTokens(data.endUserAPITokens)
    }
  }, [data?.endUserAPITokens])

  const [createApiToken, { loading: creating }] = useMutation<CreateApiTokenMutationResult>(
    CREATE_API_TOKEN,
    { context: orgScopedContext }
  )
  const [revokeApiToken, { loading: revoking }] = useMutation<RevokeApiTokenMutationResult>(
    REVOKE_API_TOKEN,
    { context: orgScopedContext }
  )

  const resetCreateForm = () => {
    setName('')
    setExpiresAt('')
  }

  const handleCreate = async () => {
    const trimmedName = name.trim()
    if (!trimmedName) {
      toast.error('请输入 Token 名称')
      return
    }

    const result = await createApiToken({
      variables: {
        name: trimmedName,
        expiresAt: toISOFromDatetimeLocal(expiresAt),
      },
    })

    const payload = result.data?.createEndUserAPIToken
    if (payload?.error) {
      toast.error(payload.error.message)
      return
    }
    if (!payload?.token || !payload.plaintext) {
      toast.error('创建 API Token 失败')
      return
    }

    setTokens((prev) => [payload.token!, ...prev])
    setPlainToken(payload.plaintext)
    setPlainTokenName(payload.token.name)
    setCreateOpen(false)
    resetCreateForm()
    toast.success('API Token 已创建')
  }

  const handleRevoke = async () => {
    if (!revokeTarget) return

    const result = await revokeApiToken({ variables: { id: revokeTarget.id } })
    const payload = result.data?.revokeEndUserAPIToken
    if (payload?.error) {
      toast.error(payload.error.message)
      return
    }

    setTokens((prev) => prev.filter((token) => token.id !== revokeTarget.id))
    setRevokeTarget(null)
    toast.success('API Token 已撤销')
  }

  return (
    <AppLayout pageTitle="API Token">
      <PageLayout maxWidth="6xl" background="card" padding="compact">
        <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <PageHeader title="API Token" spacing="compact" />
          <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1.5">
            <Plus className="size-4" />
            新建 Token
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>过期时间</TableHead>
                <TableHead>最后使用</TableHead>
                <TableHead className="w-16" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-32 text-center text-sm text-muted-foreground">
                    <Loader2 className="mr-2 inline size-4 animate-spin" />
                    加载中...
                  </TableCell>
                </TableRow>
              ) : error ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-32 text-center text-sm text-destructive">
                    加载 API Token 失败
                  </TableCell>
                </TableRow>
              ) : tokens.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-40 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <p className="text-sm text-muted-foreground">暂无 API Token</p>
                      <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1.5">
                        <Plus className="size-4" />
                        新建 Token
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                tokens.map((token) => (
                  <TableRow key={token.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <KeyRound className="size-4 text-muted-foreground" />
                        <span className="font-medium">{token.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDateTime(token.createdAt)}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDateTime(token.expiresAt)}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatLastUsedAt(token.lastUsedAt)}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-8 text-muted-foreground hover:text-destructive"
                        onClick={() => setRevokeTarget(token)}
                      >
                        <Trash2 className="size-4" />
                        <span className="sr-only">撤销 {token.name}</span>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        <Dialog
          open={createOpen}
          onOpenChange={(open) => {
            setCreateOpen(open)
            if (!open) resetCreateForm()
          }}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>新建 API Token</DialogTitle>
              <DialogDescription>
                Token 明文只会在创建成功后展示一次。
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="api-token-name">名称</Label>
                <Input
                  id="api-token-name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="例如 CI/CD Token"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-token-expires-at">过期时间（可选）</Label>
                <Input
                  id="api-token-expires-at"
                  type="datetime-local"
                  value={expiresAt}
                  onChange={(event) => setExpiresAt(event.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating}>
                取消
              </Button>
              <Button onClick={handleCreate} disabled={creating || !name.trim()}>
                {creating && <Loader2 className="mr-2 size-4 animate-spin" />}
                创建
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog open={!!plainToken} onOpenChange={(open) => !open && setPlainToken(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>API Token 已创建</DialogTitle>
              <DialogDescription>
                这是唯一一次展示完整 Token，请保存到安全位置。
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label>{plainTokenName}</Label>
              <div className="break-all rounded-md border bg-muted px-3 py-2 font-mono text-xs">
                {plainToken}
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                className="gap-1.5"
                onClick={async () => {
                  if (!plainToken) return
                  await navigator.clipboard.writeText(plainToken)
                  toast.success('API Token 已复制')
                }}
              >
                <Copy className="size-4" />
                复制
              </Button>
              <Button onClick={() => setPlainToken(null)}>完成</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <AlertDialog open={!!revokeTarget} onOpenChange={(open) => !open && setRevokeTarget(null)}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>撤销 API Token？</AlertDialogTitle>
              <AlertDialogDescription>
                撤销后该 Token 将立即失效，且不可恢复。确定撤销 {revokeTarget?.name ?? '-'} 吗？
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={revoking}>取消</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleRevoke}
                disabled={revoking}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                {revoking ? '撤销中...' : '撤销'}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </PageLayout>
    </AppLayout>
  )
}
