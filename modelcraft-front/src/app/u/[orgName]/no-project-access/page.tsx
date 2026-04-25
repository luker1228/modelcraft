import type { Metadata } from 'next'
import Link from 'next/link'
import { ShieldAlert } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Button } from '@web/components/ui/button'

export const metadata: Metadata = {
  title: '待授权',
  robots: {
    index: false,
    follow: false,
  },
}

interface NoProjectAccessPageProps {
  params: Promise<{
    orgName: string
  }>
}

export default async function NoProjectAccessPage({ params }: NoProjectAccessPageProps) {
  const { orgName } = await params

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-md border bg-background shadow-sm">
        <CardHeader className="space-y-4 px-8 pt-8 text-center">
          <div className="mx-auto flex size-10 items-center justify-center rounded-lg bg-amber-100 text-amber-700">
            <ShieldAlert className="size-5" strokeWidth={1.5} />
          </div>
          <div>
            <CardTitle className="text-2xl">暂无项目访问权限</CardTitle>
            <CardDescription className="mt-2">
              您已完成登录，但当前账号在 <span className="font-medium text-foreground">{orgName}</span>{' '}
              下暂无可访问项目。
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent className="px-8 pb-8">
          <p className="mb-6 text-sm text-muted-foreground">
            请联系管理员为您分配项目访问权限，授权完成后可重新登录进入系统。
          </p>

          <div className="flex flex-col gap-2">
            <Button asChild>
              <Link href={`/u/${orgName}/login`}>我已申请，重新登录</Link>
            </Button>
            <Button asChild variant="outline">
              <Link href={`/u/${orgName}/login`}>返回登录页</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
