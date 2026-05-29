'use client'

import { useState, useEffect, useCallback, useRef } from "react"
import { useRouter, useParams } from "next/navigation"
import { useQuery, useMutation } from "@apollo/client"
import { Button } from "@web/components/ui/button"
import { SearchInput } from "@web/components/ui/search-input"
import { Badge } from "@web/components/ui/badge"
import { Skeleton } from "@web/components/ui/skeleton"
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
import { useRegisterAICapability } from "@web/hooks/ai/use-register-ai-capability"

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

  // AI capability refs for chip highlighting
  const createProjectBtnRef = useRef<HTMLButtonElement>(null)
  useRegisterAICapability('create_project', '新建项目', createProjectBtnRef, '点击打开新建项目表单')

  useEffect(() => {
    if (pendingAction === 'nav_create_project') {
      setDialogOpen(true)
      setPendingAction(null)
    } else if (pendingAction === 'highlight_first_project') {
      setHighlightFirstProject(true)
      setPendingAction(null)
      // No auto-clear — stays until user clicks the project card
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
    router.push(`/org/${org.orgName}/dashboard`)
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
          />

            {/* Toolbar */}
            <div className="mb-6 flex items-start gap-3">
              <div className="max-w-xs flex-1 space-y-2">
                <SearchInput
                  placeholder="搜索项目..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onClear={() => setSearchTerm('')}
                  clearable
                />
                <div className="shrink-0">
                  <Button
                    ref={createProjectBtnRef}
                    onClick={handleOpenCreateDialog}
                    size="sm"
                  >
                    <Plus className="mr-1.5 size-3.5" />
                    新建项目
                  </Button>
                </div>
              </div>
              <ViewToggle value={viewMode} onValueChange={setViewMode} />
            </div>

            {/* Projects Grid/List */}
            {loading && projectsToDisplay.length === 0 ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="rounded-lg border border-border bg-card p-5">
                    <div className="mb-3 flex items-start justify-between">
                      <Skeleton className="h-5 w-2/5" />
                      <Skeleton className="size-5 rounded-md" />
                    </div>
                    <Skeleton className="mb-2 h-3.5 w-1/3" />
                    <Skeleton className="h-3.5 w-full" />
                    <Skeleton className="mt-1.5 h-3.5 w-4/5" />
                    <div className="mt-4 flex items-center justify-between border-t border-border pt-3">
                      <Skeleton className="h-5 w-12 rounded-full" />
                      <Skeleton className="h-3.5 w-20" />
                    </div>
                  </div>
                ))}
              </div>
            ) : viewMode === 'grid' ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {filteredProjects.map((project, index) => {
                  const isSpotlit = index === 0 && highlightFirstProject
                  return (
                    <div
                      key={project.id}
                      className={isSpotlit ? 'relative z-50' : undefined}
                    >
                      <ProjectCard
                        project={project}
                        onSelect={(p) => {
                          if (isSpotlit) setHighlightFirstProject(false)
                          handleSelectProject(p)
                        }}
                        onEdit={handleEditProject}
                        onDelete={handleOpenDeleteDialog}
                      />
                      {isSpotlit && (
                        <div className="pointer-events-none absolute -bottom-9 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md bg-white px-3 py-1.5 text-[12px] font-semibold text-foreground shadow-lg ring-1 ring-border">
                          👆 点击进入项目
                        </div>
                      )}
                    </div>
                  )
                })}
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

            {/* Empty State — no projects exist yet */}
            {filteredProjects.length === 0 && !loading && !searchTerm && (
              <div className="flex flex-col items-center justify-center py-24 text-center">
                <div className="mb-5 flex size-16 items-center justify-center rounded-2xl border border-border bg-card shadow-sm">
                  <FolderOpen className="size-7 text-muted-foreground/50" strokeWidth={1.5} />
                </div>
                <p className="text-[15px] font-semibold text-foreground">还没有项目</p>
                <p className="mt-1.5 max-w-xs text-[13px] leading-relaxed text-muted-foreground">
                  创建第一个项目，开始配置数据模型、权限策略和接口发布。
                </p>
                <Button className="mt-6" onClick={handleOpenCreateDialog}>
                  <Plus className="mr-1.5 size-4" />
                  新建项目
                </Button>
              </div>
            )}

            {/* Empty State — search found nothing */}
            {filteredProjects.length === 0 && !loading && searchTerm && (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <FolderOpen className="mb-4 size-9 text-muted-foreground/30" strokeWidth={1.5} />
                <p className="text-[14px] font-medium text-foreground">未找到匹配的项目</p>
                <p className="mt-1 text-[13px] text-muted-foreground">尝试调整搜索条件</p>
              </div>
            )}
        </PageLayout>
      </AppLayout>

      {/* Tutorial spotlight overlay */}
      {highlightFirstProject && (
        <div
          className="fixed inset-0 z-40 bg-black/50"
          onClick={() => setHighlightFirstProject(false)}
        />
      )}

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
