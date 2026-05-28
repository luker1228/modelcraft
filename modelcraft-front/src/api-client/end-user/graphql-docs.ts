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

/**
 * END_USER_PROJECTS — Org-scoped query for end-users.
 * Endpoint: /api/bff/graphql/end-user/org/{orgName}/
 * Returns all projects the current end-user has access to via role assignments.
 * Requires end-user Bearer token.
 */
export const END_USER_PROJECTS = gql`
  query EndUserProjects {
    endUserProjects {
      id
      slug
      title
      description
      status
      orgName
      createdAt
      updatedAt
    }
  }
`

/**
 * MY_PROJECTS — end-user 접근 가능한 프로젝트 목록 (myProjects = endUserProjects 신 이름).
 * Endpoint: /graphql/org/{orgName}/ (end-user token 필요)
 */
export const MY_PROJECTS = gql`
  query MyProjects {
    myProjects {
      id
      slug
      title
      description
      status
      orgName
      createdAt
      updatedAt
    }
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

/**
 * CREATE_USER — 통합 사용자 생성 (관리자 또는 일반 사용자).
 * isAdmin: true → 관리자, false → 일반 사용자.
 * 기존 CREATE_END_USER와 병존 (향후 CREATE_END_USER 대체 예정).
 */
export const CREATE_USER = gql`
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user {
        id
        username
        isForbidden
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

export const RESET_END_USER_PASSWORD = gql`
  mutation ResetEndUserPassword($input: ResetEndUserPasswordInput!) {
    resetEndUserPassword(input: $input) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on EndUserPasswordTooWeak {
          message
          suggestion
        }
        ... on BuiltinUserCannotBeDisabled {
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
