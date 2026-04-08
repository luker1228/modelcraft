import type { LogicalForeignKey, DbColumnInfo, CreateLogicalForeignKeyResult, DeleteLogicalForeignKeyResult } from '@/types'

// Local model type used by the model editor page
// (lighter than the shared Model type from @/types)
export interface EditorModel {
  id: string
  name: string
  title: string
  description?: string
  databaseName: string
  storageType?: string
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
  dbColumn?: DbColumnInfo
}

export interface EditorModelDetail extends EditorModel {
  fields: EditorModelField[]
}

// GraphQL response types
export interface ModelsEdge {
  node: EditorModel
}

export interface ModelsQueryData {
  models?: {
    edges: ModelsEdge[]
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

export type { LogicalForeignKey, CreateLogicalForeignKeyResult, DeleteLogicalForeignKeyResult }
