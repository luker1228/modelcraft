'use client'

import { useState, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { ApolloProvider, useQuery, useMutation } from '@apollo/client'
import { KeyRound, Plus, Trash2, Copy, Check, Eye, EyeOff, BookOpen } from 'lucide-react'
import { copyToClipboardWithCallback } from '@/shared/utils/clipboard'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
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
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { EndUserAppLayout } from '@web/components/features/layout/EndUserAppLayout'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import {
  END_USER_API_TOKENS,
  CREATE_END_USER_API_TOKEN,
  REVOKE_END_USER_API_TOKEN,
} from '@api-client/end-user/graphql-docs'

// ── Types ───────────────────────────────────────────────────────────────────

interface APIToken {
  id: string
  name: string
  createdAt: string
  expiresAt?: string | null
  lastUsedAt?: string | null
}

interface EndUserAPITokensData {
  endUserAPITokens: APIToken[]
}

interface CreateAPITokenData {
  createEndUserAPIToken: {
    token?: APIToken | null
    plaintext?: string | null
    error?: {
      __typename?: string
      message?: string
      limit?: number
    } | null
  }
}

interface RevokeAPITokenData {
  revokeEndUserAPIToken: {
    success?: boolean | null
    error?: {
      __typename?: string
      message?: string
    } | null
  }
}

// ── Helpers ─────────────────────────────────────────────────────────────────

function formatDate(iso?: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function expiresInToDate(value: string): string | undefined {
  if (value === 'never') return undefined
  const days = parseInt(value, 10)
  const d = new Date()
  d.setDate(d.getDate() + days)
  return d.toISOString()
}

// ── Copy Button ──────────────────────────────────────────────────────────────

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = () => {
    copyToClipboardWithCallback(value, () => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }
  return (
    <Button
      variant="outline"
      size="sm"
      className="h-7 gap-1.5 px-2.5 text-xs"
      onClick={handleCopy}
    >
      {copied ? (
        <Check className="size-3.5 text-emerald-500" />
      ) : (
        <Copy className="size-3.5" />
      )}
      {copied ? '已复制' : '复制'}
    </Button>
  )
}

// ── Plaintext Token Dialog ────────────────────────────────────────────────────

function PlaintextDialog({
  plaintext,
  open,
  onClose,
}: {
  plaintext: string
  open: boolean
  onClose: () => void
}) {
  const [visible, setVisible] = useState(false)

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="size-4 text-primary" />
            Token 创建成功
          </DialogTitle>
          <DialogDescription>
            请立即复制并保存此 Token。关闭后将无法再次查看。
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="rounded-md border bg-muted/40 p-3">
            <div className="flex items-center justify-between gap-2">
              <code className="min-w-0 flex-1 break-all font-mono text-xs text-foreground">
                {visible ? plaintext : '•'.repeat(Math.min(plaintext.length, 48))}
              </code>
              <div className="flex shrink-0 items-center gap-1.5">
                <Button
                  variant="ghost"
                  size="sm"
                  className="size-7 p-0"
                  onClick={() => setVisible((v) => !v)}
                  title={visible ? '隐藏' : '显示'}
                >
                  {visible ? (
                    <Eye className="size-3.5" />
                  ) : (
                    <EyeOff className="size-3.5" />
                  )}
                </Button>
                <CopyButton value={plaintext} />
              </div>
            </div>
          </div>
          <p className="flex items-start gap-1.5 text-xs text-amber-600">
            <span className="mt-px shrink-0 select-none">⚠</span>
            此 Token 仅显示一次，关闭后无法恢复，请妥善保存。
          </p>
        </div>
        <DialogFooter>
          <Button onClick={onClose}>已保存，关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Create Token Dialog ──────────────────────────────────────────────────────

function CreateTokenDialog({
  open,
  onClose,
  onCreated,
}: {
  open: boolean
  onClose: () => void
  onCreated: (plaintext: string) => void
}) {
  const [name, setName] = useState('')
  const [expires, setExpires] = useState('never')

  const [createToken, { loading }] = useMutation<CreateAPITokenData>(CREATE_END_USER_API_TOKEN)

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error('请输入 Token 名称')
      return
    }

    const expiresAt = expiresInToDate(expires)
    const { data } = await createToken({
      variables: { name: name.trim(), expiresAt: expiresAt ?? null },
    })

    const result = data?.createEndUserAPIToken
    if (result?.error) {
      const err = result.error
      if (err.__typename === 'APITokenLimitReached') {
        toast.error(`已达上限（最多 ${err.limit ?? '?'} 个 Token）`)
      } else {
        toast.error(err.message ?? '创建失败')
      }
      return
    }

    if (result?.plaintext) {
      onClose()
      setName('')
      setExpires('never')
      onCreated(result.plaintext)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>创建 API Token</DialogTitle>
          <DialogDescription>Token 用于 CLI 及 API 认证，请妥善保管。</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="token-name">名称</Label>
            <Input
              id="token-name"
              placeholder="如：my-laptop"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') void handleSubmit() }}
            />
          </div>
          <div className="space-y-1.5">
            <Label>过期时间</Label>
            <Select value={expires} onValueChange={setExpires}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="30">30 天</SelectItem>
                <SelectItem value="90">90 天</SelectItem>
                <SelectItem value="365">1 年</SelectItem>
                <SelectItem value="never">永不过期</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>
            取消
          </Button>
          <Button onClick={() => void handleSubmit()} disabled={loading}>
            {loading ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ── Revoke Confirm Dialog ─────────────────────────────────────────────────────

function RevokeDialog({
  token,
  onClose,
  onRevoked,
}: {
  token: APIToken | null
  onClose: () => void
  onRevoked: () => void
}) {
  const [revokeToken, { loading }] = useMutation<RevokeAPITokenData>(REVOKE_END_USER_API_TOKEN)

  const handleRevoke = async () => {
    if (!token) return
    const { data } = await revokeToken({ variables: { id: token.id } })
    const result = data?.revokeEndUserAPIToken
    if (result?.error) {
      toast.error(result.error.message ?? '撤销失败')
      return
    }
    toast.success(`Token「${token.name}」已撤销`)
    onRevoked()
    onClose()
  }

  return (
    <AlertDialog open={!!token} onOpenChange={(o) => { if (!o) onClose() }}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>撤销 Token</AlertDialogTitle>
          <AlertDialogDescription>
            确定要撤销 Token「
            <span className="font-mono font-medium">{token?.name}</span>
            」吗？撤销后该 Token 立即失效，使用此 Token 的 CLI 将需要重新登录。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>取消</AlertDialogCancel>
          <AlertDialogAction
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            onClick={() => void handleRevoke()}
            disabled={loading}
          >
            {loading ? '撤销中...' : '确认撤销'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

// ── Token Row Skeleton ────────────────────────────────────────────────────────

function TokenRowSkeleton() {
  return (
    <tr className="border-b last:border-0">
      <td className="px-5 py-4">
        <div className="h-3.5 w-28 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4">
        <div className="h-3 w-20 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4">
        <div className="h-3 w-20 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4">
        <div className="h-3 w-16 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4" />
    </tr>
  )
}

// ── Main Page ────────────────────────────────────────────────────────────────

function TokenPageContent({ orgName }: { orgName: string }) {
  const [createOpen, setCreateOpen] = useState(false)
  const [plaintextToken, setPlaintextToken] = useState<string | null>(null)
  const [revokeTarget, setRevokeTarget] = useState<APIToken | null>(null)

  const { data, loading, refetch } = useQuery<EndUserAPITokensData>(END_USER_API_TOKENS, {
    fetchPolicy: 'network-only',
  })

  const tokens = data?.endUserAPITokens ?? []

  return (
    <EndUserAppLayout orgName={orgName} activePage="token">
      <div className="h-full overflow-y-auto">
        <div className="space-y-8 px-10 py-8">

          {/* Page header */}
          <div className="flex items-start justify-between gap-4 pb-2">
            <div>
              <h2 className="text-xl font-semibold text-foreground">API Token 管理</h2>
              <p className="mt-1.5 text-sm text-muted-foreground">Personal Access Token 用于 CLI 和 API 认证，等同于密码，请勿分享。</p>
              <a
                href={`/end-user/${orgName}/dashboard/api-docs`}
                className="mt-2 inline-flex items-center gap-1.5 text-xs text-primary hover:underline"
              >
                <BookOpen className="size-3.5" />
                查看 API 接入文档
              </a>
            </div>
            <Button
              size="sm"
              className="mt-1 shrink-0"
              onClick={() => setCreateOpen(true)}
            >
              <Plus className="mr-1.5 size-4" />
              创建 Token
            </Button>
          </div>

          {/* Token list */}
          <div className="overflow-hidden rounded-lg border bg-card">
            {/* Table toolbar */}
            <div className="flex items-center justify-between border-b bg-[#F6F8FA] px-5 py-3">
              <span className="text-[11px] font-medium uppercase tracking-[0.07em] text-muted-foreground">
                Token 列表
              </span>
              <span className="text-xs text-muted-foreground">
                {loading ? '加载中...' : `${tokens.length} / 10 个`}
              </span>
            </div>

            {loading ? (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-[#E3E8EE] bg-white">
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">名称</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">创建时间</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">过期时间</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">最后使用</th>
                    <th className="w-12 px-5 py-4" />
                  </tr>
                </thead>
                <tbody>
                  <TokenRowSkeleton />
                  <TokenRowSkeleton />
                  <TokenRowSkeleton />
                </tbody>
              </table>
            ) : tokens.length === 0 ? (
              <div className="px-6 py-14 text-center">
                <KeyRound className="mx-auto mb-3 size-8 text-muted-foreground/30" />
                <p className="text-sm font-medium text-muted-foreground">暂无 Token</p>
                <p className="mt-1 text-xs text-muted-foreground/70">
                  创建一个 Token 用于 CLI 或 API 认证
                </p>
                <Button size="sm" className="mt-4" onClick={() => setCreateOpen(true)}>
                  <Plus className="mr-1.5 size-3.5" />
                  创建 Token
                </Button>
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-[#E3E8EE] bg-white">
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">名称</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">创建时间</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">过期时间</th>
                    <th className="px-5 py-4 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">最后使用</th>
                    <th className="w-12 px-5 py-4" />
                  </tr>
                </thead>
                <tbody>
                  {tokens.map((token, i) => {
                    const isExpired = token.expiresAt
                      ? new Date(token.expiresAt) < new Date()
                      : false
                    return (
                      <tr
                        key={token.id}
                        className={[
                          'group transition-colors hover:bg-black/[0.015]',
                          i < tokens.length - 1 ? 'border-b border-[#E3E8EE]' : '',
                        ].join(' ')}
                      >
                        <td className="px-5 py-4">
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-foreground">{token.name}</span>
                            {isExpired && (
                              <span className="rounded-sm bg-destructive/10 px-1.5 py-0.5 text-[10px] font-medium text-destructive">
                                已过期
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-5 py-4 tabular-nums text-muted-foreground">
                          {formatDate(token.createdAt)}
                        </td>
                        <td className="px-5 py-4 tabular-nums text-muted-foreground">
                          {token.expiresAt ? (
                            <span className={isExpired ? 'text-destructive' : ''}>
                              {formatDate(token.expiresAt)}
                            </span>
                          ) : (
                            <span className="text-muted-foreground/50">永不过期</span>
                          )}
                        </td>
                        <td className="px-5 py-4 tabular-nums text-muted-foreground">
                          {token.lastUsedAt ? (
                            formatDate(token.lastUsedAt)
                          ) : (
                            <span className="text-muted-foreground/50">从未使用</span>
                          )}
                        </td>
                        <td className="px-5 py-4 text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 text-muted-foreground/40 opacity-0 transition-opacity hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100"
                            title="撤销 Token"
                            onClick={() => setRevokeTarget(token)}
                          >
                            <Trash2 className="size-3.5" />
                          </Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            )}
          </div>

        </div>
      </div>

      {/* Dialogs */}
      <CreateTokenDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(pt) => setPlaintextToken(pt)}
      />
      <PlaintextDialog
        plaintext={plaintextToken ?? ''}
        open={!!plaintextToken}
        onClose={() => {
          setPlaintextToken(null)
          void refetch()
        }}
      />
      <RevokeDialog
        token={revokeTarget}
        onClose={() => setRevokeTarget(null)}
        onRevoked={() => void refetch()}
      />
    </EndUserAppLayout>
  )
}

export default function TokenPage() {
  const params = useParams<{ orgName: string }>()
  const orgName = params.orgName

  const client = useMemo(() => createEndUserOrgScopedClient(orgName), [orgName])

  return (
    <ApolloProvider client={client}>
      <TokenPageContent orgName={orgName} />
    </ApolloProvider>
  )
}
