import { gql } from '@apollo/client'

/**
 * 设置 Model RLS 策略
 */
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

/**
 * 校验 RLS 表达式
 */
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

/**
 * 设置当前项目认证变量配置（project endpoint）
 */
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
