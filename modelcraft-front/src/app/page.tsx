'use client'

import NextLink from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getToken, refreshAccessToken } from '@api-client/auth/public'
import {
  getEndUserInfoFromToken,
  getEndUserToken,
  refreshEndUserAccessToken,
} from '@api-client/end-user/end-user-auth-client'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import {
  END_USER_LOGIN_PATH,
  TENANT_LOGIN_PATH,
  TENANT_REGISTER_PATH,
} from '@shared/constants/routes'

export default function Home() {
  const router = useRouter()
  const [isCheckingSession, setIsCheckingSession] = useState(true)

  useEffect(() => {
    async function init() {
      const defaultOrgName = localStorage.getItem('defaultOrgName')
      const devToken = getToken()
      if (devToken && defaultOrgName) {
        router.replace(`/org/${defaultOrgName}/dashboard`)
        return
      }

      const endUserStore = useEndUserAuthStore.getState()
      const endUserToken = getEndUserToken()
      const endUserInfo = endUserStore.userInfo ??
        (endUserToken ? getEndUserInfoFromToken(endUserToken) : null)
      if (endUserInfo?.orgName) {
        if (endUserToken && !endUserStore.isTokenExpired()) {
          router.replace(`/end-user/${endUserInfo.orgName}/workspace`)
          return
        }

        const refreshedEndUserToken = await refreshEndUserAccessToken({
          orgName: endUserInfo.orgName,
        })
        if (refreshedEndUserToken) {
          router.replace(`/end-user/${endUserInfo.orgName}/workspace`)
          return
        }
      }

      if (!devToken) {
        const refreshedToken = await refreshAccessToken()
        if (refreshedToken) {
          const refreshedOrgName = localStorage.getItem('defaultOrgName')
          if (refreshedOrgName) {
            router.replace(`/org/${refreshedOrgName}/dashboard`)
            return
          }
        }
      }

      setIsCheckingSession(false)
    }

    void init()
  }, [router])

  if (isCheckingSession) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="space-y-4 text-center">
          <div className="mx-auto size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">正在检查登录状态...</p>
        </div>
      </div>
    )
  }

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
