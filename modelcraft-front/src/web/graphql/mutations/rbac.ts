import { gql } from '@apollo/client'

// ── Permission CRUD ────────────────────────────────────────────────────────────

/**
 * 创建终端用户权限点
 */
export const CREATE_END_USER_PERMISSION = gql`
  mutation CreateEndUserPermission($input: CreateEndUserPermissionInput!) {
    createEndUserPermission(input: $input) {
      permission {
        id
        modelId
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
        }
        ... on ModelNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
        ... on RowScopeFieldMissing {
          message
          missingField
          requiredByRowScope
        }
      }
    }
  }
`

/**
 * 删除终端用户权限点
 */
export const DELETE_END_USER_PERMISSION = gql`
  mutation DeleteEndUserPermission($id: ID!) {
    deleteEndUserPermission(id: $id) {
      success
      error {
        __typename
        ... on EndUserPermissionNotFound {
          message
        }
        ... on EndUserPermissionInUse {
          message
        }
        ... on ProjectNotFound {
          message
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
  mutation CreateEndUserBundle($input: CreateEndUserPermissionBundleInput!) {
    createEndUserPermissionBundle(input: $input) {
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
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on EndUserPermissionBundleAlreadyExists {
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
 * 更新权限包基本信息
 */
export const UPDATE_END_USER_BUNDLE = gql`
  mutation UpdateEndUserBundle($id: ID!, $input: UpdateEndUserPermissionBundleInput!) {
    updateEndUserPermissionBundle(id: $id, input: $input) {
      bundle {
        id
        name
        description
        updatedAt
      }
      error {
        __typename
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on EndUserPermissionBundleAlreadyExists {
          message
        }
        ... on InvalidInput {
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
 * 删除权限包
 */
export const DELETE_END_USER_BUNDLE = gql`
  mutation DeleteEndUserBundle($id: ID!) {
    deleteEndUserPermissionBundle(id: $id) {
      success
      error {
        __typename
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on EndUserPermissionBundleInUse {
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
 * 向权限包添加权限点
 */
export const ADD_PERMISSION_TO_BUNDLE = gql`
  mutation AddPermissionToBundle($input: AddEndUserPermissionToBundleInput!) {
    addEndUserPermissionToBundle(input: $input) {
      bundle {
        id
        name
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
      error {
        __typename
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on EndUserPermissionNotFound {
          message
        }
        ... on InvalidInput {
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
 * 从权限包移除权限点
 */
export const REMOVE_PERMISSION_FROM_BUNDLE = gql`
  mutation RemovePermissionFromBundle($input: RemoveEndUserPermissionFromBundleInput!) {
    removeEndUserPermissionFromBundle(input: $input) {
      bundle {
        id
        name
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
      error {
        __typename
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on EndUserPermissionNotFound {
          message
        }
        ... on ProjectNotFound {
          message
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
  mutation CreateEndUserRole($input: CreateEndUserRoleInput!) {
    createEndUserRole(input: $input) {
      role {
        id
        name
        description
        isImplicit
        permissionBundles {
          bundle {
            id
            name
          }
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on EndUserRoleAlreadyExists {
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
 * 删除终端用户角色（隐式角色不可删除）
 */
export const DELETE_END_USER_ROLE = gql`
  mutation DeleteEndUserRole($id: ID!) {
    deleteEndUserRole(id: $id) {
      success
      error {
        __typename
        ... on EndUserRoleNotFound {
          message
        }
        ... on EndUserImplicitRoleCannotBeModified {
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
 * 为角色添加权限包
 */
export const ASSIGN_BUNDLE_TO_ROLE = gql`
  mutation AssignBundleToRole($input: AssignBundleToEndUserRoleInput!) {
    assignBundleToEndUserRole(input: $input) {
      role {
        id
        name
        permissionBundles {
          bundle {
            id
            name
            description
          }
        }
      }
      error {
        __typename
        ... on EndUserRoleNotFound {
          message
        }
        ... on EndUserPermissionBundleNotFound {
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
 * 从角色移除权限包
 */
export const REVOKE_BUNDLE_FROM_ROLE = gql`
  mutation RevokeBundleFromRole($input: RevokeBundleFromEndUserRoleInput!) {
    revokeBundleFromEndUserRole(input: $input) {
      role {
        id
        name
        permissionBundles {
          bundle {
            id
            name
            description
          }
        }
      }
      error {
        __typename
        ... on EndUserRoleNotFound {
          message
        }
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on ProjectNotFound {
          message
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
  mutation AssignEndUserRole($input: AssignEndUserRoleInput!) {
    assignEndUserRole(input: $input) {
      endUserId
      role {
        id
        name
      }
      error {
        __typename
        ... on EndUserNotFoundInProject {
          message
        }
        ... on EndUserRoleNotFound {
          message
        }
        ... on EndUserCannotAssignImplicitRole {
          message
        }
        ... on UserRoleAlreadyAssigned {
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
 * 撤销终端用户的角色
 */
export const REVOKE_END_USER_ROLE_FROM_USER = gql`
  mutation RevokeEndUserRole($input: RevokeEndUserRoleInput!) {
    revokeEndUserRole(input: $input) {
      success
      error {
        __typename
        ... on EndUserNotFoundInProject {
          message
        }
        ... on EndUserRoleNotFound {
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
 * 直接为终端用户授予权限包
 */
export const ASSIGN_BUNDLE_TO_END_USER = gql`
  mutation AssignBundleToEndUser($input: AssignBundleToEndUserInput!) {
    assignBundleToEndUser(input: $input) {
      endUserId
      bundle {
        id
        name
      }
      error {
        __typename
        ... on EndUserNotFoundInProject {
          message
        }
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on UserBundleAlreadyAssigned {
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
 * 撤销终端用户的直接权限包授权
 */
export const REVOKE_BUNDLE_FROM_END_USER = gql`
  mutation RevokeBundleFromEndUser($input: RevokeBundleFromEndUserInput!) {
    revokeBundleFromEndUser(input: $input) {
      success
      error {
        __typename
        ... on EndUserNotFoundInProject {
          message
        }
        ... on EndUserPermissionBundleNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
