'use client'

import { useMemo } from 'react'
import { useParams } from 'next/navigation'
import { useQuery } from '@apollo/client'
import { Users } from 'lucide-react'
import { AppLayout } from '@web/components/layout/AppLayout'
import { Badge } from '@web/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { GET_ORGANIZATION_MEMBERS } from '@web/graphql/queries/user'
import { useRequireAuth } from '@web/hooks/useAuth'
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

export default function TeamPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const { isLoading: authLoading } = useRequireAuth()

  const orgScopedContext = useMemo(() => ({
    uri: `/graphql/org/${orgName}/`,
  }), [orgName])

  const { data, loading, error } = useQuery<MembersQueryData>(GET_ORGANIZATION_MEMBERS, {
    skip: !orgName || authLoading,
    context: orgScopedContext,
  })

  const members: OrganizationMember[] = data?.organizationMembers ?? []

  return (
    <AppLayout pageTitle="团队">
      <div className="h-full overflow-auto bg-white">
        <div>
          <div className="mx-auto max-w-7xl p-6">
            {/* Page Header */}
            <div className="mb-6">
              <h1 className="font-heading text-xl font-semibold tracking-tight text-foreground">
                团队成员
              </h1>
            </div>

            {/* Error State */}
            {error && !loading && (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <Users className="mb-3 size-10 text-muted-foreground/50" />
                <p className="font-sans text-sm text-muted-foreground">加载成员失败，请刷新页面重试</p>
              </div>
            )}

            {/* Table */}
            {!error && (
              <div className="rounded-lg border">
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
                    {(loading || authLoading) && (
                      Array.from({ length: 5 }).map((_, i) => (
                        <TableRow key={i}>
                          <TableCell><div className="h-4 w-24 animate-pulse rounded bg-slate-200" /></TableCell>
                          <TableCell><div className="h-4 w-40 animate-pulse rounded bg-slate-200" /></TableCell>
                          <TableCell><div className="h-5 w-16 animate-pulse rounded bg-slate-200" /></TableCell>
                          <TableCell><div className="h-5 w-16 animate-pulse rounded bg-slate-200" /></TableCell>
                          <TableCell><div className="h-4 w-20 animate-pulse rounded bg-slate-200" /></TableCell>
                        </TableRow>
                      ))
                    )}
                    {!loading && !authLoading && members.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={5}>
                          <div className="flex flex-col items-center justify-center py-12 text-center">
                            <Users className="mb-3 size-10 text-muted-foreground/50" />
                            <p className="font-sans text-sm text-muted-foreground">暂无成员</p>
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
                共 {members.length} 名成员
              </p>
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  )
}
