'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { useMutation } from '@apollo/client'
import { UPDATE_ORGANIZATION } from '@/api-client/user'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { useOrganizationStore } from '@shared/stores/organization'
import { invalidateMembershipsCache } from '@/shared/cache/memberships-cache'
import { getToken } from '@api-client/auth/public'
import { toast } from 'sonner'

interface UpdateOrgResult {
  updateOrganization: {
    organization?: { id: string; displayName?: string | null }
    error?: { message: string }
  }
}

export default function GeneralSettingsPage() {
  const params = useParams()
  const orgName = params.orgName as string

  const memberships = useOrganizationStore((state) => state.memberships)
  const loadMembershipsStore = useOrganizationStore((state) => state.loadMemberships)
  const currentOrg = memberships.find((m) => m.orgName === orgName)

  const [displayName, setDisplayName] = useState('')

  useEffect(() => {
    if (currentOrg?.displayName) {
      setDisplayName(currentOrg.displayName)
    }
  }, [currentOrg?.displayName])

  const [updateOrganization, { loading }] = useMutation<UpdateOrgResult>(UPDATE_ORGANIZATION, {
    onCompleted: (result) => {
      if (result.updateOrganization.error) {
        toast.error(result.updateOrganization.error.message)
        return
      }
      toast.success('组织名称已更新')
      // Refresh memberships so breadcrumb updates
      invalidateMembershipsCache()
      const token = getToken()
      if (token) {
        loadMembershipsStore(token, true).catch(() => {})
      }
    },
    onError: (err) => {
      toast.error(`更新失败：${err.message}`)
    },
  })

  const handleSave = () => {
    if (!displayName.trim()) return
    updateOrganization({ variables: { input: { displayName: displayName.trim() } } })
  }

  return (
    <div className="max-w-md space-y-6">
      <div className="space-y-1.5">
        <h2 className="font-sans text-sm font-semibold text-foreground">基本信息</h2>
        <p className="font-sans text-sm text-muted-foreground">修改组织的显示名称。</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="displayName" className="font-sans text-sm font-medium">
          显示名称
        </Label>
        <Input
          id="displayName"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="输入组织显示名称"
          className="font-sans"
        />
        <p className="font-sans text-xs text-muted-foreground">
          组织标识符（URL 中的名称）：<span className="font-mono">{orgName}</span>
        </p>
      </div>

      <Button
        size="sm"
        onClick={handleSave}
        disabled={loading || !displayName.trim()}
      >
        {loading ? '保存中…' : '保存'}
      </Button>
    </div>
  )
}
