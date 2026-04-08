import { useQuery, useMutation } from '@apollo/client'
import { GET_PROJECTS } from '@web/graphql/queries/project'
import { CREATE_PROJECT, UPDATE_PROJECT, DELETE_PROJECT } from '@web/graphql/mutations/project'
import { useProjectStore } from '@web/stores'
import { useGraphQLErrorHandler } from '@web/hooks/error/useGraphQLErrorHandler'
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

export function useProjectsWithErrorHandling() {
  const { setProjects, addProject, updateProject, removeProject } = useProjectStore()
  const { handleError } = useGraphQLErrorHandler()

  // 查询项目列表
  const { data, loading, error, refetch } = useQuery<GetProjectsData>(GET_PROJECTS, {
    onError: (err) => {
      handleError(err, 'GetProjects', 'query')
    },
    onCompleted: (queryData) => {
      if (queryData?.projects) {
        setProjects(queryData.projects)
      }
    },
    errorPolicy: 'all', // 即使有错误也返回部分数据
  })

  // 创建项目
  const [createProjectMutation, { loading: creating }] = useMutation<CreateProjectData>(CREATE_PROJECT, {
    onError: (err) => {
      handleError(err, 'CreateProject', 'mutation')
    },
    onCompleted: (mutationData) => {
      if (mutationData?.createProject?.project) {
        addProject(mutationData.createProject.project)
      }
    },
    refetchQueries: [{ query: GET_PROJECTS }],
  })

  // 更新项目
  const [updateProjectMutation, { loading: updating }] = useMutation<UpdateProjectData>(UPDATE_PROJECT, {
    onError: (err) => {
      handleError(err, 'UpdateProject', 'mutation')
    },
    onCompleted: (mutationData) => {
      if (mutationData?.updateProject?.project) {
        updateProject(mutationData.updateProject.project.id, mutationData.updateProject.project)
      }
    },
  })

  // 删除项目
  const [deleteProjectMutation, { loading: deleting }] = useMutation<DeleteProjectData>(DELETE_PROJECT, {
    onError: (err) => {
      handleError(err, 'DeleteProject', 'mutation')
    },
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
      // 错误已经通过onError处理了
      console.error('创建项目失败:', err)
      throw err
    }
  }

  const updateProjectById = async (name: string, input: UpdateProjectInput) => {
    try {
      const result = await updateProjectMutation({
        variables: { name, input },
      })
      return result.data?.updateProject
    } catch (err) {
      // 错误已经通过onError处理了
      console.error('更新项目失败:', err)
      throw err
    }
  }

  const deleteProject = async (name: string) => {
    try {
      const result = await deleteProjectMutation({
        variables: { name },
      })
      if (result.data?.deleteProject?.success) {
        removeProject(name)
      }
      return result.data?.deleteProject
    } catch (err) {
      // 错误已经通过onError处理了
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
    createProject,
    updateProject: updateProjectById,
    deleteProject,
    refetch,
  }
}
