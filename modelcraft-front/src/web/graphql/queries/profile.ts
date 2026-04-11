import { gql } from '@apollo/client'

function getMyUserProfileQueryDocument() {
  return `
    query MyUserProfile {
      myUserProfile {
        user {
          id
          phone
          userName
          status
          createdAt
          updatedAt
          profile {
            id
            userId
            nickname
            avatarUrl
            bio
            createdAt
            updatedAt
          }
        }
        error {
          __typename
          ... on UserNotFound {
            message
          }
          ... on ProfileNotFound {
            message
          }
        }
      }
    }
  `
}

// NOTE: profile contract 尚未进入当前 schema，先以 runtime 文档方式保留实现。
export const MY_USER_PROFILE = gql(getMyUserProfileQueryDocument())
