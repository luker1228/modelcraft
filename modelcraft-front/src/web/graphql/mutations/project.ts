import { gql } from '@apollo/client'

// 创建项目
export const CREATE_PROJECT = gql`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      project {
        id
        slug
        title
        description
        status
        orgName
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ProjectAlreadyExists {
          message
          suggestion
        }
        ... on InvalidProjectInput {
          message
          suggestion
        }
        ... on DatabaseConnectionFailed {
          message
          suggestion
        }
      }
    }
  }
`

// 更新项目
export const UPDATE_PROJECT = gql`
  mutation UpdateProject($input: UpdateProjectInput!) {
    updateProject(input: $input) {
      project {
        id
        slug
        title
        description
        status
        orgName
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ProjectNotFound {
          message
        }
        ... on InvalidProjectInput {
          message
          suggestion
        }
      }
    }
  }
`

// 删除项目
export const DELETE_PROJECT = gql`
  mutation DeleteProject($slug: String!) {
    deleteProject(slug: $slug) {
      success
      error {
        __typename
        ... on ProjectNotFound {
          message
        }
        ... on CannotDeleteDefaultProject {
          message
        }
      }
    }
  }
`

// 更新项目集群
export const UPDATE_PROJECT_CLUSTER = gql`
  mutation UpdateProjectCluster($projectSlug: String!, $input: UpdateClusterConnectionInput!) {
    updateProjectCluster(projectSlug: $projectSlug, input: $input) {
      cluster {
        id
        projectSlug
        title
        description
        connectionInfo {
          host
          port
          username
          password
        }
        status
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ClusterNotFound {
          message
        }
        ... on InvalidClusterInput {
          message
          suggestion
        }
        ... on DatabaseConnectionFailed {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

// 测试数据库连接
export const TEST_DATABASE_CONNECTION = gql`
  mutation TestDatabaseConnection($input: TestDatabaseConnectionInput!) {
    testDatabaseConnection(input: $input) {
      success
      connectionTime
      error {
        __typename
        ... on ClusterNotFound {
          message
        }
        ... on DatabaseConnectionFailed {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
