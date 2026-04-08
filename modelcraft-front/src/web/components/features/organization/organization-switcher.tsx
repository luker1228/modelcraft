'use client'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { useCurrentOrg, useOrganizations } from '@web/stores'
import { useOrganizationStore } from '@shared/stores/organization'
import { resetApolloCache } from '@web/providers/apollo-wrapper'
import { useRouter, usePathname } from 'next/navigation'
import { Building2 } from 'lucide-react'

export function OrganizationSwitcher() {
  const currentOrg = useCurrentOrg()
  const organizations = useOrganizations()
  const setCurrentOrg = useOrganizationStore((state) => state.setCurrentOrg)
  const router = useRouter()
  const pathname = usePathname()

  if (organizations.length === 0) {
    // Show current org from JWT even if store is empty
    if (currentOrg) {
      return (
        <div className="flex items-center gap-2 px-2 py-1 text-sm text-muted-foreground">
          <Building2 className="size-4" />
          <span className="max-w-[160px] truncate">{currentOrg}</span>
        </div>
      )
    }
    return null
  }

  // If only one organization, show it as text instead of selector
  if (organizations.length === 1) {
    return (
      <div className="flex items-center gap-2 px-2 py-1 text-sm text-muted-foreground">
        <Building2 className="size-4" />
        <span className="max-w-[160px] truncate">{organizations[0]}</span>
      </div>
    )
  }

  const handleOrgChange = (newOrg: string) => {
    if (newOrg === currentOrg) return

    // Update store
    setCurrentOrg(newOrg)

    // Clear Apollo cache to fetch fresh data for new org
    resetApolloCache()

    // Navigate to the same relative page under new org, or default to workspace
    if (pathname && currentOrg) {
      const orgPrefix = `/org/${currentOrg}`
      if (pathname.startsWith(orgPrefix)) {
        const relativePath = pathname.slice(orgPrefix.length) || '/workspace'
        router.push(`/org/${newOrg}${relativePath}`)
        return
      }
    }

    router.push(`/org/${newOrg}/workspace`)
  }

  return (
    <Select value={currentOrg || undefined} onValueChange={handleOrgChange}>
      <SelectTrigger className="h-9 w-[200px]">
        <div className="flex items-center gap-2">
          <Building2 className="size-4 shrink-0" />
          <SelectValue placeholder="Select organization" />
        </div>
      </SelectTrigger>
      <SelectContent>
        {organizations.map((org) => (
          <SelectItem key={org} value={org}>
            {org}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
