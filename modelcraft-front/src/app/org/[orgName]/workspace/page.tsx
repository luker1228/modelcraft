'use client'

import { useState, useEffect, useCallback, useMemo } from "react"
import { useRouter, useParams } from "next/navigation"
import { useQuery, useMutation } from "@apollo/client"
import { Button } from "@web/components/ui/button"
import { SearchInput } from "@web/components/ui/search-input"
import { Badge } from "@web/components/ui/badge"
import { Card, CardContent } from "@web/components/ui/card"
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
import { GET_PROJECTS } from "@web/graphql/queries/project"
import {
  CREATE_PROJECT,
  UPDATE_PROJECT,
  DELETE_PROJECT
} from "@web/graphql/mutations/project"
import { TEST_CLUSTER_CONNECTION } from "@web/graphql/mutations/cluster"
import {
  Plus,
  FolderOpen,
  Settings,
  MoreVertical,
  Clock,
  Database,
  Users,
  RefreshCw,
  Search
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
import { getToken } from "@bff/auth/public"

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

  const orgScopedContext = useMemo(() => {
    if (!orgName) return undefined
    return {
      uri: `/graphql/org/${orgName}/`,
    }
  }, [orgName])

  // GraphQL 查询
  const { data, loading, error: queryError, refetch } = useQuery<ProjectsQueryData>(GET_PROJECTS, {
    fetchPolicy: 'cache-and-network',
    skip: authLoading,
    context: orgScopedContext,
    onCompleted: (queryData) => {
      if (queryData?.projects) {
        setProjects(queryData.projects)
      }
    },
    onError: (error) => {
      toast.error('获取项目列表失败', {
        description: error.message || error.toString()
      })
    }
  })

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
      const suggestion = payload?.error?.suggestion
      return {
        success: false,
        message: suggestion ? `${errorMessage}: ${suggestion}` : errorMessage,
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
        <PageLayout maxWidth="full" padding="compact">
          <PageHeader title="所有项目" />

            {/* Toolbar */}
            <div className="mb-6 flex items-center justify-between gap-4">
              {/* Search */}
              <div className="max-w-md flex-1">
                <SearchInput
                  placeholder="搜索项目..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onClear={() => setSearchTerm('')}
                  clearable
                  className="border-slate-200 bg-white transition-all focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
                />
              </div>

              {/* Actions */}
              <div className="flex items-center gap-3">
                {/* View Toggle */}
                <ViewToggle value={viewMode} onValueChange={setViewMode} />

                {/* Create Project Button */}
                <Button
                  onClick={handleOpenCreateDialog}
                  size="sm"
                  className="bg-primary text-white hover:bg-primary/90"
                >
                  <Plus className="mr-1.5 size-3.5" />
                  新建项目
                </Button>
              </div>
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
                {filteredProjects.map((project) => (
                  <ProjectCard
                    key={project.id}
                    project={project}
                    onSelect={handleSelectProject}
                    onEdit={handleEditProject}
                    onDelete={handleOpenDeleteDialog}
                  />
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                {filteredProjects.map((project) => (
                  <Card
                    key={project.id}
                    className="group cursor-pointer border bg-white transition-all duration-300 hover:border-blue-100 hover:shadow-sm"
                    onClick={() => handleSelectProject(project)}
                  >
                    <CardContent className="flex items-center justify-between p-4">
                      <div className="flex flex-1 items-center gap-4">
                        <div className="flex size-10 items-center justify-center rounded-xl bg-blue-500">
                          <Database className="size-4 text-white" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="mb-1 flex items-center gap-2">
                            <span className="font-semibold text-foreground">{project.title}</span>
                            <span className="font-mono text-xs text-muted-foreground">/ {project.slug}</span>
                            <Badge
                              variant={project.status === 'ACTIVE' ? 'default' : 'secondary'}
                              className={`text-xs ${project.status === 'ACTIVE' ? 'border-0 bg-emerald-100 text-emerald-700' : ''}`}
                            >
                              {project.status === 'ACTIVE' ? '活跃' : '已归档'}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground">{project.description}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Clock className="size-3.5" />
                          <span>{formatDateSafe(project.updatedAt)}</span>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="size-8 p-0 opacity-0 transition-opacity hover:bg-selected group-hover:opacity-100"
                            >
                              <MoreVertical className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="border-slate-200 bg-white">
                            <DropdownMenuItem onClick={(e) => { e.stopPropagation(); handleEditProject(project) }}>
                              <Settings className="mr-2 size-4" />
                              设置
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem className="text-rose-600" onClick={(e) => { e.stopPropagation(); handleOpenDeleteDialog(project) }}>删除</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      </CardContent>
                    </Card>
                ))}
              </div>
            )}

            {/* Empty State */}
            {filteredProjects.length === 0 && !loading && (
              <Card className="border-2 border-dashed border-slate-300 bg-white">
                <CardContent className="p-16 text-center">
                  <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-blue-100">
                    <Search className="size-8 text-blue-600" />
                  </div>
                  <p className="mb-2 text-lg font-semibold text-foreground">未找到匹配的项目</p>
                  <p className="mb-6 text-sm text-muted-foreground">尝试调整搜索条件或创建新项目</p>
                  <Button
                    className="h-10 bg-primary font-semibold text-white hover:bg-primary/90"
                    onClick={handleOpenCreateDialog}
                  >
                    <Plus className="mr-2 size-4" />
                    创建项目
                  </Button>
                </CardContent>
              </Card>
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
