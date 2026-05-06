'use client'

import { useRouter } from 'next/navigation'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspaceProjectsTabProps {
  orgName: string
  projects: EndUserAccessibleProject[]
}

export function WorkspaceProjectsTab({ orgName, projects }: WorkspaceProjectsTabProps) {
  const router = useRouter()

  if (projects.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <span className="mb-4 text-4xl">🔒</span>
        <p className="text-base font-medium text-foreground">您暂无项目访问权限</p>
        <p className="mt-2 text-sm text-muted-foreground">请联系管理员授权</p>
      </div>
    )
  }

  return (
    <div>
      <p className="mb-6 text-sm text-muted-foreground">选择要进入的项目</p>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <div
            key={project.slug}
            className="cursor-pointer rounded-lg border bg-background p-5 transition-shadow hover:border-primary/40 hover:shadow-md"
            onClick={() => router.push(`/end-user/${orgName}/projects/${project.slug}/data`)}
          >
            <p className="font-semibold text-foreground">{project.title}</p>
            <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{project.slug}</p>
            <button
              className="mt-4 text-sm font-medium text-primary hover:underline"
              onClick={(e) => {
                e.stopPropagation()
                router.push(`/end-user/${orgName}/projects/${project.slug}/data`)
              }}
            >
              进入 →
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
