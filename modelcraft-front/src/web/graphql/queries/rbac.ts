import { gql } from '@apollo/client'

/**
 * 获取项目下所有终端用户权限点
 */
export const GET_END_USER_PERMISSIONS = gql`
  query GetEndUserPermissions($projectSlug: String!) {
    endUserPermissions(projectSlug: $projectSlug) {
      id
      modelId
      modelDisplayName
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
`

/**
 * 获取项目下所有权限包列表
 */
export const GET_END_USER_BUNDLES = gql`
  query GetEndUserBundles($projectSlug: String!) {
    endUserBundles(projectSlug: $projectSlug) {
      id
      name
      description
      createdAt
      updatedAt
      permissions {
        id
        modelId
        modelDisplayName
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
`

/**
 * 获取单个权限包详情
 */
export const GET_END_USER_BUNDLE = gql`
  query GetEndUserBundle($projectSlug: String!, $id: ID!) {
    endUserBundle(projectSlug: $projectSlug, id: $id) {
      id
      name
      description
      createdAt
      updatedAt
      permissions {
        id
        modelId
        modelDisplayName
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
`

/**
 * 获取项目下所有终端用户角色列表
 */
export const GET_END_USER_ROLES = gql`
  query GetEndUserRoles($projectSlug: String!) {
    endUserRoles(projectSlug: $projectSlug) {
      id
      name
      description
      isImplicit
      createdAt
      updatedAt
      permissionBundles {
        id
        name
        description
      }
    }
  }
`

/**
 * 获取单个终端用户角色详情
 */
export const GET_END_USER_ROLE = gql`
  query GetEndUserRole($projectSlug: String!, $id: ID!) {
    endUserRole(projectSlug: $projectSlug, id: $id) {
      id
      name
      description
      isImplicit
      createdAt
      updatedAt
      permissionBundles {
        id
        name
        description
        permissions {
          id
          modelId
          modelDisplayName
          action
          rowScope
        }
      }
    }
  }
`

/**
 * 查询某个终端用户在某个 Model 上的有效权限（三通道合并）
 */
export const GET_END_USER_EFFECTIVE_PERMISSIONS = gql`
  query GetEndUserEffectivePermissions($projectSlug: String!, $endUserId: ID!) {
    endUserEffectivePermissions(projectSlug: $projectSlug, endUserId: $endUserId) {
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
  }
`
