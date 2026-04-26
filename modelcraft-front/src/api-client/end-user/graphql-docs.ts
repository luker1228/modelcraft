import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const LIST_END_USERS = gql`
  query ListEndUsers($input: ListEndUsersInput) {
    listEndUsers(input: $input) {
      connection {
        nodes {
          id
          username
          isForbidden
          createdBy
          createdAt
          updatedAt
        }
        pageInfo {
          hasNextPage
          endCursor
        }
        totalCount
      }
      error {
        __typename
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

// ── Mutations ──────────────────────────────────────────────────────────────────

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
