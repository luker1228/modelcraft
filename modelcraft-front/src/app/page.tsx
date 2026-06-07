'use client'

import NextLink from 'next/link'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import {
  END_USER_LOGIN_PATH,
  TENANT_LOGIN_PATH,
  TENANT_REGISTER_PATH,
} from '@shared/constants/routes'

export default function Home() {
  return (
    <AuthLayout
      title="欢迎使用 ModelCraft"
      subtitle="让 AI 安全、可控地使用数据库"
      showCliPromo
    >
      <div className="flex flex-col gap-4">
        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织管理员</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              登录管理后台，管理组织、权限和数据访问。
            </p>
          </div>
          <div className="flex flex-col gap-2">
            <Button asChild className="w-full">
              <NextLink href={TENANT_LOGIN_PATH}>管理员登录</NextLink>
            </Button>
            <Button asChild variant="outline" className="w-full">
              <NextLink href={TENANT_REGISTER_PATH}>注册组织</NextLink>
            </Button>
          </div>
        </section>

        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织员工</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              通过统一入口进入你的组织空间，查看与操作授权范围内的数据。
            </p>
          </div>
          <Button asChild className="w-full">
            <NextLink href={END_USER_LOGIN_PATH}>员工登录</NextLink>
          </Button>
        </section>
      </div>
    </AuthLayout>
  )
}
