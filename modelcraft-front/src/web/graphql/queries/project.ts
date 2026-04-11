import { gql } from '@apollo/client'

// 列出数据库表（排除已导入的表，支持分页）
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

// 获取项目列表
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

// 获取单个项目
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
