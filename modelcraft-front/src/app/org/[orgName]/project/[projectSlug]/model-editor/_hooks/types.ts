import type {
  CreateEnumFieldFormValues,
  CreateLogicalForeignKeyResult,
  DbColumnInfo,
  DeleteLogicalForeignKeyResult,
  LogicalForeignKey,
  ModelEnumDomainError,
  UpdateFieldMetaFormValues,
} from '@/types'

// Local model type used by the model editor page
// (lighter than the shared Model type from @/types)
export interface EditorModel {
  id: string
  name: string
  title: string
  description?: string
  displayField?: string
  insertionOrderField?: string
  databaseName: string
  storageType?: string
  createdVia?: 'NEW' | 'IMPORTED'
}

export interface EditorModelField {
  name: string
  title: string
  format?: string
  schemaType?: string
  storageHint?: string
  nonNull?: boolean
  required?: boolean
  isPrimary?: boolean
  description?: string
  isDeprecated?: boolean
  enum?: {
    name: string
  }
  relateEnumName?: string
  enumRelationId?: string
  dbColumn?: DbColumnInfo
}

export interface EditorModelDetail extends EditorModel {
  fields: EditorModelField[]
}

// GraphQL response types
export interface ModelsQueryData {
  models?: {
    items: EditorModel[]
    hasNextPage?: boolean
  }
}

export interface ModelQueryData {
  model?: {
    model?: EditorModelDetail
    error?: { message: string }
  }
}

export interface TestConnectionResult {
  testDatabaseConnection?: {
    success: boolean
    error?: {
      message: string
    }
  }
}

export interface CreateModelResult {
  createModel?: {
    model?: { id: string }
    error?: { message: string }
  }
}

export interface DeleteModelResult {
  deleteModel?: {
    success?: boolean
    error?: { message: string }
  }
}

export interface UpdateModelMetaResult {
  updateModelMeta?: {
    model?: EditorModelDetail
    error?: { message: string }
  }
}

export interface ForeignKeysQueryData {
  logicalForeignKeys?: LogicalForeignKey[]
}

export interface EnumFieldPageState {
  saving: boolean
  error: ModelEnumDomainError | null
}

export interface CreateEnumFieldPageModel {
  defaults: CreateEnumFieldFormValues
  enumOptions: string[]
  state: EnumFieldPageState
}


export interface EditFieldImmutablePageModel {
  fieldName: string
  format: string
  relateEnumName?: string
  enumRelationId?: string
  editable: UpdateFieldMetaFormValues
  state: EnumFieldPageState
}

export type { LogicalForeignKey, CreateLogicalForeignKeyResult, DeleteLogicalForeignKeyResult }
