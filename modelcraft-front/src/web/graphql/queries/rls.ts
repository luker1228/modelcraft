import { gql } from '@apollo/client'

/**
 * 获取 Model RLS 策略
 */
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

/**
 * 获取当前项目认证变量配置（project endpoint）
 */
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
