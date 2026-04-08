'use client'

import { useQuery } from '@apollo/client'
import { GET_ORGANIZATION_MEMBERS } from '@web/graphql'
import { MembersTable } from '@web/components/features/settings/MembersTable'
import type { OrganizationMember } from '@/types'

interface MembersQueryData {
  organizationMembers: OrganizationMember[]
}

export default function MembersPage() {
  const { data, loading, error } = useQuery<MembersQueryData>(GET_ORGANIZATION_MEMBERS)

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="size-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
        <p className="text-sm text-destructive">
          Failed to load members: {error.message}
        </p>
      </div>
    )
  }

  const members = data?.organizationMembers ?? []

  return (
    <div>
      <div className="mb-4">
        <h2 className="text-lg font-semibold">Organization Members</h2>
        <p className="text-sm text-muted-foreground">
          View members of your organization and their roles.
        </p>
      </div>
      <MembersTable members={members} />
    </div>
  )
}
