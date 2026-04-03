import { gql } from '@apollo/client'

export const GET_ENUMS = gql`
  query GetEnums {
    enums {
      id
      projectSlug
      name
      displayName
      description
      isMultiSelect
      options {
        code
        label
        order
        description
      }
      createdAt
      updatedAt
    }
  }
`

export const GET_ENUM = gql`
  query GetEnum($projectSlug: String!, $name: String!) {
    enum(projectSlug: $projectSlug, name: $name) {
      enum {
        id
        projectSlug
        name
        displayName
        description
        isMultiSelect
        options {
          code
          label
          order
          description
        }
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on EnumNotFound {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`

export const GET_ENUM_REFERENCES = gql`
  query GetEnumReferences($projectSlug: String!, $name: String!) {
    enumReferences(projectSlug: $projectSlug, name: $name)
  }
`
