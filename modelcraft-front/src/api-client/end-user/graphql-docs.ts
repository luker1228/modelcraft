import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

/**
 * FIND_USERS — Org-scoped query.
 * Endpoint: /graphql/org/{orgName}/
 * Used by EndUserRef field selectors in both design and end_user workspaces.
 * Returns EndUserPublic (id, username, createdAt) with cursor-based pagination.
 */
export const FIND_USERS = gql`
  query FindUsers($where: UserWhereInput, $after: String, $first: Int) {
    findUsers(where: $where, after: $after, first: $first) {
      items {
        id
        username
        isBuiltin
        createdAt
      }
      nextCursor
      hasMore
      reqId
    }
  }
`

export const LIST_END_USERS = gql`
  query ListEndUsers($input: ListEndUsersInput) {
    listEndUsers(input: $input) {
      connection {
        nodes {
          id
          username
          isForbidden
          isBuiltin
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
        ... on InvalidInput {
          message
          suggestion
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
        isBuiltin
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
        ... on InvalidInput {
          message
          suggestion
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
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on InvalidInput {
          message
          suggestion
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
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`
