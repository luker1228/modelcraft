import { gql } from '@apollo/client'

// ── Permission CRUD ────────────────────────────────────────────────────────────

/**
 * 创建终端用户权限点
 */
export const CREATE_END_USER_PERMISSION = gql`
  mutation CreateEndUserPermission($projectSlug: String!, $input: CreateEndUserPermissionInput!) {
    createEndUserPermission(projectSlug: $projectSlug, input: $input) {
      permission {
        id
        modelId
        modelDisplayName
        action
        rowScope
        displayName
        description
        columnPolicy {
          defaultMode
          rules { fieldName mode maskPattern }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 删除终端用户权限点
 */
export const DELETE_END_USER_PERMISSION = gql`
  mutation DeleteEndUserPermission($projectSlug: String!, $id: ID!) {
    deleteEndUserPermission(projectSlug: $projectSlug, id: $id) {
      success
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

// ── Bundle CRUD ────────────────────────────────────────────────────────────────

/**
 * 创建权限包
 */
export const CREATE_END_USER_BUNDLE = gql`
  mutation CreateEndUserBundle($projectSlug: String!, $input: CreateEndUserBundleInput!) {
    createEndUserBundle(projectSlug: $projectSlug, input: $input) {
      bundle {
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
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 更新权限包基本信息
 */
export const UPDATE_END_USER_BUNDLE = gql`
  mutation UpdateEndUserBundle($projectSlug: String!, $id: ID!, $input: UpdateEndUserBundleInput!) {
    updateEndUserBundle(projectSlug: $projectSlug, id: $id, input: $input) {
      bundle {
        id
        name
        description
        updatedAt
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 删除权限包
 */
export const DELETE_END_USER_BUNDLE = gql`
  mutation DeleteEndUserBundle($projectSlug: String!, $id: ID!) {
    deleteEndUserBundle(projectSlug: $projectSlug, id: $id) {
      success
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on BundleInUse {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 向权限包添加权限点
 */
export const ADD_PERMISSION_TO_BUNDLE = gql`
  mutation AddPermissionToBundle($projectSlug: String!, $bundleId: ID!, $permissionId: ID!) {
    addPermissionToBundle(projectSlug: $projectSlug, bundleId: $bundleId, permissionId: $permissionId) {
      bundle {
        id
        name
        permissions {
          id
          modelId
          modelDisplayName
          action
          rowScope
        }
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 从权限包移除权限点
 */
export const REMOVE_PERMISSION_FROM_BUNDLE = gql`
  mutation RemovePermissionFromBundle($projectSlug: String!, $bundleId: ID!, $permissionId: ID!) {
    removePermissionFromBundle(projectSlug: $projectSlug, bundleId: $bundleId, permissionId: $permissionId) {
      bundle {
        id
        name
        permissions {
          id
          modelId
          modelDisplayName
          action
          rowScope
        }
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
      }
    }
  }
`

// ── Role CRUD ──────────────────────────────────────────────────────────────────

/**
 * 创建终端用户角色
 */
export const CREATE_END_USER_ROLE = gql`
  mutation CreateEndUserRole($projectSlug: String!, $input: CreateEndUserRoleInput!) {
    createEndUserRole(projectSlug: $projectSlug, input: $input) {
      role {
        id
        name
        description
        isImplicit
        permissionBundles {
          id
          name
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 删除终端用户角色（隐式角色不可删除）
 */
export const DELETE_END_USER_ROLE = gql`
  mutation DeleteEndUserRole($projectSlug: String!, $id: ID!) {
    deleteEndUserRole(projectSlug: $projectSlug, id: $id) {
      success
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on ImplicitRoleDeletion {
          message
          suggestion
        }
        ... on RoleInUse {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 为角色添加权限包
 */
export const ASSIGN_BUNDLE_TO_ROLE = gql`
  mutation AssignBundleToRole($projectSlug: String!, $roleId: ID!, $bundleId: ID!) {
    assignBundleToRole(projectSlug: $projectSlug, roleId: $roleId, bundleId: $bundleId) {
      role {
        id
        name
        permissionBundles {
          id
          name
          description
        }
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 从角色移除权限包
 */
export const REVOKE_BUNDLE_FROM_ROLE = gql`
  mutation RevokeBundleFromRole($projectSlug: String!, $roleId: ID!, $bundleId: ID!) {
    revokeBundleFromRole(projectSlug: $projectSlug, roleId: $roleId, bundleId: $bundleId) {
      role {
        id
        name
        permissionBundles {
          id
          name
          description
        }
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
      }
    }
  }
`

// ── User Assignment ────────────────────────────────────────────────────────────

/**
 * 为终端用户分配角色
 */
export const ASSIGN_END_USER_ROLE_TO_USER = gql`
  mutation AssignEndUserRoleToUser($projectSlug: String!, $endUserId: ID!, $roleId: ID!) {
    assignEndUserRoleToUser(projectSlug: $projectSlug, endUserId: $endUserId, roleId: $roleId) {
      assignment {
        endUserId
        role {
          id
          name
        }
        assignedAt
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 撤销终端用户的角色
 */
export const REVOKE_END_USER_ROLE_FROM_USER = gql`
  mutation RevokeEndUserRoleFromUser($projectSlug: String!, $endUserId: ID!, $roleId: ID!) {
    revokeEndUserRoleFromUser(projectSlug: $projectSlug, endUserId: $endUserId, roleId: $roleId) {
      success
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 直接为终端用户授予权限包
 */
export const ASSIGN_BUNDLE_TO_END_USER = gql`
  mutation AssignBundleToEndUser($projectSlug: String!, $endUserId: ID!, $bundleId: ID!) {
    assignBundleToEndUser(projectSlug: $projectSlug, endUserId: $endUserId, bundleId: $bundleId) {
      assignment {
        endUserId
        bundle {
          id
          name
        }
        grantedAt
      }
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

/**
 * 撤销终端用户的直接权限包授权
 */
export const REVOKE_BUNDLE_FROM_END_USER = gql`
  mutation RevokeBundleFromEndUser($projectSlug: String!, $endUserId: ID!, $bundleId: ID!) {
    revokeBundleFromEndUser(projectSlug: $projectSlug, endUserId: $endUserId, bundleId: $bundleId) {
      success
      error {
        __typename
        ... on NotFound {
          message
          suggestion
        }
      }
    }
  }
`
