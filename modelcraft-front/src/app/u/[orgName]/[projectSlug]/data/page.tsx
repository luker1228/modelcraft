'use client'

import { useParams } from 'next/navigation'
import { Button } from '@web/components/ui/button'
import { useEndUser } from '@web/hooks/end-user-auth/useRequireEndUserAuth'

export default function EndUserDataPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { user, logout } = useEndUser()

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-4xl flex-col gap-6 px-6 py-10">
      <section className="rounded-lg border border-border bg-background p-6 shadow-sm">
        <h1 className="text-2xl font-semibold text-foreground">数据管理</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          当前项目：{params.orgName} / {params.projectSlug}
        </p>
        <p className="mt-1 text-sm text-muted-foreground">
          当前用户：{user?.username || user?.id || '已登录'}
        </p>

        <div className="mt-6">
          <Button variant="outline" onClick={() => void logout()}>
            退出登录
          </Button>
        </div>
      </section>
    </main>
  )
}

