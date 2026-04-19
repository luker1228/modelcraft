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
