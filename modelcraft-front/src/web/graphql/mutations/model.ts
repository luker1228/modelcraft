import { gql } from '@apollo/client'

// 创建模型
export const CREATE_MODEL = gql`
  mutation CreateModel($input: CreateModelInput!) {
    createModel(input: $input) {
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
        ... on ModelAlreadyExists {
          message
          suggestion
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

// 更新模型元信息
export const UPDATE_MODEL = gql`
  mutation UpdateModelMeta($id: ID!, $input: UpdateModelMetaInput!) {
    updateModelMeta(id: $id, input: $input) {
      success
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

// 删除模型
export const DELETE_MODEL = gql`
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      success
      error {
        __typename
        ... on ModelNotFound {
          message
        }
        ... on CannotDeleteDeployedModel {
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

// 同步模型Schema
export const SYNC_MODEL_SCHEMA = gql`
  mutation SyncModelSchema($input: SyncModelSchemaInput!) {
    syncModelSchema(input: $input) {
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
      fieldsAdded
      fieldsSkipped
      fieldsDeleted
      deletedFields
    }
  }
`

// 修复模型
export const REPAIR_MODEL = gql`
  mutation RepairModel($input: RepairModelInput!) {
    repairModel(input: $input) {
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
      changesApplied
      detectedIssues {
        type
        description
        tableName
        fieldName
        details
      }
      executedDDL
      healthStatusBefore
      healthStatusAfter
      extraFieldsRemoved
      fieldsAdded
    }
  }
`

// 从 Schema 创建模型
export const CREATE_MODEL_FROM_SCHEMA = gql`
  mutation CreateModelFromSchema($input: CreateModelFromSchemaInput!) {
    createModelFromSchema(input: $input) {
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
    }
  }
`

// 导入模型
export const IMPORT_MODEL = gql`
  mutation ImportModel($input: ImportModelInput!) {
    importModel(input: $input) {
      modelId
      modelName
      fieldsCount
      skippedFields
    }
  }
`

// 创建分组
export const CREATE_GROUP = gql`
  mutation CreateGroup($input: CreateGroupInput!) {
    createGroup(input: $input) {
      group {
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
      error {
        __typename
        ... on GroupAlreadyExists {
          message
          suggestion
        }
        ... on InvalidGroupName {
          message
          suggestion
        }
      }
    }
  }
`

// 重命名分组
export const RENAME_GROUP = gql`
  mutation RenameGroup($input: RenameGroupInput!) {
    renameGroup(input: $input) {
      group {
        id
        name
        isVirtual
        displayOrder
      }
      error {
        __typename
        ... on GroupAlreadyExists {
          message
          suggestion
        }
        ... on InvalidGroupName {
          message
          suggestion
        }
        ... on GroupNotFound {
          message
        }
      }
    }
  }
`

// 删除分组
export const DELETE_GROUP = gql`
  mutation DeleteGroup($groupId: ID!) {
    deleteGroup(groupId: $groupId) {
      success
      error {
        __typename
        ... on GroupNotFound {
          message
        }
      }
    }
  }
`

// 重排序分组
export const REORDER_GROUP = gql`
  mutation ReorderGroup($input: ReorderGroupInput!) {
    reorderGroup(input: $input) {
      success
      error {
        __typename
        ... on GroupNotFound {
          message
        }
      }
    }
  }
`

// 移动模型到分组
export const MOVE_MODEL_TO_GROUP = gql`
  mutation MoveModelToGroup($input: MoveModelToGroupInput!) {
    moveModelToGroup(input: $input) {
      success
      error {
        __typename
        ... on GroupNotFound {
          message
        }
        ... on ModelNotFound {
          message
        }
      }
    }
  }
`

// 添加字段
export const ADD_FIELDS = gql`
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
      id
    }
  }
`

// 更新字段
export const UPDATE_FIELD = gql`
  mutation UpdateField($modelID: ID!, $fieldName: String!, $input: UpdateFieldInput!) {
    updateField(modelID: $modelID, fieldName: $fieldName, input: $input) {
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
        isDeprecated
        description
        relateFkId
        belongsToFkId
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
  }
`


// 删除字段
export const REMOVE_FIELD = gql`
  mutation RemoveField($modelID: ID!, $fieldName: String!) {
    removeField(modelID: $modelID, fieldName: $fieldName) {
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
  }
`

// 创建逻辑外键
export const CREATE_LOGICAL_FOREIGN_KEY = gql`
  mutation CreateLogicalForeignKey($input: CreateLogicalForeignKeyInput!) {
    createLogicalForeignKey(input: $input) {
      result {
        __typename
        ... on LogicalForeignKey {
          id
          pairId
          direction
          modelId
          refModelId
          sourceFields
          targetFields
        }
        ... on FKColumnsNotFoundError {
          message
        }
        ... on FKFieldCountMismatchError {
          message
        }
      }
    }
  }
`

// 删除逻辑外键
export const DELETE_LOGICAL_FOREIGN_KEY = gql`
  mutation DeleteLogicalForeignKey($pairId: String!) {
    deleteLogicalForeignKey(pairId: $pairId) {
      result {
        __typename
        ... on DeleteLogicalForeignKeySuccess {
          pairId
        }
        ... on FKNotFoundError {
          message
        }
        ... on FKPairHasRelateFieldsError {
          message
        }
      }
    }
  }
`
