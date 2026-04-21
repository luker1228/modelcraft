import { gql } from '@apollo/client'

/**
 * 创建终端用户
 */
export const CREATE_END_USER = gql`
  mutation CreateEndUser($input: CreateEndUserInput!) {
    createEndUser(input: $input) {
      endUser {
        id
        username
        isForbidden
        createdBy
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on EndUserAlreadyExists {
          message
        }
        ... on EndUserPasswordTooWeak {
          message
          suggestion
        }
        ... on ClusterNotFound {
          message
        }
        ... on InvalidInput {
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

/**
 * 更新终端用户状态（启用/禁用）
 */
export const UPDATE_END_USER_STATUS = gql`
  mutation UpdateEndUserStatus($input: UpdateEndUserStatusInput!) {
    updateEndUserStatus(input: $input) {
      endUser {
        id
        username
        isForbidden
        updatedAt
      }
      error {
        __typename
        ... on EndUserNotFound {
          message
        }
        ... on ClusterNotFound {
          message
        }
        ... on InvalidInput {
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

/**
 * 删除终端用户
 */
export const DELETE_END_USER = gql`
  mutation DeleteEndUser($input: DeleteEndUserInput!) {
    deleteEndUser(input: $input) {
      success
      error {
        __typename
        ... on EndUserNotFound {
          message
        }
        ... on ClusterNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

/**
 * 初始化项目私有库（mc_private_{projectSlug}）
 */
export const INIT_PRIVATE_DB = gql`
  mutation InitPrivateDB {
    initPrivateDB {
      success
      error {
        __typename
        ... on InitPrivateDBError {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
