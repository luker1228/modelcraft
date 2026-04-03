import { gql } from '@apollo/client'

// 创建枚举
export const CREATE_ENUM = gql`
  mutation CreateEnum($input: CreateEnumInput!) {
    createEnum(input: $input) {
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
        ... on EnumAlreadyExists {
          message
          suggestion
        }
        ... on InvalidEnumInput {
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

// 更新枚举
export const UPDATE_ENUM = gql`
  mutation UpdateEnum($name: String!, $input: UpdateEnumInput!) {
    updateEnum(name: $name, input: $input) {
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
        ... on InvalidEnumInput {
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

// 删除枚举
export const DELETE_ENUM = gql`
  mutation DeleteEnum($name: String!) {
    deleteEnum(name: $name) {
      success
      error {
        __typename
        ... on EnumNotFound {
          message
        }
        ... on CannotDeleteReferencedEnum {
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
