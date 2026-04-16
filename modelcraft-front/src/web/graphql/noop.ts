import { gql } from '@apollo/client'

export const NOOP_QUERY = gql`
  query NoopQuery {
    __typename
  }
`

export const NOOP_MUTATION = gql`
  mutation NoopMutation {
    __typename
  }
`
