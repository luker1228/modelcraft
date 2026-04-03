'use client'

import { useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useAppStore } from '@web/stores'
import { getOrgPath, getWelcomePath } from '@bff/auth/public'

// 需要项目上下文的路由
const PROTECTED_ROUTES = [
  '/app/dashboard',
  '/app/clusters',
  '/app/cms',
  '/app/enums',
  '/app/graphql-playground',
  '/app/schema-playground',
  '/app/test-crud',
]

// 公开路由（不需要项目上下文）
const PUBLIC_ROUTES = [
  '/welcome',
  '/app/guide',
]

export function useProjectContext() {
  const router = useRouter()
  const pathname = usePathname()
  const selectedProject = useAppStore((state) => state.selectedProject)
  const setSelectedProject = useAppStore((state) => state.setSelectedProject)
  const clearSelection = useAppStore((state) => state.clearSelection)

  // 检查当前路由是否需要项目上下文
  const requiresProject = PROTECTED_ROUTES.some(route => 
    pathname.startsWith(route)
  )

  const isPublicRoute = PUBLIC_ROUTES.some(route => 
    pathname.startsWith(route)
  )

  // 确保项目上下文
  const ensureProjectContext = () => {
    if (requiresProject && !selectedProject) {
      router.push(getOrgPath('/workspace'))
      return false
    }
    return true
  }

  // 清除项目上下文
  const clearProjectContext = () => {
    clearSelection()
    router.push(getWelcomePath())
  }

  // 路由守卫效果
  useEffect(() => {
    if (requiresProject && !selectedProject) {
      router.push(getOrgPath('/workspace'))
    }
  }, [pathname, selectedProject, requiresProject, router])

  return {
    selectedProject,
    setSelectedProject,
    clearProjectContext,
    ensureProjectContext,
    requiresProject,
    isPublicRoute,
  }
}
