'use client'

import { Button } from '@web/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@web/components/ui/card'
import { Skeleton } from '@web/components/ui/skeleton'
import { AppLayout } from '@web/components/features/layout/AppLayout'
import { ProfileEditForm } from '../_components'
import { useProfileEditForm, useProfilePageData, useProfilePageState } from '../_hooks'

function ProfileEditSkeleton() {
  return (
    <Card className="border bg-background">
      <CardHeader className="space-y-2">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="h-4 w-56" />
      </CardHeader>
      <CardContent className="space-y-4">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-28 w-full" />
        <Skeleton className="h-10 w-36" />
      </CardContent>
    </Card>
  )
}

function ProfileEditErrorState({
  message,
  onRetry,
  onBack,
}: {
  message: string
  onRetry: () => Promise<void>
  onBack: () => void
}) {
  return (
    <Card className="border border-destructive/30 bg-destructive/10">
      <CardHeader>
        <CardTitle className="text-base font-semibold text-destructive">资料读取失败</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm font-medium text-destructive">{message}</p>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={() => void onRetry()}>
            重新加载
          </Button>
          <Button variant="outline" onClick={onBack}>
            返回资料页
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function ProfileEditEmptyState({ onBack }: { onBack: () => void }) {
  return (
    <Card className="border border-dashed bg-background">
      <CardHeader>
        <CardTitle className="text-base font-semibold text-foreground">暂无可编辑资料</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm font-medium text-muted-foreground">当前账号还没有已创建的资料，请返回概览页重试。</p>
        <Button variant="outline" onClick={onBack}>
          返回资料页
        </Button>
      </CardContent>
    </Card>
  )
}

export default function ProfileEditPage() {
  const pageState = useProfilePageState({ mode: 'edit' })
  const pageData = useProfilePageData()

  const profileEditForm = useProfileEditForm({
    profile: pageData.profile,
    refetchProfile: pageData.refetch,
    onSuccess: pageState.goToOverview,
    onSavingChange: pageState.setSaving,
  })

  return (
    <AppLayout pageTitle="编辑个人资料">
      <div className="mx-auto w-full max-w-3xl p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-foreground">编辑个人资料</h1>
          <p className="mt-1 text-sm font-medium text-muted-foreground">更新你的昵称、头像地址与个人简介</p>
        </div>

        {pageData.loading && <ProfileEditSkeleton />}

        {!pageData.loading && pageData.error && (
          <ProfileEditErrorState
            message={pageData.error.message}
            onRetry={pageData.refetch}
            onBack={pageState.goToOverview}
          />
        )}

        {!pageData.loading && !pageData.error && !pageData.profile && (
          <ProfileEditEmptyState onBack={pageState.goToOverview} />
        )}

        {!pageData.loading && !pageData.error && pageData.profile && (
          <Card className="border bg-background">
            <CardHeader>
              <CardTitle className="text-base font-semibold text-foreground">资料表单</CardTitle>
            </CardHeader>
            <CardContent>
              <ProfileEditForm
                initialValues={profileEditForm.initialValues}
                saving={pageState.saving || profileEditForm.saving}
                submitError={profileEditForm.error}
                onSubmit={profileEditForm.submit}
                onCancel={pageState.goToOverview}
              />
            </CardContent>
          </Card>
        )}
      </div>
    </AppLayout>
  )
}
