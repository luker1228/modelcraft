'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { WorkspaceProjectsTab } from './_components/WorkspaceProjectsTab'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspacePageProps {
  params: { orgName: string }
}

export default function WorkspacePage({ params }: WorkspacePageProps) {
  const router = useRouter()
  const { orgName } = params
  const [projects, setProjects] = useState<EndUserAccessibleProject[]>([])
  const [activeTab, setActiveTab] = useState<'projects'>('projects')

  useEffect(() => {
    const raw = sessionStorage.getItem(`eu_accessible_projects_${orgName}`)
    if (raw) {
      try {
        setProjects(JSON.parse(raw) as EndUserAccessibleProject[])
      } catch {
        setProjects([])
      }
    }
  }, [orgName])

  const handleLogout = async () => {
    if (!orgName) return
    await fetch(`/api/bff/org/${orgName}/end-user/auth/logout`, {
      method: 'POST',
      credentials: 'same-origin',
    })
    router.push(`/end-user/${orgName}/login`)
  }

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      {/* 顶部栏 */}
      <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-background px-6">
        <span className="text-base font-semibold text-foreground">{orgName}</span>
        <button
          onClick={() => void handleLogout()}
          className="text-sm text-destructive hover:underline"
        >
          登出
        </button>
      </header>

      {/* Tab 导航 */}
      <nav className="flex border-b bg-background px-6">
        <button
          className={`border-b-2 px-4 py-3 text-sm font-medium transition-colors ${
            activeTab === 'projects'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
          onClick={() => setActiveTab('projects')}
        >
          Projects
        </button>
        <button
          className="cursor-not-allowed border-b-2 border-transparent px-4 py-3 text-sm text-muted-foreground/50"
          disabled
          title="即将推出"
        >
          （待定）
        </button>
      </nav>

      {/* 主内容 */}
      <main className="flex-1 p-6">
        {activeTab === 'projects' && (
          <WorkspaceProjectsTab orgName={orgName} projects={projects} />
        )}
      </main>
    </div>
  )
}
