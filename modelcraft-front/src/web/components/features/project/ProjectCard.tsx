import { Button } from "@web/components/ui/button"
import { Badge } from "@web/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@web/components/ui/dropdown-menu"
import { MoreHorizontal, Edit, Trash2, FolderOpen } from "lucide-react"
import { formatDateSafe } from "@/shared/utils"
import type { Project } from "@/types"

interface ProjectCardProps {
  project: Project
  onSelect: (project: Project) => void
  onEdit: (project: Project) => void
  onDelete: (project: Project) => void
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'ACTIVE':
    case 'active':
      return <Badge variant="success">活跃</Badge>
    case 'archived':
      return <Badge variant="secondary">已归档</Badge>
    case 'draft':
      return <Badge variant="warning">草稿</Badge>
    default:
      return <Badge variant="secondary">{status}</Badge>
  }
}

export function ProjectCard({
  project,
  onSelect,
  onEdit,
  onDelete,
}: ProjectCardProps) {
  return (
    <div
      className="group relative flex cursor-pointer flex-col rounded-lg border border-border bg-card transition-colors duration-150 hover:border-foreground/20"
      onClick={() => onSelect(project)}
    >
      {/* Card body */}
      <div className="flex flex-1 flex-col gap-2 p-5">
        {/* Title row */}
        <div className="flex items-start justify-between gap-2">
          <h3 className="line-clamp-1 text-[15px] font-semibold leading-snug text-foreground">
            {project.title}
          </h3>

          {/* Actions — stop propagation so click doesn't open project */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button
                variant="ghost"
                size="icon"
                className="size-7 flex-shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
              >
                <MoreHorizontal className="size-4" />
                <span className="sr-only">打开菜单</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onSelect(project) }}>
                <FolderOpen className="mr-2 size-4" />
                打开项目
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onEdit(project) }}>
                <Edit className="mr-2 size-4" />
                编辑
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={(e) => { e.stopPropagation(); onDelete(project) }}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="mr-2 size-4" />
                删除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Slug */}
        <p className="font-mono text-[12px] text-muted-foreground/70">
          {project.slug}
        </p>

        {/* Description */}
        {project.description ? (
          <p className="line-clamp-2 text-[13px] leading-relaxed text-muted-foreground">
            {project.description}
          </p>
        ) : (
          <p className="text-[13px] italic text-muted-foreground/40">暂无描述</p>
        )}
      </div>

      {/* Card footer */}
      <div className="flex items-center justify-between border-t border-border px-5 py-3">
        <StatusBadge status={project.status} />
        <span className="text-[12px] text-muted-foreground">
          {formatDateSafe(project.updatedAt, { month: 'short', day: 'numeric' })}
        </span>
      </div>
    </div>
  )
}
