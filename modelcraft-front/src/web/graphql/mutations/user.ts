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
        ... on InvalidRoleInput {
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
