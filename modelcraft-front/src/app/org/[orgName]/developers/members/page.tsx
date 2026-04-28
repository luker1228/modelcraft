'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@apollo/client'
import { Users } from 'lucide-react'
import { Badge } from '@web/components/ui/badge'
import { useOrgScopedContext } from '@api-client/apollo/public'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { GET_ORGANIZATION_MEMBERS } from '@/api-client/user'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import type { OrganizationMember } from '@/types'

interface MembersQueryData {
  organizationMembers: OrganizationMember[]
}

function formatDate(dateStr?: string | null): string {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString()
  } catch {
    return '-'
  }
}

function getStatusVariant(
  status: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'ACTIVE':
      return 'default'
    case 'INVITED':
      return 'secondary'
    case 'SUSPENDED':
      return 'destructive'
    default:
      return 'outline'
  }
}

export default function DevelopersMembersPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const { isLoading: authLoading } = useRequireAuth()

  const orgScopedContext = useOrgScopedContext(orgName ?? undefined)

  const { data, loading, error } = useQuery<MembersQueryData>(GET_ORGANIZATION_MEMBERS, {
    skip: !orgName || authLoading,
    context: orgScopedContext,
  })

  const members: OrganizationMember[] = data?.organizationMembers ?? []

  return (
    <div>
      {/* Error State */}
      {error && !loading && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Users className="mb-3 size-10 text-muted-foreground/50" />
          <p className="font-sans text-sm text-muted-foreground">加载开发者失败，请刷新页面重试</p>
        </div>
      )}

      {/* Table */}
      {!error && (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="font-sans font-semibold">用户名</TableHead>
                <TableHead className="font-sans font-semibold">用户 ID</TableHead>
                <TableHead className="font-sans font-semibold">角色</TableHead>
                <TableHead className="font-sans font-semibold">状态</TableHead>
                <TableHead className="font-sans font-semibold">加入时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(loading || authLoading) &&
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i}>
                    <TableCell><div className="h-4 w-24 animate-pulse rounded bg-slate-200" /></TableCell>
                    <TableCell><div className="h-4 w-40 animate-pulse rounded bg-slate-200" /></TableCell>
                    <TableCell><div className="h-5 w-16 animate-pulse rounded bg-slate-200" /></TableCell>
                    <TableCell><div className="h-5 w-16 animate-pulse rounded bg-slate-200" /></TableCell>
                    <TableCell><div className="h-4 w-20 animate-pulse rounded bg-slate-200" /></TableCell>
                  </TableRow>
                ))}
              {!loading && !authLoading && members.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className="flex flex-col items-center justify-center py-12 text-center">
                      <Users className="mb-3 size-10 text-muted-foreground/50" />
                      <p className="font-sans text-sm text-muted-foreground">暂无开发者</p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
              {!loading && !authLoading && members.map((member) => {
                const displayName = member.userName || '-'
                const avatarLetter = (member.userName || member.userID).charAt(0).toUpperCase()
                return (
                  <TableRow key={member.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex size-8 flex-shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground shadow">
                          {avatarLetter}
                        </div>
                        <span className="font-sans text-sm font-semibold text-foreground">{displayName}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">{member.userID}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{member.role?.name ?? '-'}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusVariant(member.status)}>{member.status}</Badge>
                    </TableCell>
                    <TableCell className="font-sans text-sm text-muted-foreground">
                      {formatDate(member.joinedAt)}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}
      {!loading && !authLoading && !error && (
        <p className="mt-3 font-sans text-sm text-muted-foreground">
          共 {members.length} 名开发者
        </p>
      )}
    </div>
  )
}
