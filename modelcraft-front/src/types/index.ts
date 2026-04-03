// ============================================
// Project Types
// ============================================

// 项目状态枚举
export type ProjectStatus = 'ACTIVE' | 'ARCHIVED'

export interface Project {
  id: string
  slug: string
  title: string
  description: string
  databaseName?: string
  loginUrl?: string
  status: ProjectStatus
  orgName: string
  createdAt: string
  updatedAt: string
}

// ============================================
// Cluster Types
// ============================================

export type ClusterStatus = 'ACTIVE' | 'DISABLED'

export interface DatabaseConnectionInfo {
  host: string
  port: number
  username: string
  password: string
}

export interface DatabaseCluster {
  id: string
  projectSlug: string
  title: string
  description: string
  connectionInfo: DatabaseConnectionInfo
  connectionTimeout: number
  status: ClusterStatus
  version: number
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface Database {
  name: string
}

// ============================================
// Model Types
// ============================================

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

// ============================================
// Field Types
// ============================================

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
  | 'ENUM_LABEL'

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

// ============================================
// Enum Types
// ============================================

export interface EnumOption {
  id: string
  code: string
  label: string
  order: number
  description?: string
}

export interface EnumDefinition {
  id: string
  projectSlug: string
  name: string
  displayName: string
  description?: string
  options: EnumOption[]
  isMultiSelect: boolean
  createdAt: string
  updatedAt: string
}

// ============================================
// Input Types
// ============================================

export interface ConnectionInfoInput {
  host: string
  port: number
  username: string
  password: string
  connectionTimeout?: number
}

export interface ClusterConnectionInput {
  title: string
  description?: string
  connectionInfo: ConnectionInfoInput
}

export interface UpdateClusterConnectionInput {
  title?: string
  description?: string
  connectionInfo?: ConnectionInfoInput
  skipConnectionTest?: boolean
}

export interface CreateProjectInput {
  slug: string
  title: string
  description?: string
  loginUrl?: string
  clusterInput: ClusterConnectionInput
  skipConnectionTest?: boolean
}

export interface UpdateProjectInput {
  slug: string
  title?: string
  description?: string
  loginUrl?: string
}

export interface ListProjectsInput {
  status?: ProjectStatus
}

export interface ListDatabasesInput {
  projectSlug: string
  offset?: number
  limit?: number
  search?: string
}

export interface TestDatabaseConnectionInput {
  projectSlug?: string
  connectionInfo?: ConnectionInfoInput
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

export interface EnumOptionInput {
  code: string
  label: string
  order: number
  description?: string
}

export interface EnumConfigInput {
  enumName: string
  options?: EnumOptionInput[]
  description?: string
  connectEnum: boolean
}

export interface EnumLabelConfigInput {
  sourceField: string
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
  enumLabelConfig?: EnumLabelConfigInput
}

export interface UpdateFieldInput {
  title?: string
  description?: string
  validationConfig?: ValidationConfigInput
}

export interface CreateEnumInput {
  projectSlug: string
  name: string
  displayName: string
  description?: string
  options: EnumOptionInput[]
}

export interface UpdateEnumInput {
  displayName?: string
  description?: string
  options?: EnumOptionInput[]
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

// ============================================
// User Management Types
// ============================================

export type OrganizationStatus = 'ACTIVE' | 'SUSPENDED' | 'DELETED'
export type MembershipStatus = 'ACTIVE' | 'SUSPENDED' | 'INVITED'

export interface Organization {
  id: string
  name: string
  displayName?: string
  ownerID: string
  status: OrganizationStatus
  createdAt: string
  updatedAt: string
}

export interface Role {
  id: string
  name: string
  description?: string
  permissions: string[]
  isSystem: boolean
  createdAt: string
  updatedAt: string
}

export interface OrganizationMember {
  id: string
  userID: string
  userName: string
  orgID: string
  role: Role
  status: MembershipStatus
  joinedAt?: string
  createdAt: string
}

export interface CurrentUser {
  id: string
  externalID: string
  email: string
  name: string
  organization?: Organization
  role?: Role
  permissions: string[]
}

export interface CreateRoleInput {
  name: string
  description?: string
  permissions: string[]
}

export interface UpdateOrganizationInput {
  displayName?: string
}

// ============================================
// Logical Foreign Key Types
// ============================================

export type FKDirection = 'NORMAL' | 'REVERSE'

export interface LogicalForeignKey {
  id: string
  pairId: string
  direction: FKDirection
  modelId: string
  modelName?: string
  refModelId: string
  refModelName?: string
  sourceFields: string[]
  targetFields: string[]
}

export interface CreateLogicalForeignKeyInput {
  modelId: string
  refModelId: string
  sourceFields: string[]
  targetFields: string[]
}

export interface FKColumnsNotFoundError {
  message: string
}

export interface FKFieldCountMismatchError {
  message: string
}

export interface FKNotFoundError {
  message: string
}

export interface FKPairHasRelateFieldsError {
  message: string
}

export type CreateLogicalForeignKeyResult =
  | (LogicalForeignKey & { __typename: 'LogicalForeignKey' })
  | (FKColumnsNotFoundError & { __typename: 'FKColumnsNotFoundError' })
  | (FKFieldCountMismatchError & { __typename: 'FKFieldCountMismatchError' })

export type DeleteLogicalForeignKeyResult =
  | ({ pairId: string } & { __typename: 'DeleteLogicalForeignKeySuccess' })
  | (FKNotFoundError & { __typename: 'FKNotFoundError' })
  | (FKPairHasRelateFieldsError & { __typename: 'FKPairHasRelateFieldsError' })

// ============================================
// Schema Issue Types
// ============================================

export type HealthStatus = 'HEALTHY' | 'NEEDS_REPAIR' | 'BROKEN'

export type SchemaIssueType =
  | 'TABLE_MISSING'
  | 'FIELD_MISSING'
  | 'FIELD_TYPE_MISMATCH'
  | 'FIELD_CONSTRAINT_MISMATCH'
  | 'DATABASE_MISSING'
  | 'CLUSTER_NOT_FOUND'
  | 'FIELD_HAS_DEPENDENCIES'

export interface SchemaIssue {
  type: SchemaIssueType
  description: string
  tableName: string
  fieldName?: string
  details?: string
}

// ============================================
// Permission Types
// ============================================

export interface PermissionRole {
  id: number
  name: string
  description?: string
  isSystem: boolean
  orgName: string
  createdAt: string
  updatedAt: string
}

export interface PermissionDef {
  obj: string
  act: string
}

export interface UserRoleAssignment {
  id: number
  userId: string
  roleId: number
  orgName: string
  createdAt: string
}

export interface CreateCustomRoleInput {
  name: string
  description?: string
  orgName: string
}

export interface UpdateRoleInput {
  name?: string
  description?: string
}
