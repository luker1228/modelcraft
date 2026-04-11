'use client'

import { Button } from '@web/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import { Skeleton } from '@web/components/ui/skeleton'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { ProfileOverviewPanel } from './_components'
import { useProfilePageData, useProfilePageState } from './_hooks'

function ProfileOverviewSkeleton() {
  return (
    <Card className="border bg-background">
      <CardHeader className="space-y-2">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-4 w-48" />
      </CardHeader>
      <CardContent className="space-y-4">
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
      </CardContent>
    </Card>
  )
}

function ProfileErrorState({ message, onRetry }: { message: string; onRetry: () => Promise<void> }) {
  return (
    <Card className="border border-destructive/30 bg-destructive/10">
      <CardHeader>
        <CardTitle className="text-base font-semibold text-destructive">资料加载失败</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm font-medium text-destructive">{message}</p>
        <Button variant="outline" onClick={() => void onRetry()}>
          重新加载
        </Button>
      </CardContent>
    </Card>
  )
}

function ProfileEmptyState({ onRetry }: { onRetry: () => Promise<void> }) {
  return (
    <Card className="border border-dashed bg-background">
      <CardHeader>
        <CardTitle className="text-base font-semibold text-foreground">暂无个人资料</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm font-medium text-muted-foreground">当前账号还没有可展示的个人资料信息。</p>
        <Button variant="outline" onClick={() => void onRetry()}>
          刷新
        </Button>
      </CardContent>
    </Card>
  )
}

export default function ProfilePage() {
  const pageState = useProfilePageState({ mode: 'view' })
  const pageData = useProfilePageData()

  return (
    <AppLayout pageTitle="个人资料">
      <div className="mx-auto w-full max-w-4xl p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-foreground">个人资料</h1>
          <p className="mt-1 text-sm font-medium text-muted-foreground">查看并管理你的个人账号信息</p>
        </div>

        {pageData.loading && <ProfileOverviewSkeleton />}

        {!pageData.loading && pageData.error && (
          <ProfileErrorState message={pageData.error.message} onRetry={pageData.refetch} />
        )}

        {!pageData.loading && !pageData.error && !pageData.profile && (
          <ProfileEmptyState onRetry={pageData.refetch} />
        )}

        {!pageData.loading && !pageData.error && pageData.profile && (
          <ProfileOverviewPanel profile={pageData.profile} onEdit={pageState.goToEdit} />
        )}
      </div>
    </AppLayout>
  )
}
