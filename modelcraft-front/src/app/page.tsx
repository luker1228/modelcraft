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
    <AuthLayout title="统一登录入口" subtitle="请选择适合您的登录方式" showCliPromo>
      <div className="flex flex-col gap-4">
        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织管理员</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              登录管理后台，或创建新的组织空间。
            </p>
          </div>
          <div className="flex flex-col gap-2">
            <Button asChild className="w-full">
              <NextLink href={TENANT_LOGIN_PATH}>登录管理员</NextLink>
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
              直接使用用户名和密码登录，系统会根据返回结果自动跳转到所属组织。
            </p>
          </div>
          <Button asChild className="w-full">
            <NextLink href={END_USER_LOGIN_PATH}>用户登录</NextLink>
          </Button>
        </section>
      </div>
    </AuthLayout>
  )
}
