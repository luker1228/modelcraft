import { gql } from '@apollo/client'

// Update organization display name
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
        ... on OrganizationNotFound {
          message
        }
      }
    }
  }
`

// Create a custom role
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

// Delete a custom role
export const DELETE_ROLE = gql`
  mutation DeleteRole($id: ID!) {
    deleteRole(id: $id) {
      success
      error {
        __typename
        ... on RoleNotFound {
          message
        }
        ... on CannotDeleteSystemRole {
          message
        }
      }
    }
  }
`

// Add permission to role (permission management domain)
export const ADD_PERMISSION_TO_ROLE = gql`
  mutation AddPermissionToRole($roleId: Int!, $obj: String!, $act: String!) {
    addPermissionToRole(roleId: $roleId, obj: $obj, act: $act) {
      success
      error {
        __typename
        ... on PermissionRoleNotFound {
          message
          suggestion
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

// Remove permission from role (permission management domain)
export const REMOVE_PERMISSION_FROM_ROLE = gql`
  mutation RemovePermissionFromRole($roleId: Int!, $obj: String!, $act: String!) {
    removePermissionFromRole(roleId: $roleId, obj: $obj, act: $act) {
      success
      error {
        __typename
        ... on PermissionRoleNotFound {
          message
          suggestion
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

// Create API key
export const CREATE_API_KEY = gql`
  mutation CreateApiKey($input: CreateApiKeyInput!) {
    createApiKey(input: $input) {
      result {
        id
        name
        key
        keyPrefix
        roleIDs
        createdAt
      }
      error {
        __typename
        ... on ApiKeyLimitExceeded {
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

// Update API key
export const UPDATE_API_KEY = gql`
  mutation UpdateApiKey($id: ID!, $input: UpdateApiKeyInput!) {
    updateApiKey(id: $id, input: $input) {
      apiKey {
        id
        name
        keyPrefix
        roleIDs
        lastUsedAt
        expiresAt
        revokedAt
        createdAt
      }
      error {
        __typename
        ... on ApiKeyNotFound {
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

// Revoke API key
export const REVOKE_API_KEY = gql`
  mutation RevokeApiKey($id: ID!) {
    revokeApiKey(id: $id) {
      apiKey {
        id
        revokedAt
      }
      error {
        __typename
        ... on ApiKeyNotFound {
          message
        }
      }
    }
  }
`
