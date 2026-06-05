'use client'

import { useRouter } from 'next/navigation'
import { FolderOpen } from 'lucide-react'
import { cn } from '@/shared/utils'
import type { EndUserAccessibleProject } from '@/types/end-user-auth'

interface WorkspaceProjectsTabProps {
  orgName: string
  projects: EndUserAccessibleProject[]
  loading?: boolean
}

function SkeletonRow() {
  return (
    <tr className="border-b border-[#E3E8EE]">
      <td className="px-5 py-4">
        <div className="h-3.5 w-36 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4">
        <div className="h-3 w-24 animate-pulse rounded bg-muted" />
      </td>
      <td className="px-5 py-4">
        <div className="h-3 w-10 animate-pulse rounded bg-muted" />
      </td>
      <td className="w-10 px-5 py-4" />
    </tr>
  )
}

function ProjectRow({
  project,
  orgName,
}: {
  project: EndUserAccessibleProject
  orgName: string
}) {
  const router = useRouter()

  return (
    <tr
      className="group cursor-pointer border-b border-[#E3E8EE] transition-colors last:border-0 hover:bg-black/[0.015]"
      onClick={() => router.push(`/end-user/${orgName}/projects/${project.slug}/data`)}
    >
      <td className="px-5 py-4">
        <span className="font-medium text-foreground">{project.title}</span>
      </td>
      <td className="px-5 py-4">
        <code className="font-mono text-xs text-muted-foreground">{project.slug}</code>
      </td>
      <td className="px-5 py-4">
        <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <span className="size-1.5 rounded-full bg-emerald-500" />
          可用
        </span>
      </td>
      <td className="w-10 px-5 py-4" />
    </tr>
  )
}

export function WorkspaceProjectsTab({ orgName, projects, loading }: WorkspaceProjectsTabProps) {
  return (
    <div className="space-y-8 px-10 py-8">

      {/* Page header */}
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">项目</h2>
        {!loading && projects.length > 0 && (
          <span className="text-sm text-muted-foreground">{projects.length} 个项目</span>
        )}
      </div>

      {/* List */}
      <div className="overflow-hidden rounded-lg border bg-card">
        {loading ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#E3E8EE] bg-white">
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">名称</th>
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">标识符</th>
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">状态</th>
                <th className="w-10 px-5 py-3.5" />
              </tr>
            </thead>
            <tbody>
              <SkeletonRow />
              <SkeletonRow />
              <SkeletonRow />
            </tbody>
          </table>
        ) : projects.length === 0 ? (
          <div className="px-6 py-16 text-center">
            <FolderOpen className="mx-auto mb-3 size-8 text-muted-foreground/30" strokeWidth={1.5} />
            <p className="text-sm font-medium text-muted-foreground">暂无项目访问权限</p>
            <p className="mt-1 text-xs text-muted-foreground/60">请联系管理员授予项目权限</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#E3E8EE] bg-white">
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">名称</th>
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">标识符</th>
                <th className="px-5 py-3.5 text-left text-[11px] font-medium uppercase tracking-[0.06em] text-foreground">状态</th>
                <th className="w-10 px-5 py-3.5" />
              </tr>
            </thead>
            <tbody>
              {projects.map((project) => (
                <ProjectRow key={project.slug} project={project} orgName={orgName} />
              ))}
            </tbody>
          </table>
        )}
      </div>

    </div>
  )
}
