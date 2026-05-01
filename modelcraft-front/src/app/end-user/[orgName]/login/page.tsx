import type { Metadata } from 'next'
import { EndUserOrgLoginCard } from '@web/components/features/end-user-auth/EndUserOrgLoginCard'

export const metadata: Metadata = {
  title: '终端用户登录',
  robots: {
    index: false,
    follow: false,
  },
}

interface EndUserOrgLoginPageProps {
  params: Promise<{
    orgName: string
  }>
}

/**
 * Org 级终端用户登录页（EndUser v2）。
 * 登录成功后根据可访问 Project 数量决定下一步：
 * - 1 个 Project → 直接进入 Data 页
 * - N 个 Project → 跳转选择 Project 页
 * - 0 个 Project → 显示无访问权限错误
 */
export default async function EndUserOrgLoginPage({ params }: EndUserOrgLoginPageProps) {
  const { orgName } = await params
  return <EndUserOrgLoginCard orgName={orgName} />
}
