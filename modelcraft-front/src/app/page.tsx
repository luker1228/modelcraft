'use client'

import NextLink from 'next/link'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import {
  TENANT_LOGIN_PATH,
  TENANT_REGISTER_PATH,
} from '@shared/constants/routes'

export default function Home() {
  return (
    <AuthLayout
      title="登录 ModelCraft"
      subtitle="管理 AI 数据访问权限，从这里开始。"
      variant="landing"
      showCliPromo
    >
      <Button asChild className="w-full">
        <NextLink href={TENANT_LOGIN_PATH}>管理员登录</NextLink>
      </Button>
      <Button asChild variant="outline" className="w-full">
        <NextLink href={TENANT_REGISTER_PATH}>注册组织</NextLink>
      </Button>
    </AuthLayout>
  )
}
