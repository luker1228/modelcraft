import { gql } from '@apollo/client'

export const GET_RLS_POLICIES = gql`
  query GetRlsPolicies($modelId: ID!, $orderBy: RlsPoliciesOrderBy) {
    rlsPolicies(modelId: $modelId, orderBy: $orderBy) {
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

export const VALIDATE_RLS_EXPR = gql`
  mutation ValidateRlsExpr($input: ValidateRLSExprInput!) {
    validateRLSExpr(input: $input) {
      result {
        valid
        errors {
          path
          message
          code
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
        }
        ... on InvalidRLSExpression {
          message
          suggestion
          path
        }
        ... on InvalidAuthVariable {
          message
          suggestion
          variable
        }
      }
      dryRun {
        sql
        params
        result
      }
    }
  }
`
