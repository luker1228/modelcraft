import { gql } from '@apollo/client'

/**
 * 获取终端用户列表
 */
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
