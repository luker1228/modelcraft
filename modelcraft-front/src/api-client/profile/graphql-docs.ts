import { gql } from '@apollo/client'

// NOTE: profile contract 尚未进入当前 schema，先以 runtime 文档方式保留实现。

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

function getUpdateMyProfileMutationDocument() {
  return `
    mutation UpdateMyProfile($input: UpdateMyProfileInput!) {
      updateMyProfile(input: $input) {
        profile {
          id
          userId
          nickname
          avatarUrl
          bio
          createdAt
          updatedAt
        }
        error {
          __typename
          ... on ProfileNotFound {
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
}

export const MY_USER_PROFILE = gql(getMyUserProfileQueryDocument())

export const UPDATE_MY_PROFILE = gql(getUpdateMyProfileMutationDocument())
