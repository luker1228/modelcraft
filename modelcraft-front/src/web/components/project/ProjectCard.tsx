import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@web/components/ui/card"
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
import type { Project } from "@/types"

interface ProjectCardProps {
  project: Project
  onSelect: (project: Project) => void
  onEdit: (project: Project) => void
  onDelete: (project: Project) => void
}

export function ProjectCard({
  project,
  onSelect,
  onEdit,
  onDelete,
}: ProjectCardProps) {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <Badge className="border-0 bg-gradient-to-r from-emerald-100 to-teal-100 text-xs font-semibold text-emerald-700 shadow-sm">
            活跃
          </Badge>
        )
      case 'archived':
        return (
          <Badge className="border-0 bg-gradient-to-r from-slate-100 to-slate-200 text-xs font-semibold text-muted-foreground shadow-sm">
            已归档
          </Badge>
        )
      case 'draft':
        return (
          <Badge className="border-0 bg-gradient-to-r from-amber-100 to-orange-100 text-xs font-semibold text-amber-700 shadow-sm">
            草稿
          </Badge>
        )
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  return (
    <Card className="group cursor-pointer border border-slate-200/60 bg-white/80 shadow-sm backdrop-blur-sm transition-all duration-300 hover:scale-[1.01] hover:border-slate-300 hover:bg-white hover:shadow-lg">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1" onClick={() => onSelect(project)}>
            <CardTitle className="text-lg font-semibold text-foreground transition-all group-hover:text-primary">
              {project.title}
            </CardTitle>
            <CardDescription className="mt-1 line-clamp-2 text-sm text-muted-foreground">
              {project.description || '暂无描述'}
            </CardDescription>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="size-8 opacity-0 transition-opacity hover:bg-selected group-hover:opacity-100">
                <MoreHorizontal className="size-4" />
                <span className="sr-only">打开菜单</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="border-slate-200 bg-white/95 backdrop-blur-sm">
              <DropdownMenuItem onClick={() => onSelect(project)}>
                <FolderOpen className="mr-2 size-4" />
                打开项目
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onEdit(project)}>
                <Edit className="mr-2 size-4" />
                编辑
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => onDelete(project)}
                className="text-rose-600 focus:text-rose-600"
              >
                <Trash2 className="mr-2 size-4" />
                删除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent onClick={() => onSelect(project)}>
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-2">
            {getStatusBadge(project.status)}
          </div>
          <span className="text-muted-foreground">更新于 {formatDate(project.updatedAt)}</span>
        </div>
      </CardContent>
    </Card>
  )
}
