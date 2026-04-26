import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const GET_MODEL_RLS_POLICY = gql`
  query ModelRLSPolicy($modelId: ID!) {
    modelRLSPolicy(modelId: $modelId) {
      modelId
      selectPredicate
      insertCheck
      updatePredicate
      updateCheck
      deletePredicate
      preset
      createdAt
      updatedAt
    }
  }
`

export const GET_PROJECT_AUTH_SCHEMA = gql`
  query GetProjectAuthSchema {
    projectAuthSchema {
      variables {
        name
        source
        type
      }
    }
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

export const SET_MODEL_RLS_POLICY = gql`
  mutation SetModelRLSPolicy($input: SetModelRLSPolicyInput!) {
    setModelRLSPolicy(input: $input) {
      policy {
        modelId
        selectPredicate
        insertCheck
        updatePredicate
        updateCheck
        deletePredicate
        preset
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on ModelHasNoOwnerField {
          message
          suggestion
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
    }
  }
`

export const VALIDATE_RLS_EXPR = gql`
  mutation ValidateRLSExpr($input: ValidateRLSExprInput!) {
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
        ... on ModelNotFound {
          message
        }
        ... on InvalidRLSExpression {
          message
          suggestion
          path
        }
      }
    }
  }
`

export const SET_PROJECT_AUTH_SCHEMA = gql`
  mutation SetProjectAuthSchema($input: SetProjectAuthSchemaInput!) {
    setProjectAuthSchema(input: $input) {
      authSchema {
        variables {
          name
          source
          type
        }
      }
      error {
        __typename
        ... on ProjectNotFound {
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
