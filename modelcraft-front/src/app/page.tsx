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
  // When both sessions are valid, store the destinations so UI can show "enter" buttons
  const [adminDest, setAdminDest] = useState<string | null>(null)
  const [endUserDest, setEndUserDest] = useState<string | null>(null)

  useEffect(() => {
    async function init() {
      const defaultOrgName = localStorage.getItem('defaultOrgName')
      const devToken = getToken()

      // Resolve end-user session info
      const endUserStore = useEndUserAuthStore.getState()
      const endUserToken = getEndUserToken()
      const endUserInfo =
        endUserStore.userInfo ??
        (endUserToken ? getEndUserInfoFromToken(endUserToken) : null)

      let resolvedDevOrgName: string | null = null
      let resolvedEndUserOrgName: string | null = null

      // --- Resolve dev/admin session ---
      if (devToken && defaultOrgName) {
        resolvedDevOrgName = defaultOrgName
      }
      // If resolvedDevOrgName is still null (token missing OR org name missing), attempt refresh
      if (!resolvedDevOrgName) {
        const refreshedToken = await refreshAccessToken()
        if (refreshedToken) {
          const refreshedOrgName = localStorage.getItem('defaultOrgName')
          if (refreshedOrgName) {
            resolvedDevOrgName = refreshedOrgName
          }
        }
      }

      // --- Resolve end-user session ---
      if (endUserInfo?.orgName) {
        if (endUserToken && !endUserStore.isTokenExpired()) {
          resolvedEndUserOrgName = endUserInfo.orgName
        } else {
          const refreshedEndUserToken = await refreshEndUserAccessToken({
            orgName: endUserInfo.orgName,
          })
          if (refreshedEndUserToken) {
            resolvedEndUserOrgName = endUserInfo.orgName
          }
        }
      }

      // --- Decide what to do ---
      const hasBoth = resolvedDevOrgName && resolvedEndUserOrgName

      if (hasBoth) {
        // Show selection page with "enter" buttons instead of auto-redirecting
        setAdminDest(`/org/${resolvedDevOrgName}/dashboard`)
        setEndUserDest(`/end-user/${resolvedEndUserOrgName}/dashboard`)
        setIsCheckingSession(false)
        return
      }

      if (resolvedDevOrgName) {
        router.replace(`/org/${resolvedDevOrgName}/dashboard`)
        return
      }

      if (resolvedEndUserOrgName) {
        router.replace(`/end-user/${resolvedEndUserOrgName}/dashboard`)
        return
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

  // Both sessions are active: show "enter" buttons instead of login buttons
  const hasBothSessions = adminDest && endUserDest

  return (
    <AuthLayout title="统一登录入口" subtitle={hasBothSessions ? "请选择要进入的工作区" : "请选择适合您的登录方式"} showCliPromo>
      <div className="flex flex-col gap-4">
        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织管理员</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {hasBothSessions ? '进入管理后台。' : '登录管理后台，或创建新的组织空间。'}
            </p>
          </div>

          <div className="flex flex-col gap-2">
            {hasBothSessions ? (
              <Button asChild className="w-full">
                <NextLink href={adminDest}>进入管理端</NextLink>
              </Button>
            ) : (
              <>
                <Button asChild className="w-full">
                  <NextLink href={TENANT_LOGIN_PATH}>登录管理员</NextLink>
                </Button>
                <Button asChild variant="outline" className="w-full">
                  <NextLink href={TENANT_REGISTER_PATH}>注册组织</NextLink>
                </Button>
              </>
            )}
          </div>
        </section>

        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织员工</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {hasBothSessions
                ? '进入用户工作区。'
                : '直接使用用户名和密码登录，系统会根据返回结果自动跳转到所属组织。'}
            </p>
          </div>

          <Button asChild className="w-full">
            <NextLink href={hasBothSessions ? endUserDest : END_USER_LOGIN_PATH}>
              {hasBothSessions ? '进入用户端' : '用户登录'}
            </NextLink>
          </Button>
        </section>
      </div>
    </AuthLayout>
  )
}
