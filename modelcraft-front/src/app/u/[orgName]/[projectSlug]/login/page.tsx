import { redirect } from 'next/navigation'

interface EndUserLoginPageProps {
  params: Promise<{
    orgName: string
    projectSlug: string
  }>
}

/**
 * 旧的 Project 级终端用户登录页。
 * EndUser v2 将账号管理上移到 Org 级，此页面重定向到新的 Org 级登录页。
 */
export default async function EndUserLoginPage({ params }: EndUserLoginPageProps) {
  const { orgName } = await params
  redirect(`/u/${orgName}/login`)
}
