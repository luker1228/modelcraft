import type { ClusterConnectionInput } from './cluster'

// 项目状态枚举
export type ProjectStatus = 'ACTIVE' | 'ARCHIVED'

export interface Project {
  id: string
  slug: string
  title: string
  description: string
  databaseName?: string
  loginUrl?: string
  status: ProjectStatus
  orgName: string
  createdAt: string
  updatedAt: string
}

export interface CreateProjectInput {
  slug: string
  title: string
  description?: string
  loginUrl?: string
  clusterInput: ClusterConnectionInput
  skipConnectionTest?: boolean
}

export interface UpdateProjectInput {
  slug: string
  title?: string
  description?: string
  loginUrl?: string
}

export interface ListProjectsInput {
  status?: ProjectStatus
}
