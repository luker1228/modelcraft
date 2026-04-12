/**
 * 智能路由重定向逻辑
 * 根据用户的组织和项目数量，自动跳转到最合适的页面
 */

export interface MembershipInfo {
  orgId: string
  orgName: string
  displayName: string
  role: string
  joinedAt: string
}

export interface ProjectInfo {
  id: string
  slug: string
  title: string
  status: string
}

/**
 * 决定登录后应该跳转到哪里
 */
export function getSmartRedirectUrl(
  memberships: MembershipInfo[],
  lastSelectedOrgId?: string
): string {
  // 没有组织 - 跳转到创建组织页面
  if (memberships.length === 0) {
    return '/org/create'
  }

  // 单个组织 - 直接进入该组织的工作空间
  if (memberships.length === 1) {
    const org = memberships[0]
    return `/org/${org.orgName}/workspace`
  }

  // 多个组织 - 检查是否有上次选择的组织
  if (lastSelectedOrgId) {
    const lastOrg = memberships.find(m => m.orgId === lastSelectedOrgId)
    if (lastOrg) {
      return `/org/${lastOrg.orgName}/workspace`
    }
  }

  // 多个组织且没有历史记录 - 默认进入第一个组织
  const firstOrg = memberships[0]
  return `/org/${firstOrg.orgName}/workspace`
}

/**
 * 在组织内部，决定应该显示什么
 * 注意：项目层面不做自动跳转，总是让用户在工作空间中选择
 */
export function getOrgRedirectUrl(
  orgName: string,
  projects: ProjectInfo[],
  lastSelectedProjectSlug?: string
): string {
  // 所有情况都显示工作空间，让用户自己选择项目
  return `/org/${orgName}/workspace`
}

/**
 * 保存用户的选择以便下次智能跳转
 */
export function saveUserPreferences(orgId: string, projectSlug?: string) {
  localStorage.setItem('lastSelectedOrgId', orgId)
  if (projectSlug) {
    localStorage.setItem(`lastSelectedProject_${orgId}`, projectSlug)
  }
}

/**
 * 获取用户的历史选择
 */
export function getUserPreferences(orgId?: string) {
  const lastOrgId = localStorage.getItem('lastSelectedOrgId')
  const lastProjectSlug = orgId
    ? localStorage.getItem(`lastSelectedProject_${orgId}`)
    : null

  return {
    lastOrgId,
    lastProjectSlug
  }
}
