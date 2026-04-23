import { gql } from '@apollo/client'

// Get current authenticated user info
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

// Get organizations the current user belongs to
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

// Get members of the current organization
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

// List all available roles
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

// Permission roles (permission management domain)
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

// List permissions for a specific role
export const GET_ROLE_PERMISSIONS_LIST = gql`
  query GetRolePermissionsList($roleId: Int!) {
    rolePermissionsList(roleId: $roleId) {
      obj
      act
    }
  }
`

// List API keys in current organization scope
export const GET_API_KEYS = gql`
  query GetApiKeys {
    apiKeys {
      id
      name
      keyPrefix
      roleIDs
      lastUsedAt
      expiresAt
      revokedAt
      createdAt
    }
  }
`
