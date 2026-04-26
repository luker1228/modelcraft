import type { EnumDefinition, EnumOptionInput } from './enum'

export type DbTableStatus = 'TABLE_EXISTS' | 'TABLE_MISSING' | 'CLUSTER_UNREACHABLE'

export interface ModelGroup {
  id: string
  name: string
  isVirtual: boolean
  displayOrder: string
  models: Model[]
}

export interface Model {
  id: string
  projectSlug: string
  name: string
  title: string
  description: string
  databaseName: string
  storageType: string
  fields: Field[]
  group: ModelGroup
  dbTable?: DbTableStatus
  createdAt: string
  updatedAt: string
}

export type SchemaType = 'STRING' | 'NUMBER' | 'BOOLEAN' | 'ARRAY' | 'OBJECT'

export type FormatType =
  | 'STRING'
  | 'UUID'
  | 'DATE'
  | 'DATETIME'
  | 'TIME'
  | 'NUMBER'
  | 'INTEGER'
  | 'DECIMAL'
  | 'BOOLEAN'
  | 'RELATION'
  | 'ENUM'
  | 'ENUM_ARRAY'
  | 'END_USER_REF'

export type ActualConstraintType = 'UNIQUE' | 'NOT_NULL'

export type FieldConflictAspect = 'UNIQUE_MISMATCH' | 'NOT_NULL_MISMATCH'

export interface FieldConflict {
  aspect: FieldConflictAspect
  expected: string
  actual: string
}

export interface ActualForeignKey {
  referencedTable: string
  referencedColumn: string
  constraintName: string
}

export interface DbColumnInfo {
  columnType: string
  unique: boolean
  nonNull: boolean
  defaultValue?: string
  constraints: ActualConstraintType[]
  foreignKey?: ActualForeignKey
  conflicts: FieldConflict[]
}

export interface ValidationConfig {
  minLength?: number
  maxLength?: number
  pattern?: string
  minimum?: number
  maximum?: number
}

export interface Field {
  id: string
  name: string
  title: string
  schemaType: SchemaType
  format: FormatType
  storageHint: string
  nonNull: boolean
  required: boolean
  isUnique: boolean
  isPrimary: boolean
  description?: string
  validationConfig?: ValidationConfig
  relateFkId?: string
  belongsToFkId?: string
  enum?: EnumDefinition
  dbColumn?: DbColumnInfo
  createdAt: string
  updatedAt: string
}

export interface CreateModelInput {
  projectSlug: string
  name: string
  title: string
  description?: string
  databaseName: string
}

export interface UpdateModelMetaInput {
  title?: string
  description?: string
}

export interface ModelQueryInput {
  projectSlug: string
  databaseName: string
  offset?: number
  limit?: number
  search?: string
}

export interface EnumConfigInput {
  enumName: string
  options?: EnumOptionInput[]
  description?: string
  connectEnum: boolean
}

export interface ValidationConfigInput {
  minLength?: number
  maxLength?: number
  pattern?: string
  minimum?: number
  maximum?: number
}

export interface AddFieldInput {
  name: string
  title: string
  format: FormatType
  storageHint?: string
  nonNull?: boolean
  required?: boolean
  isUnique?: boolean
  description?: string
  validationConfig?: ValidationConfigInput
  relateFkId?: string
  enumConfig?: EnumConfigInput
}

export interface UpdateFieldInput {
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}

export interface SyncModelSchemaInput {
  id: string
  schema: string
  deleteExtraFields?: boolean
}

export interface RepairModelInput {
  projectSlug: string
  modelId: string
  mode: 'DRY_RUN' | 'ADDITIVE' | 'FULL_SYNC'
  deleteExtraFields?: boolean
}

export interface CreateModelFromSchemaInput {
  projectSlug: string
  schema: string
  databaseName: string
}

export interface ReverseEngineerModelInput {
  projectSlug: string
  databaseName: string
  ddlStatement: string
  tableName: string
}

export interface CreateGroupInput {
  projectSlug: string
  name: string
}

export interface RenameGroupInput {
  projectSlug: string
  groupId: string
  newName: string
}

export interface ReorderGroupInput {
  projectSlug: string
  groupId: string
  afterGroupId?: string
}

export interface MoveModelToGroupInput {
  modelId: string
  groupId?: string
}
