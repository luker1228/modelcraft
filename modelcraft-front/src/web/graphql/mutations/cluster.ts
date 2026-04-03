import { gql } from '@apollo/client'

// 测试集群连接 - 支持已有集群和直接连接信息
export const TEST_CLUSTER_CONNECTION = gql`
  mutation TestClusterConnection($input: TestDatabaseConnectionInput!) {
    testDatabaseConnection(input: $input) {
      success
      connectionTime
      error {
        __typename
        ... on ClusterNotFound {
          message
        }
        ... on DatabaseConnectionFailed {
          message
          suggestion
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
