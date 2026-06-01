'use client'

import { useRouter } from 'next/navigation'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspaceProjectsTabProps {
  orgName: string
  projects: EndUserAccessibleProject[]
  loading?: boolean
}

function SkeletonCard() {
  return (
    <div className="animate-pulse rounded-lg border border-[#EAEAEA] bg-white p-6">
      <div className="mb-4 size-8 rounded bg-[#F0EFF0]" />
      <div className="mb-2 h-3.5 w-1/2 rounded bg-[#F0EFF0]" />
      <div className="h-3 w-1/3 rounded bg-[#F0EFF0]" />
    </div>
  )
}

function ProjectCard({
  project,
  orgName,
  index,
}: {
  project: EndUserAccessibleProject
  orgName: string
  index: number
}) {
  const router = useRouter()
  const initials = project.title
    .split(/[\s_-]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('')

  return (
    <button
      type="button"
      onClick={() => router.push(`/end-user/${orgName}/projects/${project.slug}/data`)}
      className="group flex flex-col rounded-lg border border-[#EAEAEA] bg-white p-6 text-left transition-shadow duration-150 hover:shadow-[0_2px_8px_rgba(0,0,0,0.04)]"
      style={{ animationDelay: `${index * 60}ms` }}
    >
      {/* Project icon */}
      <div className="mb-5 flex size-8 items-center justify-center rounded bg-[#F0EFF9] text-xs font-semibold text-[#4F46E5]">
        {initials || '?'}
      </div>

      {/* Title + slug */}
      <p className="truncate text-[14px] font-semibold leading-snug text-[#111111]">
        {project.title}
      </p>
      <p className="mt-1 truncate font-mono text-[12px] text-[#787774]">{project.slug}</p>

      {/* Arrow — visible on hover */}
      <div className="mt-5 flex items-center justify-between">
          <span className="inline-flex items-center gap-1.5 text-[12px] text-[#787774]">
          <span className="size-1.5 rounded-full bg-[#059669]" />
          可用
        </span>
        <span className="text-[12px] font-medium text-[#4F46E5] opacity-0 transition-opacity duration-150 group-hover:opacity-100">
          进入 →
        </span>
      </div>
    </button>
  )
}

export function WorkspaceProjectsTab({ orgName, projects, loading }: WorkspaceProjectsTabProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  if (projects.length === 0) {
    return (
      <div className="flex flex-col items-center py-24 text-center">
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#8792A2"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="mb-4"
        >
          <rect x="3" y="3" width="7" height="7" rx="1" />
          <rect x="14" y="3" width="7" height="7" rx="1" />
          <rect x="3" y="14" width="7" height="7" rx="1" />
          <rect x="14" y="14" width="7" height="7" rx="1" />
        </svg>
        <p className="text-[14px] font-medium text-[#111111]">暂无项目访问权限</p>
        <p className="mt-1 text-[13px] text-[#787774]">请联系管理员授予项目权限</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {projects.map((project, i) => (
        <ProjectCard key={project.slug} project={project} orgName={orgName} index={i} />
      ))}
    </div>
  )
}
