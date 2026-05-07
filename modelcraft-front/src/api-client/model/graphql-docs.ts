import { gql } from '@apollo/client'

// ── Queries ────────────────────────────────────────────────────────────────────

export const GET_MODELS = gql`
  query GetModels($input: ModelQueryInput) {
    models(input: $input) {
      items {
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
      hasNextPage
    }
  }
`

export const GET_MODELS_BY_DATABASE = gql`
  query GetModelsByDatabase($input: ModelQueryInput) {
    models(input: $input) {
      items {
        id
        name
        title
        databaseName
      }
    }
  }
`

export const GET_MODEL = gql`
  query GetModel($id: ID!, $withActualSchema: Boolean) {
    model(id: $id, withActualSchema: $withActualSchema) {
      model {
        id
        projectSlug
        name
        title
        description
        displayField
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
        jsonSchema
        dbTable
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

export const GET_MODEL_RECORD_WORKSPACE = gql`
  query GetModelRecordWorkspace($id: ID!) {
    model(id: $id) {
      model {
        id
        name
        title
        description
        databaseName
        jsonSchema
        fields {
          name
          isDeprecated
        }
      }
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`

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

export const GET_MODEL_JSON_SCHEMA = gql`
  query GetModelJsonSchema($id: ID!) {
    modelJsonSchema(id: $id) {
      modelId
      modelName
      schema
    }
  }
`

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

// ── Mutations ──────────────────────────────────────────────────────────────────

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
        ... on InvalidInput {
          message
          suggestion
        }
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on ModelTableAlreadyExists {
          message
          suggestion
        }
      }
    }
  }
`

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

export const DELETE_MODEL = gql`
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
          resourceType
        }
        ... on CannotDeleteDeployedModel {
          message
          suggestion
        }
      }
    }
  }
`

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
        ... on ResourceNotFound {
          message
          resourceType
        }
      }
    }
  }
`

export const DELETE_GROUP = gql`
  mutation DeleteGroup($groupId: ID!) {
    deleteGroup(groupId: $groupId) {
      success
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

export const REORDER_GROUP = gql`
  mutation ReorderGroup($input: ReorderGroupInput!) {
    reorderGroup(input: $input) {
      success
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

export const MOVE_MODEL_TO_GROUP = gql`
  mutation MoveModelToGroup($input: MoveModelToGroupInput!) {
    moveModelToGroup(input: $input) {
      success
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

export const ADD_FIELDS = gql`
  mutation AddFields($modelID: ID!, $input: [AddFieldInput!]!) {
    addFields(modelID: $modelID, input: $input) {
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
          isDeprecated
          isArray
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
        ... on InvalidInput {
          message
          suggestion
        }
      }
    }
  }
`

export const UPDATE_FIELD = gql`
  mutation UpdateField($modelID: ID!, $fieldName: String!, $input: UpdateFieldInput!) {
    updateField(modelID: $modelID, fieldName: $fieldName, input: $input) {
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
          isDeprecated
          isArray
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
        ... on InvalidInput {
          message
          suggestion
        }
        ... on FieldFormatImmutable {
          message
        }
      }
    }
  }
`

export const REMOVE_FIELD = gql`
  mutation RemoveField($modelID: ID!, $fieldName: String!) {
    removeField(modelID: $modelID, fieldName: $fieldName) {
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
        ... on InvalidInput {
          message
          suggestion
        }
        ... on FieldReferenceInUse {
          message
          suggestion
        }
      }
    }
  }
`

export const DEPRECATE_FIELD = gql`
  mutation DeprecateField($modelID: ID!, $fieldName: String!) {
    deprecateField(modelID: $modelID, fieldName: $fieldName) {
      id
    }
  }
`

export const UNDEPRECATE_FIELD = gql`
  mutation UndeprecateField($modelID: ID!, $fieldName: String!) {
    undeprecateField(modelID: $modelID, fieldName: $fieldName) {
      id
    }
  }
`

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
