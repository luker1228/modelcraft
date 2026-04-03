import { gql } from '@apollo/client'

// 查询模型列表
export const GET_MODELS = gql`
  query GetModels($input: ModelQueryInput) {
    models(input: $input) {
      edges {
        node {
          id
          projectSlug
          name
          title
          description
          databaseName
          storageType
          fields {
            name
            title
            format
            schemaType
            storageHint
            nonNull
            required
            isPrimary
            isUnique
            description
            relateFkId
            belongsToFkId
            enum {
              id
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
            }
            validationConfig {
              minLength
              maxLength
              pattern
              minimum
              maximum
            }
            createdAt
            updatedAt
          }
          group {
            id
            name
            isVirtual
            displayOrder
          }
          dbTable
          createdAt
          updatedAt
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`

// 查询单个模型
export const GET_MODEL = gql`
  query GetModel($id: ID!, $withActualSchema: Boolean) {
    model(id: $id, withActualSchema: $withActualSchema) {
      model {
        id
        projectSlug
        name
        title
        description
        databaseName
        storageType
        fields {
          name
          title
          format
          schemaType
          storageHint
          nonNull
          required
          isPrimary
          isUnique
          description
          isDeprecated
          relateFkId
          belongsToFkId
          enum {
            id
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
          }
          dbColumn {
            columnType
            unique
            nonNull
            defaultValue
            constraints
            foreignKey {
              referencedTable
              referencedColumn
              constraintName
            }
            conflicts {
              aspect
              expected
              actual
            }
          }
          validationConfig {
            minLength
            maxLength
            pattern
            minimum
            maximum
          }
          createdAt
          updatedAt
        }
        group {
          id
          name
          isVirtual
          displayOrder
        }
        dbTable
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on InvalidModelInput {
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

// 根据名称查询模型
export const GET_MODEL_BY_NAME = gql`
  query GetModelByName($name: String!, $databaseName: String!) {
    modelByName(name: $name, databaseName: $databaseName) {
      model {
        id
        projectSlug
        name
        title
        description
        databaseName
        storageType
        fields {
          name
          title
          format
          schemaType
          storageHint
          nonNull
          required
          isPrimary
          isUnique
          description
          relateFkId
          belongsToFkId
          enum {
            id
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
          }
          validationConfig {
            minLength
            maxLength
            pattern
            minimum
            maximum
          }
          createdAt
          updatedAt
        }
        group {
          id
          name
          isVirtual
          displayOrder
        }
        dbTable
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on InvalidModelInput {
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

// 获取模型分组
export const GET_MODEL_GROUPS = gql`
  query GetModelGroups {
    modelGroups {
      id
      name
      isVirtual
      displayOrder
      models {
        id
        projectSlug
        name
        title
        description
        databaseName
        storageType
        createdAt
        updatedAt
      }
    }
  }
`

// 获取模型 JSON Schema
export const GET_MODEL_JSON_SCHEMA = gql`
  query GetModelJsonSchema($id: ID!) {
    modelJsonSchema(id: $id) {
      modelId
      modelName
      schema
    }
  }
`

// 查询逻辑外键列表
export const GET_LOGICAL_FOREIGN_KEYS = gql`
  query GetLogicalForeignKeys($modelId: ID!) {
    logicalForeignKeys(modelId: $modelId) {
      id
      pairId
      direction
      modelId
      modelName
      refModelId
      refModelName
      sourceFields
      targetFields
    }
  }
`
