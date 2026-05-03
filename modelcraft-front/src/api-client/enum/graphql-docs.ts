import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

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
  query GetEnum($name: String!) {
    enum(name: $name) {
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
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const GET_ENUM_REFERENCES = gql`
  query GetEnumReferences($name: String!) {
    enumReferences(name: $name)
  }
`

// ── Mutations ──────────────────────────────────────────────────────────────────

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
        ... on InvalidInput {
          message
          suggestion
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

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
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

export const DELETE_ENUM = gql`
  mutation DeleteEnum($name: String!) {
    deleteEnum(name: $name) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on CannotDeleteReferencedEnum {
          message
          suggestion
        }
      }
    }
  }
`
