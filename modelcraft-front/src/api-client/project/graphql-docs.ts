import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const LIST_TABLES = gql`
  query ListTables($input: ListTablesInput!) {
    listTables(input: $input) {
      items {
        name
      }
      totalCount
    }
  }
`

export const GET_PROJECTS = gql`
  query GetProjects($input: ListProjectsInput) {
    projects(input: $input) {
      id
      slug
      title
      description
      status
      orgName
      createdAt
      updatedAt
    }
  }
`

export const GET_PROJECT = gql`
  query GetProject($slug: String!) {
    project(slug: $slug) {
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
      }
    }
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

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
        ... on InvalidInput {
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
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

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
        ... on InvalidInput {
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
