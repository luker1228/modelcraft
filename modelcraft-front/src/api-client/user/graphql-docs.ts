import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const GET_ME = gql`
  query GetMe {
    me {
      id
      externalID
      email
      name
      organization {
        id
        name
        displayName
        ownerID
        status
        createdAt
        updatedAt
      }
      role {
        id
        name
        description
        permissions
        isSystem
        createdAt
        updatedAt
      }
      permissions
    }
  }
`

export const GET_MY_ORGANIZATIONS = gql`
  query GetMyOrganizations {
    myOrganizations {
      id
      name
      displayName
      ownerID
      status
      createdAt
      updatedAt
    }
  }
`

export const GET_ORGANIZATION_MEMBERS = gql`
  query GetOrganizationMembers {
    organizationMembers {
      id
      userID
      userName
      orgID
      role {
        id
        name
        description
        permissions
        isSystem
      }
      status
      joinedAt
      createdAt
    }
  }
`

export const GET_ROLES = gql`
  query GetRoles {
    roles {
      id
      name
      description
      permissions
      isSystem
      createdAt
      updatedAt
    }
  }
`

export const GET_PERMISSION_ROLES = gql`
  query GetPermissionRoles($orgName: String!, $includeSystem: Boolean) {
    permissionRoles(orgName: $orgName, includeSystem: $includeSystem) {
      id
      name
      description
      isSystem
      orgName
      createdAt
      updatedAt
    }
  }
`

export const GET_ROLE_PERMISSIONS_LIST = gql`
  query GetRolePermissionsList($roleId: Int!) {
    rolePermissionsList(roleId: $roleId) {
      obj
      act
    }
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

export const UPDATE_ORGANIZATION = gql`
  mutation UpdateOrganization($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      organization {
        id
        name
        displayName
        ownerID
        status
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const CREATE_ROLE = gql`
  mutation CreateRole($input: CreateRoleInput!) {
    createRole(input: $input) {
      role {
        id
        name
        description
        permissions
        isSystem
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on RoleAlreadyExists {
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

export const DELETE_ROLE = gql`
  mutation DeleteRole($id: ID!) {
    deleteRole(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on CannotDeleteSystemRole {
          message
        }
      }
    }
  }
`

export const ADD_PERMISSION_TO_ROLE = gql`
  mutation AddPermissionToRole($roleId: Int!, $obj: String!, $act: String!) {
    addPermissionToRole(roleId: $roleId, obj: $obj, act: $act) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on PermissionSystemRoleCannotBeModified {
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

export const REMOVE_PERMISSION_FROM_ROLE = gql`
  mutation RemovePermissionFromRole($roleId: Int!, $obj: String!, $act: String!) {
    removePermissionFromRole(roleId: $roleId, obj: $obj, act: $act) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on PermissionSystemRoleCannotBeModified {
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
