'use client'

import { Button } from '@web/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@web/components/ui/avatar'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import { Separator } from '@web/components/ui/separator'
import { Pencil } from 'lucide-react'
import type { UserProfileView } from '@/types'

export interface ProfileOverviewPanelProps {
  profile: UserProfileView
  onEdit: () => void
}

function formatDateTime(dateText: string): string {
  if (!dateText) {
    return '-'
  }

  const parsed = new Date(dateText)
  if (Number.isNaN(parsed.getTime())) {
    return '-'
  }

  return parsed.toLocaleString()
}

export function ProfileOverviewPanel({ profile, onEdit }: ProfileOverviewPanelProps) {
  const fallbackText = profile.nickname.slice(0, 1).toUpperCase() || 'U'

  return (
    <Card className="border bg-background">
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <Avatar className="size-16">
            {profile.avatarUrl && <AvatarImage src={profile.avatarUrl} alt={profile.nickname} />}
            <AvatarFallback className="bg-primary text-lg font-semibold text-primary-foreground">
              {fallbackText}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col gap-1">
            <CardTitle className="text-xl font-semibold text-foreground">{profile.nickname}</CardTitle>
            <p className="text-sm font-medium text-muted-foreground">@{profile.userName}</p>
          </div>
        </div>

        <Button onClick={onEdit} className="gap-2">
          <Pencil className="size-4" />
          编辑资料
        </Button>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-1.5">
            <p className="text-sm font-medium text-muted-foreground">手机号</p>
            <p className="text-sm font-medium text-foreground">{profile.phone}</p>
          </div>
          <div className="space-y-1.5">
            <p className="text-sm font-medium text-muted-foreground">账号状态</p>
            <p className="text-sm font-medium text-foreground">{profile.status}</p>
          </div>
        </div>

        <Separator />

        <div className="space-y-1.5">
          <p className="text-sm font-medium text-muted-foreground">头像地址</p>
          <p className="break-all text-sm font-medium text-foreground">{profile.avatarUrl || '-'}</p>
        </div>

        <div className="space-y-1.5">
          <p className="text-sm font-medium text-muted-foreground">个人简介</p>
          <p className="text-sm font-medium text-foreground">{profile.bio || '暂无简介'}</p>
        </div>

        <Separator />

        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-1.5">
            <p className="text-sm font-medium text-muted-foreground">创建时间</p>
            <p className="text-sm font-medium text-foreground">{formatDateTime(profile.createdAt)}</p>
          </div>
          <div className="space-y-1.5">
            <p className="text-sm font-medium text-muted-foreground">更新时间</p>
            <p className="text-sm font-medium text-foreground">{formatDateTime(profile.updatedAt)}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
