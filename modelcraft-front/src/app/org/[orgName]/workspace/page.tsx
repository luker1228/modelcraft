'use client'

import { useState, useEffect, useCallback } from "react"
import { useRouter, useParams } from "next/navigation"
import { useQuery, useMutation } from "@apollo/client"
import { Button } from "@web/components/ui/button"
import { SearchInput } from "@web/components/ui/search-input"
import { Badge } from "@web/components/ui/badge"
import { ViewToggle, type ViewMode } from "@web/components/ui/view-toggle"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@web/components/ui/dropdown-menu"
import { ProjectCard, ProjectDialog, DeleteProjectDialog } from "@web/components/features/project"
import { AppLayout } from "@web/components/features/layout/AppLayout"
import { PageLayout, PageHeader } from "@web/components/features/layout"
import { useProjectStore } from "@web/stores/project"
import { useAppStore } from "@web/stores"
import { useOrganizationStore } from "@shared/stores/organization"
import { useRequireAuth } from "@web/hooks/auth/use-auth"
import { GET_PROJECTS } from "@/api-client/project"
import {
  CREATE_PROJECT,
  UPDATE_PROJECT,
  DELETE_PROJECT
} from "@/api-client/project"
import { TEST_CLUSTER_CONNECTION } from "@/api-client/cluster"
import {
  Plus,
  FolderOpen,
  Settings,
  CheckCircle2,
  Circle,
  Database,
  LayoutTemplate,
  Users,
  ShieldCheck,
  ArrowRight,
  MoreVertical,
  Clock,
} from "lucide-react"
import { toast } from "sonner"
import { saveUserPreferences } from "@web/routing/smart-redirect"
import { cn, formatDateSafe } from "@/shared/utils"
import type { Project } from "@/types"
import type { DatabaseConnectionInfo } from "@/types/cluster"
import type {
  ProjectFormValues,
  ProjectConnectionTestResult,
} from "@web/components/features/project/ProjectDialog"
import { getToken } from "@api-client/auth/public"
import { useOrgScopedContext } from "@api-client/apollo/public"
import { useOnboarding } from "@shared/onboarding/OnboardingContext"

/** 将后端英文数据库错误信息本地化为中文 */
function localizeDbError(msg: string): string {
  const lower = msg.toLowerCase()
  if (lower.includes('connection refused')) return '连接被拒绝，请检查主机地址和端口'
  if (lower.includes('authentication failed') || lower.includes('access denied')) return '认证失败，请检查用户名和密码'
  if (lower.includes('unknown host') || lower.includes('no such host')) return '主机地址无法解析，请检查主机名'
  if (lower.includes('timeout') || lower.includes('timed out')) return '连接超时，请检查网络和主机配置'
  if (lower.includes('数据库连接失败') || lower.includes('connect:')) {
    // 截取 "Please verify..." 之前的部分以去除英文提示
    const colonIdx = msg.indexOf(': Please')
    if (colonIdx > 0) return '数据库连接失败，请检查主机地址、端口和账号信息'
  }
  return msg
}

// Membership info from API response
interface MembershipInfo {
  orgId: string
  orgName: string
  displayName: string
  role: string
  joinedAt: string
}

interface ProjectsQueryData {
  projects: Project[]
}

interface ProjectPayloadError {
  message: string
  suggestion?: string
}

interface CreateProjectResult {
  createProject?: {
    error?: ProjectPayloadError
    project?: Project
  }
}

interface UpdateProjectResult {
  updateProject?: {
    error?: ProjectPayloadError
    project?: Project
  }
}

interface DeleteProjectResult {
  deleteProject?: {
    error?: ProjectPayloadError
    success?: boolean
  }
}

interface TestDatabaseConnectionResult {
  testDatabaseConnection?: {
    success: boolean
    connectionTime?: number
    error?: ProjectPayloadError
  }
}

