import { gql } from '@apollo/client'

/**
 * 获取项目下所有终端用户权限点
 */
export const GET_END_USER_PERMISSIONS = gql`
  query GetEndUserPermissions {
    endUserPermissions(input: {}) {
      edges {
        node {
          id
          modelId
          action
          rowScope
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
          createdAt
          updatedAt
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

/**
 * 获取项目下所有权限包列表
 */
export const GET_END_USER_BUNDLES = gql`
  query GetEndUserBundles {
    endUserPermissionBundles(input: {}) {
      edges {
        node {
          id
          name
          description
          createdAt
          updatedAt
          permissions {
            sortOrder
            permission {
              id
              modelId
              action
              rowScope
              columnPolicy {
                defaultMode
                rules {
                  fieldName
                  mode
                  maskPattern
                }
              }
            }
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

/**
 * 获取单个权限包详情
 */
export const GET_END_USER_BUNDLE = gql`
  query GetEndUserBundle($id: ID!) {
    endUserPermissionBundle(id: $id) {
      id
      name
      description
      createdAt
      updatedAt
      permissions {
        sortOrder
        permission {
          id
          modelId
          action
          rowScope
          displayName
          description
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
      }
    }
  }
`

/**
 * 获取项目下所有终端用户角色列表
 */
export const GET_END_USER_ROLES = gql`
  query GetEndUserRoles {
    endUserRoles(input: { includeImplicit: true }) {
      edges {
        node {
          id
          name
          description
          isImplicit
          createdAt
          updatedAt
          permissionBundles {
            bundle {
              id
              name
              description
            }
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`

/**
 * 获取单个终端用户角色详情
 */
export const GET_END_USER_ROLE = gql`
  query GetEndUserRole($id: ID!) {
    endUserRole(id: $id) {
      id
      name
      description
      isImplicit
      createdAt
      updatedAt
      permissionBundles {
        bundle {
          id
          name
          description
          permissions {
            sortOrder
            permission {
              id
              modelId
              action
              rowScope
            }
          }
        }
      }
    }
  }
`

/**
 * 查询某个终端用户在某个 Model 上的有效权限（三通道合并）
 */
export const GET_END_USER_EFFECTIVE_PERMISSIONS = gql`
  query GetEndUserEffectivePermissions($endUserId: ID!, $modelId: ID!) {
    effectivePermissions(input: { endUserId: $endUserId, modelId: $modelId }) {
      effectivePermissions {
        endUserId
        modelId
        grants {
          action
          rowScope
          columnPolicy {
            defaultMode
            rules {
              fieldName
              mode
              maskPattern
            }
          }
        }
      }
      error {
        __typename
        ... on EndUserNotFoundInProject {
          message
        }
        ... on ModelNotFound {
          message
        }
      }
    }
  }
`
