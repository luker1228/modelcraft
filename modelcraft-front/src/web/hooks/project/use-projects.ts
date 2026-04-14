import { useQuery, useMutation } from '@apollo/client'
import { GET_PROJECTS } from '@web/graphql/queries/project'
import { CREATE_PROJECT, UPDATE_PROJECT, DELETE_PROJECT } from '@web/graphql/mutations/project'
import { useProjectStore } from '@web/stores'
import type { Project, CreateProjectInput, UpdateProjectInput } from '@/types'

// ── GraphQL response types ──────────────────────────────────────────

interface GetProjectsData {
  projects: Project[]
}

interface CreateProjectPayload {
  project: Project | null
  error: { __typename: string; message: string } | null
}

interface UpdateProjectPayload {
  project: Project | null
  error: { __typename: string; message: string } | null
}

interface DeleteProjectPayload {
  success: boolean
  error: { __typename: string; message: string } | null
}

interface CreateProjectData {
  createProject: CreateProjectPayload
}

interface UpdateProjectData {
  updateProject: UpdateProjectPayload
}

interface DeleteProjectData {
  deleteProject: DeleteProjectPayload
}

export function useProjects() {
  const { setProjects, addProject, updateProject, removeProject } = useProjectStore()

  // 查询项目列表
  const { data, loading, error, refetch } = useQuery<GetProjectsData>(GET_PROJECTS, {
    onCompleted: (queryData) => {
      if (queryData?.projects) {
        setProjects(queryData.projects)
      }
    },
  })

  // 创建项目
  const [createProjectMutation, { loading: creating }] = useMutation<CreateProjectData>(CREATE_PROJECT, {
    onCompleted: (mutationData) => {
      if (mutationData?.createProject?.project) {
        addProject(mutationData.createProject.project)
      }
    },
    refetchQueries: [{ query: GET_PROJECTS }],
  })

  // 更新项目
  const [updateProjectMutation, { loading: updating }] = useMutation<UpdateProjectData>(UPDATE_PROJECT, {
    onCompleted: (mutationData) => {
      if (mutationData?.updateProject?.project) {
        updateProject(mutationData.updateProject.project.id, mutationData.updateProject.project)
      }
    },
  })

  // 删除项目
  const [deleteProjectMutation, { loading: deleting }] = useMutation<DeleteProjectData>(DELETE_PROJECT, {
    onCompleted: (mutationData) => {
      if (mutationData?.deleteProject?.success) {
        // 项目ID需要从变量中获取
      }
    },
    refetchQueries: [{ query: GET_PROJECTS }],
  })

  const createProject = async (input: CreateProjectInput) => {
    try {
      const result = await createProjectMutation({
        variables: { input },
      })
      return result.data?.createProject
    } catch (err) {
      console.error('创建项目失败:', err)
      throw err
    }
  }

  const updateProjectById = async (slug: string, input: UpdateProjectInput) => {
    try {
      const result = await updateProjectMutation({
        variables: { input: { ...input, slug } },
      })
      return result.data?.updateProject
    } catch (err) {
      console.error('更新项目失败:', err)
      throw err
    }
  }

  const deleteProject = async (slug: string) => {
    try {
      const result = await deleteProjectMutation({
        variables: { slug },
      })
      if (result.data?.deleteProject?.success) {
        removeProject(slug)
      }
      return result.data?.deleteProject
    } catch (err) {
      console.error('删除项目失败:', err)
      throw err
    }
  }

  return {
    projects: data?.projects || [],
    loading,
    error,
    creating,
    updating,
    deleting,
    refetch,
    createProject,
    updateProject: updateProjectById,
    deleteProject,
  }
}
