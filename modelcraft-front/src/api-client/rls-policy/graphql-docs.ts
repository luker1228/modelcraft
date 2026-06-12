import { gql } from '@apollo/client'

export const GET_RLS_POLICIES = gql`
  query GetRlsPolicies($modelId: ID!) {
    rlsPolicies(modelId: $modelId) {
      id
      policyName
      action
      role
      usingExpr
      withCheckExpr
      createdAt
      updatedAt
    }
  }
`

export const UPSERT_RLS_POLICY = gql`
  mutation UpsertRlsPolicy($modelId: ID!, $input: RlsPolicyInput!) {
    upsertRlsPolicy(modelId: $modelId, input: $input) {
      policy {
        id
        policyName
        action
        role
        usingExpr
        withCheckExpr
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ResourceNotFound {
          message
        }
      }
    }
  }
`

export const DELETE_RLS_POLICY = gql`
  mutation DeleteRlsPolicy($id: ID!) {
    deleteRlsPolicy(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
        }
      }
    }
  }
`