export default function WorkspacePage() {
  const router = useRouter()
  const params = useParams()
  const orgName = params?.orgName as string
  const { isLoading: authLoading, user } = useRequireAuth()

  // 状态管理
  const [searchTerm, setSearchTerm] = useState("")
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingProject, setEditingProject] = useState<Project | null>(null)
  const [deletingProject, setDeletingProject] = useState<Project | null>(null)
  const [mutationLoading, setMutationLoading] = useState(false)
  const [currentOrg, setCurrentOrgInfo] = useState<MembershipInfo | null>(null)
  const [memberships, setMemberships] = useState<MembershipInfo[]>([])
  const [viewMode, setViewMode] = useState<ViewMode>('grid')

  // Onboarding: auto-open create dialog if triggered from panel
  const { pendingAction, setPendingAction, syncProjects } = useOnboarding()
  const [highlightFirstProject, setHighlightFirstProject] = useState(false)

  useEffect(() => {
    if (pendingAction === 'create_project') {
      setDialogOpen(true)
      setPendingAction(null)
    } else if (pendingAction === 'highlight_first_project') {
      setHighlightFirstProject(true)
      setPendingAction(null)
      // Auto-clear highlight after 3s
      setTimeout(() => setHighlightFirstProject(false), 3000)
    }
  }, [pendingAction, setPendingAction])

  // Store
  const { projects, setProjects, setLoading } = useProjectStore()
  const { setSelectedProject } = useAppStore()

  // 组织上下文 — org access already verified by OrgLayout
  const currentOrgName = useOrganizationStore((state) => state.currentOrg)
  const setCurrentOrg = useOrganizationStore((state) => state.setCurrentOrg)
  const loadMembershipsStore = useOrganizationStore((state) => state.loadMemberships)

  // Sync current org info from store memberships
  useEffect(() => {
    if (authLoading) return
    const token = getToken()
    if (!token) return
    loadMembershipsStore(token, false).then((userMemberships: MembershipInfo[]) => {
      setMemberships(userMemberships)
      const matchingOrg = userMemberships.find((m: MembershipInfo) => m.orgName === orgName)
      if (matchingOrg) {
        setCurrentOrgInfo(matchingOrg)
        if (currentOrgName !== orgName) setCurrentOrg(orgName)
      }
    }).catch((err: unknown) => {
      console.error('[WorkspacePage] Failed to load memberships:', err)
    })
  }, [authLoading, orgName, currentOrgName, setCurrentOrg, loadMembershipsStore])

  const orgScopedContext = useOrgScopedContext(orgName ?? undefined)

  // GraphQL 查询
  const { data, loading, error: queryError, refetch } = useQuery<ProjectsQueryData>(GET_PROJECTS, {
    fetchPolicy: 'cache-and-network',
    skip: authLoading,
    context: orgScopedContext,
  })

  // 用 useEffect 监听 data/error 变化，替代废弃的 onCompleted/onError 回调
  useEffect(() => {
    if (data?.projects) {
      setProjects(data.projects)
      syncProjects(data.projects.map((p: { slug: string }) => ({ slug: p.slug })))
    }
  }, [data, setProjects, syncProjects])

  useEffect(() => {
    if (queryError) {
      toast.error('获取项目列表失败', {
        description: queryError.message || queryError.toString()
      })
    }
  }, [queryError])

  // GraphQL 变更
  const [createProjectMutation] = useMutation<CreateProjectResult>(CREATE_PROJECT, {
    refetchQueries: [{ query: GET_PROJECTS }],
    context: orgScopedContext,
  })

  const [updateProjectMutation] = useMutation<UpdateProjectResult>(UPDATE_PROJECT, {
    refetchQueries: [{ query: GET_PROJECTS }],
    context: orgScopedContext,
  })

  const [deleteProjectMutation] = useMutation<DeleteProjectResult>(DELETE_PROJECT, {
    refetchQueries: [{ query: GET_PROJECTS }],
    context: orgScopedContext,
  })

  const [testDatabaseConnectionMutation] = useMutation<TestDatabaseConnectionResult>(
    TEST_CLUSTER_CONNECTION,
    {
      context: orgScopedContext,
    },
  )

  useEffect(() => {
    setLoading(loading)
  }, [loading, setLoading])

  const projectsToDisplay: Project[] = data?.projects || projects

  const filteredProjects = projectsToDisplay.filter((project) =>
    project.title?.toLowerCase().includes(searchTerm.trim().toLowerCase()) ||
    project.description?.toLowerCase().includes(searchTerm.trim().toLowerCase())
  )

  const activeProjects = projectsToDisplay.filter(p => p.status === 'ACTIVE')

  // 选择项目并跳转
  const handleSelectProject = useCallback((project: Project) => {
    console.log('[WorkspacePage] Project selected:', {
      projectId: project.id,
      projectSlug: project.slug,
      projectTitle: project.title,
      orgName,
      targetUrl: `/org/${orgName}/project/${project.slug}`
    })

    setSelectedProject(project)

    // 保存用户选择
    if (currentOrg) {
      console.log('[WorkspacePage] Saving user preferences:', {
        orgId: currentOrg.orgId,
        projectSlug: project.slug
      })
      saveUserPreferences(currentOrg.orgId, project.slug)
    }

    console.log('[WorkspacePage] Navigating to:', `/org/${orgName}/project/${project.slug}`)
    router.push(`/org/${orgName}/project/${project.slug}`)
  }, [setSelectedProject, router, orgName, currentOrg])

  const handleEditProject = useCallback((project: Project) => {
    setEditingProject(project)
    setDialogOpen(true)
  }, [])

  const handleOpenDeleteDialog = useCallback((project: Project) => {
    setDeletingProject(project)
    setDeleteDialogOpen(true)
  }, [])

  // 切换组织
  const handleOrgSwitch = useCallback((org: MembershipInfo) => {
    setCurrentOrgInfo(org)
    localStorage.setItem('lastSelectedOrgId', org.orgId)
    localStorage.setItem('defaultOrgName', org.orgName)
    router.push(`/org/${org.orgName}/workspace`)
  }, [router])

  if (authLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-center">
          <div className="mx-auto size-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
          <p className="mt-3 text-sm text-muted-foreground">加载工作空间...</p>
        </div>
      </div>
    )
  }

  const handleSubmitProject = async (formData: ProjectFormValues) => {
    setMutationLoading(true)
    try {
      if (editingProject) {
        const { data: resultData } = await updateProjectMutation({
          variables: { input: { ...formData, slug: editingProject.slug } },
        })

        if (resultData?.updateProject?.error) {
          toast.error('更新项目失败', {
            description: resultData.updateProject.error.message
          })
          return
        }

        toast.success('项目更新成功')
      } else {
        const { data: resultData } = await createProjectMutation({
          variables: { input: formData },
        })

        if (resultData?.createProject?.error) {
          const errorMessage = resultData.createProject.error.message
          const suggestion = resultData.createProject.error.suggestion
          toast.error('创建项目失败', {
            description: suggestion ? `${errorMessage}: ${suggestion}` : errorMessage
          })
          return
        }

        toast.success('项目创建成功')
      }

      setDialogOpen(false)
      setEditingProject(null)
    } catch (err) {
      toast.error('操作失败', {
        description: err instanceof Error ? err.message : '操作失败'
      })
    } finally {
      setMutationLoading(false)
    }
  }

  const handleTestProjectConnection = async (
    connectionInfo: DatabaseConnectionInfo,
  ): Promise<ProjectConnectionTestResult> => {
    try {
      const { data: resultData } = await testDatabaseConnectionMutation({
        variables: {
          input: {
            connectionInfo,
          },
        },
      })

      const payload = resultData?.testDatabaseConnection
      if (payload?.success) {
        const connectionTime = Number(payload.connectionTime ?? 0).toFixed(2)
        return {
          success: true,
          message: `连接成功！耗时 ${connectionTime}ms`,
        }
      }

      const errorMessage = payload?.error?.message ?? "连接失败"
      return {
        success: false,
        message: localizeDbError(errorMessage),
      }
    } catch (err) {
      return {
        success: false,
        message: err instanceof Error ? err.message : "连接测试失败",
      }
    }
  }

  const handleDeleteProject = async () => {
    if (!deletingProject) return

    setMutationLoading(true)
    try {
      const { data: resultData } = await deleteProjectMutation({
        variables: { slug: deletingProject.slug },
      })

      if (resultData?.deleteProject?.error) {
        toast.error('删除项目失败', {
          description: resultData.deleteProject.error.message
        })
        return
      }

      toast.success('项目删除成功')
      setDeleteDialogOpen(false)
      setDeletingProject(null)
    } catch (err) {
      toast.error('删除失败', {
        description: err instanceof Error ? err.message : '删除失败'
      })
    } finally {
      setMutationLoading(false)
    }
  }

  const handleOpenCreateDialog = () => {
    setEditingProject(null)
    setDialogOpen(true)
  }

  const handleCloseDialog = (open: boolean) => {
    if (!open) setEditingProject(null)
    setDialogOpen(open)
  }

  const handleRefresh = () => {
    refetch()
  }

  return (
    <>
      <AppLayout pageTitle="所有项目">
        <PageLayout maxWidth="7xl">
          <PageHeader
            title="所有项目"
            spacing="compact"
            actions={
              <Button
                onClick={handleOpenCreateDialog}
                size="sm"
              >
                <Plus className="mr-1.5 size-3.5" />
                新建项目
              </Button>
            }
          />

            {/* Toolbar */}
            <div className="mb-6 flex items-center gap-3">
              <div className="max-w-xs flex-1">
                <SearchInput
                  placeholder="搜索项目..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onClear={() => setSearchTerm('')}
                  clearable
                />
              </div>
              <ViewToggle value={viewMode} onValueChange={setViewMode} />
            </div>

            {/* Projects Grid/List */}
            {loading && projectsToDisplay.length === 0 ? (
              <div className="flex items-center justify-center py-16">
                <div className="text-center">
                  <div className="mx-auto mb-3 size-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
                  <p className="text-sm text-muted-foreground">加载项目...</p>
                </div>
              </div>
            ) : viewMode === 'grid' ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {filteredProjects.map((project, index) => (
                  <div
                    key={project.id}
                    className={index === 0 && highlightFirstProject
                      ? 'rounded-lg ring-2 ring-amber-400 ring-offset-2 transition-all duration-300'
                      : undefined
                    }
                  >
                    <ProjectCard
                      project={project}
                      onSelect={handleSelectProject}
                      onEdit={handleEditProject}
                      onDelete={handleOpenDeleteDialog}
                    />
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-1.5">
                {filteredProjects.map((project) => (
                  <div
                    key={project.id}
                    className="group flex cursor-pointer items-center gap-4 rounded-lg border border-border bg-card px-4 py-3 transition-colors duration-150 hover:border-foreground/20"
                    onClick={() => handleSelectProject(project)}
                  >
                    {/* Icon */}
                    <div className="flex size-8 flex-shrink-0 items-center justify-center rounded-md bg-primary/[0.08]">
                      <FolderOpen className="size-4 text-primary" strokeWidth={1.5} />
                    </div>

                    {/* Info */}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-[14px] font-medium text-foreground">{project.title}</span>
                        <span className="font-mono text-[12px] text-muted-foreground/60">{project.slug}</span>
                        <Badge
                          variant={project.status === 'ACTIVE' ? 'success' : 'secondary'}
                        >
                          {project.status === 'ACTIVE' ? '活跃' : '已归档'}
                        </Badge>
                      </div>
                      {project.description && (
                        <p className="mt-0.5 line-clamp-1 text-[13px] text-muted-foreground">{project.description}</p>
                      )}
                    </div>

                    {/* Meta */}
                    <div className="flex flex-shrink-0 items-center gap-3">
                      <div className="flex items-center gap-1 text-[12px] text-muted-foreground">
                        <Clock className="size-3.5" strokeWidth={1.5} />
                        <span>{formatDateSafe(project.updatedAt)}</span>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 opacity-0 group-hover:opacity-100"
                          >
                            <MoreVertical className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={(e) => { e.stopPropagation(); handleEditProject(project) }}>
                            <Settings className="mr-2 size-4" />
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive"
                            onClick={(e) => { e.stopPropagation(); handleOpenDeleteDialog(project) }}
                          >
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Empty State */}
            {filteredProjects.length === 0 && !loading && (
              searchTerm ? (
                /* 搜索无结果 - 简单提示 */
                <div className="flex flex-col items-center justify-center py-20">
                  <FolderOpen className="mb-4 size-10 text-muted-foreground/30" strokeWidth={1.5} />
                  <p className="text-[14px] font-medium text-foreground">未找到匹配的项目</p>
                  <p className="mt-1 text-[13px] text-muted-foreground">尝试调整搜索条件</p>
                </div>
              ) : (
                /* 首次使用 Onboarding 引导 */
                <div className="mx-auto mt-4 w-full max-w-2xl">
                  <div className="mb-6 text-center">
                    <h2 className="text-lg font-semibold text-foreground">欢迎使用 ModelCraft</h2>
                    <p className="mt-1 text-sm text-muted-foreground">
                      按以下步骤开始，大约需要 5 分钟完成基础配置
                    </p>
                  </div>

                  <div className="space-y-3">
                    {/* Step 1 - 已完成 */}
                    <div className="flex items-start gap-4 rounded-lg border border-border bg-card p-4">
                      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full bg-emerald-100">
                        <CheckCircle2 className="size-4 text-emerald-600" strokeWidth={2} />
                      </div>
                      <div className="flex-1">
                        <p className="text-sm font-medium text-muted-foreground line-through">注册账号 &amp; 创建组织</p>
                        <p className="text-xs text-muted-foreground/60">已完成</p>
                      </div>
                    </div>

                    {/* Step 2 - 当前步骤（高亮） */}
                    <div className="flex items-start gap-4 rounded-lg border-2 border-primary/30 bg-primary/5 p-4 shadow-sm">
                      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground">
                        <Database className="size-3.5" strokeWidth={2} />
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-semibold text-foreground">创建第一个项目</p>
                          <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">当前步骤</span>
                        </div>
                        <p className="mt-0.5 text-xs text-muted-foreground">
                          连接你的 MySQL 数据库，ModelCraft 会自动将数据表暴露为 API 接口
                        </p>
                        <Button size="sm" className="mt-3" onClick={handleOpenCreateDialog}>
                          <Plus className="mr-1.5 size-3.5" />
                          新建项目
                        </Button>
                      </div>
                    </div>

                    {/* Step 3 */}
                    <div className="flex items-start gap-4 rounded-lg border border-border bg-card p-4 opacity-50">
                      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full border-2 border-muted-foreground/30">
                        <LayoutTemplate className="size-3.5 text-muted-foreground/50" strokeWidth={1.5} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">设计数据模型</p>
                        <p className="text-xs text-muted-foreground/60">
                          在项目中定义字段、逻辑外键和枚举，完善数据结构
                        </p>
                      </div>
                    </div>

                    {/* Step 4 */}
                    <div className="flex items-start gap-4 rounded-lg border border-border bg-card p-4 opacity-50">
                      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full border-2 border-muted-foreground/30">
                        <Users className="size-3.5 text-muted-foreground/50" strokeWidth={1.5} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">创建终端用户</p>
                        <p className="text-xs text-muted-foreground/60">
                          在「终端用户」页面添加可访问数据的用户账号
                        </p>
                      </div>
                    </div>

                    {/* Step 5 */}
                    <div className="flex items-start gap-4 rounded-lg border border-border bg-card p-4 opacity-50">
                      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full border-2 border-muted-foreground/30">
                        <ShieldCheck className="size-3.5 text-muted-foreground/50" strokeWidth={1.5} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">分配项目权限</p>
                        <p className="text-xs text-muted-foreground/60">
                          在项目「用户授权」页面为终端用户分配角色，用户即可通过 API 访问数据
                        </p>
                      </div>
                    </div>
                  </div>

                  <p className="mt-5 text-center text-xs text-muted-foreground">
                    完成配置后，终端用户可通过
                    <span className="mx-1 font-mono font-medium text-foreground">/end-user/{orgName}/login</span>
                    登录访问数据
                    <ArrowRight className="ml-0.5 inline size-3" />
                  </p>
                </div>
              )
            )}
        </PageLayout>
      </AppLayout>

      {/* Dialogs */}
      <ProjectDialog
        open={dialogOpen}
        onOpenChange={handleCloseDialog}
        project={editingProject}
        onSubmit={handleSubmitProject}
        onTestConnection={handleTestProjectConnection}
        loading={mutationLoading}
      />

      <DeleteProjectDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        project={deletingProject}
        onConfirm={handleDeleteProject}
        loading={mutationLoading}
      />
    </>
  )
}
