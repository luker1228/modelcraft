import { gql } from '@apollo/client'

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
          ... on InvalidProfileInput {
            message
            suggestion
          }
        }
      }
    }
  `
}

// NOTE: profile contract 尚未进入当前 schema，先以 runtime 文档方式保留实现。
export const UPDATE_MY_PROFILE = gql(getUpdateMyProfileMutationDocument())
